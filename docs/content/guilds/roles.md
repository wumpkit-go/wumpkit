---
title: "Roles & Permissions"
weight: 40
---

# Roles & Permissions

Roles define member permissions, display colors, and sidebar groupings inside a server.

wumpkit provides methods to create and manage roles, assign them to members, and work with permission bitfields.

## Creating a Role

Use `CreateRole` to create a new role:

```go
package main

import (
	"context"
	"log"

	"github.com/wumpkit/wumpkit"
)

func createStaffRole(ctx context.Context, client *wumpkit.Client, guildID wumpkit.Snowflake) {
	staffPerms := wumpkit.PermissionViewAuditLog |
		wumpkit.PermissionKickMembers |
		wumpkit.PermissionModerateMembers |
		wumpkit.PermissionManageMessages

	role, err := client.CreateRole(ctx, guildID, wumpkit.CreateRoleOptions{
		Name:        "Staff Moderator",
		Permissions: &staffPerms,
		Colors: wumpkit.RoleColors{
			PrimaryColor: wumpkit.ColorBlurple,
		},
		Hoist:       true, // display separately in server member list
		Mentionable: true,
		Reason:      "Created moderator role for new season",
	})
	if err != nil {
		log.Printf("failed to create role: %v", err)
		return
	}

	log.Printf("created role %s (ID: %s)", role.Name, role.ID)
}
```

## Assigning and Removing Member Roles

Assign or remove roles directly without needing to fetch or overwrite the whole member object:

```go
// add a role to a member
err := client.AddMemberRole(ctx, guildID, userID, roleID, wumpkit.AddMemberRoleOptions{
	Reason: "Promoted to moderator",
})

// remove a role from a member
err = client.RemoveMemberRole(ctx, guildID, userID, roleID, wumpkit.RemoveMemberRoleOptions{
	Reason: "Left staff team",
})
```

## Modifying Existing Roles

Use `ModifyRole` to rename a role, change its color, or update permissions:

```go
newPerms := wumpkit.PermissionAdministrator
name := "Senior Moderator"

_, err := client.ModifyRole(ctx, guildID, roleID, wumpkit.ModifyRoleOptions{
	Name:        &name,
	Permissions: &newPerms,
	Reason:      "Promoted role permissions",
})
```

## Deleting a Role

Use `DeleteRole` to delete a role:

```go
err := client.DeleteRole(ctx, guildID, roleID, wumpkit.DeleteRoleOptions{
	Reason: "Role no longer needed",
})
```

## Working With Permission Bitfields

Discord permissions are 64-bit integer bitmasks (`wumpkit.Permissions`). wumpkit provides helper methods to work with them cleanly:

```go
// combine permissions using bitwise OR
perms := wumpkit.PermissionSendMessages | wumpkit.PermissionEmbedLinks | wumpkit.PermissionAttachFiles

// check if specific permissions are granted
if perms.Has(wumpkit.PermissionEmbedLinks) {
	log.Println("user can send embeds")
}

// find which permissions are missing
required := wumpkit.PermissionManageGuild | wumpkit.PermissionManageChannels
missing := perms.Missing(required)
if missing != 0 {
	log.Printf("missing required permissions: %v", missing.Names())
}

// get list of human-readable permission names
for _, name := range perms.Names() {
	log.Printf("has permission: %s", name)
}
```
