// /home/krylon/go/src/github.com/blicero/scrollmaster/server/01_server_init_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 25. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-31 16:54:59 krylon>

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/model"
)

var testHost model.Host

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
	const (
		path     = "/ws/init"
		hostname = "schwarzgeraet"
	)
	var (
		err   error
		res   *http.Response
		reply model.Response
		uri   = fmt.Sprintf("http://%s%s/%s",
			addr,
			path,
			hostname)
		host01 = model.Host{
			Name: hostname,
		}
		data   []byte
		buf    *bytes.Buffer
		idstr  string
		ok     bool
		hostID int64
	)

	if srv == nil {
		t.SkipNow()
	}

	t.Logf("GET %s", uri)
	buf = bytes.NewBuffer(data)

	if res, err = client.Get(uri); err != nil {
		t.Fatalf("Error POSTing to %s: %s",
			uri,
			err.Error())
	}

	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != 200 {
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
		testHost = host01
	}
} // func TestServerHandleAgentInit(t *testing.T)

func TestServerHandleSubmitRecords(t *testing.T) {
	if srv == nil {
		t.SkipNow()
	}

	const (
		path      = "/ws/submit_records"
		recordCnt = 200
	)

	var (
		err       error
		res       *http.Response
		reply     model.Response
		records   [recordCnt]model.Record
		basestamp = time.Now().Add(time.Hour * -24)
		timeStep  = time.Second * 3
		jdata     []byte
		buf       *bytes.Buffer
		uri       = fmt.Sprintf("http://%s%s",
			addr,
			path)
	)

	for i := 0; i < recordCnt; i++ {
		records[i] = model.Record{
			HostID:  testHost.ID,
			Time:    basestamp.Add(timeStep * time.Duration(i)),
			Source:  "QA",
			Message: fmt.Sprintf("Something happened - %04d", i),
		}
	}

	if jdata, err = json.Marshal(records); err != nil {
		t.Fatalf("Failed to serialize data: %s", err.Error())
	}

	buf = bytes.NewBuffer(jdata)

	t.Logf("POST %s", uri)
	if res, err = client.Post(uri, "application/json", buf); err != nil {
		t.Fatalf("Error POSTing to %s: %s",
			uri,
			err.Error())
	}

	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != 200 {
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
		t.Fatalf("Failed to deliver log records: %s", reply.Message)
	}

	// Now, if we show up without registering first, we get rejected, don't we?
	var altClient http.Client

	for i := 0; i < recordCnt; i++ {
		records[i].Time = records[i].Time.Add(time.Hour * -24)
	}

	if jdata, err = json.Marshal(records); err != nil {
		t.Fatalf("Failed to serialize data: %s", err.Error())
	}

	buf = bytes.NewBuffer(jdata)

	t.Logf("POST %s", uri)
	if res, err = altClient.Post(uri, "application/json", buf); err != nil {
		t.Fatalf("Error POSTing to %s: %s",
			uri,
			err.Error())
	} else if res.StatusCode != 403 {
		t.Fatalf("Unexpected HTTP status %03d", res.StatusCode)
	}

	buf.Reset()
} // func TestServerHandleSubmitRecords(t *testing.T)

func TestServerHandleMostRecent(t *testing.T) {
	if srv == nil {
		t.SkipNow()
	}

	const (
		path = "/ws/most_recent"
	)

	var (
		err   error
		res   *http.Response
		reply model.Response
		buf   bytes.Buffer
		uri   = fmt.Sprintf("http://%s%s",
			addr,
			path)
	)

	t.Logf("GET %s", uri)

	if res, err = client.Get(uri); err != nil {
		t.Fatalf("Failed to GET %s: %s",
			uri,
			err.Error())
	}

	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != 200 {
		t.Fatalf("Unexpected HTTP Status %d", res.StatusCode)
	} else if _, err = io.Copy(&buf, res.Body); err != nil {
		t.Fatalf("Failed to read HTTP response body: %s",
			err.Error())
	} else if err = json.Unmarshal(buf.Bytes(), &reply); err != nil {
		t.Fatalf("Failed to decode response body: %s\n\n%s\n\n",
			err.Error(),
			buf.String())
	} else if !reply.Status {
		t.Fatalf("Request Failed: %s",
			reply.Message)
	}

	t.Logf("Response message: %s", reply.Message)

	var (
		stampStr string
		stamp    time.Time
		ok       bool
	)

	if stampStr, ok = reply.Payload["timestamp"]; !ok {
		t.Fatalf("Reply payload does not contain timestamp: %#v", reply.Payload)
	} else if stamp, err = time.Parse(common.TimestampFormatSubSecond, stampStr); err != nil {
		t.Fatalf("Cannot parse timestamp from result payload: %s (%q)",
			err.Error(),
			stampStr)
	}

	t.Logf("Timestamp: %s", stamp.Format(common.TimestampFormatSubSecond))

	if common.Debug {
		var u *url.URL

		if u, err = url.Parse(uri); err == nil {
			for idx, c := range client.Jar.Cookies(u) {
				t.Logf("Cookie %2d: %#v",
					idx+1,
					c)
			}
		} else {
			// CANTHAPPEN
			t.Logf("Strange - cannot parse URL %s: %s",
				uri,
				err.Error())
		}
	}
} // func TestServerHandleMostRecent(t *testing.T)
