package command

import "github.com/kohkimakimoto/cofu/infra/util"

type DebianCommand struct {
	LinuxCommand
}

func (c *DebianCommand) String() string {
	return "DebianCommand"
}

func (c *DebianCommand) CheckPackageIsInstalled(packagename string, version string) string {
	var cmd string
	if version != "" {
		cmd = "dpkg-query -f '${Status} ${Version}' -W " + util.ShellEscape(packagename) + " | grep -E '^(install|hold) ok installed " + util.ShellEscape(version) + "$'"
	} else {
		cmd = "dpkg-query -f '${Status}' -W " + util.ShellEscape(packagename) + " | grep -E '^(install|hold) ok installed$'"
	}

	return cmd
}

func (c *DebianCommand) GetPackageVersion(packagename string, option string) string {
	return "dpkg-query -f '${Status} ${Version}' -W " + packagename + " | sed -n 's/^install ok installed //p'"
}

func (c *DebianCommand) InstallPackage(packagename string, version string, option string) string {
	var fullPackage string
	if version != "" {
		fullPackage = packagename + "=" + version
	} else {
		fullPackage = packagename
	}
	return "DEBIAN_FRONTEND='noninteractive' apt-get -y -o Dpkg::Options::='--force-confdef' -o Dpkg::Options::='--force-confold' " + option + " install " + fullPackage
}

func (c *DebianCommand) RemovePackage(packagename string, option string) string {
	return "DEBIAN_FRONTEND='noninteractive' apt-get -y " + option + " remove " + packagename
}