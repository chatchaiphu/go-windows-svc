package fixbcst

import (
	"github.com/pkg/errors"
	"golang.org/x/sys/windows/svc/debug"
)

// Run launches the service
func Run(wl debug.Log, svcName string) error {

	s, err := setup(wl, svcName)
	if err != nil {
		return errors.Wrap(err, "setup")
	}

	// Your service should be launched as a GO routine
	go FixBcstApp(s)

	return nil
}
