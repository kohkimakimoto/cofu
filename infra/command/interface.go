package command

const (
	ServiceInit    = "init"
	ServiceSystemd = "systemd"
)

type CommandFactory interface {
	// misc
	String() string
	SetOSFamily(string)
	SetOSRelease(string)
	OSFamily() string
	OSRelease() string
	OSInfo() string
	DefaultServiceProvider() string

	// file
	CheckFileIsFile(file string) string
	CheckFileIsDirectory(file string) string
	CheckFileIsPipe(file string) string
	CheckFileIsSocket(file string) string
	CheckFileIsBlockDevice(file string) string
	CheckFileIsCharacterDevice(file string) string
	CheckFileIsSymlink(file string) string
	GetFileMode(file string) string
	GetFileOwnerUser(file string) string
	GetFileOwnerGroup(file string) string
	CheckFileIsLinkedTo(link, target string) string
	CheckFileIsLink(link string) string
	GetFileLinkTarget(link string) string
	ChangeFileMode(file string, mode string, recursive bool) string
	ChangeFileOwner(file string, owner string, group string, recursive bool) string
	ChangeFileGroup(file string, group string, recursive bool) string
	CreateFileAsDirectory(file string) string
	LinkFileTo(link string, target string, force bool) string
	RemoveFile(file string) string
	MoveFile(src, dest string) string

	// group
	CheckGroupExists(group string) string
	CheckGroupHasGid(group string, gid string) string
	GetGroupGid(group string) string
	UpdateGroupGid(group string, gid string) string
	AddGroup(group string, options map[string]string) string

	// user
	CheckUserExists(user string) string
	CheckUserBelongsToGroup(user, group string) string
	CheckUserBelongsToPrimaryGroup(user, group string) string
	CheckUserHasUid(user, uid string) string
	CheckUserHasHomeDirectory(user, pathToHome string) string
	CheckUserHasLoginShell(user, pathToShell string) string
	CheckUserHasAuthorizedKey(user, key string) string

	GetUserUid(user string) string
	GetUserGid(user string) string
	GetUserHomeDirectory(user string) string
	GetUserLoginShell(user string) string
	UpdateUserHomeDirectory(user, dir string) string
	UpdateUserLoginShell(user, shell string) string
	UpdateUserUid(user, uid string) string
	UpdateUserGid(user, gid string) string
	AddUser(user string, options map[string]string) string
	UpdateUserEncryptedPassword(user, encryptedPassword string) string
	GetUserEncryptedPassword(user string) string

	// package
	CheckPackageIsInstalled(packagename string, version string) string
	GetPackageVersion(packagename string, option string) string
	InstallPackage(packagename string, version string, option string) string
	RemovePackage(packagename string, option string) string

	// service with provider
	CheckServiceIsRunningUnderProvider(provider string, name string) string
	CheckServiceIsEnabledUnderProvider(provider string, name string) string
	StartServiceUnderProvider(provider string, name string) string
	StopServiceUnderProvider(provider string, name string) string
	RestartServiceUnderProvider(provider string, name string) string
	ReloadServiceUnderProvider(provider string, name string) string
	EnableServiceUnderProvider(provider string, name string) string
	DisableServiceUnderProvider(provider string, name string) string
}
