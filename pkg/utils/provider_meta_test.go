package utils

import (
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
	for tc, _ := range testCases {
		c := testCases[tc]
		o := TransformIdWithRegion("aws/us-east-1", c["input"].(string))
		assert.Equal(t, o, c["expected"].(string))
	}
}
