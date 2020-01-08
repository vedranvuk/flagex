// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package reflag

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"strings"
	"testing"
	"time"
)

func TestStruct(t *testing.T) {

	type (
		derivedint int64

		Solid struct {
			Surname string
		}

		Detail struct {
			Time time.Time
			User string
			Deep Solid
		}

		Sub struct {
			Nickname string `json:"FOOOO"`
			Admin    bool
			Detail   *Detail
		}

		Main struct {
			Name   string `json:"BAAAAR"`
			EMail  string
			Age    int
			Length derivedint
			// Error  *int
			Sub *Sub
		}
	)

	args := "--BAAAAR NameA --email me@net.com --age 64 --length 42 --sub --FOOOO NameB --admin --detail --user mirko --time 2020-01-02T15:04:05Z --deep --surname wut"

	main := &Main{Sub: &Sub{Detail: &Detail{}}}
	spew.Printf("Before: %+v\n", main)
	flags, err := Struct(main, strings.Split(args, " "))
	if err != nil {
		t.Fatal(err)
	}
	spew.Printf("After:  %+v\n", main)
	fmt.Println("Parsed:", flags.Parsed())
	fmt.Println("Print:")
	fmt.Println(flags.Print())
}
