package command

import (
	"fmt"
	"github.com/kohkimakimoto/cofu/infra/util"
)

type RedhatCommand struct {
	LinuxCommand
}

func (c *RedhatCommand) String() string {
	return "RedhatCommand"
}

func (c *RedhatCommand) CheckPackageIsInstalled(packagename string, version string) string {
	cmd := "rpm -q " + util.ShellEscape(packagename)
	if version != "" {
		cmd = fmt.Sprintf("%s | grep -w -- %s-%s", cmd, packagename, version)
	}

	return cmd
}

func (c *RedhatCommand) GetPackageVersion(packagename string, option string) string {
	return "rpm -q --qf '%{VERSION}-%{RELEASE}' " + packagename
}

func (c *RedhatCommand) InstallPackage(packagename string, version string, option string) string {
	var fullPackage string
	if version != "" {
		fullPackage = packagename + "-" + version
	} else {
		fullPackage = packagename
	}

	return "yum -y " + option + " install " + fullPackage
}

func (c *RedhatCommand) RemovePackage(packagename string, option string) string {
	return "yum -y " + option + " remove " + packagename
}
