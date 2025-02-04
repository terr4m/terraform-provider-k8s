package provider

import (
	"context"
	"fmt"

	"github.com/terr4m/terraform-provider-k8s/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ datasource.DataSource = &ResourceDataSource{}
var _ datasource.DataSourceWithConfigure = &ResourceDataSource{}

// NewResourceDataSource creates a new resource data source.
func NewResourceDataSource() datasource.DataSource {
	return &ResourceDataSource{}
}

// ResourceDataSource defines the data source implementation.
type ResourceDataSource struct {
	providerData *K8sProviderData
}

// ResourceDataSourceModel describes the data source data model.
type ResourceDataSourceModel struct {
	APIVersion types.String   `tfsdk:"api_version"`
	Kind       types.String   `tfsdk:"kind"`
	Namespace  types.String   `tfsdk:"namespace"`
	Name       types.String   `tfsdk:"name"`
	Object     types.Dynamic  `tfsdk:"object"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the data source metadata.
func (d *ResourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_resource", req.ProviderTypeName)
}

// Schema returns the data source schema.
func (d *ResourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "_Kubernetes_ resource TF data source.",
		Attributes: map[string]schema.Attribute{
			"api_version": schema.StringAttribute{
				MarkdownDescription: "API version of the resource to find.",
				Required:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Kind of the resource to find.",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Namespace of the resource to find; if the resource is cluster scoped this should not be set. If a namespace isn't set for a namespaced resource, the `default` namespace will be used.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the resource to find.",
				Required:            true,
			},
			"object": schema.DynamicAttribute{
				MarkdownDescription: "Resource object retrieved from the API server. The following fields are not returned; `status`, `metadata.creationTimestamp`, `metadata.generation`, `metadata.resourceVersion`, `metadata.selfLink`, `metadata.managedFields[*].time`.",
				Computed:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Read:            true,
				ReadDescription: "Timeout for reading the data source; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
			}),
		},
	}
}

// Configure configures the data source.
func (d *ResourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*K8sProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source provider data.", fmt.Sprintf("expected *K8sProviderData, got: %T", req.ProviderData))
		return
	}

	d.providerData = providerData
}

// Read reads the data source.
func (d *ResourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ResourceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ri, err := d.providerData.Client.ResourceInterface(data.APIVersion.ValueString(), data.Kind.ValueString(), data.Namespace.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure dynamic client.", err.Error())
		return
	}

	timeout, diags := data.Timeouts.Read(ctx, d.providerData.DefaultTimeouts.Read)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	o, err := ri.Get(ctx, data.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get resource.", err.Error())
		return
	}

	obj, diags := tfutils.DecodeDynamic(ctx, o.Object)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Object = obj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
