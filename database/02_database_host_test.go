// /home/krylon/go/src/github.com/blicero/scrollmaster/database/02_database_host_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 14. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-14 20:47:37 krylon>

package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/blicero/scrollmaster/model"
)

const hostCnt = 8

var hosts = make([]model.Host, 0, hostCnt)

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
		}

		hosts = append(hosts, h)
	}
} // func TestHostAdd(t *testing.T)
