package infra

import (
	"github.com/kohkimakimoto/cofu/infra/backend"
	"github.com/kohkimakimoto/cofu/infra/command"
	"github.com/kohkimakimoto/cofu/infra/detector"
)

type Infra struct {
	commandFactory command.CommandFactory
	cmd            *backend.Cmd
	detectors      []detector.Detector
}

func New() *Infra {
	i := &Infra{
		cmd:       backend.NewCmd("/bin/sh"),
		detectors: detector.DefaultDetectors,
	}

	return i
}

func (i *Infra) Command() command.CommandFactory {
	if i.commandFactory == nil {
		for _, detector := range i.detectors {
			if commandFactory := detector(i.cmd); commandFactory != nil {
				i.commandFactory = commandFactory
				return i.commandFactory
			}
		}

		panic("Couldn't detect os.")
	}

	return i.commandFactory
}

func (i *Infra) RunCommand(command string) *backend.CommandResult {
	return i.cmd.RunCommand(command)
}

func (i *Infra) BuildCommand(command string, option *backend.CommandOption) string {
	return i.cmd.BuildCommand(command, option)
}
