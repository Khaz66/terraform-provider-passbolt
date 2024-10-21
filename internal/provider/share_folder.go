package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/helper"
	"terraform-provider-passbolt/tools"
	"github.com/hashicorp/go-uuid"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &shareFolder{}
	_ resource.ResourceWithConfigure = &shareFolder{}
)

// NewshareFolder is a helper function to simplify the provider implementation.
func NewShareFolder() resource.Resource {
	return &shareFolder{}
}

// folderResource is the resource implementation.
type shareFolder struct {
	client *tools.PassboltClient
}

type shareFolderModel struct {
	ID           types.String `tfsdk:"id"`
	FolderId types.String `tfsdk:"folder_id"`
	ShareGroupId   types.String `tfsdk:"share_group_id"`
	Modify			types.Bool   `tfsdk:"modify"`
	

}

// Configure adds the provider configured client to the resource.
func (r *shareFolder) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tools.PassboltClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *shareFolder) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_share_folder"
}

// Schema defines the schema for the resource.
func (r *shareFolder) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"folder_id": schema.StringAttribute{
				Required: true,
			},
			"share_group_id": schema.StringAttribute{
				Required: true,
			},
			"modify": schema.BoolAttribute{
				Required: true,
			},
		},
	}
}

// Create a new resource.
func (r *shareFolder) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan shareFolderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	typeperm := TypePerm(plan.Modify.ValueBool())
	
		if plan.ShareGroupId.ValueString() != "" {
			var shares = []helper.ShareOperation{
				{
					Type:  typeperm,
					ARO:   "Group",
					AROID: plan.ShareGroupId.ValueString(),
				},
			}

			shareErr := helper.ShareFolder(ctx, r.client.Client, plan.FolderId.ValueString(), shares)

			if shareErr != nil {
				resp.Diagnostics.AddError("Cannot share resource", "")
				return
			}
		}
	genId, errId := uuid.GenerateUUID()
	if errId != nil {
		resp.Diagnostics.AddError("Cannot generate uuid", "")
		return
	}
	plan.ID = types.StringValue(genId)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *shareFolder) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *shareFolder) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state shareFolderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan shareFolderModel
	diagp := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diagp...)
	if resp.Diagnostics.HasError() {
		return
	}
	

	//Deletes the sharing of the resource
	if state.ShareGroupId.ValueString() != "" {
		var shares = []helper.ShareOperation{
			{
				Type:  -1,
				ARO:   "Group",
				AROID: state.ShareGroupId.ValueString(),
			},
		}

		shareErr := helper.ShareFolder(ctx, r.client.Client, state.FolderId.ValueString(), shares)

		if shareErr != nil {
			resp.Diagnostics.AddError("Cannot delete share of resource", "")
			return
		}
	}

	//Creates the sharing of the resource
	typeperm := TypePerm(plan.Modify.ValueBool())
	if plan.ShareGroupId.ValueString() != "" {
		var shares = []helper.ShareOperation{
			{
				Type:  typeperm,
				ARO:   "Group",
				AROID: plan.ShareGroupId.ValueString(),
			},
		}

		shareErr := helper.ShareFolder(ctx, r.client.Client, plan.FolderId.ValueString(), shares)

		if shareErr != nil {
			resp.Diagnostics.AddError("Cannot share resource", "")
			return
		}
	}
plan.ID = state.ID

// Set state to fully populated data
diagpl := resp.State.Set(ctx, plan)
resp.Diagnostics.Append(diagpl...)
if resp.Diagnostics.HasError() {
	return
}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *shareFolder) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state shareFolderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	
		if state.ShareGroupId.ValueString() != "" {
			var shares = []helper.ShareOperation{
				{
					Type:  -1,
					ARO:   "Group",
					AROID: state.ShareGroupId.ValueString(),
				},
			}

			shareErr := helper.ShareFolder(ctx, r.client.Client, state.FolderId.ValueString(), shares)

			if shareErr != nil {
				resp.Diagnostics.AddError("Cannot share resource", "")
				return
			}
		}
}
