// /home/krylon/go/src/krylib/datecounter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 28. 01. 2019 by Benjamin Walkenhorst
// (c) 2019 Benjamin Walkenhorst
// Time-stamp: <2019-09-13 20:57:58 krylon>

package krylib

import "time"

// DateCounter is a counter that returns consecutive days.
type DateCounter struct {
	dt time.Time
}

// NewDateCounter creates a new date counter that starts counting
// at the given date.
func NewDateCounter(y, m, d int) *DateCounter {
	return &DateCounter{
		dt: time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local),
	}
} // func NewDateCounter(y,m,d int) *DateCounter

// Next returns the next date, starting with the date given as the starting date,
// then adding one day per consecutive call.
func (dc *DateCounter) Next() time.Time {
	var next = dc.dt
	dc.dt = dc.dt.AddDate(0, 0, 1)
	return next
} // func (dc *DateCounter) Next() time.Time

// Set sets the internal counter to the given date.
func (dc *DateCounter) Set(y, m, d int) time.Time {
	dc.dt = time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
	return dc.dt
} // func (dc *DateCounter) Set(y, m, d int) time.Time
