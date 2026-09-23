package resources

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestOrganizationRoleLifecycle(t *testing.T) {
	for _, explicitEmpty := range []bool{false, true} {
		t.Run(map[bool]string{false: "copy member permissions", true: "explicit empty permissions"}[explicitEmpty], func(t *testing.T) {
			ctx := context.Background()
			role := frontegg.FronteggRole{}
			var created frontegg.OrganizationRoleCreateParams
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method + " " + r.URL.Path {
				case "GET /identity/resources/roles/v2":
					roles := []frontegg.FronteggRole{{ID: "member", Name: "Organization Member", Permissions: []string{"read-environments"}}}
					if role.ID != "" {
						roles = append(roles, role)
					}
					json.NewEncoder(w).Encode(frontegg.FronteggRolesResponse{Items: roles})
				case "POST /identity/resources/roles/v2":
					require.NoError(t, json.NewDecoder(r.Body).Decode(&created))
					role = frontegg.FronteggRole{ID: "custom", Name: created.Name, Key: created.Key, Description: created.Description, TenantID: "tenant", Permissions: created.PermissionIDs}
					json.NewEncoder(w).Encode(role)
				case "PATCH /identity/resources/roles/v1/custom":
					var v map[string]string
					require.NoError(t, json.NewDecoder(r.Body).Decode(&v))
					role.Description = v["description"]
					w.WriteHeader(http.StatusNoContent)
				case "PUT /identity/resources/roles/v1/custom/permissions":
					var v struct {
						IDs []string `json:"permissionIds"`
					}
					require.NoError(t, json.NewDecoder(r.Body).Decode(&v))
					role.Permissions = v.IDs
					w.WriteHeader(http.StatusNoContent)
				case "DELETE /identity/resources/roles/v1/custom":
					role = frontegg.FronteggRole{}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					http.Error(w, "unexpected", 500)
				}
			}))
			defer server.Close()
			client := &clients.FronteggClient{Endpoint: server.URL, HTTPClient: server.Client()}
			meta := &utils.ProviderMeta{Frontegg: client, FronteggRolesFetcher: func(ctx context.Context) (map[string][]string, error) { return frontegg.ListFronteggRoles(ctx, client) }}
			// An earlier resource can load roles before this resource is created.
			_, err := meta.GetFronteggRoles(ctx)
			require.NoError(t, err)
			config := map[string]interface{}{"name": "analytics_reader"}
			if explicitEmpty {
				config["permission_ids"] = []interface{}{}
			}
			d := schema.TestResourceDataRaw(t, OrganizationRole().Schema, config)
			d.SetId("planning")
			state := d.State()
			configured := cty.NullVal(cty.Set(cty.String))
			if explicitEmpty {
				configured = cty.SetValEmpty(cty.String)
			}
			state.RawConfig = cty.ObjectVal(map[string]cty.Value{"permission_ids": configured})
			d = OrganizationRole().Data(state)
			d.SetId("")
			require.Empty(t, organizationRoleCreate(ctx, d, meta))
			require.Equal(t, "custom", d.Id())
			require.Equal(t, "analytics_reader", created.Key)
			require.Equal(t, "member", created.BaseRoleID)
			if explicitEmpty {
				require.Empty(t, created.PermissionIDs)
			} else {
				require.Equal(t, []string{"read-environments"}, created.PermissionIDs)
			}
			roles, err := meta.GetFronteggRoles(ctx)
			require.NoError(t, err)
			require.Equal(t, []string{"custom"}, roles["analytics_reader"])
			d = schema.TestResourceDataRaw(t, OrganizationRole().Schema, map[string]interface{}{
				"name": "analytics_reader", "description": "Updated description", "permission_ids": []interface{}{"read-profile"},
			})
			d.SetId("custom")
			require.Empty(t, organizationRoleUpdate(ctx, d, meta))
			require.Equal(t, "Updated description", role.Description)
			require.Equal(t, []string{"read-profile"}, role.Permissions)
			imported := schema.TestResourceDataRaw(t, OrganizationRole().Schema, nil)
			imported.SetId("custom")
			require.Empty(t, organizationRoleRead(ctx, imported, meta))
			require.Equal(t, "analytics_reader", imported.Get("key"))
			require.Empty(t, organizationRoleDelete(ctx, d, meta))
			require.Empty(t, d.Id())
			require.Empty(t, organizationRoleRead(ctx, imported, meta))
			require.Empty(t, imported.Id())
		})
	}
}

