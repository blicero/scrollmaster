// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/02_logreader_syslog_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 16. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-16 19:51:05 krylon>

package logreader

import (
	"os"
	"path/filepath"
	"testing"
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
}
