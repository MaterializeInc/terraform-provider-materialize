package materialize

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestConnectionCreateAwsPrivateLinkQuery(t *testing.T) {
// 	r := require.New(t)

// 	b := NewConnectionAwsPrivatelinkBuilder("privatelink_conn", "schema", "database")
// 	b.PrivateLinkServiceName("com.amazonaws.us-east-1.materialize.example")
// 	b.PrivateLinkAvailabilityZones([]string{"use1-az1", "use1-az2"})
// 	r.Equal(`CREATE CONNECTION "database"."schema"."privatelink_conn" TO AWS PRIVATELINK (SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',AVAILABILITY ZONES ('use1-az1', 'use1-az2'));`, b.Create())
// }
