package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/terr4m/terraform-provider-k8s/internal/k8sutils"
	"github.com/terr4m/terraform-provider-k8s/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	_ resource.Resource                   = &ResourceResource{}
	_ resource.ResourceWithConfigure      = &ResourceResource{}
	_ resource.ResourceWithValidateConfig = &ResourceResource{}
	_ resource.ResourceWithModifyPlan     = &ResourceResource{}
	_ resource.ResourceWithImportState    = &ResourceResource{}
)

// NewResourceResource creates a new resource resource.
func NewResourceResource() resource.Resource {
	return &ResourceResource{}
}

// ResourceResource defines the resource implementation.
type ResourceResource struct {
	providerData *K8sProviderData
}

// ResourceResourceModel describes the resource data model.
type ResourceResourceModel struct {
	FieldManager  *FieldManagerModel  `tfsdk:"field_manager"`
	IgnoreFields  types.List          `tfsdk:"ignore_fields"`
	WaitOptions   *WaitOptionsModel   `tfsdk:"wait_options"`
	DeleteOptions *DeleteOptionsModel `tfsdk:"delete_options"`
	Manifest      types.Dynamic       `tfsdk:"manifest"`
	Object        types.Dynamic       `tfsdk:"object"`
	Timeouts      timeouts.Value      `tfsdk:"timeouts"`
}

// WaitOptionsModel describes the wait options.
type WaitOptionsModel struct {
	Conditions     types.List `tfsdk:"conditions"`
	FieldSelectors types.List `tfsdk:"field_selectors"`
	Rollout        types.Bool `tfsdk:"rollout"`
}

// ConditionModel describes the condition options.
type ConditionModel struct {
	Type   types.String `tfsdk:"type"`
	Status types.String `tfsdk:"status"`
}

// FieldSelectorModel describes the field selector options.
type FieldSelectorModel struct {
	Key      types.String `tfsdk:"key"`
	Value    types.String `tfsdk:"value"`
	Operator types.String `tfsdk:"operator"`
}

// DeleteOptionsModel describes the delete options.
type DeleteOptionsModel struct {
	PropagationPolicy types.String `tfsdk:"propagation_policy"`
	Wait              types.Bool   `tfsdk:"wait"`
}

// Metadata returns the resource metadata.
func (d *ResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_resource", req.ProviderTypeName)
}

// Schema returns the resource schema.
func (r *ResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "_Kubernetes_ resource TF resource.",
		Attributes: map[string]schema.Attribute{
			"field_manager": schema.SingleNestedAttribute{
				MarkdownDescription: "Field manager configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Field manager name, this will override the value set on the provider.",
						Optional:            true,
					},
					"force_conflicts": schema.BoolAttribute{
						MarkdownDescription: "If `true`, the field manager will force apply the changes by ignoring the conflicts.",
						Optional:            true,
					},
				},
			},
			"ignore_fields": schema.ListAttribute{
				MarkdownDescription: "List of fields to ignore when returned from the server.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"wait_options": schema.SingleNestedAttribute{
				MarkdownDescription: "Configure waiter options.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"conditions": schema.ListNestedAttribute{
						MarkdownDescription: "List of conditions to wait for.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "Type of condition.",
									Required:            true,
								},
								"status": schema.StringAttribute{
									MarkdownDescription: "Status of the condition.",
									Required:            true,
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"field_selectors": schema.ListNestedAttribute{
						MarkdownDescription: "List of field selectors to wait for.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									MarkdownDescription: "Key of the field.",
									Required:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "Value of the field.",
									Required:            true,
								},
								"operator": schema.StringAttribute{
									MarkdownDescription: "Operator to compare the field value; can be one of `eq` or `ne`.",
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString("eq"),
									Validators: []validator.String{
										stringvalidator.OneOfCaseInsensitive("eq", "ne"),
									},
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"rollout": schema.BoolAttribute{
						MarkdownDescription: "If the rollout should be completed before returning; this is supported on `DaemonSet`, `Deployment` & `StatefulSet` resources.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"delete_options": schema.SingleNestedAttribute{
				MarkdownDescription: "Configure delete options.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"propagation_policy": schema.StringAttribute{
						MarkdownDescription: "Propagation policy for the delete operation.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOfCaseInsensitive("Background", "Foreground", "Orphans"),
						},
					},
					"wait": schema.BoolAttribute{
						MarkdownDescription: "If the delete operation will wait for the resource to be deleted.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"manifest": schema.DynamicAttribute{
				MarkdownDescription: "_Kubernetes_ resource manifest describing the desired state of the resource.",
				Required:            true,
			},
			"object": schema.DynamicAttribute{
				MarkdownDescription: "Resource object retrieved from the API server. The following fields are not returned; `status`, `metadata.creationTimestamp`, `metadata.generation`, `metadata.resourceVersion`, `metadata.selfLink`, `metadata.managedFields[*].time`.",
				Computed:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create:            true,
				CreateDescription: "Timeout for creating the resource; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Read:              true,
				ReadDescription:   "Timeout for reading the resource; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Update:            true,
				UpdateDescription: "Timeout for updating the resource; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
				Delete:            true,
				DeleteDescription: "Timeout for deleting the resource; this defaults to the provider value if not set. This should be a string that can be [parsed as a duration] (https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as `30s` or `2h45m`. Valid time units are `s` (seconds), `m` (minutes), `h` (hours).",
			}),
		},
	}
}

