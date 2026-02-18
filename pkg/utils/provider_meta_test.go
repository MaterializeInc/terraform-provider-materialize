package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDBClientFromMeta(t *testing.T) {
	// Set up the SQL mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Wrap the sql.DB with sqlx
	dbx := sqlx.NewDb(db, "sqlmock")

	// Set up the mock ProviderMeta
	providerMeta := &ProviderMeta{
		Mode: ModeSaaS,
		DB: map[clients.Region]*clients.DBClient{
			clients.AwsUsEast1: {DB: dbx},
		},
		RegionsEnabled: map[clients.Region]bool{
			clients.AwsUsEast1: true,
		},
		DefaultRegion: clients.AwsUsEast1,
		Frontegg: &clients.FronteggClient{
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// Create a ResourceData schema with the "region" key
	resourceDataSchema := map[string]*schema.Schema{
		"region": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	// Create a ResourceData object with a valid region
	resourceData := schema.TestResourceDataRaw(t, resourceDataSchema, nil)
	err = resourceData.Set("region", "aws/us-east-1")
	require.NoError(t, err)

	// Call the GetDBClientFromMeta function
	dbClient, _, err := GetDBClientFromMeta(providerMeta, resourceData)
	require.NoError(t, err)
	assert.NotNil(t, dbClient)

	// Check that the mock expectations are met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestTransformIdWithRegion(t *testing.T) {
	testCases := []map[string]interface{}{
		{
			"input":    "aws/us-east-1:GRANT DEFAULT|SCHEMA|u1|u1|||USAGE",
			"expected": "aws/us-east-1:GRANT DEFAULT|SCHEMA|u1|u1|||USAGE",
		},
		{
			"input":    "GRANT DEFAULT|SCHEMA|u1|u1|||USAGE",
			"expected": "aws/us-east-1:GRANT DEFAULT|SCHEMA|u1|u1|||USAGE",
		},
		{
			"input":    "aws/us-east-1:u1",
			"expected": "aws/us-east-1:u1",
		},
		{
			"input":    "u1",
			"expected": "aws/us-east-1:u1",
		},
	}
	for tc := range testCases {
		c := testCases[tc]
		o := TransformIdWithRegion("aws/us-east-1", c["input"].(string))
		assert.Equal(t, o, c["expected"].(string))
	}
}

func TestGetDBClientFromMetaSelfHosted(t *testing.T) {
	// Set up the SQL mock database for self-hosted
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Wrap the sql.DB with sqlx
	dbx := sqlx.NewDb(db, "sqlmock")

	// Set up the mock ProviderMeta for self-hosted
	providerMeta := &ProviderMeta{
		Mode: ModeSelfHosted,
		DB: map[clients.Region]*clients.DBClient{
			"self-hosted": {DB: dbx},
		},
		DefaultRegion: "self-hosted",
		RegionsEnabled: map[clients.Region]bool{
			"self-hosted": true,
		},
	}

	// Test cases to verify different scenarios
	tests := []struct {
		name         string
		resourceData *schema.ResourceData
		wantRegion   clients.Region
		wantErr      bool
		errMsg       string
	}{
		{
			name: "self hosted basic",
			resourceData: schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"region": {
					Type:     schema.TypeString,
					Optional: true,
				},
			}, nil),
			wantRegion: "self-hosted",
			wantErr:    false,
		},
		{
			name: "self hosted with region set (should be overridden)",
			resourceData: func() *schema.ResourceData {
				d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"region": {
						Type:     schema.TypeString,
						Optional: true,
					},
				}, nil)
				d.Set("region", "aws/us-east-1")
				return d
			}(),
			wantRegion: "self-hosted",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient, region, err := GetDBClientFromMeta(providerMeta, tt.resourceData)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, dbClient)
			assert.Equal(t, tt.wantRegion, region)

			if tt.resourceData != nil {
				assert.Equal(t, string(tt.wantRegion), tt.resourceData.Get("region"))
			}
		})
	}

	// Check that the mock expectations are met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestProviderMetaModeHelpers(t *testing.T) {
	tests := []struct {
		name           string
		mode           ProviderMode
		wantSaaS       bool
		wantSelfHosted bool
	}{
		{
			name:           "explicit saas mode",
			mode:           ModeSaaS,
			wantSaaS:       true,
			wantSelfHosted: false,
		},
		{
			name:           "empty mode defaults to saas",
			mode:           "",
			wantSaaS:       true,
			wantSelfHosted: false,
		},
		{
			name:           "self hosted mode",
			mode:           ModeSelfHosted,
			wantSaaS:       false,
			wantSelfHosted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &ProviderMeta{Mode: tt.mode}
			assert.Equal(t, tt.wantSaaS, pm.IsSaaS())
			assert.Equal(t, tt.wantSelfHosted, pm.IsSelfHosted())
		})
	}
}

