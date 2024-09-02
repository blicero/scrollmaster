// /home/krylon/go/src/github.com/blicero/scrollmaster/agent/agent.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-02 19:56:37 krylon>

// Package agent implements the gathering and transmission of log records the the Server.
package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/common/path"
	"github.com/blicero/scrollmaster/logdomain"
	"github.com/blicero/scrollmaster/logreader"
	"github.com/blicero/scrollmaster/model"
)

// For debugging purposes, I set this value really low, for regular use, it
// should be more like a couple of minutes.
const (
	checkInterval = time.Second * 10
	maxErr        = 5
)

// Agent is the component that gathers Logrecords on a Host and transmits
// them to a Server.
type Agent struct {
	addr     string
	hostname string
	log      *log.Logger
	lock     sync.RWMutex
	active   atomic.Bool
	client   http.Client
	reader   logreader.LogReader
}

// Create creates a new Agent.
func Create(addr, logpath string) (*Agent, error) {
	var (
		err error
		ag  = &Agent{addr: addr}
	)

	if ag.log, err = common.GetLogger(logdomain.Agent); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to create Logger: %s\n",
			err.Error())
		return nil, err
	} else if ag.hostname, err = os.Hostname(); err != nil {
		ag.log.Printf("[CRITICAL] Failed to query system hostname: %s\n",
			err.Error())
		return nil, err
	} else if ag.client.Jar, err = ag.initCookieJar(); err != nil {
		return nil, err
	} else if ag.reader, err = logreader.DefaultOpener(logpath); err != nil {
		ag.log.Printf("[ERROR] Failed to create LogReader for %s: %s\n",
			logpath,
			err.Error())
		return nil, err
	}

	return ag, nil
} // func Create(addr string) (*Agent, error)

// Return the Agent's active flag.
func (ag *Agent) IsActive() bool {
	return ag.active.Load()
} // func (ag *Agent) IsActive() bool

func (ag *Agent) Run() error {
	var (
		err        error
		startStamp time.Time
	)

	ag.active.Store(true)
	defer ag.active.Store(false)

	if err = ag.reader.Init(); err != nil {
		ag.log.Printf("[ERROR] Failed to initialize LogReader: %s\n",
			err.Error())
		return err
	} else if err = ag.register(); err != nil {
		ag.log.Printf("[ERROR] Failed to register with Server at %s: %s\n",
			ag.addr,
			err.Error())
	} else if startStamp, err = ag.queryMostRecent(); err != nil {
		ag.log.Printf("[ERROR] Failed to query most recent log timestamp: %s\n",
			err.Error())
	}

	var errCnt int
	// var ticker = time.NewTicker(checkInterval)
	// defer ticker.Stop()

	for ag.active.Load() {
		var (
			queue   = make(chan model.Record)
			records = make([]model.Record, 0, 64)
		)

		go ag.reader.ReadFrom(startStamp, queue)

		for rec := range queue {
			records = append(records, rec)
		}

		startStamp = records[len(records)-1].Time

		if err = ag.submitRecords(records); err != nil {
			ag.log.Printf("[ERROR] Failed to deliver records to %s: %s\n",
				ag.addr,
				err.Error())
			if errCnt++; errCnt >= maxErr {
				ag.active.Store(false)
				ag.log.Printf("[CRITICAL] Submitting records failed %d times, giving up.\n",
					errCnt)
				return err
			}
		}

		var delay = checkInterval + time.Second*time.Duration(errCnt*errCnt)
		time.Sleep(delay)
	}

	return nil
} // func (ag *Agent) Run() error

