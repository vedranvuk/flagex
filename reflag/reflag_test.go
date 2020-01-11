// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package reflag

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vedranvuk/flagex"
)

var Verbose bool = false

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
	if Verbose {
		fmt.Printf("Before: %+v\n", main)
	}
	flags, err := Struct(main, strings.Split(args, " "))
	if err != nil {
		t.Fatal(err)
	}
	if Verbose {
		fmt.Printf("After:  %+v\n", main)
		fmt.Println("Parsed:", flags.ParseMap())
		fmt.Println("Print:")
		fmt.Println(flags.Print())
	}
}

func TestStruct2(t *testing.T) {

	type Tagged struct {
		FirstName string `json:"firstName,omitempty" reflag:"key=firstname,short=f,help=Your first name.,paramhelp=first name"`
		LastName  string `json:"lastName" reflag:"help=Your last name,paramhelp=last name"`
		Nickname  string `foo:"kickme,omitempty"`
	}

	type Child struct {
		*Tagged
		Index int
	}

	type Root struct {
		Child   Child
		Verbose bool
		Version bool `reflag:"short=V"`
	}

	type Test struct {
		Args     string
		Expected error
	}

	newData := func() *Root {
		return &Root{Child: Child{Tagged: &Tagged{"John", "Doe", "jd"}}}
	}

	tests := []Test{
		Test{"", flagex.ErrArgs},
		Test{"--verbose", nil},
		Test{"-v", nil},
		Test{"--version", nil},
		Test{"--version --verbose", nil},
		Test{"-v --version", nil},
		Test{"-v ", nil},
		Test{"-v -c -i 42", nil},
		Test{"-vci 42", flagex.ErrNotSub},
	}

	var data *Root
	var flags *flagex.Flags
	var err error
	for _, test := range tests {
		if Verbose {
			fmt.Printf("Parsing: '%s'\n", test.Args)
		}
		data = newData()
		flags, err = Struct(data, strings.Split(test.Args, " "))
		if !errors.Is(err, test.Expected) {
			t.Fatalf("fail('%s'): want '%s', got '%v'\n", test.Args, test.Expected, err)
		}
		if Verbose {
			if flags != nil {
				fmt.Println("Parsed: ", flags.ParseMap())
				fmt.Println(flags.Print())
			} else {
				fmt.Println("Parsed: <no parse>")
			}
			fmt.Println()
		}
	}

}

func init() {
	for _, v := range os.Args {
		if strings.HasPrefix(v, "-test.v") {
			Verbose = true
			return
		}
	}
}
