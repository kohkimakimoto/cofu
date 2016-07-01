package command

import "github.com/kohkimakimoto/cofu/infra/util"

type DebianCommand struct {
	LinuxCommand
}

func (c *DebianCommand) String() string {
	return "DebianCommand"
}

func (c *DebianCommand) CheckPackageIsInstalled(packagename string, version string) string {
	cmd := "test `dpkg-query -W -f='${Version}' " + util.ShellEscape(packagename) + "`"
	if version != "" {
		cmd = "test x`dpkg-query -W -f='${Version}' " + util.ShellEscape(packagename) + "` = x" + util.ShellEscape(version)
	}

	return cmd
}

func (c *DebianCommand) GetPackageVersion(packagename string, option string) string {
	return "dpkg-query -W -f='${Version}' " + util.ShellEscape(packagename)
}

func (c *DebianCommand) InstallPackage(packagename string, version string, option string) string {
	var fullPackage string
	if version != "" {
		fullPackage = packagename + "=" + version
	} else {
		fullPackage = packagename
	}

	return "apt-get -y " + option + " install " + util.ShellEscape(fullPackage)
}

func (c *DebianCommand) RemovePackage(packagename string, option string) string {
	return "apt-get -y " + option + " autoremove " + util.ShellEscape(packagename)
}
