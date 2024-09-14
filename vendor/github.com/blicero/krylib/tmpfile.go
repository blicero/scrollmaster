// -*- mode: go; coding: utf-8; -*-
//  Time-stamp: <2020-06-01 00:59:27 krylon>
// /home/krylon/go/src/krylib/tmpfile.go

// Created on 31. 05. 2020 by Benjamin Walkenhorst
// (c) 2020 Benjamin Walkenhorst

package krylib

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// RandommFilename generates filename that is randomized
// file name that should be random  enough to avoid collisions.
//
// CAVEAT: The randomness in these filenames is intended to avoid
// naming collisions by *accident*. They do *not* offer any kind of protection
// against a malicious adversary. Needless to say, they are also totally
// useless for anything that even /remotely/ anything to do with cryptography.
func RandomFilename() string {
	var now = time.Now()
	var str = strconv.Itoa(int(now.Unix()))

	var indices = rand.Perm(len(str))

	var sb strings.Builder

	for _, idx := range indices {
		var digit = str[idx]
		sb.WriteString(string(digit)) // nolint: errcheck,gosec
	}

	return sb.String()
} // func RandomFilename() string

// Local Variables:  //
// compile-command: "go generate && go vet && go build -v -p 16 && gometalinter && go test -v" //
// End: //
