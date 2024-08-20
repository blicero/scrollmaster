// /home/krylon/go/src/github.com/blicero/scrollmaster/server/server.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-20 18:06:38 krylon>

// Package server implements the server side of the application.
// It handles both talking to the Agents and the frontend.
package server

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/blicero/scrollmaster/database"
	"github.com/gorilla/mux"
)

const (
	poolSize = 4
	bufSize  = 32768
)

//go:embed assets
var assets embed.FS

// Server wraps the state required for the web interface
type Server struct {
	addr      string
	log       *log.Logger
	pool      *database.Pool
	lock      sync.RWMutex // nolint: unused,structcheck
	active    atomic.Bool
	router    *mux.Router
	tmpl      *template.Template
	web       http.Server
	mimeTypes map[string]string
}
