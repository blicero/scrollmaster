// /home/krylon/go/src/github.com/blicero/krylib/id.go
// -*- mode: go; coding: utf-8; -*-
// Created on 25. 10. 2022 by Benjamin Walkenhorst
// (c) 2022 Benjamin Walkenhorst
// Time-stamp: <2022-10-25 01:58:02 krylon>

package krylib

// ID is an identifier that is unique within a given context.
type ID int64

// INVALID_ID is an ID that does not equal any valid ID.
const INVALID_ID ID = -1
