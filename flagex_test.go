// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package flagex

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

var Verbose = false

func TestFlags(t *testing.T) {

	var (
		Args = []string{
			"--config",
			"config.json",
			"--ip",
			"10.0.0.1",
			"-v",
			"-M",
		}
	)

	f := New()
	if err := f.Def("username", "u", "specify username", "username", "guest", KindOptional); err != nil {
		t.Fatalf("Def '%s' failed ", "username")
	}
	if err := f.Def("ip", "", "specify ip address", "ip", "127.0.0.1", KindRequired); err != nil {
		t.Fatalf("Def '%s' failed ", "ip")
	}
	if err := f.Def("config", "c", "specify config file", "filename", "", KindRequired); err != nil {
		t.Fatalf("Def '%s' failed ", "config")
	}
	if err := f.Def("verbose", "v", "v for verbose", "", "", KindSwitch); err != nil {
		t.Fatalf("Def '%s' failed ", "verbose")
	}
	if err := f.Def("mode", "M", "use mode", "mode ", "best", KindOptional); err != nil {
		t.Fatalf("Def '%s' failed ", "version")
	}

	if err := f.Parse(Args); err != nil {
		t.Fatal(err)
	}

	flag, ok := f.Key("username")
	if !ok {
		t.Fatal()
	}
	if flag.Parsed() {
		t.Fatal()
	}
	if flag.Value() != "guest" {
		t.Fatal()
	}

	if _, ok = f.Short("u"); !ok {
		t.Fatal()
	}

	flag, ok = f.Key("ip")
	if !ok {
		t.Fatal()
	}
	if !flag.Parsed() {
		t.Fatal()
	}
	if flag.Value() != "10.0.0.1" {
		t.Fatal()
	}

	flag, ok = f.Key("config")
	if !ok {
		t.Fatal()
	}
	if !flag.Parsed() {
		t.Fatal()
	}
	if flag.Value() != "config.json" {
		t.Fatal()
	}

	if _, ok = f.Short("c"); !ok {
		t.Fatal()
	}

	flag, ok = f.Key("verbose")
	if !ok {
		t.Fatal()
	}
	if !flag.Parsed() {
		t.Fatal()
	}
	if v := flag.Value(); v != "" {
		t.Fatal(v)
	}

	if _, ok = f.Short("v"); !ok {
		t.Fatal()
	}

	flag, ok = f.Key("mode")
	if !ok {
		t.Fatal()
	}
	if !flag.Parsed() {
		t.Fatal()
	}
	if v := flag.Value(); v != "best" {
		t.Fatal(v)
	}

	if _, ok = f.Short("M"); !ok {
		t.Fatal("h")
	}
}

