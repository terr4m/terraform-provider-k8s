package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terr4m/terraform-provider-k8s/internal/k8sutils"
)

// Ensure K8sProvider satisfies various provider interfaces.
var _ provider.Provider = &K8sProvider{}
var _ provider.ProviderWithFunctions = &K8sProvider{}
var _ provider.ProviderWithEphemeralResources = &K8sProvider{}

// New returns a new provider implementation.
func New(version, commit string) func() provider.Provider {
	return func() provider.Provider {
		return &K8sProvider{
			version: version,
			commit:  commit,
		}
	}
}

// K8sProviderData is the data available to the resource and data sources.
type K8sProviderData struct {
	provider        *K8sProvider
	Model           *K8sProviderModel
	Client          *k8sutils.Client
	FieldManager    *k8sutils.FieldManager
	DefaultTimeouts *Timeouts
}

// Timeouts represents a set of timeouts.
type Timeouts struct {
	Create time.Duration
	Read   time.Duration
	Update time.Duration
	Delete time.Duration
}

// K8sProviderModel describes the provider data model.
type K8sProviderModel struct {
	Host                  types.String       `tfsdk:"host"`
	Username              types.String       `tfsdk:"username"`
	Password              types.String       `tfsdk:"password"`
	Insecure              types.Bool         `tfsdk:"insecure"`
	TLSServerName         types.String       `tfsdk:"tls_server_name"`
	ClientCertificate     types.String       `tfsdk:"client_certificate"`
	ClientKey             types.String       `tfsdk:"client_key"`
	ClusterCACertificate  types.String       `tfsdk:"cluster_ca_certificate"`
	ConfigPaths           types.List         `tfsdk:"config_paths"`
	ConfigContext         types.String       `tfsdk:"config_context"`
	ConfigContextAuthInfo types.String       `tfsdk:"config_context_auth_info"`
	ConfigContextCluster  types.String       `tfsdk:"config_context_cluster"`
	Token                 types.String       `tfsdk:"token"`
	ProxyURL              types.String       `tfsdk:"proxy_url"`
	Exec                  *ExecConfigModel   `tfsdk:"exec"`
	FieldManager          *FieldManagerModel `tfsdk:"field_manager"`
	Timeouts              timeouts.Value     `tfsdk:"timeouts"`
}

// ExecConfigModel configures an external command to configure the Kubernetes client.
type ExecConfigModel struct {
	APIVersion types.String `tfsdk:"api_version"`
	Command    types.String `tfsdk:"command"`
	Env        types.Map    `tfsdk:"env"`
	Args       types.List   `tfsdk:"args"`
}

// FieldManagerModel configures the field manager.
type FieldManagerModel struct {
	Name           types.String `tfsdk:"name"`
	ForceConflicts types.Bool   `tfsdk:"force_conflicts"`
}

// K8sProvider defines the provider implementation.
type K8sProvider struct {
	version string
	commit  string
}

func (p *K8sProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "k8s"
	resp.Version = p.version
}

