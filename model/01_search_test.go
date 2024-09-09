// /home/krylon/go/src/github.com/blicero/scrollmaster/model/01_search_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-09 20:39:55 krylon>

package model

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

func qEqual(q1, q2 SearchQuery) bool {
	if !slices.Equal(q1.Hosts, q2.Hosts) {
		return false
	} else if !slices.Equal(q1.Sources, q2.Sources) {
		return false
	} else if len(q1.Period) != len(q2.Period) {
		return false
	} else if len(q1.Period) == 2 && (!q1.Period[0].Equal(q2.Period[0]) || !q1.Period[1].Equal(q2.Period[1])) {
		return false
	}

	return true
} // func qEqual(q1, q2 model.SearchQuery) bool

func TestSearchSerialize(t *testing.T) {
	type testCase struct {
		q           SearchQuery
		expectError bool
	}

	var testCases = []testCase{
		{
			q: SearchQuery{
				Hosts:   []int64{1, 2, 3},
				Sources: []string{"slime"},
				Period: []time.Time{
					time.Now().Add(time.Second * -60),
					time.Now(),
				},
			},
		},
	}

	for _, c := range testCases {
		var (
			err error
			buf []byte
			re  SearchQuery
		)

		if buf, err = json.Marshal(&c.q); err != nil {
			if !c.expectError {
				t.Errorf("Failed to marshal query: %s",
					err.Error())
				continue
			}
		} else if err = json.Unmarshal(buf, &re); err != nil {
			if !c.expectError {
				t.Errorf("Failed to unmarshal marshaled query: %s",
					err.Error())
				continue
			}
		} else if !qEqual(c.q, re) {
			t.Errorf("Unmarshaled query does not equal original query:\n%#v\n\n%#v",
				c.q,
				re)
		}
	}
}
