// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/logreader.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-18 21:36:33 krylon>

// Package logreader implements the reading/parsing of log files or journald's log.
package logreader

import (
	"time"

	"github.com/blicero/scrollmaster/model"
)

// LogReader defines the interface for several implementations that access system logs as plain
// old syslog files or systemd-journald journals.
type LogReader interface {
	Init() error
	Close() error
	ReadFrom(begin time.Time, queue chan<- model.Record)
}

// ReaderOpener is a function to open a LogReader.
type ReaderOpener func(path string) (LogReader, error)

// DefaultOpener is the function to call to open a LogReader.
var DefaultOpener ReaderOpener
