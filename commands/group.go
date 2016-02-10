package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type GroupCommand struct {
	Groups []string
}

func (g *GroupCommand) configure(app *kingpin.Application) {
	grpCmd := app.Command("groups", "manipulate groups")
	grpCmd.Command("list", "Show all known groups").Action(g.list)

	add := grpCmd.Command("add", "Add a new group gonab").Action(g.add)
	add.Arg("group", "Group name to add").Required().StringsVar(&g.Groups)

	dis := grpCmd.Command("disable", "Disable a group").Action(g.disable)
	dis.Arg("group", "Group name to disable").Required().StringsVar(&g.Groups)
}

func (g *GroupCommand) list(c *kingpin.ParseContext) error {
	_, dbh := commonInit()

	groups, err := dbh.GetAllGroups()
	if err != nil {
		return fmt.Errorf("Error getting group list: %v", err)
	}

	for _, g := range groups {
		fmt.Printf("Name: %s, First %d, Last: %d\n", g.Name, g.First, g.Last)
	}
	return nil
}

func (g *GroupCommand) add(c *kingpin.ParseContext) error {
	_, dbh := commonInit()

	for _, group := range g.Groups {
		dbgroup, err := dbh.AddGroup(group)
		if err != nil {
			return fmt.Errorf("Error adding group %s: %v", group, err)
		}
		fmt.Printf("Added group %s\n", dbgroup.Name)
	}
	return nil
}

func (g *GroupCommand) disable(c *kingpin.ParseContext) error {
	_, dbh := commonInit()

	for _, group := range g.Groups {
		err := dbh.DisableGroup(group)
		if err != nil {
			return fmt.Errorf("Error disabling group %s: %v", group, err)
		}
		fmt.Printf("Disabled group %s\n", group)
	}
	return nil
}

//TODO: add delete group
