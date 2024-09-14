// /home/krylon/go/src/krylib/errors.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 09. 2019 by Benjamin Walkenhorst
// (c) 2019 Benjamin Walkenhorst
// Time-stamp: <2019-09-13 21:45:24 krylon>

package krylib

import "errors"

// ErrNotImplemented Indicates that some functionality is not implemented, yet.
var ErrNotImplemented error = errors.New("not implemented") // nolint: golint

// ErrInvalidValue Indicates that some value is invalid.
var ErrInvalidValue error = errors.New("invalid value")

// Local Variables:  //
// compile-command: "go generate && go vet && go build -v -p 16 && gometalinter && go test -v" //
// End: //
