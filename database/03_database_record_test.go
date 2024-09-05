// /home/krylon/go/src/github.com/blicero/scrollmaster/database/03_database_record_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-05 20:56:07 krylon>

package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/model"
)

const (
	recordCnt    = 100
	stepInterval = time.Second * 2
)

func TestRecordAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var baseStamp = time.Now().Add(time.Hour * -24)

	for _, h := range hosts {
		var stamp = baseStamp
		for i := 0; i < recordCnt; i++ {
			var (
				err    error
				record = model.Record{
					HostID:  h.ID,
					Time:    stamp,
					Source:  "test",
					Message: fmt.Sprintf("Test message #%03d", i+1),
				}
			)

			if err = tdb.RecordAdd(&record); err != nil {
				t.Fatalf("Error adding record #%d for Host %s: %s",
					i+1,
					h.Name,
					err.Error())
			} else if record.ID == 0 {
				t.Fatalf("Record %d for host %s was added without error, but ID was not filled in.",
					i+1,
					h.Name)
			}

			stamp = stamp.Add(stepInterval)
		}
	}
} // func TestRecordAdd(t *testing.T)

func TestRecordGetMostRecent(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var (
		err   error
		stamp time.Time
	)

	for id, host := range hosts {
		if stamp, err = tdb.RecordGetMostRecent(id); err != nil {
			t.Errorf("Failed to get most recent record timestamp for Host %s: %s",
				host.Name,
				err.Error())
		} else {
			t.Logf("Timestamp for most recent record for Host %s is %s",
				host.Name,
				stamp.Format(common.TimestampFormatSubSecond))
		}
	}
} // func TestRecordGetMostRecent(t *testing.T)

func TestRecordGetRecent(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var (
		err     error
		records []model.Record
	)

	if records, err = tdb.RecordGetRecent(2); err != nil {
		t.Errorf("Failed to get 2 recent records: %s",
			err.Error())
	} else if len(records) != 2 {
		t.Errorf("Unexpected number of records: %d (expected 2)",
			len(records))
	}
} // func TestRecordGetRecent(t *testing.T)
