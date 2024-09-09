// /home/krylon/go/src/github.com/blicero/scrollmaster/model/host.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-09 20:34:58 krylon>

package model

import (
	"regexp"
	"time"
)

var namePat = regexp.MustCompile("^([^.]+)")

// Host represents a system that we gather log records from.
type Host struct {
	ID       int64
	Name     string
	LastSeen time.Time
}

// NameShort returns the Host's name without any domain name.
// If the Host's Name is already short, it is returned as-is.
func (h *Host) NameShort() string {
	if m := namePat.FindStringSubmatch(h.Name); len(m) > 1 {
		return m[1]
	}

	return h.Name
} // func (h *Host) NameShort() string
