package action

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/tree/simple"
	"github.com/urfave/cli"
)

// MountRemove removes an existing mount
func (s *Action) MountRemove(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return fmt.Errorf("usage: gopass mount remove [alias]")
	}
	if err := s.Store.RemoveMount(c.Args()[0]); err != nil {
		color.Yellow("Failed to remove mount: %s", err)
	}
	if err := s.Store.Config().Save(); err != nil {
		return err
	}

	color.Green("Password Store %s umounted", c.Args()[0])
	return nil
}

// MountsPrint prints all existing mounts
func (s *Action) MountsPrint(c *cli.Context) error {
	if len(s.Store.Mounts()) < 1 {
		fmt.Println("No mounts")
		return nil
	}
	root := simple.New(color.GreenString(fmt.Sprintf("gopass (%s)", s.Store.Path())))
	for alias, path := range s.Store.Mounts() {
		if err := root.AddMount(alias, path); err != nil {
			fmt.Printf("Failed to add mount: %s\n", err)
		}
	}
	fmt.Fprintln(color.Output, root.Format(0))
	return nil
}

// MountsComplete will print a list of existings mount points for bash
// completion
func (s *Action) MountsComplete(*cli.Context) {
	for alias := range s.Store.Mounts() {
		fmt.Println(alias)
	}
}

// MountAdd adds a new mount
func (s *Action) MountAdd(c *cli.Context) error {
	if len(c.Args()) != 2 {
		return fmt.Errorf("usage: gopass mount add [alias] [local path]")
	}
	keys := make([]string, 0, 1)
	if k := c.String("init"); k != "" {
		keys = append(keys, k)
	}
	if err := s.Store.AddMount(c.Args()[0], c.Args()[1], keys...); err != nil {
		return err
	}
	if err := s.Store.Config().Save(); err != nil {
		return err
	}

	color.Green("Mounted %s as %s", c.Args()[0], c.Args()[1])
	return nil
}
