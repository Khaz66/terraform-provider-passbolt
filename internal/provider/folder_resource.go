package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/api"
	"terraform-provider-passbolt/tools"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &folderResource{}
	_ resource.ResourceWithConfigure = &folderResource{}
)

// NewFolderResource is a helper function to simplify the provider implementation.
func NewFolderResource() resource.Resource {
	return &folderResource{}
}

// folderResource is the resource implementation.
type folderResource struct {
	client *tools.PassboltClient
}

// created, modified, created_by, modified_by, and folder_parent_id
type foldersModelCreate struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	FolderParentId types.String `tfsdk:"folder_parent_id"`
}

// Configure adds the provider configured client to the resource.
func (r *folderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tools.PassboltClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *passboltClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *folderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

// Schema defines the schema for the resource.
func (r *folderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"folder_parent_id": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Create a new resource.
func (r *folderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan foldersModelCreate
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/*
	parent, err := r.client.Client.GetFolder(r.client.Context, plan.FolderParentId.ValueString(),nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read folder", "",
		)
		return
	}

	folders, err := r.client.Client.GetFolders(ctx, nil)
	if err != nil {
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
*/
	// Generate API request body from plan
	var folder = api.Folder{
		FolderParentID: plan.FolderParentId.ValueString(),
		Name:           plan.Name.ValueString(),
	}

	// Create new order
	cFolder, errCreate := r.client.Client.CreateFolder(r.client.Context, folder)
	if errCreate != nil {
		resp.Diagnostics.AddError(
			"Error creating folder",
			"Could not create folder, unexpected error: "+errCreate.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(cFolder.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *folderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan foldersModelCreate
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folder, err := r.client.Client.GetFolder(r.client.Context, plan.ID.ValueString(),nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read folder", "",
		)
		return
	}

	folderState := foldersModelCreate{
		ID:             types.StringValue(folder.ID),
		Name:           types.StringValue(folder.Name),
		FolderParentId: types.StringValue(folder.FolderParentID),
	}



	// Set state
	diag := resp.State.Set(ctx, folderState)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *folderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan foldersModelCreate
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state foldersModelCreate
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.FolderParentId != plan.FolderParentId {
		errMove := r.client.Client.MoveFolder(r.client.Context, state.ID.ValueString(), plan.FolderParentId.ValueString())
		if errMove != nil {
			resp.Diagnostics.AddError(
				"Unable to move folder ", "",
			)
			return
		}
	}

	// Generate API request body from plan
	var folder = api.Folder{
		FolderParentID: plan.FolderParentId.ValueString(),
		Name:           plan.Name.ValueString(),
	}

	// Create new order
	cFolder, errUpdate := r.client.Client.UpdateFolder(r.client.Context, state.ID.ValueString(), folder)
	if errUpdate != nil {
		resp.Diagnostics.AddError(
			"Error updating folder",
						"Could not update folder, unexpected error: "+errUpdate.Error(),
		)
		return
	}

	folderState := foldersModelCreate{
		ID:             types.StringValue(cFolder.ID),
		Name:           types.StringValue(cFolder.Name),
		FolderParentId: types.StringValue(cFolder.FolderParentID),
	}

	// Map response body to schema and populate Computed attribute values
	//plan.ID = types.StringValue(cFolder.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, folderState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *folderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state foldersModelCreate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Client.DeleteFolder(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Folder",
			"Could not delete Folder, unexpected error: "+err.Error(),
		)
		return
}
}
