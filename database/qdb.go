// /home/krylon/go/src/github.com/blicero/scrollmaster/database/qdb.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-03 21:16:00 krylon>

package database

import "github.com/blicero/scrollmaster/database/query"

var qdb = map[query.ID]string{
	query.HostAdd:            "INSERT INTO host (name, last_seen) VALUES (?, ?) RETURNING id",
	query.HostGetByName:      "SELECT id, last_seen FROM host WHERE name = ?",
	query.HostGetByID:        "SELECT name, last_seen FROM host WHERE id = ?",
	query.HostGetAll:         "SELECT id, name, last_seen FROM host",
	query.HostUpdateLastSeen: "UPDATE host SET last_seen = ? WHERE id = ?",
	query.RecordAdd: `
INSERT INTO record (host_id, stamp, source, message, checksum)
            VALUES (      ?,     ?,      ?,       ?,        ?)
RETURNING id
`,
	query.RecordGetByHost: `
SELECT
    id,
    stamp,
    source,
    message
FROM record
WHERE host_id = ?
ORDER BY stamp DESC
LIMIT ?
`,
	query.RecordGetByPeriod: `
SELECT
    id,
    host_id,
    stamp,
    source,
    message
FROM record
WHERE stamp BETWEEN ? AND ?
ORDER BY stamp
`,
	query.RecordGetMostRecent: `
SELECT COALESCE(MAX(stamp), 0)
FROM record
WHERE host_id = ?
`,
	query.RecordCheckExist: "SELECT COUNT(id) FROM record WHERE checksum = ?",
}
