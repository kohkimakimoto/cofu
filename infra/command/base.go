package command

import (
	"fmt"
	"github.com/kohkimakimoto/cofu/infra/util"
)

type BaseCommand struct {
	osFamily  string
	osRelease string
}

func (c *BaseCommand) SetOSFamily(v string) {
	c.osFamily = v
}

func (c *BaseCommand) SetOSRelease(v string) {
	c.osRelease = v
}

func (c *BaseCommand) OSFamily() string {
	return c.osFamily
}

func (c *BaseCommand) OSRelease() string {
	return c.osRelease
}

func (c *BaseCommand) OSInfo() string {
	return c.osFamily + c.osRelease
}

func (c *BaseCommand) String() string {
	return "BaseCommand"
}

func (c *BaseCommand) DefaultServiceProvider() string {
	return ServiceInit
}

func (c *BaseCommand) CheckFileIsFile(file string) string {
	return "test -f " + util.ShellEscape(file)
}

func (c *BaseCommand) CheckFileIsDirectory(file string) string {
	return "test -d " + util.ShellEscape(file)
}

func (c *BaseCommand) CheckFileIsPipe(file string) string {
	return "test -p " + util.ShellEscape(file)
}

func (c *BaseCommand) CheckFileIsSocket(file string) string {
	return "test -S " + util.ShellEscape(file)
}

func (c *BaseCommand) CheckFileIsBlockDevice(file string) string {
	return "test -b " + util.ShellEscape(file)
}

func (c *BaseCommand) CheckFileIsCharacterDevice(file string) string {
	return "test -c " + util.ShellEscape(file)
}

func (c *BaseCommand) CheckFileIsSymlink(file string) string {
	return "test -L " + util.ShellEscape(file)
}

func (c *BaseCommand) GetFileMode(file string) string {
	return "stat -c %a " + util.ShellEscape(file)
}

func (c *BaseCommand) GetFileOwnerUser(file string) string {
	return "stat -c %U " + util.ShellEscape(file)
}

func (c *BaseCommand) GetFileOwnerGroup(file string) string {
	return "stat -c %G " + util.ShellEscape(file)
}

func (c *BaseCommand) CheckFileIsLinkedTo(link, target string) string {
	return `test x"$(readlink ` + util.ShellEscape(link) + `)" = x"` + util.ShellEscape(target) + `"`
}

func (c *BaseCommand) CheckFileIsLink(link string) string {
	return "test -L " + util.ShellEscape(link)
}

func (c *BaseCommand) GetFileLinkTarget(link string) string {
	return "readlink " + util.ShellEscape(link)
}

func (c *BaseCommand) ChangeFileMode(file string, mode string, recursive bool) string {
	option := ""
	if recursive {
		option = "-R"
	}

	return fmt.Sprintf("chmod %s %s %s ", option, mode, util.ShellEscape(file))
}

func (c *BaseCommand) ChangeFileOwner(file string, owner string, group string, recursive bool) string {
	option := ""
	if recursive {
		option = "-R"
	}

	if group != "" {
		owner = fmt.Sprintf("%s:%s", owner, group)
	}

	return fmt.Sprintf("chown %s %s %s ", option, owner, util.ShellEscape(file))
}

func (c *BaseCommand) ChangeFileGroup(file string, group string, recursive bool) string {
	option := ""
	if recursive {
		option = "-R"
	}

	return fmt.Sprintf("chgrp %s %s %s ", option, group, util.ShellEscape(file))
}

func (c *BaseCommand) CreateFileAsDirectory(file string) string {
	return "mkdir -p " + util.ShellEscape(file)
}

func (c *BaseCommand) LinkFileTo(link string, target string, force bool) string {
	option := "-s"
	if force {
		option += "f"
	}

	return fmt.Sprintf("ln %s %s %s", option, util.ShellEscape(target), util.ShellEscape(link))
}

func (c *BaseCommand) RemoveFile(file string) string {
	return "rm -rf " + util.ShellEscape(file)
}

func (c *BaseCommand) MoveFile(src, dest string) string {
	return fmt.Sprintf("mv %s %s", util.ShellEscape(src), util.ShellEscape(dest))
}

