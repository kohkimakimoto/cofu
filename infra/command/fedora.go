package command

type FedoraCommand struct {
	RedhatCommand
}

func (c *FedoraCommand) String() string {
	return "FedoraCommand"
}
