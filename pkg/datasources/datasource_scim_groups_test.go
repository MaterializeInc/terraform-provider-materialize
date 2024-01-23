package datasources

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestDataSourceSCIMGroupsRead_Success(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		d := schema.TestResourceDataRaw(t, dataSourceSCIMGroupsSchema, nil)
		d.SetId("scim_groups")

		if err := dataSourceSCIMGroupsRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Validate the ID of the data source
		r.Equal("scim_groups", d.Id())

		// Validate the SCIM groups
		scimGroups := d.Get("groups").([]interface{})
		r.NotEmpty(scimGroups)
		r.Len(scimGroups, 1)

		groupMap := scimGroups[0].(map[string]interface{})

		// Validate each field within the first SCIM group
		r.Equal("group-1", groupMap["id"].(string))
		r.Equal("Test Group 1", groupMap["name"].(string))
		r.Equal("Description for Test Group 1", groupMap["description"].(string))
		r.Equal("{}", groupMap["metadata"].(string))
		r.Equal("frontegg", groupMap["managed_by"].(string))

		// Validate roles within the first group
		roles := groupMap["roles"].([]interface{})
		r.NotEmpty(roles)
		r.Len(roles, 1)

		roleMap := roles[0].(map[string]interface{})
		r.Equal("role-1", roleMap["id"].(string))
		r.Equal("role-key-1", roleMap["key"].(string))
		r.Equal("Role 1", roleMap["name"].(string))
		r.Equal("Role 1 Description", roleMap["description"].(string))
		r.Equal(true, roleMap["is_default"].(bool))

		// Validate users within the first group
		users := groupMap["users"].([]interface{})
		r.NotEmpty(users)
		r.Len(users, 1)

		userMap := users[0].(map[string]interface{})
		r.Equal("user-1", userMap["id"].(string))
		r.Equal("User 1", userMap["name"].(string))
		r.Equal("user1@example.com", userMap["email"].(string))
	})
}
