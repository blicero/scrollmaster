// /home/krylon/go/src/github.com/blicero/scrollmaster/database/02_database_host_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 14. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-15 19:43:57 krylon>

package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/model"
)

const hostCnt = 8

var hosts = make(map[int64]model.Host, hostCnt)

func TestHostAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	type testCase struct {
		host        model.Host
		expectError bool
	}

	var (
		err       error
		testCases = make([]testCase, hostCnt*2)
	)

	for i := 0; i < hostCnt; i++ {
		var name = fmt.Sprintf("Host%02d", i+1)
		testCases[i] = testCase{
			host: model.Host{
				Name:     name,
				LastSeen: time.Now().Add(time.Minute * -7200),
			},
		}
		testCases[i+hostCnt] = testCase{
			host: model.Host{
				Name: name,
			},
			expectError: true,
		}
	}

	for _, c := range testCases {
		h := c.host
		if err = tdb.HostAdd(&h); err != nil {
			if !c.expectError {
				t.Errorf("Failed to add Host %s: %s",
					h.Name,
					err.Error())
			}
		} else if c.expectError {
			t.Errorf("Adding Host %s to the database should have failed, but it did not.",
				h.Name)
		} else if h.ID == 0 {
			t.Errorf("Host %s was added to the database, yet its ID is 0",
				h.Name)
		} else {
			hosts[h.ID] = h
		}
	}
} // func TestHostAdd(t *testing.T)

func TestHostGetByName(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	for _, h1 := range hosts {
		var (
			err error
			h2  *model.Host
		)

		if h2, err = tdb.HostGetByName(h1.Name); err != nil {
			t.Errorf("Error looking up Host %s in database: %s",
				h1.Name,
				err.Error())
		} else if h2 == nil {
			t.Errorf("Did not find Host %s in database",
				h1.Name)
		}
	}
} // func TestHostGetByName(t *testing.T)

func TestHostSetLastSeen(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var (
		stamp = time.Unix(time.Now().Unix(), 0)
		err   error
	)

	tdb.Begin() // nolint: errcheck

	for _, h := range hosts {
		if err = tdb.HostUpdateLastSeen(&h, stamp); err != nil {
			tdb.Rollback() // nolint: errcheck
			t.Fatalf("Error setting LastSeen on Host %s: %s",
				h.Name,
				err.Error())
		} else {
			hosts[h.ID] = h
		}
	}

	tdb.Commit() // nolint: errcheck

	var (
		dbHosts []model.Host
		ok      bool
	)

	if dbHosts, err = tdb.HostGetAll(); err != nil {
		t.Fatalf("Failed to get all Hosts from database: %s",
			err.Error())
	}

	for _, h1 := range dbHosts {
		if _, ok = hosts[h1.ID]; !ok {
			t.Errorf("Missing host %s", h1.Name)
		} else if !h1.LastSeen.Equal(stamp) {
			t.Errorf("Unexpected LastSeen stamp on Host %s: %s (expected %s)",
				h1.Name,
				h1.LastSeen.Format(common.TimestampFormat),
				stamp.Format(common.TimestampFormat))
		}
	}
} // func TestHostSetLastSeen(t *testing.T)
