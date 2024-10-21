package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/passbolt/go-passbolt/helper"
	"terraform-provider-passbolt/tools"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &passwordDataSource{}
	_ datasource.DataSourceWithConfigure = &passwordDataSource{}
)

// NewpasswordDataSource is a helper function to simplify the provider implementation.
func NewPasswordDataSource() datasource.DataSource {
	return &passwordDataSource{}
}

// coffeesDataSource is the data source implementation.
type passwordDataSource struct {
	client *tools.PassboltClient
}


// Configure adds the provider configured client to the data source.
func (d *passwordDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

// Metadata returns the data source type name.
func (d *passwordDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password"
}

// Schema defines the schema for the data source.
func (d *passwordDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"uri": schema.StringAttribute{
				Optional: true,
			},
			"folder_parent_id": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *passwordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var conf passwordModel
	diags := req.Config.Get(ctx, &conf)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	folderParentID, name, username, uri, password, description , err  := helper.GetResource(ctx, d.client.Client, conf.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read password ", "",
		)
		return
	}

	passwordState := passwordModel{
		ID:             conf.ID,
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
