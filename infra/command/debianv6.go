package command

import "github.com/kohkimakimoto/cofu/infra/util"

type DebianV6Command struct {
	DebianCommand
}

func (c *DebianV6Command) String() string {
	return "DebianV6Command"
}

func (c *DebianV6Command) CheckServiceIsEnabledUnderProvider(provider string, name string) string {
	return "ls /etc/rc2.d/S[0-9][0-9]" + util.ShellEscape(name)
}

func (c *DebianV6Command) EnableServiceUnderProvider(provider string, name string) string {
	return "insserv " + util.ShellEscape(name)
}

func (c *DebianV6Command) DisableServiceUnderProvider(provider string, name string) string {
	return "insserv -r " + util.ShellEscape(name)
}
