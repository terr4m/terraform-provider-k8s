package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceDataSource(t *testing.T) {
	t.Run("cluster_scoped_resource", func(t *testing.T) {
		name := "cluster-admin"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`data "k8s_resource" "test" {
  api_version = "rbac.authorization.k8s.io/v1"
  kind        = "ClusterRole"
  name        = "%s"
}`, name),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("data.k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue("data.k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
					},
				},
			},
		})
	})

	t.Run("namespaced_resource", func(t *testing.T) {
		namespace := "default"
		name := "default"

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`data "k8s_resource" "test" {
  api_version = "v1"
  kind        = "ServiceAccount"
  namespace   = "%s"
  name        = "%s"
}`, namespace, name),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("data.k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue("data.k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
						statecheck.ExpectKnownValue("data.k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
					},
				},
			},
		})
	})
}
