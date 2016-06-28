package cofu

import "fmt"

type Notification struct {
	DefinedInResource  *Resource
	Action             string
	TargetResourceDesc string
	Timing             string
}

func (n *Notification) Delayed() bool {
	if n.Timing == "delay" || n.Timing == "delayed" {
		return true
	} else {
		return false
	}
}

func (n *Notification) Immediately() bool {
	if n.Timing == "immediately" {
		return true
	} else {
		return false
	}
}

func (n *Notification) Validate() error {
	if n.Timing == "delay" || n.Timing == "delayed" || n.Timing == "immediately" {
		return nil
	}

	return fmt.Errorf("'%s' is not valid notification timing. (Valid option is delayed or immediately)", n.Timing)
}

func (n *Notification) Run() error {
	targetResource := n.DefinedInResource.App.FindOneResource(n.TargetResourceDesc)
	if targetResource == nil {
		return fmt.Errorf("Not found target resource '%s'", n.TargetResourceDesc)
	}

	return targetResource.Run(n.Action)
}
