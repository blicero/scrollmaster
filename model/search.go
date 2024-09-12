// /home/krylon/go/src/github.com/blicero/scrollmaster/model/search.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-12 11:11:11 krylon>

package model

import (
	"regexp"
	"slices"
	"time"
)

// SearchQuery wraps the parameters for searching the log.
type SearchQuery struct {
	Hosts   []int64          `json:"hosts"`
	Sources []string         `json:"sources"`
	Period  []time.Time      `json:"period"`
	Terms   []*regexp.Regexp `json:"terms"`
}

// Match checks if a given Record r matches the criteria of the SearchQuery.
func (q *SearchQuery) Match(r *Record) bool {
	if len(q.Period) == 2 && (r.Time.Before(q.Period[0]) || r.Time.After(q.Period[1])) {
		return false
	} else if len(q.Sources) > 0 && !slices.Contains(q.Sources, r.Source) {
		return false
	} else if len(q.Hosts) > 0 && !slices.Contains(q.Hosts, r.HostID) {
		return false
	}

	var match bool

	for _, pat := range q.Terms {
		if pat.MatchString(r.Message) {
			match = true
			break
		}
	}

	if len(q.Terms) > 0 && !match {
		return false
	}

	return true
} // func (q *SearchQuery) Match(r *Record) bool

// Search represents a search, including the Query and the list of IDs
// it returned.
type Search struct {
	ID        int64
	Timestamp time.Time
	Query     SearchQuery
	Results   []int64
	Count     int64
}