func TestOrganizationRoleRejectsBuiltinsAndPreservesStateOnError(t *testing.T) {
	for _, status := range []int{200, 403, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "GET", r.Method)
				if status != 200 {
					http.Error(w, "unavailable", status)
					return
				}
				json.NewEncoder(w).Encode(frontegg.FronteggRolesResponse{Items: []frontegg.FronteggRole{{ID: "builtin", Name: "Organization Admin"}}})
			}))
			defer server.Close()
			meta := &utils.ProviderMeta{Frontegg: &clients.FronteggClient{Endpoint: server.URL, HTTPClient: server.Client()}}
			d := schema.TestResourceDataRaw(t, OrganizationRole().Schema, nil)
			d.SetId("builtin")
			require.True(t, organizationRoleRead(context.Background(), d, meta).HasError())
			require.Equal(t, "builtin", d.Id())
			require.True(t, organizationRoleDelete(context.Background(), d, meta).HasError())
			require.Equal(t, "builtin", d.Id())
		})
	}
}

func TestOrganizationRoleReservedNames(t *testing.T) {
	for _, name := range []string{"Admin", "Member", "Organization Admin", "Organization Member", "MaterializePlatformAdmin", "MaterializePlatform", "mZ_system", "PG_reader", "external_team", "PUBLIC", "Current_User", ""} {
		_, errs := validateOrganizationRoleName(name, "name")
		require.NotEmpty(t, errs, name)
	}
	_, errs := validateOrganizationRoleName("Organization Analytics", "name")
	require.Empty(t, errs)
	d := schema.TestResourceDataRaw(t, OrganizationRole().Schema, map[string]interface{}{"name": "reader"})
	require.True(t, organizationRoleCreate(context.Background(), d, &utils.ProviderMeta{Mode: utils.ModeSelfHosted}).HasError())
}

func TestOrganizationRoleTerraformLifecycle(t *testing.T) {
	var mu sync.Mutex
	var role *frontegg.FronteggRole
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case "GET":
			roles := []frontegg.FronteggRole{{ID: "member", Name: "Organization Member", Permissions: []string{"read-environments"}}}
			if role != nil {
				roles = append(roles, *role)
			}
			json.NewEncoder(w).Encode(frontegg.FronteggRolesResponse{Items: roles})
		case "POST":
			var params frontegg.OrganizationRoleCreateParams
			require.NoError(t, json.NewDecoder(r.Body).Decode(&params))
			role = &frontegg.FronteggRole{ID: "custom", Name: params.Name, Key: params.Key, TenantID: "tenant", Description: params.Description, Permissions: params.PermissionIDs}
			json.NewEncoder(w).Encode(role)
		case "PATCH":
			var params map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&params))
			role.Description = params["description"]
			w.WriteHeader(204)
		case "PUT":
			var params struct {
				IDs []string `json:"permissionIds"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&params))
			role.Permissions = params.IDs
			w.WriteHeader(204)
		case "DELETE":
			role = nil
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected request %s", r.Method)
		}
	}))
	defer server.Close()
	meta := &utils.ProviderMeta{Frontegg: &clients.FronteggClient{Endpoint: server.URL, HTTPClient: server.Client()}}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){"materialize": func() (*schema.Provider, error) {
			return &schema.Provider{ResourcesMap: map[string]*schema.Resource{"materialize_organization_role": OrganizationRole()}, ConfigureContextFunc: func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) { return meta, nil }}, nil
		}},
		Steps: []resource.TestStep{
			{Config: `resource "materialize_organization_role" "test" { name = "analytics_reader" }`, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("materialize_organization_role.test", "key", "analytics_reader"),
				resource.TestCheckResourceAttr("materialize_organization_role.test", "permission_ids.#", "1"),
			)},
			{ResourceName: "materialize_organization_role.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"base_role_name"}},
			{Config: `resource "materialize_organization_role" "test" {
    name = "analytics_reader"
    description = "Updated"
    permission_ids = []
   }`, Check: resource.TestCheckResourceAttr("materialize_organization_role.test", "permission_ids.#", "0")},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	require.Nil(t, role)
}
