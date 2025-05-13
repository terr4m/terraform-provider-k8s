package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccServerVersionDataSource(t *testing.T) {
	config := `data "k8s_server_version" "test" {}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("major"), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("minor"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("git_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("git_commit"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("git_tree_state"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("build_date"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("go_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("compiler"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.k8s_server_version.test", tfjsonpath.New("platform"), knownvalue.NotNull()),
				},
			},
		},
	})
}
