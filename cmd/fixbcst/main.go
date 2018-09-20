// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"fixe_bcst/app/fixbcst"

	"github.com/pkg/errors"
)

// This is the name you will use for the NET START command
const svcName = "FixBcst"

// This is the name that will appear in the Services control panel
const svcNameLong = "FixBcst Service"

func svcLauncher() error {

	err := fixbcst.Run(elog, svcName)
	if err != nil {
		return errors.Wrap(err, "fixbcst.run")
	}

	return nil
}
