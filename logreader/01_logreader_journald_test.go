// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/01_logreader_journald_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 18. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-18 21:43:42 krylon>

package logreader

import "testing"

var rdr LogReader

func TestOpenReader(t *testing.T) {
	var err error

	if rdr, err = DefaultOpener("test"); err != nil {
		rdr = nil
		t.Fatalf("Failed to open LogReader: %s",
			err.Error())
	} else if rdr == nil {
		t.Fatal("DefaultOpener returned no error, but no LogReader, either")
	}
} // func TestOpenReader(t *testing.T)
