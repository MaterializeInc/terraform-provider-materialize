package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Define the resource schema and methods.
type clusterResource struct {
	client *utils.ProviderData
}

func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "materialize_cluster_2"
	// resp.TypeName = req.ProviderTypeName + "_cluster_2"
}

type ClusterStateModelV0 struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Size                       types.String `tfsdk:"size"`
	ReplicationFactor          types.Int64  `tfsdk:"replication_factor"`
	Disk                       types.Bool   `tfsdk:"disk"`
	AvailabilityZones          types.List   `tfsdk:"availability_zones"`
	IntrospectionInterval      types.String `tfsdk:"introspection_interval"`
	IntrospectionDebugging     types.Bool   `tfsdk:"introspection_debugging"`
	IdleArrangementMergeEffort types.Int64  `tfsdk:"idle_arrangement_merge_effort"`
	OwnershipRole              types.String `tfsdk:"ownership_role"`
	Comment                    types.String `tfsdk:"comment"`
	Region                     types.String `tfsdk:"region"`
}

func ClusterSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The Cluster ID",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name":                          NewObjectNameSchema("cluster", true, true),
		"comment":                       NewCommentSchema(false),
		"ownership_role":                NewOwnershipRoleSchema(),
		"size":                          NewSizeSchema("managed cluster", false, false, []string{"replication_factor", "availability_zones"}),
		"replication_factor":            NewReplicationFactorSchema(),
		"disk":                          NewDiskSchema(false),
		"availability_zones":            NewAvailabilityZonesSchema(),
		"introspection_interval":        NewIntrospectionIntervalSchema(false, []string{"size"}),
		"introspection_debugging":       NewIntrospectionDebuggingSchema(false, []string{"size"}),
		"idle_arrangement_merge_effort": NewIdleArrangementMergeEffortSchema(false, []string{"size"}),
		"region":                        NewRegionSchema(),
	}
}

func (r *clusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: ClusterSchema(),
	}
}

func (r *clusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// client, ok := req.ProviderData.(*provider.ProviderData)
	client, ok := req.ProviderData.(*utils.ProviderData)

	// Verbously log the reg.ProviderData
	log.Printf("[DEBUG] ProviderData contents: %+v\n", fmt.Sprintf("%+v", req.ProviderData))

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *utils.ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Implement Create method to store the cluster name in the state.
func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Initialize and retrieve values from the request's plan.
	var state ClusterStateModelV0
	diags := req.Plan.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	metaDb, region, err := utils.NewGetDBClientFromMeta(r.client, state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get DB client", err.Error())
		return
	}

	o := materialize.MaterializeObject{ObjectType: "CLUSTER", Name: state.Name.ValueString()}
	b := materialize.NewClusterBuilder(metaDb, o)

	// Managed cluster options.
	if !state.Size.IsNull() {
		size := state.Size.ValueString()

		b.Size(size)

		if !state.ReplicationFactor.IsNull() {
			r := int(state.ReplicationFactor.ValueInt64())
			b.ReplicationFactor(&r)
		}

		if strings.HasSuffix(size, "cc") || strings.HasSuffix(size, "C") {
			// DISK option not supported for cluster sizes ending in cc or C.
			log.Printf("[WARN] disk option not supported for cluster size %s, disk is always enabled", size)
			b.Disk(true)
		} else if !state.Disk.IsNull() {
			b.Disk(state.Disk.ValueBool())
		}

		if !state.AvailabilityZones.IsNull() && len(state.AvailabilityZones.Elements()) > 0 {
			f := make([]string, len(state.AvailabilityZones.Elements()))
			for i, elem := range state.AvailabilityZones.Elements() {
				f[i] = elem.(types.String).ValueString()
			}
			b.AvailabilityZones(f)
		}

		if !state.IntrospectionInterval.IsNull() {
			b.IntrospectionInterval(state.IntrospectionInterval.ValueString())
		}

		if !state.IntrospectionDebugging.IsNull() && state.IntrospectionDebugging.ValueBool() {
			b.IntrospectionDebugging()
		}

		if !state.IdleArrangementMergeEffort.IsNull() {
			b.IdleArrangementMergeEffort(int(state.IdleArrangementMergeEffort.ValueInt64()))
		}
	}

	// Create the resource.
	if err := b.Create(); err != nil {
		resp.Diagnostics.AddError("Failed to create the cluster", err.Error())
		return
	}

	// Ownership.
	// TODO: Fix failing error
	// if !state.OwnershipRole.IsNull() {
	// 	ownership := materialize.NewOwnershipBuilder(metaDb, o)

	// 	if err := ownership.Alter(state.OwnershipRole.ValueString()); err != nil {
	// 		log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
	// 		b.Drop()
	// 		resp.Diagnostics.AddError("Failed to set ownership", err.Error())
	// 		return
	// 	}
	// }

	// Object comment.
	if !state.Comment.IsNull() {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(state.Comment.ValueString()); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			resp.Diagnostics.AddError("Failed to add comment", err.Error())
			return
		}
	}

	// Set ID.
	i, err := materialize.ClusterId(metaDb, o)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set resource ID", err.Error())
		return
	}

	// After all operations are successful and you have the cluster ID:
	clusterID := utils.TransformIdWithRegion(string(region), i)

	// Update the ID in the state and set the entire state in the response
	state.ID = types.StringValue(clusterID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Implementation for Read operation
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Implementation for Update operation
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Implementation for Delete operation
}
