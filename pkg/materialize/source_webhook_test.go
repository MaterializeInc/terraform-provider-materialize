package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceWebhook = MaterializeObject{Name: "webhook_source", SchemaName: "schema", DatabaseName: "database"}

func TestSourceWebhookCreateExposeHeaders(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."webhook_source" IN CLUSTER "cluster" FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADER 'timestamp' AS ts INCLUDE HEADER 'x-event-type' AS event_type;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		var includeHeader = []HeaderStruct{
			{
				Header: "timestamp",
				Alias:  "ts",
			},
			{
				Header: "x-event-type",
				Alias:  "event_type",
			},
		}

		b := NewSourceWebhookBuilder(db, sourceWebhook)
		b.ClusterName("cluster")
		b.BodyFormat("JSON")
		b.IncludeHeader(includeHeader)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceWebhookCreateIncludeHeaders(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."webhook_source" IN CLUSTER "cluster" FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADERS \(NOT 'authorization', NOT 'x-api-key'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		var excludeHeaders = []string{"authorization", "x-api-key"}

		b := NewSourceWebhookBuilder(db, sourceWebhook)
		b.ClusterName("cluster")
		b.BodyFormat("JSON")
		b.IncludeHeaders(true)
		b.ExcludeHeaders(excludeHeaders)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceWebhookCreateValidated(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."webhook_source" IN CLUSTER "cluster" FROM WEBHOOK BODY FORMAT JSON CHECK \( WITH \(HEADERS, BODY AS request_body, SECRET "database"."schema"."my_webhook_shared_secret"\) decode\(headers->'x-signature', 'base64'\) = hmac\(request_body, my_webhook_shared_secret, 'sha256'\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		var checkOptions = []CheckOptionsStruct{
			{
				Field: FieldStruct{Headers: true},
			},
			{
				Field: FieldStruct{Body: true},
				Alias: "request_body",
			},
			{
				Field: FieldStruct{
					Secret: IdentifierSchemaStruct{
						DatabaseName: "database",
						SchemaName:   "schema",
						Name:         "my_webhook_shared_secret",
					},
				},
			},
		}

		b := NewSourceWebhookBuilder(db, sourceWebhook)
		b.ClusterName("cluster")
		b.BodyFormat("JSON")
		b.CheckOptions(checkOptions)
		b.CheckExpression("decode(headers->'x-signature', 'base64') = hmac(request_body, my_webhook_shared_secret, 'sha256')")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceWebhookCreateSegment(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."webhook_source" IN CLUSTER "cluster" FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADER 'event-type' AS event_type INCLUDE HEADERS CHECK \( WITH \(BODY BYTES, HEADERS, SECRET "database"."schema"."my_webhook_shared_secret" AS secret BYTES\) decode\(headers->'x-signature', 'hex'\) = hmac\(body, secret, 'sha1'\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		var includeHeader = []HeaderStruct{
			{
				Header: "event-type",
				Alias:  "event_type",
			},
		}
		var checkOptions = []CheckOptionsStruct{
			{
				Field: FieldStruct{Body: true},
				Bytes: true,
			},
			{
				Field: FieldStruct{Headers: true},
			},
			{
				Field: FieldStruct{
					Secret: IdentifierSchemaStruct{
						DatabaseName: "database",
						SchemaName:   "schema",
						Name:         "my_webhook_shared_secret",
					},
				},
				Alias: "secret",
				Bytes: true,
			},
		}

		b := NewSourceWebhookBuilder(db, sourceWebhook)
		b.ClusterName("cluster")
		b.BodyFormat("JSON")
		b.IncludeHeader(includeHeader)
		b.IncludeHeaders(true)
		b.CheckOptions(checkOptions)
		b.CheckExpression("decode(headers->'x-signature', 'hex') = hmac(body, secret, 'sha1')")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceWebhookCreateRudderstack(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."webhook_source" IN CLUSTER "cluster" FROM WEBHOOK BODY FORMAT JSON CHECK \( WITH \(HEADERS, BODY AS request_body, SECRET "database"."schema"."my_webhook_shared_secret"\) headers->'authorization' = rudderstack_shared_secret\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		var checkOptions = []CheckOptionsStruct{
			{
				Field: FieldStruct{Headers: true},
			},
			{
				Field: FieldStruct{Body: true},
				Alias: "request_body",
			},
			{
				Field: FieldStruct{
					Secret: IdentifierSchemaStruct{
						DatabaseName: "database",
						SchemaName:   "schema",
						Name:         "my_webhook_shared_secret",
					},
				},
			},
		}

		b := NewSourceWebhookBuilder(db, sourceWebhook)
		b.ClusterName("cluster")
		b.BodyFormat("JSON")
		b.CheckOptions(checkOptions)
		b.CheckExpression("headers->'authorization' = rudderstack_shared_secret")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
