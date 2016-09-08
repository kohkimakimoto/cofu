package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
)

var ResourceTypes = []*cofu.ResourceType{
	Directory,
	Execute,
	File,
	Git,
	Group,
	Link,
	LuaFunction,
	SoftwarePackage,
	Service,
	RemoteFile,
	Template,
	User,
}