package clients

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionString(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf", nil)
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C+serializable&sslmode=require`, c)
}

func TestConnectionStringTesting(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "disable", "tf", nil)
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C+serializable&sslmode=disable`, c)
}

func TestConnectionStringWithOptions(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf", map[string]string{
		"search_path": "public,extra",
		"cluster":     "quickstart",
	})
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C+serializable+--cluster%3Dquickstart+--search_path%3Dpublic%2Cextra&sslmode=require`, c)
}

func TestConnectionStringOptionEscaping(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf", map[string]string{
		"application_name": "my app",
	})
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C+serializable+--application_name%3Dmy%5C+app&sslmode=require`, c)
}

func TestNewDBClientFailure(t *testing.T) {
	r := require.New(t)

	client, diags := NewDBClient("localhost", "user", "pass", 6875, "database", "tf-provider", "v0.1.0", "invalid-sslmode", nil)
	r.NotEmpty(diags)
	r.Nil(client)
}
