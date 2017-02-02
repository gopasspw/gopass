package action

import "github.com/urfave/cli"

// Fsck checks the store integrity
func (s *Action) Fsck(c *cli.Context) error {
	check := c.Bool("check")
	force := c.Bool("force")
	if check {
		force = false
	}
	return s.Store.Fsck(check, force)
}
