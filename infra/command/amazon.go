package command

type AmazonCommand struct {
	RedhatCommand
}

func (c *AmazonCommand) String() string {
	return "AmazonCommand"
}
