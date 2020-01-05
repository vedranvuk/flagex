// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package reflag

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"strings"
	"testing"
)

func TestStruct(t *testing.T) {

	type (
		Sub struct {
			Name  string `json:"name"`
			Admin bool   `json:"admin"`
		}

		Main struct {
			Name  string
			EMail string
			Sub   *Sub `json:"sub"`
		}
	)

	args := "--name NameA --sub --name NameB --admin"

	main := &Main{Sub: &Sub{}}
	spew.Printf("Before:%+v\n", main)
	flags, err := FromStruct(main, strings.Split(args, " "))
	if err != nil {
		t.Fatal(err)
	}
	spew.Printf("After:%+v\n", main)
	fmt.Println(flags.Print())

}