func (c *BaseCommand) CheckPackageIsInstalled(packagename string, version string) string {
	panic("Unsupported method")
}

func (c *BaseCommand) GetPackageVersion(packagename string, option string) string {
	panic("Unsupported method")
}

func (c *BaseCommand) InstallPackage(packagename string, version string, option string) string {
	panic("Unsupported method")
}

func (c *BaseCommand) RemovePackage(packagename string, option string) string {
	panic("Unsupported method")
}

// service

func (c *BaseCommand) CheckServiceIsRunningUnderProvider(provider string, name string) string {
	switch provider {
	case ServiceInit:
		return "service " + util.ShellEscape(name) + " status"
	case ServiceSystemd:
		return "systemctl is-active " + util.ShellEscape(name)
	default:
		panic("Unsupported provider: " + provider)
	}
}

func (c *BaseCommand) CheckServiceIsEnabledUnderProvider(provider string, name string) string {
	switch provider {
	case ServiceInit:
		return "chkconfig --list " + util.ShellEscape(name) + " | grep 3:on"
	case ServiceSystemd:
		return "systemctl --quiet is-enabled " + util.ShellEscape(name)
	default:
		panic("Unsupported provider: " + provider)
	}
}

func (c *BaseCommand) StartServiceUnderProvider(provider string, name string) string {
	switch provider {
	case ServiceInit:
		return "service " + util.ShellEscape(name) + " start"
	case ServiceSystemd:
		return "systemctl start " + util.ShellEscape(name)
	default:
		panic("Unsupported provider: " + provider)
	}
}

func (c *BaseCommand) StopServiceUnderProvider(provider string, name string) string {
	switch provider {
	case ServiceInit:
		return "service " + util.ShellEscape(name) + " stop"
	case ServiceSystemd:
		return "systemctl stop " + util.ShellEscape(name)
	default:
		panic("Unsupported provider: " + provider)
	}
}

func (c *BaseCommand) RestartServiceUnderProvider(provider string, name string) string {
	switch provider {
	case ServiceInit:
		return "service " + util.ShellEscape(name) + " restart"
	case ServiceSystemd:
		return "systemctl restart " + util.ShellEscape(name)
	default:
		panic("Unsupported provider: " + provider)
	}
}

func (c *BaseCommand) ReloadServiceUnderProvider(provider string, name string) string {
	switch provider {
	case ServiceInit:
		return "service " + util.ShellEscape(name) + " reload"
	case ServiceSystemd:
		return "systemctl reload " + util.ShellEscape(name)
	default:
		panic("Unsupported provider: " + provider)
	}
}

func (c *BaseCommand) EnableServiceUnderProvider(provider string, name string) string {
	switch provider {
	case ServiceInit:
		return "chkconfig " + util.ShellEscape(name) + " on"
	case ServiceSystemd:
		return "systemctl enable " + util.ShellEscape(name)
	default:
		panic("Unsupported provider: " + provider)
	}
}

func (c *BaseCommand) DisableServiceUnderProvider(provider string, name string) string {
	switch provider {
	case ServiceInit:
		return "chkconfig " + util.ShellEscape(name) + " off"
	case ServiceSystemd:
		return "systemctl disable " + util.ShellEscape(name)
	default:
		panic("Unsupported provider: " + provider)
	}
}

// group

func (c *BaseCommand) CheckGroupExists(group string) string {
	return fmt.Sprintf("getent group %s", util.ShellEscape(group))
}

func (c *BaseCommand) CheckGroupHasGid(group string, gid string) string {
	return fmt.Sprintf("getent group %s | cut -f 3 -d ':' | grep -w -- %s", util.ShellEscape(group), util.ShellEscape(gid))
}

func (c *BaseCommand) GetGroupGid(group string) string {
	return fmt.Sprintf("getent group %s | cut -f 3 -d ':'", util.ShellEscape(group))
}

func (c *BaseCommand) UpdateGroupGid(group string, gid string) string {
	return fmt.Sprintf("groupmod -g %s %s", util.ShellEscape(gid), util.ShellEscape(group))
}

func (c *BaseCommand) AddGroup(group string, options map[string]string) string {
	cmd := "groupadd"

	if gid, ok := options["gid"]; ok {
		cmd += " -g " + util.ShellEscape(gid)
	}

	cmd += " " + util.ShellEscape(group)
	return cmd
}

