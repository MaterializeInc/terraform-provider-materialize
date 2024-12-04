package clients

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionString(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf")
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C+serializable&sslmode=require`, c)
}

func TestConnectionStringTesting(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "disable", "tf")
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C+serializable&sslmode=disable`, c)
}

func TestNewDBClientFailure(t *testing.T) {
	r := require.New(t)

	client, diags := NewDBClient("localhost", "user", "pass", 6875, "database", "tf-provider", "v0.1.0", "invalid-sslmode")
	r.NotEmpty(diags)
	r.Nil(client)
}
