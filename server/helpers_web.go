// /home/krylon/go/src/pepper/web/helpers_web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2019 by Benjamin Walkenhorst
// (c) 2019 Benjamin Walkenhorst
// Time-stamp: <2024-08-25 18:54:21 krylon>
//
// Helper functions for use by the HTTP request handlers

package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blicero/scrollmaster/common"
)

func errJSON(msg string) []byte { // nolint: unused,deadcode
	var res = fmt.Sprintf(
		`
{
    "Status": false,
    "Message": %q
    "Timestamp": %q,
}`,
		jsonEscape(msg),
		time.Now().Format(common.TimestampFormatSubSecond),
	)

	return []byte(res)
} // func errJSON(msg string) []byte

func jsonEscape(i string) string { // nolint: unused
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
}

// func getMimeType(path string) (string, error) {
// 	var (
// 		fh      *os.File
// 		err     error
// 		buffer  [512]byte
// 		byteCnt int
// 	)

// 	if fh, err = os.Open(path); err != nil {
// 		return "", err
// 	}

// 	defer fh.Close() // nolint: errcheck

// 	if byteCnt, err = fh.Read(buffer[:]); err != nil {
// 		return "", fmt.Errorf("cannot read from %s: %s",
// 			path,
// 			err.Error())
// 	}

// 	return http.DetectContentType(buffer[:byteCnt]), nil
// }
// func getMimeType(path string) (string, error)

func generateHostKey() (string, error) {
	var (
		err       error
		data      [keyLength]byte
		hex       string
		bytesRead int
	)

	if bytesRead, err = rand.Reader.Read(data[:]); err != nil {
		return "", err
	} else if bytesRead != keyLength {
		return "", fmt.Errorf("RNG returned insufficient data (%d/%d bytes)",
			bytesRead,
			keyLength)
	} else if hex, err = common.GetChecksum(data[:]); err != nil {
		return "", err
	}

	return hex, nil
} // func generateHostKey() (string, error)

func (srv *Server) baseData(title string, r *http.Request) tmplDataBase { // nolint: unused
	return tmplDataBase{
		Title: title,
		Debug: common.Debug,
		URL:   r.URL.String(),
	}
} // func (srv *Server) baseData(title string, r *http.Request) tmplDataBase
