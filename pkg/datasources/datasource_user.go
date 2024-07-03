package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func User() *schema.Resource {
	return &schema.Resource{
		ReadContext: userDataSourceRead,

		Description: `The user data source allows you to retrieve information about a user in your Materialize organization.`,

		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The email address of the user to retrieve.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique (UUID) identifier of the user.",
			},
			"verified": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user's email address has been verified.",
			},
			"metadata": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Additional metadata associated with the user.",
			},
			"auth_provider": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The authentication provider for the user.",
			},
		},
	}
}

func userDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*utils.ProviderMeta).Frontegg

	email := d.Get("email").(string)

	params := frontegg.QueryUsersParams{
		Email: email,
		Limit: 1,
	}

	users, err := frontegg.GetUsers(ctx, client, params)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(users) == 0 {
		return diag.Errorf("No user found with email: %s", email)
	}

	user := users[0]

	d.SetId(user.ID)
	d.Set("email", user.Email)
	d.Set("verified", user.Verified)
	d.Set("metadata", user.Metadata)
	d.Set("auth_provider", user.Provider)

	return nil
}
