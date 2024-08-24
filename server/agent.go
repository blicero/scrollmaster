// /home/krylon/go/src/github.com/blicero/scrollmaster/server/agent.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-25 00:51:55 krylon>

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/blicero/scrollmaster/database"
	"github.com/blicero/scrollmaster/model"
)

// Methods for handling the Agents

// /ws/init
func (srv *Server) handleAgentInit(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	var (
		err          error
		db           *database.Database
		buf          bytes.Buffer
		body         []byte
		host, dbhost *model.Host
		msg          string
		res          model.Response
	)

	if _, err = io.Copy(&buf, r.Body); err != nil {
		res.Message = fmt.Sprintf("Failed to read HTTP request body: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n",
			res.Message)
		goto SEND_RESPONSE
	}

	body = buf.Bytes()
	host = new(model.Host)

	if err = json.Unmarshal(body, host); err != nil {
		msg = fmt.Sprintf("Failed to decode payload: %s", err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		res.Message = msg
		goto SEND_RESPONSE
	}

	buf.Reset()
	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if host.ID == 0 {
		if err = db.HostAdd(host); err != nil {
			res.Message = fmt.Sprintf("Adding Host to database failed: %s",
				err.Error())
			srv.log.Printf("[ERROR] %s\n", res.Message)
			goto SEND_RESPONSE
		}
	} else {
		if dbhost, err = db.HostGetByID(host.ID); err != nil {
			res.Message = fmt.Sprintf("Error looking up Host %d in database: %s",
				host.ID,
				err.Error())
		} else if dbhost.Name == host.Name {

		}
	}

SEND_RESPONSE:
	res.Timestamp = time.Now()
	var rbuf []byte
	if rbuf, err = json.Marshal(&res); err != nil {
		srv.log.Printf("[ERROR] Error serializing response: %s\n",
			err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(200)
	if _, err = w.Write(rbuf); err != nil {
		msg = fmt.Sprintf("Failed to send result: %s",
			err.Error())
		srv.log.Println("[ERROR] " + msg)
	}
} // func (srv *Server) handleAgentInit(w http.ResponseWriter, r *http.Request)
