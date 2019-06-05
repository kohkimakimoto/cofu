package cofu

import (
	"fmt"
	"github.com/kardianos/osext"
	"os"
)

var BinPath string

func init() {
	binPath, err := osext.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		binPath = "cofu"
	}
	BinPath = binPath
}
