// /home/krylon/go/src/github.com/blicero/scrollmaster/server/01_server_init_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 25. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-25 23:37:05 krylon>

package server

import (
	"fmt"
	"testing"

	"github.com/blicero/scrollmaster/common"
)

const testPort = common.Port + 2

var srv *Server

func TestServerCreate(t *testing.T) {
	var (
		err  error
		addr string
	)

	addr = fmt.Sprintf("[::1]:%d", testPort)

	if srv, err = Create(addr); err != nil {
		srv = nil
		t.Fatalf("Error creating Server: %s",
			err.Error())
	}
} // func TestServerCreate(t *testing.T)
