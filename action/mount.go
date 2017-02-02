package action

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/tree"
	"github.com/urfave/cli"
)

// MountRemove removes an existing mount
func (s *Action) MountRemove(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return fmt.Errorf("usage: gopass mount remove [alias]")
	}
	if err := s.Store.RemoveMount(c.Args()[0]); err != nil {
		return err
	}
	if err := writeConfig(s.Store); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Password Store %s umounted", c.Args()[0]))
	return nil
}

// MountsPrint prints all existing mounts
func (s *Action) MountsPrint(c *cli.Context) error {
	if len(s.Store.Mount) < 1 {
		fmt.Println("No mounts")
		return nil
	}
	root := tree.New(color.GreenString(fmt.Sprintf("gopass (%s)", s.Store.Path)))
	for alias, path := range s.Store.Mount {
		if err := root.AddMount(alias, path); err != nil {
			fmt.Printf("Failed to add mount: %s\n", err)
		}
	}
	fmt.Println(root.Format())
	return nil
}

// MountsComplete will print a list of existings mount points for bash
// completion
func (s *Action) MountsComplete(*cli.Context) {
	for alias := range s.Store.Mount {
		fmt.Println(alias)
	}
	return
}

// MountAdd adds a new mount
func (s *Action) MountAdd(c *cli.Context) error {
	if len(c.Args()) != 2 {
		return fmt.Errorf("usage: gopass mount add [alias] [path]")
	}
	keys := make([]string, 0, 1)
	if k := c.String("init"); k != "" {
		keys = append(keys, k)
	}
	if err := s.Store.AddMount(c.Args()[0], c.Args()[1], keys...); err != nil {
		return err
	}
	if err := writeConfig(s.Store); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Mounted %s as %s", c.Args()[0], c.Args()[1]))
	return nil
}
