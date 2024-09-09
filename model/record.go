// /home/krylon/go/src/github.com/blicero/scrollmaster/model/record.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-09 20:35:33 krylon>

package model

import (
	"fmt"
	"time"

	"github.com/blicero/scrollmaster/common"
)

// Record is one record from a system log.
type Record struct {
	ID      int64
	HostID  int64
	Time    time.Time
	Source  string
	Message string
	cksum   string
}

// Checksum returns a hash value of the record that (hopefully) uniquely identifies it.
func (r *Record) Checksum() string {
	if r.cksum != "" {
		return r.cksum
	}

	var raw string = fmt.Sprintf("%d##%d##%s##%s##%s",
		r.ID,
		r.HostID,
		r.Time.Format(common.TimestampFormatSubSecond),
		r.Source,
		r.Message)
	var (
		err    error
		result string
	)

	if result, err = common.GetChecksum([]byte(raw)); err != nil {
		panic(err)
	}

	r.cksum = result

	return result
} // func (r *Record) Checksum() string

// RecordSlice is a slice of Records that can be sorted.
type RecordSlice []Record

func (r RecordSlice) Len() int {
	return len(r)
} // func (r RecordSlice) Len() int

func (r RecordSlice) Less(i, j int) bool {
	return r[i].Time.Before(r[j].Time)
} // func (r RecordSlice) Less(i, j int) bool

func (r RecordSlice) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
} // func (r RecordSlice) Swap(i, j int)
