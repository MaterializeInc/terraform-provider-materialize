package resources

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
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
			FronteggRoles: map[string]string{
				"Admin":  "1",
				"Member": "2",
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
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
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
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
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