// Configure configures the resource.
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

// ValidateConfig validates the resource config.
func (r *ResourceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.Config.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	if data.Manifest.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest is required.", "expected non-null manifest")
		return
	}

	o, ok := data.Manifest.UnderlyingValue().(types.Object)
	if !ok {
		resp.Diagnostics.AddAttributeError(path.Root("manifest"), "Manifest type not object.", "expected object manifest type")
		return
	}

	attrs := o.Attributes()

	for _, v := range []string{"apiVersion", "kind", "metadata"} {
		if _, ok := attrs[v]; !ok {
			resp.Diagnostics.AddAttributeError(path.Root("manifest").AtName(v), "Missing required attribute.", fmt.Sprintf("expected manifest to have attribute %q", v))
		}
	}

	for _, v := range k8sutils.ServerSideFields() {
		if _, ok := attrs[v]; ok {
			resp.Diagnostics.AddAttributeError(path.Root("manifest").AtName(v), "Forbidden attribute.", fmt.Sprintf("forbidden manifest attribute %q", v))
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if data.WaitOptions != nil && data.WaitOptions.Rollout.ValueBool() {
		if attr, ok := attrs["kind"]; ok {
			kind, ok := attr.(types.String)
			if !ok {
				resp.Diagnostics.AddError("Failed to check kind.", "expected kind to be a string")
				return
			}

			if !slices.Contains(k8sutils.RolloutKinds(), kind.ValueString()) {
				resp.Diagnostics.AddAttributeError(path.Root("wait").AtName("rollout"), "Unsupported kind for rollout.", "rollout requires the kind to be one of DaemonSet, Deployment or StatefulSet")
			}
		}
	}
}

// ModifyPlan modifies the resource plan.
func (r *ResourceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan ResourceResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...); resp.Diagnostics.HasError() {
		return
	}

	if !plan.Object.IsUnknown() {
		return
	}

	var ignoreFields path.Expressions
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("creationTimestamp"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("deletionGracePeriodSeconds"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("deletionTimestamp"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("finalizers"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("generateName"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("generation"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("managedFields")) // .AtAnyListIndex().AtName("time"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("minReadySeconds"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("ownerReferences"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("paused"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("resourceVersion"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("selfLink"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("status"))

	var unknownFields path.Expressions
	unknownFields = unknownFields.Append(path.MatchRoot("metadata").AtName("annotations"))

	if !req.State.Raw.IsNull() {
		resp.RequiresReplace = append(resp.RequiresReplace,
			path.Root("manifest").AtName("kind"),
			path.Root("manifest").AtName("metadata").AtName("name"),
		)

		m, ok := plan.Manifest.UnderlyingValue().(types.Object)
		if !ok {
			resp.Diagnostics.AddError("Failed to access manifest.", "expected manifest to be an object")
		}
		attr := m.Attributes()

		meta, ok := attr["metadata"].(types.Object)
		if !ok {
			resp.Diagnostics.AddError("Failed to access metadata.", "expected metadata to be an object")
		}
		metaAttr := meta.Attributes()

		ns, ok := metaAttr["namespace"]
		if ok && !ns.IsNull() {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("manifest").AtName("metadata").AtName("namespace"))
		}
	} else {
		unknownFields = unknownFields.Append(path.MatchRoot("metadata").AtName("uid"))
	}

	if !tfutils.IsFullyKnown(plan.Manifest) {
		return
	}

	m, err := tfutils.EncodeDynamicObject(ctx, plan.Manifest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to encode plan manifest.", err.Error())
		return
	}
	u := &unstructured.Unstructured{Object: m}
	namespace := u.GetNamespace()

	gvkExists, err := r.providerData.Client.CheckGVK(u.GetAPIVersion(), u.GetKind())
	if err != nil {
		resp.Diagnostics.AddError("Failed to check GVK.", err.Error())
		return
	}

	checkNamespace, err := k8sutils.NamespaceEmptyOrExists(ctx, r.providerData.Client, namespace)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check namespace.", err.Error())
		return
	}

	if gvkExists && checkNamespace {
		gvk, err := k8sutils.ParseGVK(u.GetAPIVersion(), u.GetKind())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse GVK.", err.Error())
			return
		}

		ri, err := r.providerData.Client.ResourceInterface(gvk, namespace, true)
		if err != nil {
			resp.Diagnostics.AddError("Failed to configure dynamic client.", err.Error())
			return
		}

		timeout, diags := plan.Timeouts.Read(ctx, r.providerData.DefaultTimeouts.Read)
		if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
			return
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		opts := metav1.ApplyOptions{
			FieldManager: r.providerData.FieldManager.Name,
			Force:        r.providerData.FieldManager.ForceConflicts,
			DryRun:       []string{metav1.DryRunAll},
		}

		if plan.FieldManager != nil {
			if !plan.FieldManager.Name.IsNull() {
				opts.FieldManager = plan.FieldManager.Name.ValueString()
			}

			if !plan.FieldManager.ForceConflicts.IsNull() {
				opts.Force = plan.FieldManager.ForceConflicts.ValueBool()
			}
		}

		o, err := ri.Apply(ctx, u.GetName(), u, opts)
		if err != nil {
			resp.Diagnostics.AddError("Failed to dry-run apply manifest.", err.Error())
			return
		}

		// a := o.Object
		// if err = k8sutils.RemoveServerSideFields(opts.FieldManager, m, a, false, false); err != nil {
		// 	resp.Diagnostics.AddError("Failed to remove server-side fields.", err.Error())
		// 	return
		// }

		ogv, err := r.providerData.Client.GetGVOpenAPISchemaLookup(gvk)
		if err != nil {
			resp.Diagnostics.AddError("Failed to get schema lookup.", err.Error())
			return
		}

		sc, err := k8sutils.GetOpenAPISchema(ogv, gvk)
		if err != nil {
			resp.Diagnostics.AddError("Failed to get schema.", err.Error())
			return
		}

		obj, err := tfutils.DecodeDynamicWithTypeAndUnknowns(ctx, sc, ignoreFields, unknownFields, o.Object)
		if err != nil {
			resp.Diagnostics.AddError("Failed to decode value.", err.Error())
			return
		}

		plan.Object = obj

		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

// Create creates the resource.
func (r *ResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	m, err := tfutils.EncodeDynamicObject(ctx, data.Manifest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to encode manifest.", err.Error())
		return
	}
	u := &unstructured.Unstructured{Object: m}

	name := u.GetName()
	gvk, err := k8sutils.ParseGVK(u.GetAPIVersion(), u.GetKind())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse GVK.", err.Error())
		return
	}

	ri, err := r.providerData.Client.ResourceInterface(gvk, u.GetNamespace(), true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure dynamic client.", err.Error())
		return
	}

	timeout, diags := data.Timeouts.Create(ctx, r.providerData.DefaultTimeouts.Create)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err = ri.Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		resp.Diagnostics.AddError("Resource already exists.", fmt.Sprintf("resource %q already exists.", name))
		return
	} else if !errors.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to determine if resource already exists.", err.Error())
		return
	}

	opts := metav1.ApplyOptions{
		FieldManager: r.providerData.FieldManager.Name,
		Force:        r.providerData.FieldManager.ForceConflicts,
	}

	if data.FieldManager != nil {
		if !data.FieldManager.Name.IsNull() {
			opts.FieldManager = data.FieldManager.Name.ValueString()
		}

		if !data.FieldManager.ForceConflicts.IsNull() {
			opts.Force = data.FieldManager.ForceConflicts.ValueBool()
		}
	}

	var watcher watch.Interface
	if data.WaitOptions != nil && (!data.WaitOptions.Conditions.IsNull() || !data.WaitOptions.FieldSelectors.IsNull() || data.WaitOptions.Rollout.ValueBool()) {
		fsl, diags := getFieldSelectorStrings(ctx, data.WaitOptions.FieldSelectors)
		if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
			return
		}

		watcher, err = k8sutils.GetWatcher(ctx, ri, name, fsl...)
		if err != nil {
			resp.Diagnostics.AddError("Failed to watch resource.", err.Error())
			return
		}
		defer watcher.Stop()
	}

	o, err := ri.Apply(ctx, name, u, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to apply manifest.", err.Error())
		return
	}

	if watcher != nil {
		conds, diags := getConditionsMap(ctx, data.WaitOptions.Conditions)
		if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
			return
		}

		u, err := k8sutils.WatchForAddedModified(ctx, watcher, conds, data.WaitOptions.Rollout.ValueBool(), gvk.Kind)
		if err != nil {
			resp.Diagnostics.AddError("Failed to wait for resource.", err.Error())
			return
		}
		o = u
	}

	// a := o.Object
	// if err = k8sutils.RemoveServerSideFields(opts.FieldManager, m, a, false, false); err != nil {
	// 	resp.Diagnostics.AddError("Failed to remove server-side fields.", err.Error())
	// 	return
	// }

	ogv, err := r.providerData.Client.GetGVOpenAPISchemaLookup(gvk)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get schema lookup.", err.Error())
		return
	}

	sc, err := k8sutils.GetOpenAPISchema(ogv, gvk)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get schema.", err.Error())
		return
	}

	var ignoreFields path.Expressions
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("creationTimestamp"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("deletionGracePeriodSeconds"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("deletionTimestamp"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("finalizers"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("generateName"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("generation"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("managedFields")) // .AtAnyListIndex().AtName("time"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("minReadySeconds"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("ownerReferences"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("paused"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("resourceVersion"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("selfLink"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("status"))

	obj, err := tfutils.DecodeDynamicWithType(ctx, sc, ignoreFields, o.Object)
	if err != nil {
		resp.Diagnostics.AddError("Failed to decode value.", err.Error())
		return
	}

	data.Object = obj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the resource state.
func (r *ResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	m, err := tfutils.EncodeDynamicObject(ctx, data.Manifest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to encode manifest.", err.Error())
		return
	}
	u := &unstructured.Unstructured{Object: m}

	gvk, err := k8sutils.ParseGVK(u.GetAPIVersion(), u.GetKind())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse GVK.", err.Error())
		return
	}

	ri, err := r.providerData.Client.ResourceInterface(gvk, u.GetNamespace(), true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure dynamic client.", err.Error())
		return
	}

	timeout, diags := data.Timeouts.Read(ctx, r.providerData.DefaultTimeouts.Read)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	o, err := ri.Get(ctx, u.GetName(), metav1.GetOptions{})
	if !errors.IsNotFound(err) && err != nil {
		resp.Diagnostics.AddError("Failed to get resource.", err.Error())
		return
	}

	if errors.IsNotFound(err) {
		resp.Diagnostics.AddWarning("Failed to get resource.", err.Error())
		data.Object = types.DynamicNull()
	} else {
		// var fieldManager string
		// if data.FieldManager != nil && !data.FieldManager.Name.IsNull() {
		// 	fieldManager = data.FieldManager.Name.ValueString()
		// } else {
		// 	fieldManager = r.providerData.FieldManager.Name
		// }

		// a := o.Object
		// if err = k8sutils.RemoveServerSideFields(fieldManager, m, a, false, false); err != nil {
		// 	resp.Diagnostics.AddError("Failed to remove server-side fields.", err.Error())
		// 	return
		// }

		ogv, err := r.providerData.Client.GetGVOpenAPISchemaLookup(gvk)
		if err != nil {
			resp.Diagnostics.AddError("Failed to get schema lookup.", err.Error())
			return
		}

		sc, err := k8sutils.GetOpenAPISchema(ogv, gvk)
		if err != nil {
			resp.Diagnostics.AddError("Failed to get schema.", err.Error())
			return
		}

		var ignoreFields path.Expressions
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("creationTimestamp"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("deletionGracePeriodSeconds"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("deletionTimestamp"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("finalizers"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("generateName"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("generation"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("managedFields")) // .AtAnyListIndex().AtName("time"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("minReadySeconds"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("ownerReferences"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("paused"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("resourceVersion"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("selfLink"))
		ignoreFields = ignoreFields.Append(path.MatchRoot("status"))

		obj, err := tfutils.DecodeDynamicWithType(ctx, sc, ignoreFields, o.Object)
		if err != nil {
			resp.Diagnostics.AddError("Failed to decode value.", err.Error())
			return
		}

		data.Object = obj
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *ResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	m, err := tfutils.EncodeDynamicObject(ctx, data.Manifest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to encode manifest.", err.Error())
		return
	}
	u := &unstructured.Unstructured{Object: m}

	name := u.GetName()
	gvk, err := k8sutils.ParseGVK(u.GetAPIVersion(), u.GetKind())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse GVK.", err.Error())
		return
	}

	ri, err := r.providerData.Client.ResourceInterface(gvk, u.GetNamespace(), true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure dynamic client.", err.Error())
		return
	}

	opts := metav1.ApplyOptions{
		FieldManager: r.providerData.FieldManager.Name,
		Force:        r.providerData.FieldManager.ForceConflicts,
	}

	if data.FieldManager != nil {
		if !data.FieldManager.Name.IsNull() {
			opts.FieldManager = data.FieldManager.Name.ValueString()
		}

		if !data.FieldManager.ForceConflicts.IsNull() {
			opts.Force = data.FieldManager.ForceConflicts.ValueBool()
		}
	}

	timeout, diags := data.Timeouts.Update(ctx, r.providerData.DefaultTimeouts.Update)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var watcher watch.Interface
	if data.WaitOptions != nil && (!data.WaitOptions.Conditions.IsNull() || !data.WaitOptions.FieldSelectors.IsNull() || data.WaitOptions.Rollout.ValueBool()) {
		fsl, diags := getFieldSelectorStrings(ctx, data.WaitOptions.FieldSelectors)
		if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
			return
		}

		watcher, err = k8sutils.GetWatcher(ctx, ri, name, fsl...)
		if err != nil {
			resp.Diagnostics.AddError("Failed to watch resource.", err.Error())
			return
		}
		defer watcher.Stop()
	}

	o, err := ri.Apply(ctx, name, u, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to apply manifest.", err.Error())
		return
	}

	if watcher != nil {
		conds, diags := getConditionsMap(ctx, data.WaitOptions.Conditions)
		if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
			return
		}

		u, err := k8sutils.WatchForAddedModified(ctx, watcher, conds, data.WaitOptions.Rollout.ValueBool(), gvk.Kind)
		if err != nil {
			resp.Diagnostics.AddError("Failed to wait for resource.", err.Error())
			return
		}
		o = u
	}

	// a := o.Object
	// if err = k8sutils.RemoveServerSideFields(opts.FieldManager, m, a, false, false); err != nil {
	// 	resp.Diagnostics.AddError("Failed to remove server-side fields.", err.Error())
	// 	return
	// }

	ogv, err := r.providerData.Client.GetGVOpenAPISchemaLookup(gvk)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get schema lookup.", err.Error())
		return
	}

	sc, err := k8sutils.GetOpenAPISchema(ogv, gvk)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get schema.", err.Error())
		return
	}

	var ignoreFields path.Expressions
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("creationTimestamp"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("deletionGracePeriodSeconds"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("deletionTimestamp"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("finalizers"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("generateName"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("generation"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("managedFields")) // .AtAnyListIndex().AtName("time"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("minReadySeconds"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("ownerReferences"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("paused"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("resourceVersion"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("metadata").AtName("selfLink"))
	ignoreFields = ignoreFields.Append(path.MatchRoot("status"))

	obj, err := tfutils.DecodeDynamicWithType(ctx, sc, ignoreFields, o.Object)
	if err != nil {
		resp.Diagnostics.AddError("Failed to decode value.", err.Error())
		return
	}

	data.Object = obj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *ResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	m, err := tfutils.EncodeDynamicObject(ctx, data.Manifest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to encode manifest.", err.Error())
		return
	}
	u := &unstructured.Unstructured{Object: m}

	gvk, err := k8sutils.ParseGVK(u.GetAPIVersion(), u.GetKind())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse GVK.", err.Error())
		return
	}

	ri, err := r.providerData.Client.ResourceInterface(gvk, u.GetNamespace(), true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure dynamic client.", err.Error())
		return
	}

	timeout, diags := data.Timeouts.Delete(ctx, r.providerData.DefaultTimeouts.Delete)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	opts := metav1.DeleteOptions{}

	var watcher watch.Interface
	if data.DeleteOptions != nil {
		if !data.DeleteOptions.PropagationPolicy.IsNull() {
			opts.PropagationPolicy = k8sutils.Ptr(metav1.DeletionPropagation(data.DeleteOptions.PropagationPolicy.ValueString()))
		}

		if data.DeleteOptions.Wait.ValueBool() {
			watcher, err = k8sutils.GetWatcher(ctx, ri, u.GetName())
			if err != nil {
				resp.Diagnostics.AddError("Failed to watch resource.", err.Error())
				return
			}
			defer watcher.Stop()
		}
	}

	err = ri.Delete(ctx, u.GetName(), opts)
	if errors.IsGone(err) || errors.IsNotFound(err) {
		resp.Diagnostics.AddWarning("Resource already deleted.", err.Error())
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to delete resource.", err.Error())
		return
	}

	if watcher != nil {
		err = k8sutils.WatchForDelete(ctx, watcher)
		if err != nil {
			resp.Diagnostics.AddError("Failed to wait for resource to be deleted.", err.Error())
			return
		}
	}
}

// ImportState imports the resource state.
func (r *ResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "|")

	if len(idParts) != 4 || len(idParts[0]) == 0 || len(idParts[1]) == 0 || len(idParts[2]) == 0 {
		resp.Diagnostics.AddError(
			"Unexpected import identifier.", fmt.Sprintf("expected import identifier with format: apiVersion|kind|name|namespace; got %q", req.ID),
		)
		return
	}

	var a map[string]any
	if len(idParts[3]) > 0 {
		a = map[string]any{
			"apiVersion": idParts[0],
			"kind":       idParts[1],
			"metadata": map[string]interface{}{
				"namespace": idParts[3],
				"name":      idParts[2],
			},
		}
	} else {
		a = map[string]any{
			"apiVersion": idParts[0],
			"kind":       idParts[1],
			"metadata": map[string]interface{}{
				"name": idParts[2],
			},
		}
	}

	obj, err := tfutils.DecodeDynamic(ctx, nil, a)
	if err != nil {
		resp.Diagnostics.AddError("Failed to decode import identifier.", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("manifest"), &obj)...)
}

// // updateObject updates the object to make sure unknown values.
// func updateObjectUnknowns(ctx context.Context, obj types.Dynamic, removePaths map[string]any) (types.Dynamic, error) {
// 	tfObj, err := obj.UnderlyingValue().ToTerraformValue(ctx)
// 	if err != nil {
// 		return types.Dynamic{}, fmt.Errorf("failed to convert object to Terraform value: %w", err)
// 	}

// 	updatedTfObj, err := tftypes.Transform(tfObj, func(ap *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
// 		if _, ok := removePaths[ap.String()]; ok {
// 			return tftypes.Value{}, nil
// 		}
// 		return v, nil
// 	})
// 	if err != nil {
// 		return types.Dynamic{}, fmt.Errorf("failed to transform Terraform object: %w", err)
// 	}

// 	updatedAttr, err := obj.Type(ctx).ValueFromTerraform(ctx, updatedTfObj)
// 	if err != nil {
// 		return types.Dynamic{}, fmt.Errorf("failed to convert Terraform value to object: %w", err)
// 	}

// 	return types.DynamicValue(updatedAttr), nil
// }

// getConditionsMap returns a map of conditions from a list of condition input objects.
func getConditionsMap(ctx context.Context, l types.List) (map[string]string, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	var arr []ConditionModel
	if diagnostics.Append(l.ElementsAs(ctx, &arr, false)...); diagnostics.HasError() {
		return nil, diagnostics
	}

	res := make(map[string]string, len(arr))
	for _, v := range arr {
		res[v.Type.ValueString()] = v.Status.ValueString()
	}

	return res, diagnostics
}

// getFieldSelectorStrings returns a list of field selector strings from a list of field selector input objects.
func getFieldSelectorStrings(ctx context.Context, l types.List) ([]string, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	var arr []FieldSelectorModel
	if diagnostics.Append(l.ElementsAs(ctx, &arr, false)...); diagnostics.HasError() {
		return nil, diagnostics
	}

	res := make([]string, len(arr))
	for i, v := range arr {
		if v.Operator.ValueString() != "eq" {
			res[i] = fmt.Sprintf("%s!=%s", v.Key.ValueString(), v.Value.ValueString())
		} else {
			res[i] = fmt.Sprintf("%s==%s", v.Key.ValueString(), v.Value.ValueString())
		}
	}

	return res, diagnostics
}
