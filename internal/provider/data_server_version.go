package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerVersionDataSource{}

// NewServerVersionDataSource creates a new server version data source.
func NewServerVersionDataSource() datasource.DataSource {
	return &ServerVersionDataSource{}
}

// ServerVersionDataSource defines the data source implementation.
type ServerVersionDataSource struct {
	providerData *K8sProviderData
}

// ServerVersionDataSourceModel describes the data source data model.
type ServerVersionDataSourceModel struct {
	Major        types.String `tfsdk:"major"`
	Minor        types.String `tfsdk:"minor"`
	GitVersion   types.String `tfsdk:"git_version"`
	GitCommit    types.String `tfsdk:"git_commit"`
	GitTreeState types.String `tfsdk:"git_tree_state"`
	BuildDate    types.String `tfsdk:"build_date"`
	GoVersion    types.String `tfsdk:"go_version"`
	Compiler     types.String `tfsdk:"compiler"`
	Platform     types.String `tfsdk:"platform"`
}

// Metadata returns the data source metadata.
func (d *ServerVersionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_server_version", req.ProviderTypeName)
}

// Schema returns the data source schema.
func (d *ServerVersionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "_Kubernetes_ server version data source.",
		Attributes: map[string]schema.Attribute{
			"major": schema.StringAttribute{
				MarkdownDescription: "Major version of the server.",
				Computed:            true,
			},
			"minor": schema.StringAttribute{
				MarkdownDescription: "Minor version of the server.",
				Computed:            true,
			},
			"git_version": schema.StringAttribute{
				MarkdownDescription: "Git version of the server.",
				Computed:            true,
			},
			"git_commit": schema.StringAttribute{
				MarkdownDescription: "Git commit of the server.",
				Computed:            true,
			},
			"git_tree_state": schema.StringAttribute{
				MarkdownDescription: "Git tree state of the server.",
				Computed:            true,
			},
			"build_date": schema.StringAttribute{
				MarkdownDescription: "Build date of the server.",
				Computed:            true,
			},
			"go_version": schema.StringAttribute{
				MarkdownDescription: "Go version of the server.",
				Computed:            true,
			},
			"compiler": schema.StringAttribute{
				MarkdownDescription: "Compiler of the server.",
				Computed:            true,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "Platform of the server.",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source.
func (d *ServerVersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*K8sProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source provider data.", fmt.Sprintf("expected *http.Client, got: %T", req.ProviderData))
		return
	}

	d.providerData = providerData
}

// Read reads the data source.
func (d *ServerVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerVersionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	discoveryClient, err := d.providerData.Client.DiscoveryClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure discovery client.", err.Error())
		return
	}

	version, err := discoveryClient.ServerVersion()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get server version.", err.Error())
		return
	}

	data.Major = types.StringValue(version.Major)
	data.Minor = types.StringValue(version.Minor)
	data.GitVersion = types.StringValue(version.GitVersion)
	data.GitCommit = types.StringValue(version.GitCommit)
	data.GitTreeState = types.StringValue(version.GitTreeState)
	data.BuildDate = types.StringValue(version.BuildDate)
	data.GoVersion = types.StringValue(version.GoVersion)
	data.Compiler = types.StringValue(version.Compiler)
	data.Platform = types.StringValue(version.Platform)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
