// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/logreader_syslog.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-15 20:23:20 krylon>

package logreader

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/logdomain"
	"github.com/blicero/scrollmaster/model"
)

type logfile struct {
	path string
	fh   *os.File
}

type SyslogReader struct {
	log   *log.Logger
	err   error
	files []logfile
}

var mpat = regexp.MustCompile(`(i)^(\w+ \d+ \d+:\d+:\d+) (\S+) (.*)$`)

func CreateSyslogReader(path ...string) (LogReader, error) {
	var (
		err error
		rdr = &SyslogReader{}
	)

	if rdr.log, err = common.GetLogger(logdomain.LogReader); err != nil {
		return nil, err
	}

	rdr.files = make([]logfile, len(path))

	for idx, logpath := range path {
		rdr.files[idx].path = logpath
	}

	return rdr, nil
} // func CreateSyslogReader(path string) (LogReader, error)

func (r *SyslogReader) Init() error {
	var (
		err error
	)

	for idx := range r.files {
		if r.files[idx].fh, err = os.Open(r.files[idx].path); err != nil {
			r.log.Printf("[ERROR] Failed to open %s: %s\n",
				r.files[idx].path,
				err.Error())
			for i := 0; i < idx; i++ {
				r.files[idx].fh.Close() // nolint: errcheck
				r.files[idx].fh = nil
			}
			return err
		}
	}

	return nil
} // func (r *SyslogReader) Init() error

func (r *SyslogReader) Close() error {
	for idx := range r.files {
		r.files[idx].fh.Close() // nolint: errcheck
		r.files[idx].fh = nil
	}

	return nil
} // func (r *SyslogReader) Close() error

// IsError returns the Reader's error state
func (r *SyslogReader) IsError() (bool, error) {
	return (r.err != nil), r.err
} // func (r *JournaldReader) IsError() (bool, error)

// ReadFrom reads Journal entries beginning a the given time stamp.
// Records are fed to the channel passed as the second argument.
// Upon returning, the method will close the channel.
func (r *SyslogReader) ReadFrom(begin time.Time, max int, queue chan<- model.Record) {
	defer close(queue)

	var (
		err error
		cnt int64
	)

	for _, lf := range r.files {
		var (
			sc *bufio.Scanner
		)

		sc = bufio.NewScanner(lf.fh)

		for sc.Scan() {
			var (
				line string
				rec  model.Record
				m    []string
			)

			line = sc.Text()

			if m = mpat.FindStringSubmatch(line); m != nil {
				// parse timestamp
			}
		}
	}
} // func (r *SyslogReader) ReadFrom(begin time.Time, max int, queue chan<- model.Record)
