package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceTableWebhook = MaterializeObject{Name: "webhook_table", SchemaName: "schema", DatabaseName: "database"}

func TestSourceTableWebhookCreateExposeHeaders(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."webhook_table"
			FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADER 'timestamp' AS ts
			INCLUDE HEADER 'x-event-type' AS event_type;`,
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

		b := NewSourceTableWebhookBuilder(db, sourceTableWebhook)
		b.BodyFormat("JSON")
		b.IncludeHeader(includeHeader)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableWebhookCreateIncludeHeaders(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."webhook_table"
			FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADERS \(NOT 'authorization', NOT 'x-api-key'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableWebhookBuilder(db, sourceTableWebhook)
		b.BodyFormat("JSON")
		b.IncludeHeaders(IncludeHeadersStruct{
			Not: []string{"authorization", "x-api-key"},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableWebhookCreateValidated(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."webhook_table"
			FROM WEBHOOK BODY FORMAT JSON CHECK
			\( WITH \(HEADERS, BODY AS request_body, SECRET "database"."schema"."my_webhook_shared_secret"\)
			decode\(headers->'x-signature', 'base64'\) = hmac\(request_body, my_webhook_shared_secret, 'sha256'\)\);`,
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

		b := NewSourceTableWebhookBuilder(db, sourceTableWebhook)
		b.BodyFormat("JSON")
		b.CheckOptions(checkOptions)
		b.CheckExpression("decode(headers->'x-signature', 'base64') = hmac(request_body, my_webhook_shared_secret, 'sha256')")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableWebhookCreateSegment(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."webhook_table"
			FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADER 'event-type' AS event_type INCLUDE HEADERS CHECK
			\( WITH \(BODY BYTES, HEADERS, SECRET "database"."schema"."my_webhook_shared_secret" AS secret BYTES\)
			decode\(headers->'x-signature', 'hex'\) = hmac\(body, secret, 'sha1'\)\);`,
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

		b := NewSourceTableWebhookBuilder(db, sourceTableWebhook)
		b.BodyFormat("JSON")
		b.IncludeHeader(includeHeader)
		b.IncludeHeaders(IncludeHeadersStruct{All: true})
		b.CheckOptions(checkOptions)
		b.CheckExpression("decode(headers->'x-signature', 'hex') = hmac(body, secret, 'sha1')")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableWebhookCreateRudderstack(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."webhook_table" FROM WEBHOOK BODY FORMAT JSON CHECK \( WITH \(HEADERS, BODY AS request_body, SECRET "database"."schema"."my_webhook_shared_secret"\) headers->'authorization' = rudderstack_shared_secret\);`,
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

		b := NewSourceTableWebhookBuilder(db, sourceTableWebhook)
		b.BodyFormat("JSON")
		b.CheckOptions(checkOptions)
		b.CheckExpression("headers->'authorization' = rudderstack_shared_secret")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableWebhookRename(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER TABLE "database"."schema"."webhook_table" RENAME TO "database"."schema"."new_webhook_table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableWebhookBuilder(db, sourceTableWebhook)
		if err := b.Rename("new_webhook_table"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableWebhookDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP TABLE "database"."schema"."webhook_table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableWebhookBuilder(db, sourceTableWebhook)
		if err := b.Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