func TestProviderMeta_ValidateSaaSOnly_SelfHosted(t *testing.T) {
	r := require.New(t)

	providerMeta := &ProviderMeta{
		Mode: ModeSelfHosted,
	}

	diags := providerMeta.ValidateSaaSOnly("materialize_app_password")
	r.True(diags.HasError())

	r.Contains(diags[0].Summary, "materialize_app_password is only available in Materialize Cloud (SaaS) environments")
	r.Contains(diags[0].Summary, "You are currently using self-hosted authentication mode")
	r.Contains(diags[0].Summary, "you need to switch to Materialize Cloud (SaaS) mode")
}

func TestProviderMeta_ValidateSaaSOnly_SaaS(t *testing.T) {
	r := require.New(t)

	providerMeta := &ProviderMeta{
		Mode: ModeSaaS,
	}

	diags := providerMeta.ValidateSaaSOnly("materialize_app_password")
	r.False(diags.HasError())
}

func TestProviderMeta_ValidateSaaSOnly_EmptyMode(t *testing.T) {
	r := require.New(t)

	// Empty mode should default to SaaS
	providerMeta := &ProviderMeta{
		Mode: "",
	}

	diags := providerMeta.ValidateSaaSOnly("materialize_app_password")
	r.False(diags.HasError())
}

func TestProviderMeta_GetFronteggRoles_AlreadyLoaded(t *testing.T) {
	r := require.New(t)

	// When FronteggRoles is already set, it should return without calling fetcher
	providerMeta := &ProviderMeta{
		FronteggRoles: map[string]string{
			"Admin":  "role-1",
			"Member": "role-2",
		},
	}

	roles, err := providerMeta.GetFronteggRoles(context.Background())
	r.NoError(err)
	r.Equal("role-1", roles["Admin"])
	r.Equal("role-2", roles["Member"])
}

func TestProviderMeta_GetFronteggRoles_NoFetcher(t *testing.T) {
	r := require.New(t)

	// When no fetcher is configured, it should return an error
	providerMeta := &ProviderMeta{}

	_, err := providerMeta.GetFronteggRoles(context.Background())
	r.Error(err)
	r.Contains(err.Error(), "frontegg roles fetcher not configured")
}

func TestProviderMeta_GetFronteggRoles_LazyLoad(t *testing.T) {
	r := require.New(t)

	fetchCount := 0
	providerMeta := &ProviderMeta{
		FronteggRolesFetcher: func(ctx context.Context) (map[string]string, error) {
			fetchCount++
			return map[string]string{
				"Admin":  "fetched-role-1",
				"Member": "fetched-role-2",
			}, nil
		},
	}

	// First call should fetch roles
	roles, err := providerMeta.GetFronteggRoles(context.Background())
	r.NoError(err)
	r.Equal("fetched-role-1", roles["Admin"])
	r.Equal(1, fetchCount)

	// Second call should return cached roles without fetching again
	roles, err = providerMeta.GetFronteggRoles(context.Background())
	r.NoError(err)
	r.Equal("fetched-role-1", roles["Admin"])
	r.Equal(1, fetchCount) // Still 1, fetcher not called again
}

func TestProviderMeta_GetFronteggRoles_FetcherError(t *testing.T) {
	r := require.New(t)

	providerMeta := &ProviderMeta{
		FronteggRolesFetcher: func(ctx context.Context) (map[string]string, error) {
			return nil, errors.New("network error")
		},
	}

	_, err := providerMeta.GetFronteggRoles(context.Background())
	r.Error(err)
	r.Contains(err.Error(), "unable to fetch Frontegg roles")
	r.Contains(err.Error(), "Organization Admin permissions")
}

func TestProviderMeta_GetFronteggRoles_RetryAfterError(t *testing.T) {
	r := require.New(t)

	callCount := 0
	providerMeta := &ProviderMeta{
		FronteggRolesFetcher: func(ctx context.Context) (map[string]string, error) {
			callCount++
			if callCount == 1 {
				return nil, errors.New("transient error")
			}
			return map[string]string{
				"Admin":  "role-1",
				"Member": "role-2",
			}, nil
		},
	}

	// First call fails
	_, err := providerMeta.GetFronteggRoles(context.Background())
	r.Error(err)
	r.Equal(1, callCount)

	// Second call should retry and succeed (not cached on error)
	roles, err := providerMeta.GetFronteggRoles(context.Background())
	r.NoError(err)
	r.Equal("role-1", roles["Admin"])
	r.Equal(2, callCount)

	// Third call should use cached value
	roles, err = providerMeta.GetFronteggRoles(context.Background())
	r.NoError(err)
	r.Equal("role-1", roles["Admin"])
	r.Equal(2, callCount) // Still 2, fetcher not called again after success
}
