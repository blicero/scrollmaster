// /home/krylon/go/src/github.com/blicero/scrollmaster/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-08-15 19:18:32 krylon>

package main

import (
	"fmt"

	"github.com/blicero/scrollmaster/common"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	fmt.Println("IMPLEMENT ME!")
} // func main()
