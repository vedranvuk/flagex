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
		Detail struct {
			Time time.Time
			User string
		}

		Sub struct {
			Name   string `json:"name"`
			Admin  bool   `json:"admin"`
			Detail *Detail
		}

		Main struct {
			Name  string
			EMail string
			Age   int
			Sub   *Sub `json:"sub"`
		}
	)

	args := "--name NameA --email me@net.com --age 64 --sub --name NameB --admin --detail --user mirko"

	main := &Main{Sub: &Sub{Detail: &Detail{}}}
	spew.Printf("Before:%+v\n", main)
	flags, err := FromStruct(main, strings.Split(args, " "))
	if err != nil {
		t.Fatal(err)
	}
	spew.Printf("After:%+v\n", main)
	fmt.Println(flags.Print())

}
