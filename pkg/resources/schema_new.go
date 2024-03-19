package resources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewObjectNameSchema(resource string, required, forceNew bool) schema.StringAttribute {
	attr := schema.StringAttribute{
		Description: fmt.Sprintf("The identifier for the %s.", resource),
		Required:    required,
		Optional:    !required,
	}
	if forceNew {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	}
	return attr
}

func NewCommentSchema(forceNew bool) schema.StringAttribute {
	attr := schema.StringAttribute{
		Description: "**Public Preview** Comment on an object in the database.",
		Optional:    true,
	}
	if forceNew {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	}
	return attr
}

func NewOwnershipRoleSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "The ownership role of the object.",
		Optional:    true,
		Computed:    true,
	}
}

func NewSizeSchema(resource string, required bool, forceNew bool, alsoRequires []string) schema.StringAttribute {
	expressions := make([]path.Expression, len(alsoRequires))
	for i, req := range alsoRequires {
		expressions[i] = path.MatchRoot(req)
	}

	attr := schema.StringAttribute{
		Description: fmt.Sprintf("The size of the %s.", resource),
		Required:    required,
		Optional:    !required,
		Validators: []validator.String{
			stringvalidator.OneOf(replicaSizes...),
			stringvalidator.AlsoRequires(expressions...),
		},
	}
	if forceNew {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	}
	return attr
}

func NewDiskSchema(forceNew bool) schema.BoolAttribute {
	attr := schema.BoolAttribute{
		Description: "**Deprecated**. This attribute is maintained for backward compatibility with existing configurations. New users should use 'cc' sizes for disk access. Disk replicas are deprecated and will be removed in a future release. The `disk` attribute will be enabled by default for 'cc' clusters",
		Optional:    true,
		Computed:    true,
	}
	if forceNew {
		attr.PlanModifiers = []planmodifier.Bool{boolplanmodifier.RequiresReplace()}
	}
	return attr
}

func NewIntrospectionIntervalSchema(forceNew bool, alsoRequires []string) schema.StringAttribute {
	expressions := make([]path.Expression, len(alsoRequires))
	for i, req := range alsoRequires {
		expressions[i] = path.MatchRoot(req)
	}

	attr := schema.StringAttribute{
		Description: "The interval at which to collect introspection data.",
		Optional:    true,
		Computed:    true,
		Default:     stringdefault.StaticString("1m"),
		Validators: []validator.String{
			stringvalidator.AlsoRequires(expressions...),
		},
	}
	if forceNew {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	}
	return attr
}

func NewIntrospectionDebuggingSchema(forceNew bool, alsoRequires []string) schema.BoolAttribute {
	expressions := make([]path.Expression, len(alsoRequires))
	for i, req := range alsoRequires {
		expressions[i] = path.MatchRoot(req)
	}

	attr := schema.BoolAttribute{
		Description: "Whether to introspect the gathering of the introspection data.",
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(false),
		Validators: []validator.Bool{
			boolvalidator.AlsoRequires(expressions...),
		},
	}
	if forceNew {
		attr.PlanModifiers = []planmodifier.Bool{boolplanmodifier.RequiresReplace()}
	}
	return attr
}

func NewIdleArrangementMergeEffortSchema(forceNew bool, alsoRequires []string) schema.Int64Attribute {
	expressions := make([]path.Expression, len(alsoRequires))
	for i, req := range alsoRequires {
		expressions[i] = path.MatchRoot(req)
	}

	attr := schema.Int64Attribute{
		Description: "The amount of effort to exert compacting arrangements during idle periods. This is an unstable option! It may be changed or removed at any time.",
		Optional:    true,
		Validators: []validator.Int64{
			int64validator.AlsoRequires(expressions...),
		},
	}
	if forceNew {
		attr.PlanModifiers = []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	}
	return attr
}

func NewRegionSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "The region to use for the resource connection. If not set, the default region is used.",
		Optional:    true,
		Computed:    true,
		// PlanModifiers: []planmodifier.String{
		// 	stringplanmodifier.RequiresReplace(),
		// },
	}
}

func NewReplicationFactorSchema() schema.Int64Attribute {
	return schema.Int64Attribute{
		Description: "The number of replicas of each dataflow-powered object to maintain.",
		Optional:    true,
		Computed:    true,
		Validators: []validator.Int64{
			int64validator.AlsoRequires(path.MatchRoot("size")),
		},
	}
}

func NewAvailabilityZonesSchema() schema.ListAttribute {
	return schema.ListAttribute{
		Description: "The specific availability zones of the cluster.",
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Validators: []validator.List{
			listvalidator.AlsoRequires(path.MatchRoot("size")),
		},
	}
}
