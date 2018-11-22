package command

type DebianV9Command struct {
	DebianCommand
}

func (c *DebianV9Command) String() string {
	return "DebianV9Command"
}

func (c *DebianV9Command) DefaultServiceProvider() string {
	return ServiceSystemd
}
