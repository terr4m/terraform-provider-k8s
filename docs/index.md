---
page_title: "K8s Provider - terraform-provider-k8s"
subcategory: ""
description: |-
  The K8s provider provides a way to manage Kubernetes resources using Terraform. It maps to the Kubernetes API using server-side-apply and field management.
---

# K8s Provider

The K8s provider provides a way to manage _Kubernetes_ resources using _Terraform_. It maps to the _Kubernetes_ API using server-side-apply and field management.

## Example Usage

```terraform
provider "k8s" {
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `client_certificate` (String) PEM-encoded client certificate for TLS authentication. Can be set with the `KUBE_CLIENT_CERT_DATA` environment variable.
- `client_key` (String) PEM-encoded client certificate key for TLS authentication. Can be set with the `KUBE_CLIENT_KEY_DATA` environment variable.
- `cluster_ca_certificate` (String) PEM-encoded root certificates bundle for TLS authentication. Can be set with the `KUBE_CLUSTER_CA_CERT_DATA` environment variable.
- `config_context` (String) Context to choose from the kube config file. Can be set with the `KUBE_CTX`environment variable.
- `config_context_auth_info` (String) Authentication info context of the kube config (name of the kube config user, --user flag in kubectl). Can be set with the `KUBE_CTX_AUTH_INFO` environment variable.
- `config_context_cluster` (String) Cluster context of the kube config (name of the kube config cluster, --cluster flag in kubectl). Can be set with the `KUBE_CTX_CLUSTER` environment variable.
- `config_paths` (List of String) List of paths to the kube config file. Can be set with the `KUBE_CONFIG_PATHS` environment variable.
- `exec` (Attributes) Exec configuration for Kubernetes authentication (see [below for nested schema](#nestedatt--exec))
- `field_manager` (Attributes) Field manager configuration. (see [below for nested schema](#nestedatt--field_manager))
- `host` (String) The hostname (in form of URI) of _Kubernetes_ master. Can be set with the `KUBE_HOST` environment variable.
- `insecure` (Boolean) Whether server should be accessed without verifying the TLS certificate. Can be set with the `KUBE_INSECURE` environment variable.
- `password` (String) The password to use for HTTP basic authentication when accessing the _Kubernetes_ master endpoint. Can be set with the `KUBE_PASSWORD` environment variable.
- `proxy_url` (String) URL to the proxy to be used for all API requests. Can be set with the `KUBE_PROXY_URL` environment variable.
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))
- `tls_server_name` (String) Server name passed to the server for SNI and is used in the client to check server certificates against. Can be set with the `KUBE_TLS_SERVER_NAME` environment variable.
- `token` (String) Token to authenticate a service account. Can be set with the `KUBE_TOKEN` environment variable.
- `username` (String) The username to use for HTTP basic authentication when accessing the _Kubernetes_ master endpoint. Can be set with the `KUBE_USER` environment variable.

<a id="nestedatt--exec"></a>
### Nested Schema for `exec`

Required:

- `api_version` (String) API version for the exec plugin.
- `command` (String) Command to run for _Kubernetes_ exec plugin.

Optional:

- `args` (List of String) Arguments for the exec plugin.
- `env` (Map of String) Environment variables for the exec plugin.


<a id="nestedatt--field_manager"></a>
### Nested Schema for `field_manager`

Optional:

- `force_conflicts` (Boolean) If `true`, the field manager will force apply the changes by ignoring the conflicts.
- `name` (String) Field manager name.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) Timeout for resource creation; defaults to `10m`. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).
- `delete` (String) Timeout for resource deletion; defaults to `10m`. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).
- `read` (String) Timeout for resource or data source reads; defaults to `10m`. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).
- `update` (String) Timeout for resource update; defaults to `10m`. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).
