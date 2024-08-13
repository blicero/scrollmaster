// /home/krylon/go/src/github.com/blicero/scrollmaster/model/record.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-13 21:14:30 krylon>

package model

import (
	"fmt"
	"time"

	"github.com/blicero/scrollmaster/common"
)

type Record struct {
	ID      int64
	HostID  int64
	Time    time.Time
	Source  string
	Message string
}

// Checksum returns a hash value of the record that (hopefully) uniquely identifies it.
func (r *Record) Checksum() string {
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

	return result
} // func (r *Record) Checksum() string
