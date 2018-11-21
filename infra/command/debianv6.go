package command

// see https://github.com/hnakamur/cofu/blob/support_debian_and_ubuntu_in_specinfra_way/infra/command/debianv6.go

import "github.com/kohkimakimoto/cofu/infra/util"

type DebianV6Command struct {
	DebianCommand
}

func (c *DebianV6Command) String() string {
	return "DebianV6Command"
}

func (c *DebianV6Command) CheckServiceIsEnabledUnderProvider(provider string, name string) string {
	// Until everything uses Upstart, this needs an OR.
	level := "3"
	return "ls /etc/rc" + level + ".d/ | grep -- '^S.." + util.ShellEscape(name) + `$' || grep '^\s*start on' /etc/init/` + util.ShellEscape(name) + ".conf"
}

func (c *DebianV6Command) EnableServiceUnderProvider(provider string, name string) string {
	return "update-rc.d " + util.ShellEscape(name) + " defaults"
}

func (c *DebianV6Command) DisableServiceUnderProvider(provider string, name string) string {
	return "update-rc.d " + util.ShellEscape(name) + " remove"
}
