package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/helper"
	"terraform-provider-passbolt/tools"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &passwordResource{}
	_ resource.ResourceWithConfigure = &passwordResource{}
)

// NewPasswordResource is a helper function to simplify the provider implementation.
func NewPasswordResource() resource.Resource {
	return &passwordResource{}
}

// folderResource is the resource implementation.
type passwordResource struct {
	client *tools.PassboltClient
}

type passwordModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Username     types.String `tfsdk:"username"`
	Uri          types.String `tfsdk:"uri"`
	FolderParentId types.String `tfsdk:"folder_parent_id"`
	Password     types.String `tfsdk:"password"`
	Description     types.String `tfsdk:"description"`
}

// Configure adds the provider configured client to the resource.
func (r *passwordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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
func (r *passwordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

// Schema defines the schema for the resource.
func (r *passwordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"username": schema.StringAttribute{
				Required: true,
			},
			"uri": schema.StringAttribute{
				Optional: true,
			},
			"folder_parent_id": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Create a new resource.
func (r *passwordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan passwordModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceTypes, err := r.client.Client.GetResourceTypes(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Cannot get resource types", "")
		return
	}

	for _, resourceType := range resourceTypes {
		if resourceType.Slug == "password-and-description" {
			//		plan.ResourceTypeId = types.StringValue(resourceType.ID)
		}
	}

	folders, errFolder := r.client.Client.GetFolders(ctx, nil)
	if errFolder != nil {
		resp.Diagnostics.AddError("Cannot get folders", "")
		return
	}

	var folderId string
	if !plan.FolderParentId.IsUnknown() && !plan.FolderParentId.IsNull() {
		for _, folder := range folders {
			if folder.ID == plan.FolderParentId.ValueString() {
				folderId = folder.ID
			}
		}
	}

	resourceId, err := helper.CreateResource(ctx, r.client.Client, folderId, plan.Name.ValueString(), plan.Username.ValueString(), plan.Uri.ValueString(), plan.Password.ValueString(), plan.Description.ValueString())
/*
	var groupId string
	if !plan.ShareGroup.IsUnknown() && !plan.FolderParentId.IsNull() {
		groups, _ := r.client.Client.GetGroups(ctx, nil)

		for _, group := range groups {
			if group.Name == plan.ShareGroup.ValueString() {
				groupId = group.ID
			}
		}

		if groupId != "" {
			var shares = []helper.ShareOperation{
				{
					Type:  7,
					ARO:   "Group",
					AROID: groupId,
				},
			}

			shareErr := helper.ShareResource(ctx, r.client.Client, resourceId, shares)

			if shareErr != nil {
				resp.Diagnostics.AddError("Cannot share resource", "")
			}
		}
	}
*/
	plan.ID = types.StringValue(resourceId)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *passwordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan passwordModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	folderParentID, name, username, uri, password, description , err  := helper.GetResource(ctx, r.client.Client, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read password ", "",
		)
		return
	}

	passwordState := passwordModel{
		ID:             plan.ID,
		Name:           types.StringValue(name),
		Username: 		types.StringValue(username),
		FolderParentId: types.StringValue(folderParentID),
		Uri:          types.StringValue(uri),
		Description:     types.StringValue(description),
		Password:		types.StringValue(password),
	}



	// Set state
	diag := resp.State.Set(ctx, passwordState)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *passwordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan passwordModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state passwordModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	errUpd  := helper.UpdateResource(ctx, r.client.Client, state.ID.ValueString(), plan.Name.ValueString(), plan.Username.ValueString(), plan.Uri.ValueString(), plan.Password.ValueString(), plan.Description.ValueString())
	if errUpd != nil {
		resp.Diagnostics.AddError(
			"Unable to update password ", "",
		)
		return
	}

	if state.FolderParentId != plan.FolderParentId {
		errMove := helper.MoveResource(ctx, r.client.Client, state.ID.ValueString(),plan.FolderParentId.ValueString())
		if errMove != nil {
			resp.Diagnostics.AddError(
				"Unable to move password ", "",
			)
			return
		}
	}

	passwordState := passwordModel{
		ID:             state.ID,
		Name:           plan.Name,
		Username: 		plan.Username,
		FolderParentId: plan.FolderParentId,
		Uri:          plan.Uri,
		Description:     plan.Description,
		Password:		plan.Password,
	}

	// Set state
	diag := resp.State.Set(ctx, passwordState)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
	
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *passwordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state passwordModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	err := r.client.Client.DeleteResource(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting password",
			"Could not delete password, unexpected error: "+err.Error(),
		)
		return
	}
}
