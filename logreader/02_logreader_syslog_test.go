// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/02_logreader_syslog_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 16. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-16 21:03:49 krylon>

package logreader

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/model"
)

func TestSyslogreaderOpen(t *testing.T) {
	var (
		err  error
		path []string
		dirh *os.File
	)

	if dirh, err = os.Open("testdata"); err != nil {
		rdr = nil
		t.Fatalf("Error opening directory testdata: %s",
			err.Error())
	}

	defer dirh.Close()

	if path, err = dirh.Readdirnames(-1); err != nil {
		t.Fatalf("Failed to read list of logfiles: %s",
			err.Error())
	}

	for idx, logfile := range path {
		path[idx] = filepath.Join("testdata", logfile)
	}

	if rdr, err = CreateSyslogReader(path...); err != nil {
		rdr = nil
		t.Fatalf("Failed to create SyslogReader: %s",
			err.Error())
	} else if err = rdr.Init(); err != nil {
		t.Fatalf("Failed to open logfiles: %s",
			err.Error())
	}
} // func TestSyslogreaderOpen(t *testing.T)

func TestSyslogReaderRead(t *testing.T) {
	if rdr == nil {
		t.SkipNow()
	}

	var (
		cnt   int
		queue chan model.Record
		begin time.Time
	)

	queue = make(chan model.Record)
	// begin = time.Now().Add(time.Hour * -2)

	go rdr.ReadFrom(begin, 0, queue)

	for record := range queue {
		cnt++
		t.Logf("Record #%4d: %s / %s / %s\n",
			cnt,
			record.Time.Format(common.TimestampFormatSubSecond),
			record.Source,
			record.Message)
	}

	if cnt == 0 {
		t.Errorf("SyslogReader returned %d records, we did expect a little more than that.",
			cnt)
	}
} // func TestSyslogReaderRead(t *testing.T)

func TestSyslogReaderClose(t *testing.T) {
	if rdr == nil {
		t.SkipNow()
	}

	var err error

	if err = rdr.Close(); err != nil {
		t.Errorf("Error closing LogReader: %s",
			err.Error())
	}

	rdr = nil
} // func TestSyslogReaderClose(t *testing.T)