func (c *BaseCommand) CheckUserExists(user string) string {
	return "id " + util.ShellEscape(user)
}

func (c *BaseCommand) CheckUserBelongsToGroup(user, group string) string {
	return "id " + util.ShellEscape(user) + " | sed 's/ context=.*//g' | cut -f 4 -d '=' | grep -- " + util.ShellEscape(group)
}

func (c *BaseCommand) CheckUserBelongsToPrimaryGroup(user, group string) string {
	return "id -gn " + util.ShellEscape(user) + " | grep ^" + group + "$"
}

func (c *BaseCommand) CheckUserHasUid(user, uid string) string {
	regexp := "^uid=" + uid + "("
	return "id " + util.ShellEscape(user) + " | grep -- " + regexp
}

func (c *BaseCommand) CheckUserHasHomeDirectory(user, pathToHome string) string {
	return "getent passwd " + util.ShellEscape(user) + " | cut -f 6 -d ':' | grep -w -- " + util.ShellEscape(pathToHome)
}

func (c *BaseCommand) CheckUserHasLoginShell(user, pathToShell string) string {
	return "getent passwd " + util.ShellEscape(user) + " | cut -f 7 -d ':' | grep -w -- " + util.ShellEscape(pathToShell)
}

func (c *BaseCommand) CheckUserHasAuthorizedKey(user, key string) string {
	panic("Unsupported method")
}

func (c *BaseCommand) GetUserUid(user string) string {
	return "id -u " + util.ShellEscape(user)
}

func (c *BaseCommand) GetUserGid(user string) string {
	return "id -g " + util.ShellEscape(user)
}

func (c *BaseCommand) GetUserHomeDirectory(user string) string {
	return "getent passwd " + util.ShellEscape(user) + " | cut -f 6 -d ':'"
}

func (c *BaseCommand) GetUserLoginShell(user string) string {
	return "getent passwd " + util.ShellEscape(user) + " | cut -f 7 -d ':'"
}

func (c *BaseCommand) UpdateUserHomeDirectory(user, dir string) string {
	return "usermod -d " + util.ShellEscape(dir) + " " + util.ShellEscape(user)
}

func (c *BaseCommand) UpdateUserLoginShell(user, shell string) string {
	return "usermod -s " + util.ShellEscape(shell) + " " + util.ShellEscape(user)
}

func (c *BaseCommand) UpdateUserUid(user, uid string) string {
	return "usermod -u " + util.ShellEscape(uid) + " " + util.ShellEscape(user)
}

func (c *BaseCommand) UpdateUserGid(user, gid string) string {
	return "usermod -g " + util.ShellEscape(gid) + " " + util.ShellEscape(user)
}

func (c *BaseCommand) AddUser(user string, options map[string]string) string {
	cmd := "useradd"

	if gid, ok := options["gid"]; ok && gid != "" {
		cmd += " -g " + util.ShellEscape(gid)
	}

	if homeDirectory, ok := options["home_directory"]; ok && homeDirectory != "" {
		cmd += " -d " + util.ShellEscape(homeDirectory)
	}

	if password, ok := options["password"]; ok && password != "" {
		cmd += " -p " + util.ShellEscape(password)
	}

	if shell, ok := options["shell"]; ok && shell != "" {
		cmd += " -s " + util.ShellEscape(shell)
	}

	if create_home, ok := options["create_home"]; ok && create_home != "" {
		cmd += " -m"
	}

	if system_user, ok := options["system_user"]; ok && system_user != "" {
		cmd += " -r"
	}

	if uid, ok := options["uid"]; ok && uid != "" {
		cmd += " -u " + util.ShellEscape(uid)
	}

	cmd += " " + util.ShellEscape(user)

	return cmd
}

func (c *BaseCommand) UpdateUserEncryptedPassword(user, encryptedPassword string) string {
	return `echo " + ` + user + `:` + encryptedPassword + `" | chpasswd -e`
}
func (c *BaseCommand) GetUserEncryptedPassword(user string) string {
	return "getent shadow " + util.ShellEscape(user) + " | cut -f 2 -d ':'"
}
