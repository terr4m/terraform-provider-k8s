package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestAccResourceResource(t *testing.T) {
	t.Run("invalid_manifest", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{{
				Config: `resource "k8s_resource" "test" {
  manifest = {
    kind = "ConfigMap"
    metadata = {
      name = "example"
    }
  }
}`,
				ExpectError: regexp.MustCompile("Missing required attribute|manifest.apiVersion is required|expected manifest to have attribute"),
			}},
		})
	})

	t.Run("configmap_create_update", func(t *testing.T) {
		name := "test-resource"
		namespace := "default"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata = {
      namespace = %q
      name      = %q
    }
    data = {
      foo = "bar"
    }
  }
}`,
						namespace, name),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("managed_fields_fingerprint"), knownvalue.NotNull()),
					},
				},
				{
					Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata = {
      namespace = %q
      name      = %q
    }
    data = {
      foo = "baz"
    }
  }
}`,
						namespace, name),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("managed_fields_fingerprint"), knownvalue.NotNull()),
					},
				},
			},
		})
	})

	t.Run("cluster_role_create", func(t *testing.T) {
		name := "test-resource-cluster"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{{
				Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "rbac.authorization.k8s.io/v1"
    kind       = "ClusterRole"
    metadata = {
      name = %q
    }
    rules = [{
      apiGroups = ["*"]
      resources = ["*"]
      verbs     = ["get"]
    }]
  }
}`,
					name),
				ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("managed_fields_fingerprint"), knownvalue.NotNull())},
			}},
		})
	})

	t.Run("managed_field_drift_detected", func(t *testing.T) {
		name := "test-resource-drift-managed"
		namespace := "default"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigMapResourceConfig(namespace, name, map[string]string{"foo": "bar"}),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("managed_fields_fingerprint"), knownvalue.NotNull()),
					},
				},
				{
					PreConfig: testAccMutateConfigMap(t, namespace, name, func(cm *corev1.ConfigMap) {
						cm.Data["foo"] = "drifted"
					}),
					Config:      testAccConfigMapResourceConfig(namespace, name, map[string]string{"foo": "bar"}),
					ExpectError: regexp.MustCompile("xxx"),
				},
			},
		})
	})

	t.Run("unrelated_field_change_ignored", func(t *testing.T) {
		name := "test-resource-drift-unrelated"
		namespace := "default"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigMapResourceConfig(namespace, name, map[string]string{"foo": "bar"}),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("managed_fields_fingerprint"), knownvalue.NotNull()),
					},
				},
				{
					PreConfig: testAccMutateConfigMap(t, namespace, name, func(cm *corev1.ConfigMap) {
						if cm.Data == nil {
							cm.Data = map[string]string{}
						}
						cm.Data["baz"] = "qux"
					}),
					Config: testAccConfigMapResourceConfig(namespace, name, map[string]string{"foo": "bar"}),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					}},
				},
			},
		})
	})
}

func testAccConfigMapResourceConfig(namespace, name string, data map[string]string) string {
	var dataBody strings.Builder
	for key, value := range data {
		dataBody.WriteString(fmt.Sprintf("      %s = %q\n", key, value))
	}

	return fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata = {
      namespace = %q
      name      = %q
    }
    data = {
%s    }
  }
}`,
		namespace, name, dataBody.String())
}

func testAccMutateConfigMap(t *testing.T, namespace, name string, mutate func(*corev1.ConfigMap)) func() {
	t.Helper()

	return func() {
		loader := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			t.Fatalf("failed to build kube client config: %v", err)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			t.Fatalf("failed to build kubernetes client: %v", err)
		}

		cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("failed to get configmap %s/%s: %v", namespace, name, err)
		}

		mutate(cm)

		if _, err := clientset.CoreV1().ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("failed to update configmap %s/%s: %v", namespace, name, err)
		}
	}
}
