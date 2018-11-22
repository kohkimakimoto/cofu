package command

type UbuntuCommand struct {
	DebianCommand
}

func (c *UbuntuCommand) String() string {
	return "UbuntuCommand"
}
