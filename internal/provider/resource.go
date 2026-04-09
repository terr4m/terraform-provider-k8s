package provider

import (
	"context"
	"fmt"

	"github.com/terr4m/terraform-provider-k8s/internal/k8sutils"
	"github.com/terr4m/terraform-provider-k8s/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &ResourceResource{}
	_ resource.ResourceWithConfigure      = &ResourceResource{}
	_ resource.ResourceWithValidateConfig = &ResourceResource{}
	_ resource.ResourceWithModifyPlan     = &ResourceResource{}
)

// NewResourceResource creates a new managed kubernetes resource.
func NewResourceResource() resource.Resource {
	return &ResourceResource{}
}

// ResourceResource defines the resource implementation.
type ResourceResource struct {
	providerData *K8sProviderData
}

// ResourceResourceModel describes the resource data model.
type ResourceResourceModel struct {
	FieldManager             *FieldManagerModel `tfsdk:"field_manager"`
	Manifest                 types.Dynamic      `tfsdk:"manifest"`
	ManagedFieldsFingerprint types.String       `tfsdk:"managed_fields_fingerprint"`
	Timeouts                 timeouts.Value     `tfsdk:"timeouts"`
}

func (r *ResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_resource", req.ProviderTypeName)
}

func (r *ResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "_Kubernetes_ resource managed with server-side apply.",
		Attributes: map[string]schema.Attribute{
			"field_manager": schema.SingleNestedAttribute{
				MarkdownDescription: "Field manager configuration for this resource, overriding provider defaults.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Field manager name.",
						Optional:            true,
					},
					"force_conflicts": schema.BoolAttribute{
						MarkdownDescription: "Whether to force conflicts during server-side apply.",
						Optional:            true,
					},
				},
			},
			"manifest": schema.DynamicAttribute{
				MarkdownDescription: "Partial resource manifest describing the desired managed subset.",
				Required:            true,
			},
			"managed_fields_fingerprint": schema.StringAttribute{
				MarkdownDescription: "Computed fingerprint of the live object projected onto the managed manifest subset.",
				Computed:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create:            true,
				CreateDescription: "Timeout for creating the resource; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Read:              true,
				ReadDescription:   "Timeout for reading the resource; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Update:            true,
				UpdateDescription: "Timeout for updating the resource; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Delete:            true,
				DeleteDescription: "Timeout for deleting the resource; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
			}),
		},
	}
}

func (r *ResourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*K8sProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource provider data.", fmt.Sprintf("expected *K8sProviderData, got: %T", req.ProviderData))
		return
	}

	r.providerData = providerData
}

func (r *ResourceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.Config.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	if data.Manifest.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest is required.", "expected non-null manifest")
		return
	}
	if data.Manifest.IsUnknown() || !tfutils.IsFullyKnown(data.Manifest) {
		return
	}

	manifest, err := tfutils.EncodeDynamicObject(data.Manifest)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest type not object.", err.Error())
		return
	}

	for _, validationErr := range k8sutils.ValidateManifest(manifest) {
		attributePath := path.Root("manifest")
		switch validationErr.Path {
		case "", "manifest":
		case "apiVersion":
			attributePath = attributePath.AtName("apiVersion")
		case "kind":
			attributePath = attributePath.AtName("kind")
		case "metadata":
			attributePath = attributePath.AtName("metadata")
		case "metadata.name":
			attributePath = attributePath.AtName("metadata").AtName("name")
		case "status":
			attributePath = attributePath.AtName("status")
		default:
			switch validationErr.Path {
			case "metadata.creationTimestamp":
				attributePath = attributePath.AtName("metadata").AtName("creationTimestamp")
			case "metadata.generation":
				attributePath = attributePath.AtName("metadata").AtName("generation")
			case "metadata.managedFields":
				attributePath = attributePath.AtName("metadata").AtName("managedFields")
			case "metadata.resourceVersion":
				attributePath = attributePath.AtName("metadata").AtName("resourceVersion")
			case "metadata.selfLink":
				attributePath = attributePath.AtName("metadata").AtName("selfLink")
			case "metadata.uid":
				attributePath = attributePath.AtName("metadata").AtName("uid")
			}
		}

		resp.Diagnostics.AddAttributeError(attributePath, validationErr.Summary, validationErr.Detail)
	}
}

