// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/logreader_syslog.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-16 21:05:04 krylon>

package logreader

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/logdomain"
	"github.com/blicero/scrollmaster/model"
)

type logfile struct {
	path string
	fh   *os.File
}

// SyslogReader is a LogReader that reads files created by the syslog daemon
// commonly used on *BSD (and some Linux distros, I suppose).
type SyslogReader struct {
	log   *log.Logger
	err   error
	files []logfile
}

var (
	mpat = regexp.MustCompile(`^(\w{3}\s+\d+\s+\d+:\d+:\d+)\s+(\S+)\s+(.*)$`)
	spat = regexp.MustCompile(`^(\w+)\[\d+\]:\s+(.*)$`)
)

func init() {
	if runtime.GOOS != "linux" {
		DefaultOpener = CreateSyslogReader
	}
}

// CreateSyslogReader creates and returns a SyslogReader that reads the given
// logfiles.
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

// Init opens the logfiles.
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

// Close closes all the opened logfiles.
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
	)

	for _, lf := range r.files {
		var (
			cnt int
			sc  *bufio.Scanner
		)

		r.log.Printf("[TRACE] Reading from %s\n", lf.path)

		sc = bufio.NewScanner(lf.fh)

		for sc.Scan() {
			var (
				line = sc.Text()
				rec  model.Record
				m    []string
			)

			if m = mpat.FindStringSubmatch(line); m != nil {
				// parse timestamp
				if rec.Time, err = time.Parse("Jan  2 15:04:05", m[1]); err != nil {
					r.log.Printf("[ERROR] Cannot parse timestamp %q: %s\n",
						m[1],
						err.Error())
					continue
				} else if rec.Time.Before(begin) {
					r.log.Printf("[TRACE] Record timestamp is too old: %s < %s\n",
						rec.Time.Format(common.TimestampFormat),
						begin.Format(common.TimestampFormat))
					continue
				}

				var source = spat.FindStringSubmatch(m[3])
				if source == nil {
					rec.Source = m[2]
					rec.Message = m[3]
				} else {
					rec.Source = source[1]
					rec.Message = source[2]
				}

				queue <- rec
				if cnt++; cnt >= max {
					break
				}
			} else {
				r.log.Printf("[TRACE] Failed to parse line: %s\n",
					line)
			}
		}

		r.log.Printf("[TRACE] Read %d records from %s\n",
			cnt,
			lf.path)
	}
} // func (r *SyslogReader) ReadFrom(begin time.Time, max int, queue chan<- model.Record)
