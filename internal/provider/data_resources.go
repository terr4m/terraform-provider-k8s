package provider

import (
	"context"
	"fmt"

	"github.com/terr4m/terraform-provider-k8s/internal/k8sutils"
	"github.com/terr4m/terraform-provider-k8s/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ datasource.DataSource = &ResourcesDataSource{}
var _ datasource.DataSourceWithConfigure = &ResourcesDataSource{}

// NewResourcesDataSource creates a new server version data source.
func NewResourcesDataSource() datasource.DataSource {
	return &ResourcesDataSource{}
}

// ResourcesDataSource defines the data source implementation.
type ResourcesDataSource struct {
	providerData *K8sProviderData
}

// ResourcesDataSourceModel describes the data source data model.
type ResourcesDataSourceModel struct {
	APIVersion    types.String   `tfsdk:"api_version"`
	Kind          types.String   `tfsdk:"kind"`
	Namespace     types.String   `tfsdk:"namespace"`
	FieldSelector types.String   `tfsdk:"field_selector"`
	LabelSelector types.String   `tfsdk:"label_selector"`
	Limit         types.Number   `tfsdk:"limit"`
	Objects       types.Dynamic  `tfsdk:"objects"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the data source metadata.
func (d *ResourcesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_resources", req.ProviderTypeName)
}

// Schema returns the data source schema.
func (d *ResourcesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "_Kubernetes_ resources TF data source.",
		Attributes: map[string]schema.Attribute{
			"api_version": schema.StringAttribute{
				MarkdownDescription: "API version of the resources to find.",
				Required:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Kind of the resources to find.",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Namespace of the resources to find.",
				Optional:            true,
			},
			"field_selector": schema.StringAttribute{
				MarkdownDescription: "Field selector for the resources to find.",
				Optional:            true,
			},
			"label_selector": schema.StringAttribute{
				MarkdownDescription: "Label selector for the resources to find.",
				Optional:            true,
			},
			"limit": schema.NumberAttribute{
				MarkdownDescription: "Limit the number of resources to find.",
				Optional:            true,
			},
			"objects": schema.DynamicAttribute{
				MarkdownDescription: "List of resource objects retrieved from the API server. The following object fields are not returned; `status`, `metadata.creationTimestamp`, `metadata.generation`, `metadata.resourceVersion`, `metadata.selfLink`, `metadata.managedFields[*].time`.",
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
func (d *ResourcesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ResourcesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ri, err := d.providerData.Client.ResourceInterface(data.APIVersion.ValueString(), data.Kind.ValueString(), data.Namespace.ValueString(), false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure dynamic client.", err.Error())
		return
	}

	limit := int64(0)
	if !data.Limit.IsNull() {
		limit, _ = data.Limit.ValueBigFloat().Int64()
	}

	opts := metav1.ListOptions{
		FieldSelector: data.FieldSelector.ValueString(),
		LabelSelector: data.LabelSelector.ValueString(),
		Limit:         limit,
	}

	timeout, diags := data.Timeouts.Read(ctx, d.providerData.DefaultTimeouts.Read)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	l, err := ri.List(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list resources.", err.Error())
		return
	}

	col, diags := tfutils.DecodeDynamic(ctx, k8sutils.UnstructuredListToObjects(l))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Objects = col

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
