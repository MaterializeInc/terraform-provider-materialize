package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	// Import necessary packages
)

var dataSourceSCIMGroupsSchema = map[string]*schema.Schema{
	"groups": {
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the group. This is a unique identifier for the group. ",
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name of the group.",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The description of the group.",
				},
				"metadata": {
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    true,
					Description: "The metadata of the group.",
				},
				"roles": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The ID of the role. This is a unique identifier for the role.",
							},
							"key": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The key of the role.",
							},
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The name of the role.",
							},
							"description": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The description of the role.",
							},
							"is_default": {
								Type:        schema.TypeBool,
								Computed:    true,
								Description: "Indicates whether the role is the default role.",
							},
						},
					},
				},
				"users": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The ID of the user. This is a unique identifier for the user.",
							},
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The name of the user.",
							},
							"email": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The email of the user.",
							},
						},
					},
				},
				"managed_by": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the user who manages the group.",
				},
			},
		},
	},
}

// Group represents the structure of a group in the response.
type ScimGroup struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Metadata    string     `json:"metadata"`
	Roles       []ScimRole `json:"roles"`
	Users       []ScimUser `json:"users"`
	ManagedBy   string     `json:"managedBy"`
}

// Role represents the structure of a role within a group.
type ScimRole struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

// User represents the structure of a user within a group.
type ScimUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SCIMGroupsResponse represents the overall structure of the response from the SCIM groups API.
type SCIMGroupsResponse struct {
	Groups []ScimGroup `json:"groups"`
}

// SCIMGroups data source function
func SCIMGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIMGroupsRead,
		Schema:      dataSourceSCIMGroupsSchema,
	}
}

// Read function for SCIM groups data source
func dataSourceSCIMGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	endpoint := fmt.Sprintf("%s/frontegg/identity/resources/groups/v1?_groupsRelations=rolesAndUsers", client.Endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Add authorization header to the request
	req.Header.Add("Authorization", "Bearer "+client.Token)

	// Send the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error reading SCIM groups: status %d, response: %s", resp.StatusCode, sb.String())
	}

	// Decode the JSON response
	var groupsResponse SCIMGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&groupsResponse); err != nil {
		return diag.FromErr(err)
	}

	// Map the response to the schema
	if err := d.Set("groups", flattenGroups(groupsResponse.Groups)); err != nil {
		return diag.FromErr(err)
	}

	// Set the ID of the data source
	d.SetId("scim_groups")

	return nil
}

// Helper function to flatten the groups data
func flattenGroups(groups []ScimGroup) []interface{} {
	var flattenedGroups []interface{}
	for _, group := range groups {
		flattenedGroup := map[string]interface{}{
			"id":          group.ID,
			"name":        group.Name,
			"description": group.Description,
			"metadata":    group.Metadata,
			"roles":       flattenRoles(group.Roles),
			"users":       flattenUsers(group.Users),
			"managed_by":  group.ManagedBy,
		}
		flattenedGroups = append(flattenedGroups, flattenedGroup)
	}
	return flattenedGroups
}

// Helper function to flatten the roles data
func flattenRoles(roles []ScimRole) []interface{} {
	var flattenedRoles []interface{}
	for _, role := range roles {
		flattenedRole := map[string]interface{}{
			"id":          role.ID,
			"key":         role.Key,
			"name":        role.Name,
			"description": role.Description,
			"is_default":  role.IsDefault,
		}
		flattenedRoles = append(flattenedRoles, flattenedRole)
	}
	return flattenedRoles
}

// Helper function to flatten the users data
func flattenUsers(users []ScimUser) []interface{} {
	var flattenedUsers []interface{}
	for _, user := range users {
		flattenedUser := map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		}
		flattenedUsers = append(flattenedUsers, flattenedUser)
	}
	return flattenedUsers
}