func (p *K8sProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "K8s provider.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The hostname (in form of URI) of _Kubernetes_ master. Can be set with the `KUBE_HOST` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username to use for HTTP basic authentication when accessing the _Kubernetes_ master endpoint. Can be set with the `KUBE_USER` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password to use for HTTP basic authentication when accessing the _Kubernetes_ master endpoint. Can be set with the `KUBE_PASSWORD` environment variable.",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Whether server should be accessed without verifying the TLS certificate. Can be set with the `KUBE_INSECURE` environment variable.",
				Optional:            true,
			},
			"tls_server_name": schema.StringAttribute{
				MarkdownDescription: "Server name passed to the server for SNI and is used in the client to check server certificates against. Can be set with the `KUBE_TLS_SERVER_NAME` environment variable.",
				Optional:            true,
			},
			"client_certificate": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded client certificate for TLS authentication. Can be set with the `KUBE_CLIENT_CERT_DATA` environment variable.",
				Optional:            true,
			},
			"client_key": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded client certificate key for TLS authentication. Can be set with the `KUBE_CLIENT_KEY_DATA` environment variable.",
				Optional:            true,
			},
			"cluster_ca_certificate": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded root certificates bundle for TLS authentication. Can be set with the `KUBE_CLUSTER_CA_CERT_DATA` environment variable.",
				Optional:            true,
			},
			"config_paths": schema.ListAttribute{
				MarkdownDescription: "List of paths to the kube config file. Can be set with the `KUBE_CONFIG_PATHS` environment variable.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"config_context": schema.StringAttribute{
				MarkdownDescription: "Context to choose from the kube config file. Can be set with the `KUBE_CTX`environment variable.",
				Optional:            true,
			},
			"config_context_auth_info": schema.StringAttribute{
				MarkdownDescription: "Authentication info context of the kube config (name of the kube config user, --user flag in kubectl). Can be set with the `KUBE_CTX_AUTH_INFO` environment variable.",
				Optional:            true,
			},
			"config_context_cluster": schema.StringAttribute{
				MarkdownDescription: "Cluster context of the kube config (name of the kube config cluster, --cluster flag in kubectl). Can be set with the `KUBE_CTX_CLUSTER` environment variable.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Token to authenticate a service account. Can be set with the `KUBE_TOKEN` environment variable.",
				Optional:            true,
			},
			"proxy_url": schema.StringAttribute{
				MarkdownDescription: "URL to the proxy to be used for all API requests. Can be set with the `KUBE_PROXY_URL` environment variable.",
				Optional:            true,
			},
			"exec": schema.SingleNestedAttribute{
				MarkdownDescription: "Exec configuration for Kubernetes authentication",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"api_version": schema.StringAttribute{
						MarkdownDescription: "API version for the exec plugin.",
						Required:            true,
					},
					"command": schema.StringAttribute{
						MarkdownDescription: "Command to run for _Kubernetes_ exec plugin.",
						Required:            true,
					},
					"env": schema.MapAttribute{
						MarkdownDescription: "Environment variables for the exec plugin.",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"args": schema.ListAttribute{
						MarkdownDescription: "Arguments for the exec plugin.",
						ElementType:         types.StringType,
						Optional:            true,
					},
				},
			},
			"field_manager": schema.SingleNestedAttribute{
				MarkdownDescription: "Field manager configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Field manager name.",
						Optional:            true,
					},
					"force_conflicts": schema.BoolAttribute{
						MarkdownDescription: "If `true`, the field manager will force apply the changes by ignoring the conflicts.",
						Optional:            true,
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create:            true,
				CreateDescription: "Timeout for resource creation; defaults to `10m`. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Read:              true,
				ReadDescription:   "Timeout for resource or data source reads; defaults to `10m`. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Update:            true,
				UpdateDescription: "Timeout for resource update; defaults to `10m`. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Delete:            true,
				DeleteDescription: "Timeout for resource deletion; defaults to `10m`. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
			}),
		},
	}
}

// Configure configures the provider.
func (p *K8sProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	if req.ClientCapabilities.DeferralAllowed && !req.Config.Raw.IsFullyKnown() {
		resp.Deferred = &provider.Deferred{
			Reason: provider.DeferredReasonProviderConfigUnknown,
		}
	}

	// Load the provider config
	model := &K8sProviderModel{}
	if resp.Diagnostics.Append(req.Config.Get(ctx, model)...); resp.Diagnostics.HasError() {
		return
	}

	// Create a Kubernetes REST client configuration.
	restConfig, diags := getRestClientConfig(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the field manager config
	fieldManagerName := "terraform-provider-k8s"
	fieldManagerForceConflicts := false
	if model.FieldManager != nil {
		if !model.FieldManager.Name.IsNull() {
			fieldManagerName = model.FieldManager.Name.ValueString()
		}

		if !model.FieldManager.ForceConflicts.IsNull() {
			fieldManagerForceConflicts = model.FieldManager.ForceConflicts.ValueBool()
		}
	}

	// Lookup timeouts
	createTimeout, diags := model.Timeouts.Create(ctx, 10*time.Minute)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	readTimeout, diags := model.Timeouts.Read(ctx, 10*time.Minute)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	updateTimeout, diags := model.Timeouts.Update(ctx, 10*time.Minute)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	deleteTimeout, diags := model.Timeouts.Delete(ctx, 10*time.Minute)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	// Configure provider data
	providerData := &K8sProviderData{
		provider:     p,
		Model:        model,
		Client:       k8sutils.NewClient(restConfig),
		FieldManager: k8sutils.NewFieldManager(fieldManagerName, fieldManagerForceConflicts),
		DefaultTimeouts: &Timeouts{
			Create: createTimeout,
			Read:   readTimeout,
			Update: updateTimeout,
			Delete: deleteTimeout,
		},
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *K8sProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceResource,
	}
}

func (p *K8sProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *K8sProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewResourceDataSource,
		NewResourcesDataSource,
		NewServerVersionDataSource,
	}
}

func (p *K8sProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}
