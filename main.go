// /home/krylon/go/src/github.com/blicero/scrollmaster/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-03 15:15:00 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/server"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	fmt.Println("IMPLEMENT ME!")

	const defaultAddr = "::1"

	var (
		err  error
		addr string
		mode string
		port int
	)

	flag.StringVar(
		&addr,
		"address",
		defaultAddr,
		"The IP address to either listen on or connect to.")
	flag.StringVar(
		&mode,
		"mode",
		"",
		"Are we the server or the agent?")
	flag.IntVar(
		&port,
		"port",
		common.Port,
		"The TCP port to listen on or connect to")

	flag.Parse()

	if err = common.InitApp(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Cannot initialize directory %s: %s\n",
			common.BaseDir,
			err.Error(),
		)
		os.Exit(1)
	}

	addr = fmt.Sprintf("[%s]:%d",
		addr,
		port)

	switch strings.ToLower(mode) {
	case "server":
		// Be servile
		runServer(addr, port)
	case "agent":
		// Display some agency
		runAgent(addr, port)
	default:
		fmt.Fprintf(
			os.Stderr,
			"Invalid mode %q (must be either \"agent\" or \"server\")\n",
			mode)
		os.Exit(1)
	}
} // func main()

func runServer(addr string, port int) {
	var (
		err   error
		srv   *server.Server
		laddr string
	)

	laddr = fmt.Sprintf("[%s]:%d",
		addr,
		port)

	if srv, err = server.Create(laddr); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to create Server: %s\n",
			err.Error())
		os.Exit(2)
	}

	srv.ListenAndServe()
}

func runAgent(addr string, port int) {
	fmt.Println("IMPLEMENT ME!")
}
