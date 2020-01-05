// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package reflag

import (
	"fmt"
	"strings"
	"testing"
)

func TestStruct(t *testing.T) {

	type (
		Sub struct {
			Name string `json:"name"`
		}

		Main struct {
			Name  string
			EMail string
			Sub   *Sub `json:"sub"`
		}
	)

	args := "--name NameA --sub --name NameB"

	main := &Main{Sub: &Sub{}}
	flags, err := FromStruct(main, strings.Split(args, " "))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(flags.Print())

	fmt.Println("Before:", args)

}
