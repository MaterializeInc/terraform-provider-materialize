package clients

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/stretchr/testify/require"
)

func TestConnectionString(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf", nil)
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C%20serializable&sslmode=require`, c)
}

func TestConnectionStringTesting(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "disable", "tf", nil)
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C%20serializable&sslmode=disable`, c)
}

func TestConnectionStringWithOptions(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf", map[string]string{
		"search_path": "public,extra",
		"cluster":     "quickstart",
	})
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C%20serializable%20--cluster%3Dquickstart%20--search_path%3Dpublic%2Cextra&sslmode=require`, c)
}

func TestConnectionStringOptionEscaping(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf", map[string]string{
		"application_name": "my app",
	})
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C%20serializable%20--application_name%3Dmy%5C%20app&sslmode=require`, c)
}

// Guard against regressions in escapeOptionToken: backslashes MUST be escaped
// before spaces, otherwise an input like `a\ b` (backslash + space) would
// double-escape the newly added backslash and produce a malformed token.
func TestConnectionStringBackslashAndSpaceEscaping(t *testing.T) {
	r := require.New(t)
	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf", map[string]string{
		"search_path": `a\b c`,
	})
	// The options value after escaping is `--search_path=a\\b\ c`, which the
	// URL encoder renders with %5C for each backslash and %20 for the space.
	r.Equal(`postgres://user:pass@host:6875/database?application_name=tf&options=--transaction_isolation%3Dstrict%5C%20serializable%20--search_path%3Da%5C%5Cb%5C%20c&sslmode=require`, c)
}

func TestNewDBClientFailure(t *testing.T) {
	r := require.New(t)

	client, diags := NewDBClient("localhost", "user", "pass", 6875, "database", "tf-provider", "v0.1.0", "invalid-sslmode", nil)
	r.NotEmpty(diags)
	r.Nil(client)
}

// pgx v5 does not read "+" back as a space in query values, so the connection
// string must percent-encode the spaces in the options string. Otherwise the
// server sees a malformed value, ignores the option, and the session silently
// falls back to its default isolation level.
func TestConnectionStringOptionsSurviveParsing(t *testing.T) {
	r := require.New(t)

	c := buildConnectionString("host", "user", "pass", 6875, "database", "require", "tf", map[string]string{
		"search_path": "my schema",
	})

	cfg, err := pgconn.ParseConfig(c)
	r.NoError(err)
	r.Equal(`--transaction_isolation=strict\ serializable --search_path=my\ schema`, cfg.RuntimeParams["options"])
}
