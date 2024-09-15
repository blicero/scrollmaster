// /home/krylon/go/src/github.com/blicero/scrollmaster/logreader/logreader_syslog.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-15 15:33:19 krylon>

package logreader

import (
	"log"
	"os"

	"github.com/blicero/scrollmaster/model"
)

type logfile struct {
	path string
	fh   *os.File
}

type SyslogReader struct {
	log   *log.Logger
	queue chan<- model.Record
	err   error
	files []logfile
}

func CreateSyslogReader(path string) (LogReader, error) {

} // func CreateSyslogReader(path string) (LogReader, error)
