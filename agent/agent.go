// /home/krylon/go/src/github.com/blicero/scrollmaster/agent/agent.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-31 15:44:13 krylon>

// Package agent implements the gathering and transmission of log records the the Server.
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/common/path"
	"github.com/blicero/scrollmaster/logdomain"
	"github.com/blicero/scrollmaster/logreader"
)

// Agent is the component that gathers Logrecords on a Host and transmits
// them to a Server.
type Agent struct {
	addr   string
	log    *log.Logger
	client http.Client
	reader logreader.LogReader
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
