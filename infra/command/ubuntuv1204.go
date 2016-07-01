package command

import "github.com/kohkimakimoto/cofu/infra/util"

type UbuntuV1204Command struct {
	UbuntuCommand
}

func (c *UbuntuV1204Command) String() string {
	return "UbuntuV1204Command"
}

func (c *UbuntuV1204Command) CheckServiceIsEnabledUnderProvider(provider string, name string) string {
	return "ls /etc/rc2.d/S[0-9][0-9]" + util.ShellEscape(name)
}

func (c *UbuntuV1204Command) EnableServiceUnderProvider(provider string, name string) string {
	return "/usr/lib/insserv/insserv " + util.ShellEscape(name)
}

func (c *UbuntuV1204Command) DisableServiceUnderProvider(provider string, name string) string {
	return "/usr/lib/insserv/insserv -r " + util.ShellEscape(name)
}
