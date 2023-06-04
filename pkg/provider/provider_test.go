package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestConnectionString(t *testing.T) {
	r := require.New(t)
	c := connectionString("host", "user", "pass", 6875, "database", false)
	r.Equal(`postgres://user:pass@host:6875/database?sslmode=require`, c)
}

func TestConnectionStringTesting(t *testing.T) {
	r := require.New(t)
	c := connectionString("host", "user", "pass", 6875, "database", true)
	r.Equal(`postgres://user:pass@host:6875/database?sslmode=disable`, c)
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

var testAccProvider = Provider()
var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"materialize": func() (*schema.Provider, error) { return testAccProvider, nil },
}

func testAccPreCheck(t *testing.T) {

}
