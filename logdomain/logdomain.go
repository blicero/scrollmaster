// /home/krylon/go/src/github.com/blicero/scrollmaster/logdomain/logdomain.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-13 20:41:04 krylon>

// Package logdomain provides symbolic constants to identify the various
// pieces of the application that need to do logging.
package logdomain

//go:generate stringer -type=ID

// ID is an id...
type ID uint8

// These constants represent the pieces of the application that need to log stuff.
const (
	Common ID = iota
	Client
	Database
)

// AllDomains returns a slice of all the valid values for ID.
func AllDomains() []ID {
	return []ID{
		Common,
		Client,
		Database,
	}
} // func AllDomains() []ID