func (ag *Agent) initCookieJar() (*cookiejar.Jar, error) {
	var (
		err              error
		cookiepath, ustr string
		opt              *cookiejar.Options
		jar              *cookiejar.Jar
		buf              bytes.Buffer
		fh               *os.File
		cookies          []*http.Cookie
		uri              *url.URL
	)

	ag.lock.Lock()
	defer ag.lock.Unlock()

	ustr = fmt.Sprintf("http://%s/ws/init",
		ag.addr)

	if uri, err = url.Parse(ustr); err != nil {
		ag.log.Printf("[CANTHAPPEN] Failed to parse URL %s: %s\n",
			ustr,
			err.Error())
		return nil, err
	}

	opt = &cookiejar.Options{}
	if jar, err = cookiejar.New(opt); err != nil {
		ag.log.Printf("[CANTHAPPEN] Failed to create empty cookiejar: %s\n",
			err.Error())
		return nil, err
	}

	cookiepath = common.Path(path.Cookiejar)

	if fh, err = os.Open(cookiepath); err != nil {
		if os.IsNotExist(err) {
			goto END
		}
		ag.log.Printf("[ERROR] Cannot open stored cookies at %s: %s\n",
			cookiepath,
			err.Error())
		return nil, err
	}

	defer fh.Close() // nolint: errcheck

	if _, err = io.Copy(&buf, fh); err != nil {
		ag.log.Printf("[ERROR] Failed to read stored cookies: %s\n",
			err.Error())
		return nil, err
	} else if err = json.Unmarshal(buf.Bytes(), &cookies); err != nil {
		ag.log.Printf("[ERROR] Failed to parse stored cookies: %s\n\n%s\n",
			err.Error(),
			buf.String())
		return nil, err
	}

	jar.SetCookies(uri, cookies)

END:
	return jar, nil
} // func (ag *Agent) initCookieJar() (*cookiejar.Jar, error)

func (ag *Agent) saveCookieJar() error { // nolint: unused
	var (
		err     error
		buf     []byte
		fh      *os.File
		cookies []*http.Cookie
		ustr    string
		uri     *url.URL
		path    = common.Path(path.Cookiejar)
	)

	ag.lock.Lock()
	defer ag.lock.Unlock()

	ustr = fmt.Sprintf("http://%s/ws/init",
		ag.addr)

	if uri, err = url.Parse(ustr); err != nil {
		ag.log.Printf("[CANTHAPPEN] Failed to parse URL %s: %s\n",
			ustr,
			err.Error())
		return err
	}

	cookies = ag.client.Jar.Cookies(uri)

	if buf, err = json.Marshal(cookies); err != nil {
		ag.log.Printf("[ERROR] Failed serialize cookies: %s\n",
			err.Error())
		return err
	} else if fh, err = os.Create(path); err != nil {
		ag.log.Printf("[ERROR] Failed to open cookie store at %s: %s\n",
			path,
			err.Error())
		return err
	}

	defer fh.Close() // nolint: errcheck

	if _, err = fh.Write(buf); err != nil {
		ag.log.Printf("[ERROR] Failed to write cookies to %s: %s\n",
			path,
			err.Error())
	}

	return nil
} // func (ag *Agent) saveCookieJar() error

func (ag *Agent) register() error {
	const uriBase = "/ws/init"
	var (
		err   error
		buf   bytes.Buffer
		res   *http.Response
		reply model.Response
		ustr  string
	)

	ag.lock.RLock()
	defer ag.lock.RUnlock()

	ustr = fmt.Sprintf("http://%s%s/%s",
		ag.addr,
		uriBase,
		ag.hostname)

	if res, err = ag.client.Get(ustr); err != nil {
		ag.log.Printf("[ERROR] HTTP Protocol level error GETting %s: %s\n",
			ustr,
			err.Error())
		return err
	}

	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != 200 {
		ag.log.Printf("[ERROR] Unhappy HTTP status code %3d\n",
			res.StatusCode)

	}

	if _, err = io.Copy(&buf, res.Body); err != nil {
		ag.log.Printf("[ERROR] Failed to read HTTP response body: %s\n",
			err.Error())
		return err
	} else if err = json.Unmarshal(buf.Bytes(), &reply); err != nil {
		ag.log.Printf("[ERROR] Failed to parse response from server: %s\n\n%s\n",
			err.Error(),
			buf.String())
		return err
	} else if !reply.Status {
		ag.log.Printf("[ERROR] Server %s says request failed: %s\n",
			ag.addr,
			reply.Message)

	}

	ag.log.Printf("[DEBUG] Server %s says %q\n",
		ag.addr,
		reply.Message)

	return nil
} // func (ag *Agent) register() error

