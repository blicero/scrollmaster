// /home/krylon/go/src/github.com/blicero/scrollmaster/database/qinit.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-12 11:08:59 krylon>

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
        checksum        TEXT UNIQUE NOT NULL,
        FOREIGN KEY (host_id) REFERENCES host (id)
            ON UPDATE RESTRICT
            ON DELETE CASCADE,
        CHECK (host_id > 0)
) STRICT
`,
	"CREATE INDEX record_host_idx ON record (host_id)",
	"CREATE INDEX record_stamp_idx ON record (stamp)",
	"CREATE INDEX record_source_idx ON record (source)",
	"CREATE UNIQUE INDEX record_ck_idx ON record (checksum)",

	`
CREATE view readable AS
SELECT
    r.id,
    h.name,
    datetime(r.stamp, 'unixepoch') AS stamp,
    r.source,
    r.message
FROM record r
INNER JOIN host h ON r.host_id = h.id`,

	`
CREATE TABLE search (
    id			INTEGER PRIMARY KEY,
    timestamp		INTEGER NOT NULL,
    query               TEXT NOT NULL,
    results             TEXT NOT NULL DEFAULT '[]',
    cnt                 INTEGER NOT NULL DEFAULT 0,
    CHECK (json_valid(query) > 0 AND json_valid(results) > 0),
    CHECK (cnt >= 0)
) STRICT
`,
	"CREATE INDEX search_time_idx ON search (timestamp)",
}
