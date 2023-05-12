package materialize

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestConnectionSshTunnelCreateQuery(t *testing.T) {
// 	r := require.New(t)

// 	b := NewConnectionSshTunnelBuilder("ssh_conn", "schema", "database")
// 	b.SSHHost("localhost")
// 	b.SSHPort(123)
// 	b.SSHUser("user")
// 	r.Equal(`CREATE CONNECTION "database"."schema"."ssh_conn" TO SSH TUNNEL (HOST 'localhost', USER 'user', PORT 123);`, b.Create())
// }