func (ag *Agent) queryMostRecent() (time.Time, error) {
	const uriBase = "/ws/most_recent"
	var (
		err   error
		buf   bytes.Buffer
		res   *http.Response
		reply model.Response
		ustr  string
		stamp time.Time
	)

	ag.lock.RLock()
	defer ag.lock.RUnlock()

	ustr = fmt.Sprintf("http://%s%s",
		ag.addr,
		uriBase)

	if res, err = ag.client.Get(ustr); err != nil {
		ag.log.Printf("[ERROR] HTTP Protocol level error GETting %s: %s\n",
			ustr,
			err.Error())
		return stamp, err
	}

	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != 200 {
		ag.log.Printf("[ERROR] Unhappy HTTP status code %3d\n",
			res.StatusCode)

	}

	if _, err = io.Copy(&buf, res.Body); err != nil {
		ag.log.Printf("[ERROR] Failed to read HTTP response body: %s\n",
			err.Error())
		return stamp, err
	} else if err = json.Unmarshal(buf.Bytes(), &reply); err != nil {
		ag.log.Printf("[ERROR] Failed to parse response from server: %s\n\n%s\n",
			err.Error(),
			buf.String())
		return stamp, err
	} else if !reply.Status {
		err = fmt.Errorf("Server %s says request failed: %s",
			ag.addr,
			reply.Message)
		ag.log.Printf("[ERROR] %s\n", err.Error())
		return stamp, err
	}

	var (
		stampStr string
		ok       bool
	)

	if stampStr, ok = reply.Payload["timestamp"]; !ok {
		err = errors.New("Response Payload did not include timestamp")
		ag.log.Printf("[ERROR] %s\n", err.Error())
		return stamp, err
	} else if stamp, err = time.Parse(common.TimestampFormatSubSecond, stampStr); err != nil {
		ag.log.Printf("[ERROR] Cannot parse time stamp from payload %q: %s\n",
			stampStr,
			err.Error())
		return stamp, err
	}

	return stamp, err
} // func (ag *Agent) queryMostRecent() (time.Time, error)

func (ag *Agent) submitRecords(records []model.Record) error {
	const uriBase = "/ws/submit_records"
	var (
		err   error
		data  []byte
		buf   *bytes.Buffer
		res   *http.Response
		reply model.Response
		ustr  string
	)

	ustr = fmt.Sprintf("http://%s%s",
		ag.addr,
		uriBase)

	if data, err = json.Marshal(records); err != nil {
		ag.log.Printf("[ERROR] Failed to serialize log records: %s\n",
			err.Error())
		return err
	}

	buf = bytes.NewBuffer(data)

	if res, err = ag.client.Post(ustr, "application/json", buf); err != nil {
		ag.log.Printf("[ERROR] Failed to upload %d log records to %s: %s\n",
			len(records),
			ustr,
			err.Error())
		return err
	}

	defer res.Body.Close() // nolint: errcheck

	buf.Reset()
	if _, err = io.Copy(buf, res.Body); err != nil {
		ag.log.Printf("[ERROR] Trouble reading HTTP response body: %s\n",
			err.Error())
		return err
	} else if err = json.Unmarshal(buf.Bytes(), &reply); err != nil {
		ag.log.Printf("[ERROR] Failed to parse response: %s\n\n%s\n",
			err.Error(),
			buf.String())
		return err
	}

	if res.StatusCode != 200 {
		ag.log.Printf("[ERROR] Unexpected HTTP status code %d\n", res.StatusCode)
	}

	if !reply.Status {
		ag.log.Printf("[ERROR] Submitting %d log records to %s failed: %s\n",
			len(records),
			ustr,
			reply.Message)
		return errors.New(reply.Message)
	}

	return nil
} // func (ag *Agent) submitRecords() error
