package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceResource(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		t.Run("missing_api_version", func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: `resource "k8s_resource" "test" {
  manifest = {
    kind        = "ConfigMap"
    metadata = {
      namespace = "default"
      name      = "test"
    }
    data = {}
  }
}`,
						ExpectError: regexp.MustCompile(".+"),
					},
				},
			})
		})

		t.Run("empty_api_version", func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: `resource "k8s_resource" "test" {
  manifest = {
    apiVersion = ""
    kind        = "ConfigMap"
    metadata = {
      namespace = "default"
      name      = "test"
    }
    data = {}
  }
}`,
						ExpectError: regexp.MustCompile(".+"),
					},
				},
			})
		})

		t.Run("missing_kind", func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: `resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "v1"
    metadata = {
      namespace = "default"
      name      = "test"
    }
    data = {}
  }
}`,
						ExpectError: regexp.MustCompile(".+"),
					},
				},
			})
		})

		t.Run("empty_kind", func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: `resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "v1"
    kind        = ""
    metadata = {
      namespace = "default"
      name      = "test"
    }
    data = {}
  }
}`,
						ExpectError: regexp.MustCompile(".+"),
					},
				},
			})
		})
	})

	t.Run("clusterrole", func(t *testing.T) {
		apiVersion := "rbac.authorization.k8s.io/v1"
		kind := "ClusterRole"

		t.Run("create", func(t *testing.T) {
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      name      = "%s"
    }
  }
}`, apiVersion, kind, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
						},
					},
				},
			})
		})

		t.Run("create_update", func(t *testing.T) {
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      name      = "%s"
    }
  }
}`, apiVersion, kind, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
						},
					},
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      name      = "%s"
    }
    rules = [{
      apiGroups = ["*"]
      resources = ["*"]
      verbs = ["*"]
    }]
  }
}`, apiVersion, kind, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("rules"), knownvalue.ListSizeExact(1)),
						},
					},
				},
			})
		})
	})

	t.Run("configmap", func(t *testing.T) {
		apiVersion := "v1"
		kind := "ConfigMap"

		t.Run("create", func(t *testing.T) {
			namespace := "default"
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      namespace = "%s"
      name      = "%s"
    }
    data = {
      foo = "bar"
    }
  }
}`, apiVersion, kind, namespace, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("data").AtMapKey("foo"), knownvalue.StringExact("bar")),
						},
					},
				},
			})
		})

		t.Run("use_default_namespace", func(t *testing.T) {
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      name      = "%s"
    }
    data = {
      foo = "bar"
    }
  }
}`, apiVersion, kind, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact("default")),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("data").AtMapKey("foo"), knownvalue.StringExact("bar")),
						},
					},
				},
			})
		})

		t.Run("create_update", func(t *testing.T) {
			namespace := "default"
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      namespace = "%s"
      name      = "%s"
    }
    data = {
      foo = "bar"
    }
  }
}`, apiVersion, kind, namespace, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("data").AtMapKey("foo"), knownvalue.StringExact("bar")),
						},
					},
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      namespace = "%s"
      name      = "%s"
    }
    data = {
      test = "test"
    }
  }
}`, apiVersion, kind, namespace, name),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction("k8s_resource.test", plancheck.ResourceActionUpdate),
							},
						},
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("data").AtMapKey("test"), knownvalue.StringExact("test")),
						},
					},
				},
			})
		})

		t.Run("create_in_new_namespace", func(t *testing.T) {
			namespace := "test"
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "ns" {
  manifest = {
    apiVersion = "v1"
    kind       = "Namespace"
    metadata = {
      name = "%s"
    }
  }
}

resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      namespace = k8s_resource.ns.object.metadata.name
      name      = "%s"
    }
    data = {
      foo = "bar"
    }
  }
}`, namespace, apiVersion, kind, name),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectUnknownValue("k8s_resource.test", tfjsonpath.New("object")),
							},
						},
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("data").AtMapKey("foo"), knownvalue.StringExact("bar")),
						},
					},
				},
			})
		})
	})

	t.Run("deployment", func(t *testing.T) {
		apiVersion := "apps/v1"
		kind := "Deployment"

		t.Run("create", func(t *testing.T) {
			namespace := "default"
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      namespace = "%s"
      name      = "%s"
      labels = {
        "app.kubernetes.io/name" = "test"
      }
    }
    spec = {
      replicas = 1
      selector = {
        matchLabels = {
          "app.kubernetes.io/name" = "test"
        }
      }
      template = {
        metadata = {
          labels = {
            "app.kubernetes.io/name" = "test"
          }
        }
        spec = {
          containers = [{
            name = "nginx"
            image = "docker.io/nginx:latest"
            ports = [{
              containerPort = 80
            }]
            readinessProbe = {
              httpGet = {
                path = "/"
                port = 80
              }
              initialDelaySeconds = 10
            }
          }]
        }
      }
    }
  }
}`, apiVersion, kind, namespace, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
						},
					},
				},
			})
		})

		t.Run("create_wait_for_rollout", func(t *testing.T) {
			namespace := "default"
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  wait_options = {
    rollout = true
  }

  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      namespace = "%s"
      name      = "%s"
      labels = {
        "app.kubernetes.io/name" = "test"
      }
    }
    spec = {
      replicas = 1
      selector = {
        matchLabels = {
          "app.kubernetes.io/name" = "test"
        }
      }
      template = {
        metadata = {
          labels = {
            "app.kubernetes.io/name" = "test"
          }
        }
        spec = {
          containers = [{
            name = "nginx"
            image = "docker.io/nginx:latest"
            ports = [{
              containerPort = 80
            }]
            readinessProbe = {
              httpGet = {
                path = "/"
                port = 80
              }
              initialDelaySeconds = 10
            }
          }]
        }
      }
    }
  }
}`, apiVersion, kind, namespace, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
						},
					},
				},
			})
		})
	})

	t.Run("secret", func(t *testing.T) {
		apiVersion := "v1"
		kind := "Secret"

		t.Run("create", func(t *testing.T) {
			namespace := "default"
			name := "test"

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "k8s_resource" "test" {
  manifest = {
    apiVersion = "%s"
    kind        = "%s"
    metadata = {
      namespace = "%s"
      name      = "%s"
    }
    data = {
      foo = base64encode("bar")
    }
  }
}`, apiVersion, kind, namespace, name),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object"), knownvalue.NotNull()),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("apiVersion"), knownvalue.StringExact(apiVersion)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("kind"), knownvalue.StringExact(kind)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("namespace"), knownvalue.StringExact(namespace)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("metadata").AtMapKey("name"), knownvalue.StringExact(name)),
							statecheck.ExpectKnownValue("k8s_resource.test", tfjsonpath.New("object").AtMapKey("data").AtMapKey("foo"), knownvalue.NotNull()),
						},
					},
				},
			})
		})
	})
}
