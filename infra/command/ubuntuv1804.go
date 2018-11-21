package command

type UbuntuV1804Command struct {
	UbuntuCommand
}

func (c *UbuntuV1804Command) String() string {
	return "UbuntuV1804Command"
}

func (c *UbuntuV1804Command) DefaultServiceProvider() string {
	return ServiceSystemd
}
