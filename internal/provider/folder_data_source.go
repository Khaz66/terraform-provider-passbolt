package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-passbolt/tools"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &folderDataSource{}
	_ datasource.DataSourceWithConfigure = &folderDataSource{}
)

// NewFolderDataSource is a helper function to simplify the provider implementation.
func NewFolderDataSource() datasource.DataSource {
	return &folderDataSource{}
}

// coffeesDataSource is the data source implementation.
type folderDataSource struct {
	client *tools.PassboltClient
}

type folderModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	FolderParentId types.String `tfsdk:"folder_parent_id"`
}

// Configure adds the provider configured client to the data source.
func (d *folderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *folderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

// Schema defines the schema for the data source.
func (d *folderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},			
			"folder_parent_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *folderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var conf folderModel
	diags := req.Config.Get(ctx, &conf)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	folder, err := d.client.Client.GetFolder(d.client.Context, conf.ID.ValueString(),nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read folder", "",
		)
		return
	}

	folderState := folderModel{
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
