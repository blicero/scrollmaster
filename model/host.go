// /home/krylon/go/src/github.com/blicero/scrollmaster/model/host.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-13 21:00:53 krylon>

package model

import "time"

type Host struct {
	ID       int64
	Name     string
	LastSeen time.Time
}