func TestMux(t *testing.T) {

	type FlagItem struct {
		Key       string
		ShortKey  string
		Help      string
		ParamHelp string
		Default   string
		Kind      FlagKind
		Sub       *Flags
	}

	var PackageItems = []FlagItem{
		{
			"list",
			"l",
			"list packages",
			"",
			"",
			KindOptional,
			nil,
		},
		{
			"export",
			"e",
			"export package list",
			"",
			"",
			KindOptional,
			nil,
		},
		{
			"csv",
			"c",
			"use csv format",
			"",
			"",
			KindOptional,
			nil,
		},
	}
	var PackageExcl = []string{
		"list",
		"export",
	}

	var DatabaseItems = []FlagItem{
		{
			"clean",
			"c",
			"clean database",
			"",
			"",
			KindOptional,

			nil,
		},
		{
			"backup",
			"b",
			"backup database",
			"",
			"",
			KindOptional,

			nil,
		},
	}
	var DatabaseExcl = []string{
		"clean",
		"backup",
	}

	var SyncItems = []FlagItem{
		{
			"install",
			"i",
			"install a package",
			"package name",
			"",
			KindOptional,

			nil,
		},
		{
			"uninstall",
			"u",
			"uninstall a package",
			"package name",
			"",
			KindOptional,

			nil,
		},
		{
			"update",
			"b",
			"update packages",
			"package name",
			"",
			KindOptional,
			nil,
		},
		{
			"target",
			"t",
			"sync target (required)",
			"targetname",
			"",
			KindRequired,
			nil,
		},
		{
			"mode",
			"m",
			"sync mode",
			"mode",
			"best",
			KindRequired,
			nil,
		},
		{
			"verbose",
			"v",
			"verbose output",
			"",
			"",
			KindSwitch,
			nil,
		},
	}

	var SyncExcl = []string{
		"install",
		"uninstall",
		"update",
	}

	pkg := New()
	for _, fi := range PackageItems {
		pkg.Def(fi.Key, fi.ShortKey, fi.Help, fi.ParamHelp, fi.Default, fi.Kind)
	}
	pkg.Exclusive(PackageExcl...)

	dbs := New()
	for _, fi := range DatabaseItems {
		dbs.Def(fi.Key, fi.ShortKey, fi.Help, fi.ParamHelp, fi.Default, fi.Kind)
	}
	dbs.Exclusive(DatabaseExcl...)

	snc := New()
	for _, fi := range SyncItems {
		snc.Def(fi.Key, fi.ShortKey, fi.Help, fi.ParamHelp, fi.Default, fi.Kind)
	}
	snc.Exclusive(SyncExcl...)

	var RootItems = []FlagItem{
		{
			"packages",
			"P",
			"work with packages",
			"",
			"",
			KindOptional,
			pkg,
		},
		{
			"database",
			"D",
			"work with database",
			"",
			"",
			KindOptional,
			dbs,
		},
		{
			"sync",
			"S",
			"package sync",
			"",
			"",
			KindOptional,
			snc,
		},
		{
			"verbose",
			"v",
			"verbose output",
			"",
			"",
			KindSwitch,
			nil,
		},
	}

	flag := New()
	flag.Sub(RootItems[0].Key, RootItems[0].ShortKey, RootItems[0].Help, RootItems[0].Sub)
	flag.Sub(RootItems[1].Key, RootItems[1].ShortKey, RootItems[1].Help, RootItems[1].Sub)
	flag.Sub(RootItems[2].Key, RootItems[2].ShortKey, RootItems[2].Help, RootItems[2].Sub)
	flag.Switch(RootItems[3].Key, RootItems[3].ShortKey, RootItems[3].Help)

	type TestItem struct {
		Args        string
		ExpectedErr error
	}

	var TestItems = []TestItem{
		{"", ErrNoArgs},
		{"-", ErrNotFound},
		{"--", ErrNotFound},
		{"-P", ErrSub},
		{"-P -l", nil},
		{"-P -e", nil},
		{"-P -l -c", nil},
		{"-P -e -c", nil},
		{"-P -l -e", ErrExclusive},
		{"-D -c", nil},
		{"-D -b", nil},
		{"-D -c -b", ErrExclusive},
		{"-S -i", ErrRequired},
		{"-S -t -i", ErrReqVal},
		{"-S -t target -m improved -u", nil},
		{"-S -t target -m new -i", nil},
		{"-S -t target -b", ErrRequired},
		{"-S -t target -i -u", ErrExclusive},
		{"-S -t target -i -b", ErrExclusive},
		{"-S -t target -u -b", ErrExclusive},
		{"-S -t target -i -v extra", ErrSwitch},
		{"-S -i -v -t", ErrReqVal},
		{"-S -? -v", ErrNotFound},
		{"-S -v -?", ErrSwitch},
		{"-S -? -!", ErrNotFound},
		{"-S", ErrSub},
		{"-?", ErrNotFound},
		{"-S -t target -m new -v -v", ErrDuplicate},
		{"-S", ErrSub},
		{"-S -S", ErrNotFound},
		{"-S -S -S", ErrNotFound},
		{"-? -?", ErrNotFound},
		{"-S -!", ErrNotFound},
		{"-S -t -v", ErrReqVal},
		{"-S -m mode -v", ErrRequired},
		{"-S -t -m mode -v", ErrReqVal},
		{"-S -t target -v", ErrRequired},
		{"-S -t target -m -v", ErrReqVal},
		{"-S -t target -m mode -v", nil},
		{"-S --target target -m mode -v", nil},
		{"-S -t target --mode mode -v", nil},
		{"-S --target target --mode mode -v", nil},
		{"-S --target target --mode mode -v extra", ErrSwitch},
		{"-S --target target --mode mode -v -v", ErrDuplicate},
		{"-S --target target --mode mode -v --target target", ErrDuplicate},
		{"-S --target target --mode mode -v --mode mode", ErrDuplicate},
		{"-Svi --mode mode --target target", nil},
		{"-Plc", nil},
		{"-Dcb", ErrExclusive},
		{"-Svt target -m mode", nil},
		{"-Svt target -m", ErrReqVal},
		{"-Svm mode --target", ErrReqVal},
		{"-Svm --?", ErrRequired},
		{"-Sv --mode any --target best", nil},
		{"-Sv --mode --? --target --!", nil},
		{"-Sib", ErrExclusive},
		{"-S -i -b", ErrExclusive},
		{"-vSi", ErrNotSub},
		{"-v -Sit target -m mode", nil},
		{"-v -Si -t target -m mode", nil},
		{"-PDS", ErrNotFound},
		{"-PSD", ErrNotFound},
		{"-SPD", ErrNotFound},
		{"-SDP", ErrNotFound},
		{"-DSP", ErrNotFound},
		{"-DPS", ErrNotFound},
	}

	for i := 0; i < len(TestItems); i++ {
		err := flag.Parse(strings.Split(TestItems[i].Args, " "))
		if Verbose {
			fmt.Printf("Testing: '%s'\n", TestItems[i].Args)
		}
		if !errors.Is(err, TestItems[i].ExpectedErr) {
			log.Fatalf("'%s': expected '%v', got '%v'", TestItems[i].Args, TestItems[i].ExpectedErr, err)
		}
		if Verbose {
			fmt.Printf("Result:  '%v'\n", err)
			fmt.Printf("Parsed:  '%#v'\n", flag.ParseMap())
			fmt.Println()
		}
	}
	if Verbose {
		fmt.Println(flag.String())
	}
}

func TestParsed(t *testing.T) {
	f := New()
	f.Switch("a", "a", "a")
	f.Switch("b", "b", "b")
	f.Switch("c", "c", "c")

	f.Parse([]string{"-a", "-b", "-c"})
	if !f.Parsed() {
		t.Fatal("parsed failed")
	}

	f.Parse([]string{"-a", "-c"})
	if !f.Parsed("a", "c") {
		t.Fatal("parsed failed")
	}
	if f.Parsed("b") {
		t.Fatal("parsed failed")
	}
	if f.Parsed("x") {
		t.Fatal("parsed failed")
	}

	f.Parse([]string{"-b"})
	if !f.Parsed("b") {
		t.Fatal("parsed failed")
	}
	if f.Parsed("a") {
		t.Fatal("parsed failed")
	}

	if f.Parsed("c") {
		t.Fatal("parsed failed")
	}
	if f.Parsed("a", "c") {
		t.Fatal("parsed failed")
	}
}

func TestValue(t *testing.T) {
	f := New()
	f.Opt("test", "t", "a test switch", "string", "defval")
	f.Parse([]string{"-t", "notdefval"})
	if f.Value("test") != "notdefval" {
		t.Fatal("Value() failed")
	}
	if f.Value("doesnotexist") != "" {
		t.Fatal("Value() failed")
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
