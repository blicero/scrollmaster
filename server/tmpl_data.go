// /home/krylon/go/src/ticker/web/tmpl_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 05. 2020 by Benjamin Walkenhorst
// (c) 2020 Benjamin Walkenhorst
// Time-stamp: <2024-06-10 18:43:07 krylon>
//
// This file contains data structures to be passed to HTML templates.

package server

import (
	"crypto/sha512"
	"fmt"
	"time"

	"github.com/blicero/donkey/common"

	"github.com/hashicorp/logutils"
)

type message struct { // nolint: unused
	Timestamp time.Time
	Level     logutils.LogLevel
	Message   string
}

func (m *message) TimeString() string { // nolint: unused
	return m.Timestamp.Format(common.TimestampFormat)
} // func (m *Message) TimeString() string

func (m *message) Checksum() string { // nolint: unused
	var str = m.Timestamp.Format(common.TimestampFormat) + "##" +
		string(m.Level) + "##" +
		m.Message

	var hash = sha512.New()
	hash.Write([]byte(str)) // nolint: gosec,errcheck

	var cksum = hash.Sum(nil)
	var ckstr = fmt.Sprintf("%x", cksum)

	return ckstr
} // func (m *message) Checksum() string

type tmplDataBase struct { // nolint: unused
	Title      string
	Messages   []message
	Debug      bool
	TestMsgGen bool
	URL        string
}

type tmplDataIndex struct { // nolint: unused,deadcode
	tmplDataBase
}

// Local Variables:  //
// compile-command: "go generate && go vet && go build -v -p 16 && gometalinter && go test -v" //
// End: //
