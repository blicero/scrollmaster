// /home/krylon/go/src/github.com/blicero/scrollmaster/database/04_database_search_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-12 19:04:49 krylon>

package database

import (
	"slices"
	"testing"
	"time"

	"github.com/blicero/scrollmaster/model"
)

func TestSearchExecute(t *testing.T) {
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
} // func TestSearchExecute(t *testing.T)

func TestSearchAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var s = model.Search{
		Timestamp: time.Now(),
		Query: model.SearchQuery{
			Hosts: []int64{hosts[1].ID},
		},
		Results: make([]int64, 0, recordCnt),
	}
	var q = make(chan model.Record)

	go tdb.RecordSearch(&s.Query, q)

	for r := range q {
		s.Results = append(s.Results, r.ID)
	}

	if len(s.Results) != recordCnt {
		t.Fatalf("Unexpected number of results from Search: want %d got %d",
			recordCnt,
			len(s.Results))
	}

	var err error

	if err = tdb.SearchAdd(&s); err != nil {
		t.Fatalf("Error adding Search to database: %s", err.Error())
	} else if s.ID == 0 {
		t.Fatal("SearchAdd did not set the Search ID")
	}

	var results []model.Record

	if results, err = tdb.SearchGetResults(s.ID, 0, -1); err != nil {
		t.Fatalf("Error getting search results: %s", err.Error())
	} else if len(results) != recordCnt {
		t.Fatalf("Unexpected number of results: got %d want %d",
			len(results),
			recordCnt)
	}

	var idlist = make([]int64, len(s.Results))

	for idx, r := range results {
		idlist[idx] = r.ID
	}

	slices.Sort(s.Results)
	slices.Sort(idlist)

	if !slices.Equal(s.Results, idlist) {
		t.Errorf("Unexpected list of Records return from Search:\n%#v\n%#v\n",
			idlist,
			s.Results)
	}
} // func TestSearchAdd(t *testing.T)
