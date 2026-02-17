package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var connectionKafkaSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"kafka_broker": {
		Description:   "The Kafka broker's configuration.",
		Type:          schema.TypeList,
		ConflictsWith: []string{"aws_privatelink"},
		AtLeastOneOf:  []string{"kafka_broker", "aws_privatelink"},
		Optional:      true,
		MinItems:      1,
		ForceNew:      false,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"broker": {
					Description: "The Kafka broker, in the form of `host:port`.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"target_group_port": {
					Description: "The port of the target group associated with the Kafka broker.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"availability_zone": {
					Description: "The availability zone of the Kafka broker.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"privatelink_connection": IdentifierSchema(IdentifierSchemaParams{
					Elem:        "privatelink_connection",
					Description: "The AWS PrivateLink connection name in Materialize.",
					Required:    false,
					ForceNew:    false,
				}),
				"ssh_tunnel": IdentifierSchema(IdentifierSchemaParams{
					Elem:        "ssh_tunnel",
					Description: "The name of an SSH tunnel connection to route network traffic through by default.",
					Required:    false,
					ForceNew:    false,
				}),
			},
		},
	},
	"aws_privatelink": {
		Description:   "AWS PrivateLink configuration. Conflicts with `kafka_broker`.",
		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{"kafka_broker"},
		AtLeastOneOf:  []string{"kafka_broker", "aws_privatelink"},
		MinItems:      1,
		MaxItems:      1,
		ForceNew:      false,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"privatelink_connection": IdentifierSchema(IdentifierSchemaParams{
					Elem:        "privatelink_connection",
					Description: "The AWS PrivateLink connection name in Materialize.",
					Required:    true,
					ForceNew:    false,
				}),
				"privatelink_connection_port": {
					Description: "The port of the AWS PrivateLink connection.",
					Type:        schema.TypeInt,
					Required:    true,
					ForceNew:    false,
				},
			},
		},
	},
	"aws_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_connection",
		Description: "The AWS connection to use for IAM authentication.",
		Required:    false,
		ForceNew:    false,
	}),
	"security_protocol": {
		Description:  "The security protocol to use: `PLAINTEXT`, `SSL`, `SASL_PLAINTEXT`, or `SASL_SSL`.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     false,
		ValidateFunc: validation.StringInSlice(securityProtocols, true),
		StateFunc: func(val any) string {
			return strings.ToUpper(val.(string))
		},
	},
	"progress_topic": {
		Description: "The name of a topic that Kafka sinks can use to track internal consistency metadata.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"progress_topic_replication_factor": {
		Description: "The replication factor to use when creating the Kafka progress topic (if the Kafka topic does not already exist).",
		Type:        schema.TypeInt,
		Optional:    true,
		ForceNew:    true,
	},
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Kafka broker.", false, false),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Kafka broker.", false, false),
	"ssl_key": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssl_key",
		Description: "The client key for the Kafka broker.",
		Required:    false,
		ForceNew:    false,
	}),
	"sasl_mechanisms": {
		Description:  "The SASL mechanism for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(saslMechanisms, true),
		RequiredWith: []string{"sasl_username", "sasl_password"},
		StateFunc: func(val any) string {
			return strings.ToUpper(val.(string))
		},
		ForceNew: false,
	},
	"sasl_username": ValueSecretSchema("sasl_username", "The SASL username for the Kafka broker.", false, false),
	"sasl_password": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "sasl_password",
		Description: "The SASL password for the Kafka broker.",
		Required:    false,
		ForceNew:    false,
	}),
	"ssh_tunnel": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssh_tunnel",
		Description: "The default SSH tunnel configuration for the Kafka brokers.",
		Required:    false,
		ForceNew:    false,
	}),
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionKafka() *schema.Resource {
	return &schema.Resource{
		Description: "A Kafka connection establishes a link to a Kafka cluster.",

		CreateContext: connectionKafkaCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionKafkaUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionKafkaSchema,
	}
}

func connectionKafkaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.BaseConnection, Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionKafkaBuilder(metaDb, o)

	if v, ok := d.GetOk("kafka_broker"); ok {
		brokers := materialize.GetKafkaBrokersStruct(v)
		b.KafkaBrokers(brokers)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		privatelink := materialize.GetAwsPrivateLinkConnectionStruct(v)
		b.KafkaAwsPrivateLink(privatelink)
	}

	if v, ok := d.GetOk("aws_connection"); ok {
		awsConn := materialize.GetIdentifierSchemaStruct(v)
		b.AwsConnection(awsConn)
	}

	if v, ok := d.GetOk("security_protocol"); ok {
		b.KafkaSecurityProtocol(v.(string))
	}

	if v, ok := d.GetOk("progress_topic"); ok {
		b.KafkaProgressTopic(v.(string))
	}

	if v, ok := d.GetOk("progress_topic_replication_factor"); ok {
		b.KafkaProgressTopicReplicationFactor(v.(int))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(v)
		b.KafkaSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(v)
		b.KafkaSSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		key := materialize.GetIdentifierSchemaStruct(v)
		b.KafkaSSLKey(key)
	}

	if v, ok := d.GetOk("sasl_mechanisms"); ok {
		b.KafkaSASLMechanisms(v.(string))
	}

	if v, ok := d.GetOk("sasl_username"); ok {
		sasl_username := materialize.GetValueSecretStruct(v)
		b.KafkaSASLUsername(sasl_username)
	}

	if v, ok := d.GetOk("sasl_password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(v)
		b.KafkaSASLPassword(pass)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.KafkaSSHTunnel(conn)
	}

	if v, ok := d.GetOk("validate"); ok {
		b.Validate(v.(bool))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if diags := applyOwnership(d, metaDb, o, b); diags != nil {
		return diags
	}

	// object comment
	if diags := applyComment(d, metaDb, o, b); diags != nil {
		return diags
	}

	// set id
	i, err := materialize.ConnectionId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return connectionRead(ctx, d, meta)
}

func connectionKafkaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	validate := d.Get("validate").(bool)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.BaseConnection, Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}

	b := materialize.NewConnectionKafkaBuilder(metaDb, o)
	options := map[string]interface{}{}
	resetOptions := []string{}
	addResetOption := func(option string) {
		for _, existingOption := range resetOptions {
			if existingOption == option {
				return
			}
		}
		resetOptions = append(resetOptions, option)
	}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: materialize.BaseConnection, Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewConnection(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("kafka_broker") {
		_, newBrokers := d.GetChange("kafka_broker")
		kafkaBrokers := materialize.GetKafkaBrokersStruct(newBrokers)
		b.KafkaBrokers(kafkaBrokers)
		builderBrokersString := b.BuildBrokersString()

		if builderBrokersString != "" {
			options["BROKERS"] = materialize.RawSQL(fmt.Sprintf("(%s)", builderBrokersString))
			addResetOption("AWS PRIVATELINK")
		} else {
			addResetOption("BROKERS")
		}
	}

	if d.HasChange("aws_privatelink") {
		_, newAwsPrivatelink := d.GetChange("aws_privatelink")
		awsPrivateLink := materialize.GetAwsPrivateLinkConnectionStruct(newAwsPrivatelink)
		b.KafkaAwsPrivateLink(awsPrivateLink)

		awsPrivateLinkString := b.BuildAwsPrivateLinkString()

		if awsPrivateLinkString != "" {
			options["AWS PRIVATELINK"] = materialize.RawSQL(awsPrivateLinkString)
			addResetOption("BROKERS")
		} else if !d.HasChange("kafka_broker") {
			addResetOption("AWS PRIVATELINK")
		}
	}

	if d.HasChange("security_protocol") {
		_, newProtocol := d.GetChange("security_protocol")
		if newProtocol != "" {
			options["SECURITY PROTOCOL"] = newProtocol.(string)
		} else {
			addResetOption("SECURITY PROTOCOL")
		}
	}

	if d.HasChange("ssl_certificate_authority") {
		_, newSslCa := d.GetChange("ssl_certificate_authority")
		if newSslCa == nil || len(newSslCa.([]interface{})) == 0 {
			addResetOption("SSL CERTIFICATE AUTHORITY")
		} else {
			options["SSL CERTIFICATE AUTHORITY"] = materialize.GetValueSecretStruct(newSslCa)
		}
	}

	if d.HasChange("ssl_certificate") || d.HasChange("ssl_key") {
		newSslCert := d.Get("ssl_certificate")
		newSslKey := d.Get("ssl_key")

		if newSslCert != nil && len(newSslCert.([]interface{})) > 0 && newSslKey != nil && len(newSslKey.([]interface{})) > 0 {
			options["SSL CERTIFICATE"] = materialize.GetValueSecretStruct(newSslCert)
			options["SSL KEY"] = materialize.GetIdentifierSchemaStruct(newSslKey)
		} else {
			addResetOption("SSL CERTIFICATE")
			addResetOption("SSL KEY")
		}
	}

	if d.HasChange("sasl_mechanisms") {
		oldMechanisms, newMechanisms := d.GetChange("sasl_mechanisms")
		if newMechanisms != "" {
			options["SASL MECHANISMS"] = newMechanisms.(string)
		} else if oldMechanisms != "" {
			addResetOption("SASL MECHANISMS")
		}
	}

	if d.HasChange("sasl_username") || d.HasChange("sasl_password") {
		newSaslUsername := d.Get("sasl_username")
		newSaslPassword := d.Get("sasl_password")

		if newSaslUsername != nil && len(newSaslUsername.([]interface{})) > 0 && newSaslPassword != nil && len(newSaslPassword.([]interface{})) > 0 {
			options["SASL USERNAME"] = materialize.GetValueSecretStruct(newSaslUsername)
			options["SASL PASSWORD"] = materialize.GetIdentifierSchemaStruct(newSaslPassword)
		} else {
			addResetOption("SASL USERNAME")
			addResetOption("SASL PASSWORD")
		}
	}

	if d.HasChange("aws_connection") {
		_, newAwsConn := d.GetChange("aws_connection")
		if newAwsConn == nil || len(newAwsConn.([]interface{})) == 0 {
			addResetOption("AWS CONNECTION")
		} else {
			options["AWS CONNECTION"] = materialize.GetIdentifierSchemaStruct(newAwsConn)
		}
	}

	// Apply the changes
	if len(options) > 0 || len(resetOptions) > 0 {
		if err := b.Alter(options, resetOptions, false, validate); err != nil {
			// Reverting to old values if alter fails
			if d.HasChange("security_protocol") {
				oldValue, _ := d.GetChange("security_protocol")
				d.Set("security_protocol", oldValue)
			}
			if d.HasChange("ssl_certificate_authority") {
				oldValue, _ := d.GetChange("ssl_certificate_authority")
				d.Set("ssl_certificate_authority", oldValue)
			}
			if d.HasChange("ssl_certificate") {
				oldValue, _ := d.GetChange("ssl_certificate")
				d.Set("ssl_certificate", oldValue)
			}
			if d.HasChange("ssl_key") {
				oldValue, _ := d.GetChange("ssl_key")
				d.Set("ssl_key", oldValue)
			}
			if d.HasChange("sasl_mechanisms") {
				oldValue, _ := d.GetChange("sasl_mechanisms")
				d.Set("sasl_mechanisms", oldValue)
			}
			if d.HasChange("sasl_username") {
				oldValue, _ := d.GetChange("sasl_username")
				d.Set("sasl_username", oldValue)
			}
			if d.HasChange("sasl_password") {
				oldValue, _ := d.GetChange("sasl_password")
				d.Set("sasl_password", oldValue)
			}
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ssh_tunnel") {
		oldTunnel, newTunnel := d.GetChange("ssh_tunnel")
		b := materialize.NewConnection(metaDb, o)
		if newTunnel == nil || len(newTunnel.([]interface{})) == 0 {
			if err := b.AlterDrop([]string{"SSH TUNNEL"}, validate); err != nil {
				d.Set("ssh_tunnel", oldTunnel)
				return diag.FromErr(err)
			}
		} else {
			tunnel := materialize.GetIdentifierSchemaStruct(newTunnel)
			options := map[string]interface{}{
				"SSH TUNNEL": tunnel,
			}
			if err := b.Alter(options, nil, false, validate); err != nil {
				d.Set("ssh_tunnel", oldTunnel)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(metaDb, o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return connectionRead(ctx, d, meta)
}
