package command

type LinuxCommand struct {
	BaseCommand
}

func (c *LinuxCommand) String() string {
	return "LinuxCommand"
}
