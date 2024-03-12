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
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Username       types.String `tfsdk:"username"`
	Uri            types.String `tfsdk:"uri"`
	ResourceTypeId types.String `tfsdk:"resource_type_id"`
	FolderParent   types.String `tfsdk:"folder_parent"`
	Password       types.String `tfsdk:"password"`
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
				Required: true,
			},
			"resource_type_id": schema.StringAttribute{
				Computed: true,
			},
			"folder_parent": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
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
			plan.ResourceTypeId = types.StringValue(resourceType.ID)
		}
	}

	var pubKey, _, errKey = r.client.Client.GetPublicKey(ctx)
	if errKey != nil {
		resp.Diagnostics.AddError("Cannot get public key", "")
		return
	}

	var enc, errEnc = r.client.Client.EncryptMessageWithPublicKey(pubKey, plan.Password.ValueString())

	if errEnc != nil {
		resp.Diagnostics.AddError("Cannot encrypt message", "")
		return
	}

	var secrets = []api.Secret{
		{
			Data: enc,
		},
	}

	folders, errFolder := r.client.Client.GetFolders(ctx, nil)
	if errFolder != nil {
		resp.Diagnostics.AddError("Cannot get folders", "")
		return
	}

	var folderId string
	if !plan.FolderParent.IsUnknown() && !plan.FolderParent.IsNull() {
		for _, folder := range folders {
			if folder.Name == plan.FolderParent.ValueString() {
				folderId = folder.ID
			}
		}
	}

	var password = api.Resource{
		FolderParentID: folderId,
		Name:           plan.Name.ValueString(),
		Username:       plan.Username.ValueString(),
		URI:            plan.Uri.ValueString(),
		ResourceTypeID: plan.ResourceTypeId.ValueString(),
		Secrets:        secrets,
	}

	cPassword, err := r.client.Client.CreateResource(ctx, password)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating password",
			"Could not create password, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(cPassword.ID)
	plan.ResourceTypeId = types.StringValue(cPassword.ResourceTypeID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *passwordResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *passwordResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
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
