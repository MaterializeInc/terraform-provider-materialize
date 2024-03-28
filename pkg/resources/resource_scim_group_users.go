package resources

import (
	"context"
	"fmt"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var ScimGroupUsersSchema = map[string]*schema.Schema{
	"group_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the SCIM group.",
	},
	"users": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "The set of user IDs to assign to the SCIM group.",
	},
}

func SCIM2GroupUsers() *schema.Resource {
	return &schema.Resource{
		CreateContext: scimGroupUsersCreate,
		ReadContext:   scimGroupUsersRead,
		UpdateContext: scimGroupUsersUpdate,
		DeleteContext: scimGroupUsersDelete,
		Schema:        ScimGroupUsersSchema,
		Description:   "The materialize_scim_group_users resource allows managing users within a SCIM group in Frontegg.",
	}
}

func scimGroupUsersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	userIDs := expandStringSet(d.Get("users").(*schema.Set))

	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	err = frontegg.AddUsersToGroup(ctx, client, groupID, userIDs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error adding users to SCIM group: %s", err))
	}

	d.SetId(groupID)
	return scimGroupUsersRead(ctx, d, meta)
}

func scimGroupUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)

	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	group, err := frontegg.GetSCIMGroupByID(ctx, client, groupID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching SCIM group: %s", err))
	}

	var userIDs []interface{}
	for _, user := range group.Users {
		userIDs = append(userIDs, user.ID)
	}

	if err := d.Set("users", userIDs); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func scimGroupUsersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	oldUserIDs := expandStringSet(d.Get("users").(*schema.Set))

	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	// Get the current users assigned to the group
	group, err := frontegg.GetSCIMGroupByID(ctx, client, groupID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching SCIM group: %s", err))
	}

	// Extract user IDs from group.Users
	var existingUserIDs []string
	for _, user := range group.Users {
		existingUserIDs = append(existingUserIDs, user.ID)
	}

	// Determine the user IDs that need to be removed
	var removedUserIDs []string
	for _, userID := range existingUserIDs {
		if !stringSetContains(oldUserIDs, userID) {
			removedUserIDs = append(removedUserIDs, userID)
		}
	}

	// Check if removedUserIDs is not empty
	if len(removedUserIDs) > 0 {
		// Remove the users that were removed from the group only if removedUserIDs is not empty
		err = frontegg.RemoveUsersFromGroup(ctx, client, groupID, removedUserIDs)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error removing users from SCIM group: %s", err))
		}
	}

	// Determine the new users to be added by filtering out existing users
	var newUserIDs []string
	for _, newUserID := range oldUserIDs {
		if !stringSetContains(existingUserIDs, newUserID) {
			newUserIDs = append(newUserIDs, newUserID)
		}
	}

	log.Printf("[DEBUG] Adding users to SCIM group: %v", newUserIDs)
	// Check if newUserIDs is not empty
	if len(newUserIDs) > 0 {
		// Add the new users to the group only if newUserIDs is not empty
		err = frontegg.AddUsersToGroup(ctx, client, groupID, newUserIDs)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error adding users to SCIM group: %s", err))
		}
	}

	return scimGroupUsersRead(ctx, d, meta)
}

func scimGroupUsersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	userIDs := expandStringSet(d.Get("users").(*schema.Set))

	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	err = frontegg.RemoveUsersFromGroup(ctx, client, groupID, userIDs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error removing users from SCIM group: %s", err))
	}

	// Forcing deletion by setting an empty ID
	d.SetId("")

	return nil
}

// Helper function to check if a string set contains a specific string
func stringSetContains(set []string, str string) bool {
	for _, s := range set {
		if s == str {
			return true
		}
	}
	return false
}
