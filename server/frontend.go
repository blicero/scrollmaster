// /home/krylon/go/src/github.com/blicero/scrollmaster/server/frontend.go
// -*- mode: go; coding: utf-8; -*-
// Created on 05. 09. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-11 20:12:22 krylon>
//
// This file contains handlers etc. having to do with the web-based frontend.

package server

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/blicero/scrollmaster/database"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func (srv *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)
	const tmplName = "main"
	var (
		err  error
		msg  string
		tmpl *template.Template
		db   *database.Database
		sess *sessions.Session
		data = tmplDataIndex{
			tmplDataBase: tmplDataBase{
				Title: "Main",
				Debug: true,
				URL:   r.URL.EscapedPath(),
			},
		}
	)

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if sess, err = srv.store.Get(r, sessionNameFrontend); err != nil {
		msg = fmt.Sprintf("Error getting client session from session store: %s",
			err.Error())
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		msg = fmt.Sprintf("Could not find template %q", tmplName)
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Hosts, err = db.HostGetAll(); err != nil {
		msg = fmt.Sprintf("Failed to query all Hosts from database: %s", err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
	}

	if err = sess.Save(r, w); err != nil {
		srv.log.Printf("[ERROR] Failed to set session cookie: %s\n",
			err.Error())
	}
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	if err = tmpl.Execute(w, &data); err != nil {
		msg = fmt.Sprintf("Error rendering template %q: %s",
			tmplName,
			err.Error())
		srv.sendErrorMessage(w, msg)
	}
} // func (srv *Server) handleMain(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleLogRecent(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	const tmplName = "log"
	var (
		err       error
		msg, nstr string
		tmpl      *template.Template
		db        *database.Database
		cnt       int64
		sess      *sessions.Session
		vars      map[string]string
		data      = tmplDataLog{
			tmplDataBase: tmplDataBase{
				Title: "Main",
				Debug: true,
				URL:   r.URL.EscapedPath(),
			},
		}
	)

	vars = mux.Vars(r)
	db = srv.pool.Get()
	defer srv.pool.Put(db)

	nstr = vars["cnt"]
	if nstr == "" {
		nstr = "500"
	}

	if cnt, err = strconv.ParseInt(nstr, 10, 64); err != nil {
		msg = fmt.Sprintf("Failed to parse number of records to display (%q): %s",
			nstr,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if sess, err = srv.store.Get(r, sessionNameFrontend); err != nil {
		msg = fmt.Sprintf("Error getting client session from session store: %s",
			err.Error())
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		msg = fmt.Sprintf("Could not find template %q", tmplName)
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Hosts, err = db.HostGetAll(); err != nil {
		msg = fmt.Sprintf("Failed to query all Hosts from database: %s", err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Records, err = db.RecordGetRecent(cnt); err != nil {
		msg = fmt.Sprintf("Failed to get query %d most recent records from database: %s",
			cnt,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	data.Hostnames = make(map[int64]string, len(data.Hosts))
	for _, h := range data.Hosts {
		data.Hostnames[h.ID] = h.NameShort()
	}

	data.Sources = make([]string, 0, 16)
	var srcMap = make(map[string]bool)

	for _, r := range data.Records {
		var ok bool
		if _, ok = srcMap[r.Source]; !ok {
			srcMap[r.Source] = true
			data.Sources = append(data.Sources, r.Source)
		}
	}

	slices.Sort(data.Sources)

	if err = sess.Save(r, w); err != nil {
		srv.log.Printf("[ERROR] Failed to set session cookie: %s\n",
			err.Error())
	}
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	if err = tmpl.Execute(w, &data); err != nil {
		msg = fmt.Sprintf("Error rendering template %q: %s",
			tmplName,
			err.Error())
		srv.sendErrorMessage(w, msg)
	}
} // func (srv *Server)  handleLogRecent(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)
	const tmplName = "search"
	var (
		err  error
		msg  string
		tmpl *template.Template
		db   *database.Database
		sess *sessions.Session
		data = tmplDataSearch{
			tmplDataBase: tmplDataBase{
				Title: "Main",
				Debug: true,
				URL:   r.URL.EscapedPath(),
			},
			Begin: time.Unix(0, 0),
			End:   time.Now(),
		}
	)

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if sess, err = srv.store.Get(r, sessionNameFrontend); err != nil {
		msg = fmt.Sprintf("Error getting client session from session store: %s",
			err.Error())
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		msg = fmt.Sprintf("Could not find template %q", tmplName)
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Hosts, err = db.HostGetAll(); err != nil {
		msg = fmt.Sprintf("Failed to query all Hosts from database: %s", err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Sources, err = db.RecordGetSources(); err != nil {
		msg = fmt.Sprintf("Failed to query all record sources from database: %s", err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Searches, err = db.SearchGetAllID(); err != nil {
		msg = fmt.Sprintf("Failed to query all search IDs: %s", err.Error())
		srv.log.Printf("[ERROR] %s\n",
			msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	if err = sess.Save(r, w); err != nil {
		srv.log.Printf("[ERROR] Failed to set session cookie: %s\n",
			err.Error())
	}
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	if err = tmpl.Execute(w, &data); err != nil {
		msg = fmt.Sprintf("Error rendering template %q: %s",
			tmplName,
			err.Error())
		srv.sendErrorMessage(w, msg)
	}
} // func (srv *Server) handleSearch(w http.ResponseWriter, r *http.Request)
