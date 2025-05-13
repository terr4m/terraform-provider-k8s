package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourcesDataSource(t *testing.T) {
	t.Run("cluster_scoped_resource", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `data "k8s_resources" "test" {
  api_version = "v1"
  kind        = "Node"
}`,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects"), knownvalue.ListSizeExact(1)),
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects").AtSliceIndex(0).AtMapKey("kind"), knownvalue.StringExact("Node")),
					},
				},
			},
		})
	})

	t.Run("namespaced_resource", func(t *testing.T) {
		namespace := "default"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`data "k8s_resources" "test" {
  api_version = "v1"
  kind        = "ServiceAccount"
  namespace   = "%s"
}`, namespace),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects"), knownvalue.ListSizeExact(1)),
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects").AtSliceIndex(0).AtMapKey("kind"), knownvalue.StringExact("ServiceAccount")),
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects").AtSliceIndex(0).AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
					},
				},
			},
		})
	})

	t.Run("with_field_selector", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `data "k8s_resources" "test" {
  api_version    = "rbac.authorization.k8s.io/v1"
  kind           = "ClusterRole"
	field_selector = "metadata.name=cluster-admin"
}`,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects"), knownvalue.ListSizeExact(1)),
					},
				},
			},
		})
	})

	t.Run("with_field_selector_for_namespace", func(t *testing.T) {
		namespace := "kube-public"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`data "k8s_resources" "test" {
  api_version    = "v1"
  kind           = "ServiceAccount"
	field_selector = "metadata.namespace==%s"
}`, namespace),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects"), knownvalue.ListSizeExact(1)),
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects").AtSliceIndex(0).AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
					},
				},
			},
		})
	})

	t.Run("with_label_selector", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `data "k8s_resources" "test" {
  api_version    = "v1"
  kind           = "Node"
	label_selector = "kubernetes.io/os!=linux"
}`,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects"), knownvalue.ListSizeExact(0)),
					},
				},
			},
		})
	})

	t.Run("with_limit", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `data "k8s_resources" "test" {
  api_version = "rbac.authorization.k8s.io/v1"
  kind        = "ClusterRole"
	limit       = 1
}`,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("data.k8s_resources.test", tfjsonpath.New("objects"), knownvalue.ListSizeExact(1)),
					},
				},
			},
		})
	})
}
