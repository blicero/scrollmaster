// /home/krylon/go/src/github.com/blicero/scrollmaster/database/qdb.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-12 19:23:43 krylon>

package database

import "github.com/blicero/scrollmaster/database/query"

var qdb = map[query.ID]string{
	query.HostAdd:            "INSERT INTO host (name, last_seen) VALUES (?, ?) RETURNING id",
	query.HostGetByName:      "SELECT id, last_seen FROM host WHERE name = ?",
	query.HostGetByID:        "SELECT name, last_seen FROM host WHERE id = ?",
	query.HostGetAll:         "SELECT id, name, last_seen FROM host ORDER BY name",
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
	query.RecordGetRecent: `
SELECT
    id,
    host_id,
    stamp,
    source,
    message
FROM record
ORDER BY stamp DESC
LIMIT ?
`,
	query.RecordGetSources: `
SELECT
    source,
    COUNT(source) AS cnt
FROM record
GROUP BY source
ORDER BY source`,
	query.SearchAdd: `
INSERT INTO search (timestamp, query, results, cnt)
            VALUES (        ?,     ?,       ?,   ?)
RETURNING id
`,
	query.SearchGetByID: `
SELECT
    timestamp,
    query,
    results,
    cnt
FROM search
WHERE id = ?
`,
	query.SearchDelete: "DELETE FROM search WHERE id = ?",
	query.SearchGetResults: `
WITH idlist (id) AS (
	SELECT value
	FROM search s, json_each(s.results)
	WHERE s.id = ?
)

SELECT
        i.id,
	r.host_id,
	r.stamp,
	r.source,
	r.message
FROM idlist i
INNER JOIN record r ON i.id = r.id
ORDER BY r.stamp
LIMIT ?
OFFSET ?
`,
	query.SearchGetAllID:       "SELECT id, cnt FROM search",
	query.SearchGetResultCount: "SELECT cnt FROM search WHERE id = ?",
}
