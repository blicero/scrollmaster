// /home/krylon/go/src/github.com/blicero/scrollmaster/server/ajax.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-12 20:52:15 krylon>

// This file has handlers for Ajax calls

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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

const searchPageCount = 250

func (srv *Server) handleAjaxSearchCreate(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	var (
		err    error
		msg    string
		sess   *sessions.Session
		db     *database.Database
		buf    bytes.Buffer
		search model.Search
		res    = model.Response{
			Payload: make(map[string]string),
		}
		hstatus int = 200
		q       chan model.Record
		data    tmplDataSearchResults
		hosts   []model.Host
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

	if _, err = io.Copy(&buf, r.Body); err != nil {
		res.Message = fmt.Sprintf("Failed to copy Request body: %s",
			err.Error())
		hstatus = 500
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	} else if err = json.Unmarshal(buf.Bytes(), &search.Query); err != nil {
		res.Message = fmt.Sprintf("Failed to parse search query: %s", err.Error())
		hstatus = 500
		srv.log.Printf("[ERROR] %s\n\n%s\n",
			res.Message,
			buf.String())
		goto SEND_RESPONSE
	} else if common.Debug {
		var patterns = make([]string, len(search.Query.Terms))
		for idx, pat := range search.Query.Terms {
			patterns[idx] = pat.String()
		}
		srv.log.Printf("[DEBUG] Search Terms = %#v\n",
			patterns)
	}

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if hosts, err = db.HostGetAll(); err != nil {
		res.Message = fmt.Sprintf("Failed to query all Hosts from database: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 500
		goto SEND_RESPONSE
	}

	data.Hostnames = make(map[int64]string, len(hosts))
	for _, h := range hosts {
		data.Hostnames[h.ID] = h.Name
	}

	q = make(chan model.Record)
	go db.RecordSearch(&search.Query, q)
	search.Results = make([]int64, 0, 32)

	for r := range q {
		search.Results = append(search.Results, r.ID)
	}

	search.Timestamp = time.Now()

	if err = db.SearchAdd(&search); err != nil {
		res.Message = fmt.Sprintf("Failed to create Search: %s",
			err.Error())
		hstatus = 500
		goto SEND_RESPONSE
	}

	res.Message = fmt.Sprintf("Search was performed and persisted successfully, yielding %d records",
		len(search.Results))
	res.Status = true
	res.Payload["id"] = strconv.FormatInt(search.ID, 10)
	res.Payload["cnt"] = strconv.FormatInt(int64(len(search.Results)), 10)

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
} // func (srv *Server) handleAjaxSearchCreate(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleAjaxSearchLoad(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	const tmplName = "search_results"
	var (
		err               error
		msg               string
		sess              *sessions.Session
		sid, page, offset int64
		db                *database.Database
		buf               bytes.Buffer
		tmpl              *template.Template
		res               = model.Response{
			Payload: make(map[string]string),
		}
		hstatus int = 200
		data    tmplDataSearchResults
		hosts   []model.Host
		vars    map[string]string
	)

	if sess, err = srv.store.Get(r, sessionNameFrontend); err != nil {
		res.Message = fmt.Sprintf(
			"Error getting/creating session %s: %s",
			sessionNameFrontend,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		sess = nil
		hstatus = 403
		goto SEND_RESPONSE
	} else if common.Debug {
		msg = dumpSession(sess)
		srv.log.Printf("[DEBUG] Existing session %s\n", msg)
	}

	vars = mux.Vars(r)

	if sid, err = strconv.ParseInt(vars["id"], 10, 64); err != nil {
		res.Message = fmt.Sprintf("Cannot parse Search ID %q: %s",
			vars["id"],
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 400
		goto SEND_RESPONSE
	} else if page, err = strconv.ParseInt(vars["page"], 10, 64); err != nil {
		res.Message = fmt.Sprintf("Cannot parse Page number %q: %s",
			vars["page"],
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 400
		goto SEND_RESPONSE
	} else if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		res.Message = fmt.Sprintf("Cannot find template %s",
			tmplName)
		srv.log.Printf("[CRITICAL] %s\n", res.Message)
		hstatus = 500
		goto SEND_RESPONSE
	}

	data.ID = sid
	data.Page = page
	offset = searchPageCount * (page - 1)
	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if hosts, err = db.HostGetAll(); err != nil {
		res.Message = fmt.Sprintf("Failed to query all Hosts from database: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 500
		goto SEND_RESPONSE
	} else if data.ResultCountTotal, err = db.SearchGetResultCount(sid); err != nil {
		res.Message = fmt.Sprintf("Failed to query number of results for Search %d: %s",
			sid,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 500
		goto SEND_RESPONSE
	}

	data.Hostnames = make(map[int64]string, len(hosts))
	for _, h := range hosts {
		data.Hostnames[h.ID] = h.Name
	}

	data.MaxPage = data.ResultCountTotal / searchPageCount
	if data.ResultCountTotal%searchPageCount != 0 {
		data.MaxPage++
	}

	if data.Records, err = db.SearchGetResults(sid, offset, searchPageCount); err != nil {
		res.Message = fmt.Sprintf("Error fetching Results for Search #%d: %s",
			sid,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 500
		goto SEND_RESPONSE
	} else if err = tmpl.Execute(&buf, &data); err != nil {
		res.Message = fmt.Sprintf("Error rendering results: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		hstatus = 500
		goto SEND_RESPONSE
	}

	res.Payload["results"] = buf.String()
	res.Message = "Success"
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
} // func (srv *Server) handleAjaxSearchLoad(w http.ResponseWriter, r *http.Request)
