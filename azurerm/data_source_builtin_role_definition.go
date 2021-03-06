package azurerm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceArmBuiltInRoleDefinition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmBuiltInRoleDefinitionRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			// Computed
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"not_actions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"data_actions": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
						"not_data_actions": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"assignable_scopes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceArmBuiltInRoleDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).roleDefinitionsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	if name == "VirtualMachineContributor" {
		name = "Virtual Machine Contributor"
	}
	filter := fmt.Sprintf("roleName eq '%s'", name)
	roleDefinitions, err := client.List(ctx, "", filter)
	if err != nil {
		return fmt.Errorf("Error loading Role Definition List: %+v", err)
	}
	if len(roleDefinitions.Values()) != 1 {
		return fmt.Errorf("Error loading Role Definition List: could not find role '%s'", name)
	}

	roleDefinitionId := *roleDefinitions.Values()[0].ID

	d.SetId(roleDefinitionId)

	role, err := client.GetByID(ctx, roleDefinitionId)
	if err != nil {
		return fmt.Errorf("Error loading Role Definition: %+v", err)
	}

	if props := role.RoleDefinitionProperties; props != nil {
		d.Set("name", props.RoleName)
		d.Set("description", props.Description)
		d.Set("type", props.RoleType)

		permissions := flattenRoleDefinitionDataSourcePermissions(props.Permissions)
		if err := d.Set("permissions", permissions); err != nil {
			return err
		}

		assignableScopes := flattenRoleDefinitionDataSourceAssignableScopes(props.AssignableScopes)
		if err := d.Set("assignable_scopes", assignableScopes); err != nil {
			return err
		}
	}

	return nil
}

func flattenRoleDefinitionDataSourcePermissions(input *[]authorization.Permission) []interface{} {
	permissions := make([]interface{}, 0)
	if input == nil {
		return permissions
	}

	for _, permission := range *input {
		output := make(map[string]interface{}, 0)

		actions := make([]string, 0)
		if permission.Actions != nil {
			for _, action := range *permission.Actions {
				actions = append(actions, action)
			}
		}
		output["actions"] = actions

		dataActions := make([]interface{}, 0)
		if permission.DataActions != nil {
			for _, dataAction := range *permission.DataActions {
				dataActions = append(dataActions, dataAction)
			}
		}
		output["data_actions"] = schema.NewSet(schema.HashString, dataActions)

		notActions := make([]string, 0)
		if permission.NotActions != nil {
			for _, action := range *permission.NotActions {
				notActions = append(notActions, action)
			}
		}
		output["not_actions"] = notActions

		notDataActions := make([]interface{}, 0)
		if permission.NotDataActions != nil {
			for _, dataAction := range *permission.NotDataActions {
				notDataActions = append(notDataActions, dataAction)
			}
		}
		output["not_data_actions"] = schema.NewSet(schema.HashString, notDataActions)

		permissions = append(permissions, output)
	}

	return permissions
}

func flattenRoleDefinitionDataSourceAssignableScopes(input *[]string) []interface{} {
	scopes := make([]interface{}, 0)
	if input == nil {
		return scopes
	}

	for _, scope := range *input {
		scopes = append(scopes, scope)
	}

	return scopes
}
