// /home/krylon/go/src/github.com/blicero/scrollmaster/database/query/id.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-03 21:15:37 krylon>

//go:generate stringer -type=ID

// Package query defines symbolic constants to reference database queries.
package query

// ID Represents a database query.
type ID uint8

const (
	HostAdd ID = iota
	HostGetByName
	HostGetByID
	HostGetAll
	HostUpdateLastSeen
	RecordAdd
	RecordGetByHost
	RecordGetByPeriod
	RecordGetMostRecent
	RecordCheckExist
)
