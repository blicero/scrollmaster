// /home/krylon/go/src/github.com/blicero/scrollmaster/database/04_database_search_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-09 20:22:30 krylon>

package database

import (
	"testing"

	"github.com/blicero/scrollmaster/model"
)

func TestSearch(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	type testCase struct {
		q   model.SearchQuery
		cnt int64
	}

	var testCases = []testCase{
		{
			q:   model.SearchQuery{},
			cnt: int64(len(hosts) * recordCnt),
		},
		{
			q: model.SearchQuery{
				Hosts: []int64{hosts[1].ID},
			},
			cnt: recordCnt,
		},
	}

	for _, c := range testCases {
		var (
			q   = make(chan model.Record)
			cnt int64
		)

		go tdb.RecordSearch(&c.q, q)

		for range q {
			cnt++
		}

		if cnt != c.cnt {
			t.Errorf("Expected %d records, got %d",
				c.cnt,
				cnt)
		}
	}
} // func TestSearch(t *testing.T)
