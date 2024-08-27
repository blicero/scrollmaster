// /home/krylon/go/src/github.com/blicero/scrollmaster/server/01_server_init_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 25. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-27 15:21:48 krylon>

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

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
	time.Sleep(time.Second)
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
		host01 = model.Host{
			Name: "schwarzgeraet.local",
		}
		data   []byte
		buf    *bytes.Buffer
		idstr  string
		ok     bool
		hostID int64
	)

	t.Logf("POST %s", uri)

	if srv == nil {
		t.SkipNow()
	} else if data, err = json.Marshal(&host01); err != nil {
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
	} else if !reply.Status {
		t.Fatalf("Failed to initialize Agent session for Host %s: %s",
			host01.Name,
			reply.Message)
	} else if idstr, ok = reply.Payload["ID"]; !ok {
		t.Errorf("Expected Host ID in Payload (%#v)",
			reply.Payload)
	} else if hostID, err = strconv.ParseInt(idstr, 10, 64); err != nil {
		t.Errorf("Could not parse ID (%q): %s",
			idstr,
			err.Error())
	} else if hostID < 1 {
		t.Errorf("Server replied invalid hostID: %d", hostID)
	} else {
		host01.ID = hostID
	}

	// Next we try registering an unknown Server:
	var host02 = model.Host{
		Name: "zappelwurst.local",
		ID:   42,
	}

	if data, err = json.Marshal(&host02); err != nil {
		t.Fatalf("Error serializing Host: %s", err.Error())
	}

	buf.Reset()
	//io.Copy(buf,
	buf.Write(data)

	if res, err = client.Post(uri, "application/json", buf); err != nil {
		t.Fatalf("Error POSTing to %s: %s",
			uri,
			err.Error())
	} else if res.StatusCode != 200 {
		t.Fatalf("Unexpected HTTP status %03d", res.StatusCode)
	} else if _, err = io.Copy(buf, res.Body); err != nil {
		t.Fatalf("Error reading response body: %s", err.Error())
	} else if err = json.Unmarshal(buf.Bytes(), &reply); err != nil {
		t.Fatalf("Error decoding server reply: %s\n\n%s\n",
			err.Error(),
			buf.String())
	} else if reply.Status {
		t.Fatalf("Registering Host %s should have failed: %s",
			host02.Name,
			reply.Message)
	} else {
		t.Logf("Registering Host %s has failed as expected: %s",
			host02.Name,
			reply.Message)
	}
} // func TestServerHandleAgentInit(t *testing.T)
