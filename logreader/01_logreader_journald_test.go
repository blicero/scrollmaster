// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/01_logreader_journald_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 18. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-16 20:55:24 krylon>

package logreader

import (
	"testing"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/model"
)

func TestReaderOpen(t *testing.T) {
	var err error

	if rdr, err = DefaultOpener("test"); err != nil {
		rdr = nil
		t.Fatalf("Failed to open LogReader: %s",
			err.Error())
	} else if rdr == nil {
		t.Fatal("DefaultOpener returned no error, but no LogReader, either")
	} else if err = rdr.Init(); err != nil {
		rdr = nil
		t.Fatalf("Error initializing LogReader: %s",
			err.Error())
	}
} // func TestReaderOpen(t *testing.T)

func TestReaderRead(t *testing.T) {
	if rdr == nil {
		t.SkipNow()
	}

	var (
		cnt   int
		queue chan model.Record
		begin time.Time
	)

	queue = make(chan model.Record)
	begin = time.Now().Add(time.Hour * -2)

	go rdr.ReadFrom(begin, 0, queue)

	for record := range queue {
		cnt++
		t.Logf("Record #%4d: %s / %s / %s\n",
			cnt,
			record.Time.Format(common.TimestampFormatSubSecond),
			record.Source,
			record.Message)
	}
} // func TestReaderRead(t *testing.T)

func TestReaderClose(t *testing.T) {
	if rdr == nil {
		t.SkipNow()
	}

	var err error

	if err = rdr.Close(); err != nil {
		t.Errorf("Error closing LogReader: %s",
			err.Error())
	}

	rdr = nil
} // func TestReaderClose(t *testing.T)
