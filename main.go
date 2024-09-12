// /home/krylon/go/src/github.com/blicero/scrollmaster/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-12 20:19:43 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/blicero/scrollmaster/agent"
	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/server"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	const defaultAddr = "::1"

	var (
		err      error
		addr     string
		mode     string
		basePath string
		port     int
	)

	flag.StringVar(
		&addr,
		"address",
		defaultAddr,
		"The IP address to either listen on or connect to")
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
	flag.StringVar(
		&basePath,
		"basedir",
		common.BaseDir,
		"The base directory to store application-specific files",
	)

	flag.Parse()

	if basePath != common.BaseDir {
		if err = common.SetBaseDir(basePath); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Error setting base directory to %s: %s\n",
				basePath,
				err.Error())
			os.Exit(1)
		}
	}

	if err = common.InitApp(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Cannot initialize directory %s: %s\n",
			common.BaseDir,
			err.Error(),
		)
		os.Exit(1)
	}

	switch strings.ToLower(mode) {
	case "server":
		// Be servile
		runServer(addr, port)
	case "agent":
		// Show some agency
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
	var (
		err     error
		ag      *agent.Agent
		srvAddr string
	)

	srvAddr = fmt.Sprintf("[%s]:%d",
		addr,
		port)

	if ag, err = agent.Create(srvAddr, ""); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Error creating Agent: %s\n",
			err.Error())
		os.Exit(2)
	}

	ag.Run()
} // func runAgent(addr string, port int)
