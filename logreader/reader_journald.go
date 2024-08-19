// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/reader_journald.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-19 19:53:49 krylon>

//go:build linux

package logreader

import (
	"errors"
	"log"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/logdomain"
	"github.com/blicero/scrollmaster/model"
	"github.com/coreos/go-systemd/sdjournal"
)

func init() {
	DefaultOpener = CreateJournaldReader
}

// JournaldReader reads from systemd's journald log.
type JournaldReader struct {
	log     *log.Logger
	queue   chan<- model.Record
	journal *sdjournal.Journal
}

// CreateJournaldReader creates a JournaldReader. Duh.
func CreateJournaldReader(path string) (LogReader, error) {
	var (
		err error
		rdr = new(JournaldReader)
	)

	if rdr.log, err = common.GetLogger(logdomain.LogReader); err != nil {
		return nil, err
	}

	return rdr, nil
} // func CreateJournaldReader() (*JournaldReader, error)

// Init opens the Journal.
func (r *JournaldReader) Init() error {
	var err error

	if r.journal, err = sdjournal.NewJournal(); err != nil {
		r.log.Printf("[ERROR] Cannot open Journal: %s", err.Error())
		return err
	} else if r.journal == nil {
		var msg = "sdjournal.NewJournal did not return an error, but no Journal, either"
		r.log.Printf("[ERROR] %s", msg)
		return errors.New(msg)
	}

	r.log.Println("[INFO] Journal was opened successfully.")

	return nil
} // func (r *JournaldReader) Init() error

// Close closes the Journal.
func (r *JournaldReader) Close() error {
	var err error

	if r.journal == nil {
		r.log.Println("[ERROR] Close was called, but Journal is not open.")
		return nil
	} else if err = r.journal.Close(); err != nil {
		r.log.Printf("[ERROR] Failed to close Journal: %s\n", err.Error())
		return err
	}

	r.journal = nil
	return nil
} // func (r *JournaldReader) Close() error

// ReadFrom reads Journal entries beginning a the given time stamp.
// Records are fed to the channel passed as the second argument.
// Upon returning, the method will close the channel.
func (r *JournaldReader) ReadFrom(begin time.Time, queue chan<- model.Record) {
	if r.journal == nil {
		r.log.Println("[CRITICAL] ReadFrom was called on unopened Journal")
		panic("ReadFrom was called on unopened Journal")
	}

	r.queue = queue

	defer close(queue)

	var (
		err    error
		step   uint64
		bstamp uint64 = uint64(begin.Unix()) * 1_000_000
	)

	if err = r.journal.SeekRealtimeUsec(bstamp); err != nil {
		r.log.Printf("[ERROR] Failed to seek log to specified time: %s\n",
			err.Error())
		return
	}

	for step, err = r.journal.Next(); err == nil && step > 0; step, err = r.journal.Next() {
		var (
			rec   model.Record
			entry *sdjournal.JournalEntry
		)

		if entry, err = r.journal.GetEntry(); err != nil {
			r.log.Printf("[ERROR] Failed to read from Journal: %s\n",
				err.Error())
			continue
		} else if entry.RealtimeTimestamp < bstamp {
			continue
		}

		rec = model.Record{
			Time:    time.Unix(int64(entry.RealtimeTimestamp), 0),
			Source:  entry.Fields["_COMM"],
			Message: entry.Fields["MESSAGE"],
		}

		queue <- rec

	}
} // func (r *JournaldReader) ReadFrom(begin time.Time, queue chan<- model.Record)
