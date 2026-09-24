package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func OrganizationRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: organizationRoleCreate,
		ReadContext:   organizationRoleRead,
		UpdateContext: organizationRoleUpdate,
		DeleteContext: organizationRoleDelete,
		Importer:      &schema.ResourceImporter{StateContext: organizationRoleImport},
		Description:   "Manages a custom organization role in Materialize Cloud. Organization roles define a user's account-level permissions and belong to the account associated with the provider credentials. When assigned to a user, the role's name is included in the user's JWT roles claim and automatically mapped to an existing database role with the same name. Create the database role separately. Materialize creates two reserved, built-in roles: Organization Admin (key MaterializePlatformAdmin) and Organization Member (key MaterializePlatform). You cannot edit or delete them with this resource, but you can assign them to SCIM groups with materialize_scim_group_roles using Admin or Member. Requires organization role management to be enabled and an Organization Admin app password.",
		Schema: map[string]*schema.Schema{
			"name":           {Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validateOrganizationRoleName, Description: "Name of the custom role. When assigned to a user, this name maps to an existing database role with the same name."},
			"key":            {Type: schema.TypeString, Computed: true, Description: "Identifier included in the JWT roles claim. For roles created by this resource, it matches the role name."},
			"description":    {Type: schema.TypeString, Optional: true, Description: "Description of the organization role."},
			"base_role_name": {Type: schema.TypeString, Optional: true, Default: "Member", ForceNew: true, ValidateFunc: validation.StringIsNotEmpty, Description: "Existing organization role used to determine the new role's level. Defaults to Member. When permission_ids is omitted, copies this role's permissions at creation. Later changes to the base role are not propagated. Imports assume Member because Frontegg does not return the role used at creation; configuring another base role after import replaces the role."},
			"permission_ids": {Type: schema.TypeSet, Optional: true, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Organization permission IDs assigned to the role. If omitted on creation, copies the base role's permissions. These permissions do not grant privileges on database objects. Use an explicit empty set for no organization permissions."},
		},
	}
}

func organizationRoleImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	// Frontegg does not return the base role used at creation. Seed the
	// default so an imported role does not immediately plan a replacement.
	if err := d.Set("base_role_name", "Member"); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func validateOrganizationRoleName(value interface{}, key string) ([]string, []error) {
	name := value.(string)
	lower := strings.ToLower(name)
	switch lower {
	case "", "admin", "member", "organization admin", "organization member", "materializeplatformadmin", "materializeplatform", "public", "current_user", "current_role", "session_user", "user", "none":
		return nil, []error{fmt.Errorf("%s must not use a reserved role name", key)}
	}
	for _, prefix := range []string{"mz_", "pg_", "external_"} {
		if strings.HasPrefix(lower, prefix) {
			return nil, []error{fmt.Errorf("%s must not start with %s", key, prefix)}
		}
	}
	return nil, nil
}

func organizationRoleMeta(meta interface{}) (*utils.ProviderMeta, diag.Diagnostics) {
	p, err := utils.GetProviderMeta(meta)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if diags := p.ValidateSaaSOnly("materialize_organization_role"); diags.HasError() {
		return nil, diags
	}
	return p, nil
}

func organizationRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diags := organizationRoleMeta(meta)
	if diags.HasError() {
		return diags
	}
	roles, err := frontegg.FetchFronteggRoles(ctx, p.Frontegg)
	if err != nil {
		return diag.FromErr(err)
	}
	var base *frontegg.FronteggRole
	for i := range roles {
		if frontegg.RoleName(roles[i].Name) == d.Get("base_role_name").(string) {
			if base != nil {
				return diag.Errorf("base organization role name is ambiguous")
			}
			base = &roles[i]
		}
	}
	if base == nil {
		return diag.Errorf("base organization role %q not found", d.Get("base_role_name"))
	}
	permissions := base.Permissions
	// Raw config distinguishes an explicitly empty set from an omitted
	// Optional+Computed attribute. GetOk treats both as the zero value.
	configured, configDiags := d.GetRawConfigAt(cty.GetAttrPath("permission_ids"))
	if configDiags.HasError() {
		return configDiags
	}
	if !configured.IsNull() {
		permissions = expandStringSet(d.Get("permission_ids").(*schema.Set))
	}
	if permissions == nil {
		permissions = []string{}
	}
	name := d.Get("name").(string)
	role, err := frontegg.CreateOrganizationRole(ctx, p.Frontegg, frontegg.OrganizationRoleCreateParams{
		Name: name, Key: name, Description: d.Get("description").(string), BaseRoleID: base.ID, PermissionIDs: permissions,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(role.ID)
	p.InvalidateFronteggRoles()
	return organizationRoleRead(ctx, d, meta)
}

func organizationRoleFetch(ctx context.Context, p *utils.ProviderMeta, id string) (*frontegg.FronteggRole, error) {
	roles, err := frontegg.FetchFronteggRoles(ctx, p.Frontegg)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if roles[i].ID == id {
			if roles[i].TenantID == "" {
				return nil, fmt.Errorf("role %s is not tenant-scoped; built-in roles cannot be managed by this resource", id)
			}
			return &roles[i], nil
		}
	}
	return nil, nil
}

func organizationRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diags := organizationRoleMeta(meta)
	if diags.HasError() {
		return diags
	}
	role, err := organizationRoleFetch(ctx, p, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if role == nil {
		d.SetId("")
		return nil
	}
	for key, value := range map[string]interface{}{"name": role.Name, "key": role.Key, "description": role.Description, "permission_ids": role.Permissions} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func organizationRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diags := organizationRoleMeta(meta)
	if diags.HasError() {
		return diags
	}
	role, err := organizationRoleFetch(ctx, p, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if role == nil {
		d.SetId("")
		return nil
	}
	if d.HasChange("description") {
		if err := frontegg.UpdateOrganizationRole(ctx, p.Frontegg, d.Id(), d.Get("description").(string)); err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("permission_ids") {
		permissions := expandStringSet(d.Get("permission_ids").(*schema.Set))
		if permissions == nil {
			permissions = []string{}
		}
		if err := frontegg.SetOrganizationRolePermissions(ctx, p.Frontegg, d.Id(), permissions); err != nil {
			return diag.FromErr(err)
		}
	}
	p.InvalidateFronteggRoles()
	return organizationRoleRead(ctx, d, meta)
}

func organizationRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diags := organizationRoleMeta(meta)
	if diags.HasError() {
		return diags
	}
	role, err := organizationRoleFetch(ctx, p, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if role != nil {
		err = frontegg.DeleteOrganizationRole(ctx, p.Frontegg, d.Id())
		if err != nil && !clients.IsNotFoundError(err) {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	p.InvalidateFronteggRoles()
	return nil
}
