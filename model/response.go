// /home/krylon/go/src/github.com/blicero/scrollmaster/model/response.go
// -*- mode: go; coding: utf-8; -*-
// Created on 25. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-25 00:15:30 krylon>

package model

import "time"

// Response is what the Server sends to the Agent after handling a request.
type Response struct {
	Timestamp time.Time
	Status    bool
	Message   string
}
