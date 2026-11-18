package provider

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kybornos/terraform-provider-kybornos/internal/client"
)

var _ resource.Resource = (*VMResource)(nil)
var _ resource.ResourceWithConfigure = (*VMResource)(nil)
var _ resource.ResourceWithImportState = (*VMResource)(nil)

type VMResource struct{ c *client.Client }
type vmModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Size types.String `tfsdk:"size"`
}

func NewVMResource() resource.Resource { return &VMResource{} }
func (r *VMResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "kybornos_vm"
}
func (r *VMResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "A Kybornos virtual machine. Name/size changes replace the VM.", Attributes: map[string]schema.Attribute{
		"id":   schema.StringAttribute{Computed: true, Description: "Stable VM identifier."},
		"name": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"size": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
	}}
}
func (r *VMResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Invalid provider data", "Expected *client.Client")
		return
	}
	r.c = c
}
func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m vmModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	vm, err := r.c.Create(ctx, m.Name.ValueString(), m.Size.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create VM failed", err.Error())
		return
	}
	m.ID = types.StringValue(vm.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m vmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	vm, err := r.c.Get(ctx, m.ID.ValueString())
	if err != nil {
		var apiErr *client.Error
		if errors.As(err, &apiErr) && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VM failed", err.Error())
		return
	}
	if vm.ID == "" {
		resp.Diagnostics.AddError("Invalid VM response", "Missing VM id")
		return
	}
	m.ID = types.StringValue(vm.ID)
	m.Name = types.StringValue(vm.Name)
	m.Size = types.StringValue(vm.Size)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *VMResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported VM update", "Changes to name or size require replacement.")
}
func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m vmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.c.Delete(ctx, m.ID.ValueString())
	if err != nil {
		var apiErr *client.Error
		if errors.As(err, &apiErr) && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("Delete VM failed", err.Error())
	}
}
func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
