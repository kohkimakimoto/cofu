package command

import (
	"github.com/kohkimakimoto/cofu/infra/util"
)

type DarwinCommand struct {
	BaseCommand
}

func (c *DarwinCommand) String() string {
	return "DarwinCommand"
}

func (c *DarwinCommand) GetFileMode(file string) string {
	return "stat -f%Lp " + util.ShellEscape(file)
}

func (c *DarwinCommand) GetFileOwnerUser(file string) string {
	return "stat -f %Su " + util.ShellEscape(file)
}

func (c *DarwinCommand) GetFileOwnerGroup(file string) string {
	return "stat -f %Sg #{escape(file)}" + util.ShellEscape(file)
}

func (c *DarwinCommand) CheckFileIsLinkedTo(link, target string) string {
	return "stat -f %Y " + util.ShellEscape(link) + " | grep -- " + util.ShellEscape(target)
}