func (r *ResourceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan ResourceResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...); resp.Diagnostics.HasError() {
		return
	}

	if !req.State.Raw.IsNull() {
		resp.RequiresReplace = append(resp.RequiresReplace,
			path.Root("manifest").AtName("kind"),
			path.Root("manifest").AtName("metadata").AtName("name"),
		)

		var state ResourceResourceModel
		if resp.Diagnostics.Append(req.State.Get(ctx, &state)...); resp.Diagnostics.HasError() {
			return
		}

		manifest, ok := state.Manifest.UnderlyingValue().(types.Object)
		if !ok {
			resp.Diagnostics.AddError("Failed to access state manifest.", "expected manifest to be an object")
			return
		}

		metadataValue, ok := manifest.Attributes()["metadata"]
		if !ok {
			resp.Diagnostics.AddError("Failed to access state metadata.", "expected manifest.metadata to be set")
			return
		}

		metadata, ok := metadataValue.(types.Object)
		if !ok {
			resp.Diagnostics.AddError("Failed to access state metadata.", "expected metadata to be an object")
			return
		}

		if namespace, ok := metadata.Attributes()["namespace"]; ok && !namespace.IsNull() {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("manifest").AtName("metadata").AtName("namespace"))
		}
	}

	if plan.Manifest.IsUnknown() || !tfutils.IsFullyKnown(plan.Manifest) {
		plan.ManagedFieldsFingerprint = types.StringUnknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	manifest, err := tfutils.EncodeDynamicObject(plan.Manifest)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest type not object.", err.Error())
		return
	}

	fingerprint, err := k8sutils.DesiredManifestFingerprint(manifest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fingerprint manifest.", err.Error())
		return
	}

	plan.ManagedFieldsFingerprint = types.StringValue(fingerprint)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *ResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	manifest, err := tfutils.EncodeDynamicObject(data.Manifest)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest type not object.", err.Error())
		return
	}

	timeout, diags := data.Timeouts.Create(ctx, r.providerData.DefaultTimeouts.Create)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var fieldManager *k8sutils.FieldManager
	if data.FieldManager != nil {
		fieldManager = &k8sutils.FieldManager{}
		if !data.FieldManager.Name.IsNull() {
			fieldManager.Name = data.FieldManager.Name.ValueString()
		}
		if !data.FieldManager.ForceConflicts.IsNull() {
			fieldManager.ForceConflicts = data.FieldManager.ForceConflicts.ValueBool()
		}
	}

	fingerprint, err := k8sutils.NewApplyManager(r.providerData.Client, k8sutils.FieldManager{
		Name:           r.providerData.FieldManager.Name,
		ForceConflicts: r.providerData.FieldManager.ForceConflicts,
	}).Create(ctx, manifest, fieldManager)
	if err != nil {
		resp.Diagnostics.AddError("Failed to apply manifest.", err.Error())
		return
	}

	data.ManagedFieldsFingerprint = types.StringValue(fingerprint)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	manifest, err := tfutils.EncodeDynamicObject(data.Manifest)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest type not object.", err.Error())
		return
	}

	timeout, diags := data.Timeouts.Read(ctx, r.providerData.DefaultTimeouts.Read)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := k8sutils.NewApplyManager(r.providerData.Client, k8sutils.FieldManager{
		Name:           r.providerData.FieldManager.Name,
		ForceConflicts: r.providerData.FieldManager.ForceConflicts,
	}).Read(ctx, manifest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get resource.", err.Error())
		return
	}
	if !result.Found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ManagedFieldsFingerprint = types.StringValue(result.Fingerprint)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	manifest, err := tfutils.EncodeDynamicObject(data.Manifest)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest type not object.", err.Error())
		return
	}

	timeout, diags := data.Timeouts.Update(ctx, r.providerData.DefaultTimeouts.Update)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var fieldManager *k8sutils.FieldManager
	if data.FieldManager != nil {
		fieldManager = &k8sutils.FieldManager{}
		if !data.FieldManager.Name.IsNull() {
			fieldManager.Name = data.FieldManager.Name.ValueString()
		}
		if !data.FieldManager.ForceConflicts.IsNull() {
			fieldManager.ForceConflicts = data.FieldManager.ForceConflicts.ValueBool()
		}
	}

	fingerprint, err := k8sutils.NewApplyManager(r.providerData.Client, k8sutils.FieldManager{
		Name:           r.providerData.FieldManager.Name,
		ForceConflicts: r.providerData.FieldManager.ForceConflicts,
	}).Update(ctx, manifest, fieldManager)
	if err != nil {
		resp.Diagnostics.AddError("Failed to apply manifest.", err.Error())
		return
	}

	data.ManagedFieldsFingerprint = types.StringValue(fingerprint)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	manifest, err := tfutils.EncodeDynamicObject(data.Manifest)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest type not object.", err.Error())
		return
	}

	timeout, diags := data.Timeouts.Delete(ctx, r.providerData.DefaultTimeouts.Delete)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = k8sutils.NewApplyManager(r.providerData.Client, k8sutils.FieldManager{
		Name:           r.providerData.FieldManager.Name,
		ForceConflicts: r.providerData.FieldManager.ForceConflicts,
	}).Delete(ctx, manifest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete resource.", err.Error())
	}
}
