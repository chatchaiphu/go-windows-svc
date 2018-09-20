package fixbcst

import (
	"golang.org/x/sys/windows/svc/debug"
)

// if setup returns an error, the service doesn't start
func setup(wl debug.Log, svcName string) (server, error) {
	var s server

	s.winlog = wl

	// Note: any logging here goes to Windows App Log
	// I suggest you setup local logging
	//s.winlog.Info(1, fmt.Sprintf("%s: setup ()", svcName))

	// read configuration
	// configure more logging

	return s, nil
}
