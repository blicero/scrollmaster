// /home/krylon/go/src/github.com/blicero/scrollmaster/database/01_database_create_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 14. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-14 19:12:21 krylon>

package database

import (
	"database/sql"
	"testing"

	"github.com/blicero/scrollmaster/common"
)

var tdb *Database

func TestCreateDatabase(t *testing.T) {
	var err error

	if tdb, err = Open(common.DbPath); err != nil {
		tdb = nil
		t.Fatalf("Error opening database: %s", err.Error())
	}
} // func TestCreateDatabase(t *testing.T)

func TestQueryPrepare(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var (
		err  error
		stmt *sql.Stmt
	)

	for qid := range qdb {
		if stmt, err = tdb.getQuery(qid); err != nil {
			t.Errorf("Failed to prepare query %s: %s",
				qid,
				err.Error())
		} else if stmt == nil {
			t.Errorf("getQuery(%s) did not return an error, but no statement handle, either",
				qid)
		}
	}
} // func TestQueryPrepare(t *testing.T)
