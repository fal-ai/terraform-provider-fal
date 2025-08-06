package fal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fal-ai/terraform-provider-fal/internal/fal"
	"github.com/fal-ai/terraform-provider-fal/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	defaultBranch   = "main"
	defaultStrategy = "rolling"
	defaultAuthMode = "private"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &AppResource{}
	_ resource.ResourceWithImportState = &AppResource{}
)

func NewAppResource() resource.Resource {
	return &AppResource{}
}

// AppResource defines the resource implementation.
type AppResource struct {
	client *fal.Client
}

type SSH struct {
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	PrivateKey types.String `tfsdk:"private_key"`
}

type HTTP struct {
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	InsecureHTTPAllowed  types.Bool   `tfsdk:"allow_insecure_http"`
	CertificateAuthority types.String `tfsdk:"certificate_authority"`
}

type Git struct {
	URL    types.String `tfsdk:"url"`
	Branch types.String `tfsdk:"branch"`
	SSH    *SSH         `tfsdk:"ssh"`
	HTTP   *HTTP        `tfsdk:"http"`
}

// AppResourceModel describes the resource data model.
type AppResourceModel struct {
	RevisionID types.String `tfsdk:"revision_id"`

	Name       types.String `tfsdk:"name"`
	Entrypoint types.String `tfsdk:"entrypoint"`
	Strategy   types.String `tfsdk:"strategy"`
	AuthMode   types.String `tfsdk:"auth_mode"`
	Git        types.Object `tfsdk:"git"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (r *AppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *AppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "App resource",

		Attributes: map[string]schema.Attribute{
			"revision_id": schema.StringAttribute{
				MarkdownDescription: "The app's revision id",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The app's name",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entrypoint": schema.StringAttribute{
				MarkdownDescription: "The app deployment entrypoint",
				Required:            true,
			},
			"strategy": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The app deployment strategy. Defaults to `%s`.", defaultStrategy),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.OneOf("rolling", "recreate"),
				},
				Default: stringdefault.StaticString(defaultStrategy),
			},
			"auth_mode": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The app auth mode.  Defaults to `%s`.", defaultAuthMode),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.OneOf("public", "private"),
				},
				Default: stringdefault.StaticString(defaultAuthMode),
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp for when the app was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp for the last time the app was updated",
				Computed:            true,
			},
			"git": schema.SingleNestedAttribute{
				Description: "Configuration block with settings for fal's app deployer.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "URL of git repository to bootstrap from.",
						Required:    true,
						Validators: []validator.String{
							validators.URLScheme("http", "https", "ssh"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"branch": schema.StringAttribute{
						Description: fmt.Sprintf("Branch in repository to use with deployment. Defaults to `%s`.", defaultBranch),
						Optional:    true,
					},
					"ssh": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Description: "Username for Git SSH server.",
								Optional:    true,
							},
							"password": schema.StringAttribute{
								Description: "Password for private key.",
								Optional:    true,
								Sensitive:   true,
							},
							"private_key": schema.StringAttribute{
								Description: "Private key used for authenticating to the Git SSH server.",
								Optional:    true,
								Sensitive:   true,
							},
						},
						Optional: true,
					},
					"http": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Description: "Username for basic authentication.",
								Optional:    true,
							},
							"password": schema.StringAttribute{
								Description: "Password for basic authentication.",
								Optional:    true,
								Sensitive:   true,
							},
							"allow_insecure_http": schema.BoolAttribute{
								Description: "Allows HTTP Git URL connections.",
								Optional:    true,
							},
							"certificate_authority": schema.StringAttribute{
								Description: "Certificate authority to validate self-signed certificates.",
								Optional:    true,
							},
						},
						Optional: true,
					},
				},
				Required: true,
			},
		},
	}
}

func (r *AppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	_client, ok := req.ProviderData.(*fal.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *client.Client, got: %T. Please report this issue to the provider developers.",
		)

		return
	}
	r.client = _client
}

func (r *AppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.deployApp(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.readApp(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// This indicates the app is not found and we should remove it from our state
	if data.Name.IsNull() && data.RevisionID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AppResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.deployApp(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appName := data.Name.ValueString()

	err := r.client.Delete(ctx, appName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to delete app, got error: "+err.Error())
		return
	}
}

func (r *AppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data AppResourceModel
	data.Name = types.StringValue(req.ID)

	r.readApp(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.IsNull() && data.RevisionID.IsNull() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) readApp(ctx context.Context, data *AppResourceModel, diags *diag.Diagnostics) {
	name := data.Name.ValueString()

	apps, err := r.client.List(ctx)
	if err != nil {
		diags.AddError("Client Error", "Unable to read apps, got error: "+err.Error())
		return
	}

	var app *fal.App
	for _, a := range apps {
		if a.Alias != name {
			continue
		}
		app = a
		break
	}

	// Check if the app was deleted
	if app == nil {
		data.Name = types.StringNull()
		data.RevisionID = types.StringNull()
		return
	}

	if app.Alias != "" {
		data.Name = types.StringValue(app.Alias)
	}

	if app.Revision != "" {
		data.RevisionID = types.StringValue(app.Revision)
	}

	if app.AuthMode != "" {
		data.AuthMode = types.StringValue(strings.ToLower(app.AuthMode))
	}
}

func (r *AppResource) deployApp(ctx context.Context, data *AppResourceModel, diags *diag.Diagnostics) {
	gd := gitFromResourceModel(ctx, data)

	git, err := gd.Client()
	if err != nil {
		diags.AddError("Client Error", "Unable to get git client, got error: "+err.Error())
		return
	}

	repoURL := gd.RepositoryURL().String()
	res, err := r.client.Deploy(ctx, git, repoURL, &fal.DeployOpts{
		Entrypoint: data.Entrypoint.ValueString(),
		Strategy:   fal.DeployStrategy(data.Strategy.ValueString()),
		AuthMode:   fal.AuthMode(data.AuthMode.ValueString()),
	})
	if err != nil {
		diags.AddError("Client Error", "Unable to deploy app, got error: "+err.Error())
		return
	}
	data.Name = types.StringValue(res.FunctionName)
	data.RevisionID = types.StringValue(res.Revision)

	now := time.Now().Format(time.RFC3339)

	if data.CreatedAt.IsUnknown() || data.CreatedAt.IsNull() {
		data.CreatedAt = types.StringValue(now)
	}

	data.UpdatedAt = types.StringValue(now)
}
