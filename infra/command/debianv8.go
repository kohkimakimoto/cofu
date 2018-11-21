package command

type DebianV8Command struct {
	DebianCommand
}

func (c *DebianV8Command) String() string {
	return "DebianV8Command"
}

func (c *DebianV8Command) DefaultServiceProvider() string {
	return ServiceSystemd
}
