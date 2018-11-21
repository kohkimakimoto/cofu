package command

type UbuntuV1604Command struct {
	UbuntuCommand
}

func (c *UbuntuV1604Command) String() string {
	return "UbuntuV1604Command"
}

func (c *UbuntuV1604Command) DefaultServiceProvider() string {
	return ServiceSystemd
}