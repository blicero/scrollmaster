// /home/krylon/go/src/github.com/blicero/scrollmaster/server/agent.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-05 20:28:43 krylon>

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/database"
	"github.com/blicero/scrollmaster/model"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Methods for handling the Agents

// /ws/init/$hostname
func (srv *Server) handleAgentInit(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	var (
		err         error
		db          *database.Database
		buf         bytes.Buffer
		hostname    string
		host        *model.Host
		msg, status string
		statusRaw   any
		res         model.Response
		sess        *sessions.Session
		ok          bool
		hstatus     int = 200
	)

	vars := mux.Vars(r)
	hostname = vars["hostname"]

	res.Payload = make(map[string]string)
	buf.Reset()
	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if host, err = db.HostGetByName(hostname); err != nil {
		res.Message = fmt.Sprintf("Failed to lookup host %s in database: %s",
			hostname,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 500
	} else if host == nil {
		srv.log.Printf("[INFO] Register Host %s in the database\n",
			hostname)

		host = &model.Host{Name: hostname}

		if err = db.HostAdd(host); err != nil {
			res.Message = fmt.Sprintf("Adding Host to database failed: %s",
				err.Error())
			srv.log.Printf("[ERROR] %s\n", res.Message)
			goto SEND_RESPONSE
		}
		res.Payload["ID"] = strconv.FormatInt(host.ID, 10)
	}

	if err = db.HostUpdateLastSeen(host, time.Now()); err != nil {
		res.Message = fmt.Sprintf("Failed to update last contact timestamp on Host %s: %s",
			hostname,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 500
		goto SEND_RESPONSE
	}

	if sess, err = srv.store.Get(r, sessionNameAgent); err != nil {
		res.Message = fmt.Sprintf(
			"Error getting/creating session %s: %s",
			sessionNameAgent,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		sess = nil
		goto SEND_RESPONSE
	} else if common.Debug {
		msg = dumpSession(sess)
		srv.log.Printf("[DEBUG] Existing session for Host %s (%d):\n%s\n",
			host.Name,
			host.ID,
			msg)
	}

	if statusRaw, ok = sess.Values["status"]; !ok {
		srv.log.Println("[DEBUG] No field \"status\" found in session")
	} else if status, ok = statusRaw.(string); ok && status == "ok" {
		res.Message = "Welcome back"
		res.Status = true
		res.Payload["status"] = "ok"
		goto SEND_RESPONSE
	} else if status != "ok" {
		res.Message = fmt.Sprintf("Invalid session status for host %s: %s",
			hostname, status)
		srv.log.Printf("[ERROR] %s\n",
			res.Message)
		hstatus = 403
		goto SEND_RESPONSE
	}

	sess.Values["status"] = "ok"
	sess.Values["host"] = host.ID

	res.Status = true
	res.Message = "Welcome aboard, buddy"

SEND_RESPONSE:
	if sess != nil {
		srv.log.Printf("[DEBUG] Save session for host %s\n",
			host.Name)
		if err = sess.Save(r, w); err != nil {
			srv.log.Printf("[ERROR] Failed to set session cookie: %s\n",
				err.Error())
		}
	}
	res.Timestamp = time.Now()
	var rbuf []byte
	if rbuf, err = json.Marshal(&res); err != nil {
		srv.log.Printf("[ERROR] Error serializing response: %s\n",
			err.Error())
		rbuf = errJSON(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(hstatus)
	if _, err = w.Write(rbuf); err != nil {
		msg = fmt.Sprintf("Failed to send result: %s",
			err.Error())
		srv.log.Println("[ERROR] " + msg)
	}
} // func (srv *Server) handleAgentInit(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleGetMostRecent(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	var (
		err         error
		hstatus     int = 200
		db          *database.Database
		hostID      int64
		msg, status string
		raw         any
		res         model.Response
		sess        *sessions.Session
		ok          bool
		timestamp   time.Time
	)

	if sess, err = srv.store.Get(r, sessionNameAgent); err != nil {
		res.Message = fmt.Sprintf(
			"Error getting/creating session %s: %s",
			sessionNameAgent,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		sess = nil
		hstatus = 403
		goto SEND_RESPONSE
	} else if common.Debug {
		msg = dumpSession(sess)
		srv.log.Printf("[DEBUG] Existing session %s\n", msg)
	}

	if raw, ok = sess.Values["status"]; !ok {
		res.Message = "No session status"
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 403
		goto SEND_RESPONSE
	} else if status, ok = raw.(string); !ok {
		res.Message = fmt.Sprintf("Cannot decode session status: %#v", raw)
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if status != "ok" {
		res.Message = fmt.Sprintf("Invalid session status %q", status)
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if raw, ok = sess.Values["host"]; !ok {
		res.Message = "No host ID in session"
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if hostID, ok = raw.(int64); !ok {
		res.Message = fmt.Sprintf("Invalid type for Host ID in session: %T (%#v)",
			raw,
			raw)
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	}

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if timestamp, err = db.RecordGetMostRecent(hostID); err != nil {
		res.Message = fmt.Sprintf("Error looking up records for Host %d: %s",
			hostID,
			err.Error())
	}

	res.Message = "Success"
	res.Status = true
	res.Payload = map[string]string{
		"timestamp": timestamp.Format(common.TimestampFormatSubSecond),
	}

SEND_RESPONSE:
	if sess != nil {
		if err = sess.Save(r, w); err != nil {
			srv.log.Printf("[ERROR] Failed to set session cookie: %s\n",
				err.Error())
		}
	}
	res.Timestamp = time.Now()
	var rbuf []byte
	if rbuf, err = json.Marshal(&res); err != nil {
		srv.log.Printf("[ERROR] Error serializing response: %s\n",
			err.Error())
		rbuf = errJSON(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(hstatus)
	if _, err = w.Write(rbuf); err != nil {
		msg = fmt.Sprintf("Failed to send result: %s",
			err.Error())
		srv.log.Println("[ERROR] " + msg)
	}
} // func (srv *Server) handleGetMostRecent(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleSubmitRecords(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	var (
		err          error
		hstatus      int = 200
		db           *database.Database
		buf          bytes.Buffer
		body         []byte
		hostID       int64
		host         *model.Host
		msg, status  string
		data         model.RecordSlice
		raw          any
		res          model.Response
		sess         *sessions.Session
		ok, txStatus bool
	)

	if sess, err = srv.store.Get(r, sessionNameAgent); err != nil {
		res.Message = fmt.Sprintf(
			"Error getting/creating session %s: %s",
			sessionNameAgent,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		sess = nil
		hstatus = 403
		goto SEND_RESPONSE
	} else if common.Debug {
		msg = dumpSession(sess)
		srv.log.Printf("[DEBUG] Existing session %s\n", msg)
	}

	if raw, ok = sess.Values["status"]; !ok {
		res.Message = "No session status"
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 403
		goto SEND_RESPONSE
	} else if status, ok = raw.(string); !ok {
		res.Message = fmt.Sprintf("Cannot decode session status: %#v", raw)
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if status != "ok" {
		res.Message = fmt.Sprintf("Invalid session status %q", status)
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if raw, ok = sess.Values["host"]; !ok {
		res.Message = "No host ID in session"
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if hostID, ok = raw.(int64); !ok {
		res.Message = fmt.Sprintf("Invalid type for Host ID in session: %T (%#v)",
			raw,
			raw)
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	}

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if err = db.Begin(); err != nil {
		res.Message = fmt.Sprintf("Error starting database transaction: %s\n",
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	}

	defer func() {
		var e error

		if txStatus {
			if e = db.Commit(); e != nil {
				srv.log.Printf("[ERROR] Error committing transaction: %s\n", e.Error())
			}
		} else {
			if e = db.Rollback(); e != nil {
				srv.log.Printf("[ERROR] Error rolling back transaction: %s\n", e.Error())
			}
		}
	}()

	if host, err = db.HostGetByID(hostID); err != nil {
		res.Message = fmt.Sprintf("Error looking up host %d in database", hostID)
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if host == nil {
		res.Message = fmt.Sprintf("Could not find host %d in database", hostID)
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if err = db.HostUpdateLastSeen(host, time.Now()); err != nil {
		res.Message = fmt.Sprintf("[ERROR] Cannot update LastSeen timestamp on Host %s (%d): %s\n",
			host.Name,
			host.ID,
			err.Error())
	}

	if _, err = io.Copy(&buf, r.Body); err != nil {
		res.Message = fmt.Sprintf("Failed to read HTTP request body: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n",
			res.Message)
		goto SEND_RESPONSE
	}

	body = buf.Bytes()

	if err = json.Unmarshal(body, &data); err != nil {
		msg = fmt.Sprintf("Failed to decode payload: %s", err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		res.Message = msg
		goto SEND_RESPONSE
	}

	srv.log.Printf("[DEBUG] Agent on %s delivered %d log records\n",
		host.Name,
		len(data))

	for idx, rec := range data {
		rec.HostID = host.ID
		var exist bool

		if exist, err = db.RecordCheckExist(rec.Checksum()); err != nil {
			srv.log.Printf("[ERROR] Failed to check if record with checksum %q exists: %s\n",
				rec.Checksum(),
				err.Error())
		} else if exist {
			// srv.log.Printf("[DEBUG] Skipping duplicate log record from %s / %s / %s\n",
			// 	host.Name,
			// 	rec.Source,
			// 	rec.Time.Format(common.TimestampFormat))
			continue
		} else if err = db.RecordAdd(&rec); err != nil {
			res.Message = fmt.Sprintf("Failed to add Record #%d: %s",
				idx,
				err.Error())
			srv.log.Printf("[ERROR] Failed to add log record for %s: %s\n",
				host.Name,
				err.Error())
			goto SEND_RESPONSE
		}
	}

	txStatus = true
	res.Status = true

SEND_RESPONSE:
	if sess != nil {
		if err = sess.Save(r, w); err != nil {
			srv.log.Printf("[ERROR] Failed to set session cookie: %s\n",
				err.Error())
		}
	}
	res.Timestamp = time.Now()
	var rbuf []byte
	if rbuf, err = json.Marshal(&res); err != nil {
		srv.log.Printf("[ERROR] Error serializing response: %s\n",
			err.Error())
		rbuf = errJSON(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(hstatus)
	if _, err = w.Write(rbuf); err != nil {
		msg = fmt.Sprintf("Failed to send result: %s",
			err.Error())
		srv.log.Println("[ERROR] " + msg)
	}
} // func (srv *Server) handleSubmitRecords(w http.ResponseWriter, r *http.Request)
