// /home/krylon/go/src/github.com/blicero/scrollmaster/server/01_server_init_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 25. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-26 09:48:07 krylon>

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/blicero/scrollmaster/model"
)

func TestServerCreate(t *testing.T) {
	var err error

	addr = fmt.Sprintf("[::1]:%d", testPort)

	if srv, err = Create(addr); err != nil {
		srv = nil
		t.Fatalf("Error creating Server: %s",
			err.Error())
	}

	go srv.ListenAndServe()
} // func TestServerCreate(t *testing.T)

func TestServerHandleAgentInit(t *testing.T) {
	const path = "/ws/init"
	var (
		err   error
		res   *http.Response
		reply model.Response
		uri   = fmt.Sprintf("http://%s%s",
			addr,
			path)
		host = model.Host{
			Name: "schwarzgeraet.local",
		}
		data []byte
		buf  *bytes.Buffer
	)

	t.Logf("POST %s", uri)

	if srv == nil {
		t.SkipNow()
	} else if data, err = json.Marshal(&host); err != nil {
		t.Fatalf("Error serializing Host: %s", err.Error())
	}

	buf = bytes.NewBuffer(data)

	if res, err = client.Post(uri, "application/json", buf); err != nil {
		t.Fatalf("Error POSTing to %s: %s",
			uri,
			err.Error())
	} else if res.StatusCode != 200 {
		t.Fatalf("Unexpected HTTP status %03d", res.StatusCode)
	}

	buf.Reset()

	if _, err = io.Copy(buf, res.Body); err != nil {
		t.Fatalf("Error reading response body: %s", err.Error())
	} else if err = json.Unmarshal(buf.Bytes(), &reply); err != nil {
		t.Fatalf("Error decoding server reply: %s\n\n%s\n",
			err.Error(),
			buf.String())
	}
} // func TestServerHandleAgentInit(t *testing.T)
