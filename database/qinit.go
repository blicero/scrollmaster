// /home/krylon/go/src/github.com/blicero/scrollmaster/database/qinit.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-14 19:14:00 krylon>

package database

var qInit = []string{
	`
CREATE TABLE host (
	id		INTEGER PRIMARY KEY,
        name            TEXT UNIQUE NOT NULL,
        last_seen       INTEGER NOT NULL DEFAULT 0
) STRICT
`,
	"CREATE UNIQUE INDEX host_name_idx ON host (name)",

	`
CREATE TABLE record (
	id		INTEGER PRIMARY KEY,
        host_id		INTEGER NOT NULL,
        stamp           INTEGER NOT NULL DEFAULT 0,
        source          TEXT NOT NULL,
        message         TEXT NOT NULL,
        FOREIGN KEY (host_id) REFERENCES host (id)
            ON UPDATE RESTRICT
            ON DELETE CASCADE
) STRICT
`,
	"CREATE INDEX record_host_idx ON record (host_id)",
	"CREATE INDEX record_stamp_idx ON record (stamp)",
	"CREATE INDEX record_source_idx ON record (source)",
}
