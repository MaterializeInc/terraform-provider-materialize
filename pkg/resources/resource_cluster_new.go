package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// Define the resource schema and methods.
type clusterResource struct {
	client *utils.ProviderData
}

func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_2"
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

	client, ok := req.ProviderData.(*utils.ProviderData)

	// Verbously log the reg.ProviderData for debugging purposes.
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
	state.Region = types.StringValue(string(region))

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
			state.Disk = types.BoolValue(true)
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
	if !state.OwnershipRole.IsNull() && state.OwnershipRole.ValueString() != "" {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(state.OwnershipRole.ValueString()); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			resp.Diagnostics.AddError("Failed to set ownership", err.Error())
			return
		}
	}

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

	// Update the ID in the state
	state.ID = types.StringValue(clusterID)

	// After the cluster is successfully created, read its current state
	readState, _ := r.read(ctx, &state, false)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the state with the freshly read information
	diags = resp.State.Set(ctx, readState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClusterStateModelV0

	// Retrieve the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the lower-case read function to get the updated state
	updatedState, _ := r.read(ctx, &state, false)

	// Set the updated state in the response
	diags = resp.State.Set(ctx, updatedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClusterStateModelV0
	var state ClusterStateModelV0
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	metaDb, region, err := utils.NewGetDBClientFromMeta(r.client, state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get DB client", err.Error())
		return
	}
	state.Region = types.StringValue(string(region))

	o := materialize.MaterializeObject{ObjectType: "CLUSTER", Name: state.Name.ValueString()}
	b := materialize.NewClusterBuilder(metaDb, o)

	// Update cluster attributes if they have changed
	if state.OwnershipRole.ValueString() != plan.OwnershipRole.ValueString() && plan.OwnershipRole.ValueString() != "" {
		ownershipBuilder := materialize.NewOwnershipBuilder(metaDb, o)
		if err := ownershipBuilder.Alter(plan.OwnershipRole.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to update ownership role", err.Error())
			return
		}
	}

	if state.Size.ValueString() != plan.Size.ValueString() {
		if err := b.Resize(plan.Size.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to resize the cluster", err.Error())
			return
		}
	}

	// Handle changes in the 'disk' attribute
	if state.Disk.ValueBool() != plan.Disk.ValueBool() {
		if strings.HasSuffix(state.Size.ValueString(), "cc") || strings.HasSuffix(state.Size.ValueString(), "C") {
			// DISK option not supported for cluster sizes ending in cc or C.
			log.Printf("[WARN] disk option not supported for cluster size %s, disk is always enabled", state.Size.ValueString())
			state.Disk = types.BoolValue(true)
		} else {
			if err := b.SetDisk(plan.Disk.ValueBool()); err != nil {
				resp.Diagnostics.AddError("Failed to update disk setting", err.Error())
				return
			}
		}
	}

	// Handle changes in the 'replication_factor' attribute
	if state.ReplicationFactor.ValueInt64() != plan.ReplicationFactor.ValueInt64() {
		if err := b.SetReplicationFactor(int(plan.ReplicationFactor.ValueInt64())); err != nil {
			resp.Diagnostics.AddError("Failed to update replication factor", err.Error())
			return
		}
	}

	// Handle changes in the 'availability_zones' attribute
	if !state.AvailabilityZones.Equal(plan.AvailabilityZones) && len(plan.AvailabilityZones.Elements()) > 0 {
		azs := make([]string, len(plan.AvailabilityZones.Elements()))
		for i, elem := range plan.AvailabilityZones.Elements() {
			azs[i] = elem.(types.String).ValueString()
		}
		if err := b.SetAvailabilityZones(azs); err != nil {
			resp.Diagnostics.AddError("Failed to update availability zones", err.Error())
			return
		}
	}

	// Handle changes in the 'introspection_interval' attribute
	if state.IntrospectionInterval.ValueString() != plan.IntrospectionInterval.ValueString() {
		if err := b.SetIntrospectionInterval(plan.IntrospectionInterval.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to update introspection interval", err.Error())
			return
		}
	}

	// Handle changes in the 'introspection_debugging' attribute
	if state.IntrospectionDebugging.ValueBool() != plan.IntrospectionDebugging.ValueBool() {
		if err := b.SetIntrospectionDebugging(plan.IntrospectionDebugging.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Failed to update introspection debugging", err.Error())
			return
		}
	}

	// Handle changes in the 'idle_arrangement_merge_effort' attribute
	if state.IdleArrangementMergeEffort.ValueInt64() != plan.IdleArrangementMergeEffort.ValueInt64() {
		if err := b.SetIdleArrangementMergeEffort(int(plan.IdleArrangementMergeEffort.ValueInt64())); err != nil {
			resp.Diagnostics.AddError("Failed to update idle arrangement merge effort", err.Error())
			return
		}
	}

	// After updating the cluster, read its current state
	updatedState, _ := r.read(ctx, &plan, false)
	// Update the state with the freshly read information
	diags = resp.State.Set(ctx, updatedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve the current state
	var state ClusterStateModelV0
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	metaDb, _, err := utils.NewGetDBClientFromMeta(r.client, state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get DB client", err.Error())
		return
	}

	o := materialize.MaterializeObject{ObjectType: "CLUSTER", Name: state.Name.ValueString()}
	b := materialize.NewClusterBuilder(metaDb, o)

	// Drop the cluster
	if err := b.Drop(); err != nil {
		resp.Diagnostics.AddError("Failed to delete the cluster", err.Error())
		return
	}

	// After successful deletion, clear the state by setting ID to empty
	state.ID = types.String{}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) read(ctx context.Context, data *ClusterStateModelV0, dryRun bool) (*ClusterStateModelV0, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	metaDb, region, err := utils.NewGetDBClientFromMeta(r.client, data.Region.ValueString())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get DB client",
			Detail:   err.Error(),
		})
		return data, diags
	}

	clusterID := data.ID.ValueString()
	clusterDetails, err := materialize.ScanCluster(metaDb, utils.ExtractId(clusterID))
	if err != nil {
		if err == sql.ErrNoRows {
			data.ID = types.String{}
			data.Name = types.String{}
			data.Size = types.String{}
			data.ReplicationFactor = types.Int64{}
			data.Disk = types.Bool{}
			data.AvailabilityZones = types.List{}
			data.IntrospectionInterval = types.String{}
			data.IntrospectionDebugging = types.Bool{}
			data.IdleArrangementMergeEffort = types.Int64{}
			data.OwnershipRole = types.String{}
			data.Comment = types.String{}
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to read the cluster",
				Detail:   err.Error(),
			})
		}
		return data, diags
	}

	// Set the values from clusterDetails to data, checking for null values.
	data.ID = types.StringValue(clusterID)
	data.Name = types.StringValue(getNullString(clusterDetails.ClusterName))
	data.ReplicationFactor = types.Int64Value(clusterDetails.ReplicationFactor.Int64)
	data.Disk = types.BoolValue(clusterDetails.Disk.Bool)
	data.OwnershipRole = types.StringValue(getNullString(clusterDetails.OwnerName))

	// Handle the Size attribute
	if clusterDetails.Size.Valid && clusterDetails.Size.String != "" {
		data.Size = types.StringValue(clusterDetails.Size.String)
	} else {
		data.Size = types.StringNull()
	}

	// Handle the Comment attribute
	if clusterDetails.Comment.Valid && clusterDetails.Comment.String != "" {
		data.Comment = types.StringValue(clusterDetails.Comment.String)
	} else {
		data.Comment = types.StringNull()
	}

	regionStr := string(region)
	if regionStr != "" {
		data.Region = types.StringValue(regionStr)
	} else {
		data.Region = types.StringNull()
	}

	// Handle the AvailabilityZones which is a slice of strings.
	azValues := make([]attr.Value, len(clusterDetails.AvailabilityZones))
	for i, az := range clusterDetails.AvailabilityZones {
		azValues[i] = types.StringValue(az)
	}

	azList, _ := types.ListValue(types.StringType, azValues)

	data.AvailabilityZones = azList

	return data, diags
}

// getNullString checks if the sql.NullString is valid and returns the string or an empty string if not.
func getNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
