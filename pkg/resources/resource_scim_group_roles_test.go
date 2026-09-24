package resources

import (
	"context"
	"encoding/json"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestScimGroupRoleResourceCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"group_id": "test-group-id",
		"roles":    []interface{}{"Admin", "Member"},
	}
	d := schema.TestResourceDataRaw(t, ScimGroupRoleSchema, in)
	r.NotNil(d)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
			FronteggRoles: map[string][]string{
				"Admin":  {"1"},
				"Member": {"2"},
			},
		}

		if err := scimGroupRoleCreate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after create
		r.Equal("test-group-id", d.Id())
	})
}

func TestScimGroupRoleResourceRead(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		d := schema.TestResourceDataRaw(t, ScimGroupRoleSchema, nil)
		d.SetId("mock-group-id")

		if err := scimGroupRoleRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		r.Equal("mock-group-id", d.Id())
	})
}

func TestScimGroupRoleResourceDelete(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
			FronteggRoles: map[string][]string{
				"Admin":  {"1"},
				"Member": {"2"},
			},
		}

		d := schema.TestResourceDataRaw(t, ScimGroupRoleSchema, nil)
		d.SetId("mock-group-id")
		d.Set("group_id", "test-group-id")

		if err := scimGroupRoleDelete(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after delete
		r.Empty(d.Id())
	})
}

func TestScimGroupRolesCustomNamesAndUpdate(t *testing.T) {
	group := frontegg.ScimGroup{ID: "group", Roles: []frontegg.ScimRole{{ID: "member", Name: "Organization Member"}, {ID: "old", Name: "Organization Old"}}}
	var removed, added []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(group)
			return
		}
		var body struct {
			IDs []string `json:"roleIds"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		switch r.Method {
		case "DELETE":
			removed = append(removed, body.IDs...)
			group.Roles = group.Roles[:1]
			w.WriteHeader(200)
		case "POST":
			added = append(added, body.IDs...)
			group.Roles = append(group.Roles, frontegg.ScimRole{ID: "new", Name: "Organization Analytics"})
			w.WriteHeader(201)
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()
	meta := &utils.ProviderMeta{Frontegg: &clients.FronteggClient{Endpoint: server.URL, HTTPClient: server.Client()}, FronteggRoles: map[string][]string{"Member": {"member"}, "Organization Analytics": {"new"}}}
	d := schema.TestResourceDataRaw(t, ScimGroupRoleSchema, map[string]interface{}{"group_id": "group", "roles": []interface{}{"Member", "Organization Analytics"}})
	d.SetId("group")
	require.Empty(t, scimGroupRoleUpdate(context.Background(), d, meta))
	require.Equal(t, []string{"old"}, removed)
	require.ElementsMatch(t, []string{"member", "new"}, added)
	require.ElementsMatch(t, []interface{}{"Member", "Organization Analytics"}, d.Get("roles").(*schema.Set).List())
}

// scimRoleRemovalServer answers the group role removal with status and records
// which role IDs it was asked to remove.
func scimRoleRemovalServer(t *testing.T, status int) (*httptest.Server, func() ([]string, int)) {
	var mu sync.Mutex
	var removed []string
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Method != http.MethodDelete || !strings.HasSuffix(r.URL.Path, "/roles") {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		calls++
		var body struct {
			RoleIds []string `json:"roleIds"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		removed = append(removed, body.RoleIds...)
		w.WriteHeader(status)
	}))
	return srv, func() ([]string, int) { mu.Lock(); defer mu.Unlock(); return removed, calls }
}

func scimGroupRoleDeleteData(t *testing.T, roles ...interface{}) *schema.ResourceData {
	d := schema.TestResourceDataRaw(t, ScimGroupRoleSchema, map[string]interface{}{
		"group_id": "test-group-id",
		"roles":    roles,
	})
	d.SetId("test-group-id")
	return d
}

// A role deleted in Frontegg takes its group assignment with it, so the destroy
// should remove what is left and succeed rather than get stuck on the lookup.
func TestScimGroupRoleResourceDeleteSkipsMissingRole(t *testing.T) {
	r := require.New(t)
	srv, recorded := scimRoleRemovalServer(t, http.StatusOK)
	defer srv.Close()

	providerMeta := &utils.ProviderMeta{
		Frontegg:      &clients.FronteggClient{Endpoint: srv.URL, HTTPClient: srv.Client()},
		FronteggRoles: map[string][]string{"Admin": {"1"}}, // Member is gone
	}
	d := scimGroupRoleDeleteData(t, "Admin", "Member")

	r.False(scimGroupRoleDelete(context.TODO(), d, providerMeta).HasError())
	r.Empty(d.Id())
	removed, calls := recorded()
	r.Equal(1, calls)
	r.Equal([]string{"1"}, removed)
}

func TestScimGroupRoleResourceDeleteAllRolesMissing(t *testing.T) {
	r := require.New(t)
	srv, recorded := scimRoleRemovalServer(t, http.StatusOK)
	defer srv.Close()

	providerMeta := &utils.ProviderMeta{
		Frontegg:      &clients.FronteggClient{Endpoint: srv.URL, HTTPClient: srv.Client()},
		FronteggRoles: map[string][]string{},
	}
	d := scimGroupRoleDeleteData(t, "Admin", "Member")

	r.False(scimGroupRoleDelete(context.TODO(), d, providerMeta).HasError())
	r.Empty(d.Id())
	_, calls := recorded()
	r.Equal(0, calls, "nothing left to remove, so no request should be made")
}

func TestScimGroupRoleResourceDeleteGroupGone(t *testing.T) {
	r := require.New(t)
	srv, _ := scimRoleRemovalServer(t, http.StatusNotFound)
	defer srv.Close()

	providerMeta := &utils.ProviderMeta{
		Frontegg:      &clients.FronteggClient{Endpoint: srv.URL, HTTPClient: srv.Client()},
		FronteggRoles: map[string][]string{"Admin": {"1"}},
	}
	d := scimGroupRoleDeleteData(t, "Admin")

	r.False(scimGroupRoleDelete(context.TODO(), d, providerMeta).HasError())
	r.Empty(d.Id())
}

// An ambiguous name could remove the wrong role, so that still has to fail.
func TestScimGroupRoleResourceDeleteAmbiguousRoleFails(t *testing.T) {
	r := require.New(t)
	srv, recorded := scimRoleRemovalServer(t, http.StatusOK)
	defer srv.Close()

	providerMeta := &utils.ProviderMeta{
		Frontegg:      &clients.FronteggClient{Endpoint: srv.URL, HTTPClient: srv.Client()},
		FronteggRoles: map[string][]string{"Admin": {"1", "2"}},
	}
	d := scimGroupRoleDeleteData(t, "Admin")

	diags := scimGroupRoleDelete(context.TODO(), d, providerMeta)
	r.True(diags.HasError())
	r.Contains(diags[0].Summary, "ambiguous")
	_, calls := recorded()
	r.Equal(0, calls)
}
