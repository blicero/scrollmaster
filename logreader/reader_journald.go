// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/reader_journald.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-15 20:59:52 krylon>

//go:build linux

package logreader

import (
	"log"

	"github.com/blicero/scrollmaster/model"
	"github.com/coreos/go-systemd/sdjournal"
)

type JournaldReader struct {
	log     *log.Logger
	queue   <-chan model.Record
	journal *sdjournal.Journal
}
