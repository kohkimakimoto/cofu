package command

type RedhatV7Command struct {
	RedhatCommand
}

func (c *RedhatV7Command) String() string {
	return "RedhatV7Command"
}

func (c *RedhatV7Command) DefaultServiceProvider() string {
	return ServiceSystemd
}
