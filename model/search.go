// /home/krylon/go/src/github.com/blicero/scrollmaster/model/search.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-09 20:35:55 krylon>

package model

import (
	"slices"
	"time"
)

// SearchQuery wraps the parameters for searching the log.
type SearchQuery struct {
	Hosts   []int64     `json:"hosts"`
	Sources []string    `json:"sources"`
	Period  []time.Time `json:"period"`
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

	return true
} // func (q *SearchQuery) Match(r *Record) bool