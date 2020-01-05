package flagex

import (
	"errors"
	"log"
	"strings"
	"testing"
)

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
	if err := f.Def("username", "u", "specify username", "guest", KindOptional); err != nil {
		t.Fatalf("Def '%s' failed ", "username")
	}
	if err := f.Def("ip", "", "specify ip address", "127.0.0.1", KindRequired); err != nil {
		t.Fatalf("Def '%s' failed ", "ip")
	}
	if err := f.Def("config", "c", "specify config file", "", KindRequired); err != nil {
		t.Fatalf("Def '%s' failed ", "config")
	}
	if err := f.Def("verbose", "v", "v for verbose", "", KindSwitch); err != nil {
		t.Fatalf("Def '%s' failed ", "verbose")
	}
	if err := f.Def("mode", "M", "use mode", "best", KindOptional); err != nil {
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
		Key      string
		ShortKey string
		Help     string
		Default  string
		Kind     FlagKind
		Sub      *Flags
	}

	var PackageItems = []FlagItem{
		FlagItem{
			"list",
			"l",
			"list packages",
			"",
			KindOptional,
			nil,
		},
		FlagItem{
			"export",
			"e",
			"export package list",
			"",
			KindOptional,
			nil,
		},
		FlagItem{
			"csv",
			"c",
			"use csv format",
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
		FlagItem{
			"clean",
			"c",
			"clean database",
			"",
			KindOptional,

			nil,
		},
		FlagItem{
			"backup",
			"b",
			"backup database",
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
		FlagItem{
			"install",
			"i",
			"install a package",
			"",
			KindOptional,

			nil,
		},
		FlagItem{
			"uninstall",
			"u",
			"uninstall a package",
			"",
			KindOptional,

			nil,
		},
		FlagItem{
			"update",
			"b",
			"update packages",
			"",
			KindOptional,
			nil,
		},
		FlagItem{
			"target",
			"t",
			"sync target (required)",
			"",
			KindRequired,
			nil,
		},
		FlagItem{
			"mode",
			"m",
			"sync mode",
			"best",
			KindRequired,
			nil,
		},
		FlagItem{
			"verbose",
			"v",
			"verbose output",
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
		pkg.Def(fi.Key, fi.ShortKey, fi.Help, fi.Default, fi.Kind)
	}
	pkg.Exclusive(PackageExcl...)

	dbs := New()
	for _, fi := range DatabaseItems {
		dbs.Def(fi.Key, fi.ShortKey, fi.Help, fi.Default, fi.Kind)
	}
	dbs.Exclusive(DatabaseExcl...)

	snc := New()
	for _, fi := range SyncItems {
		snc.Def(fi.Key, fi.ShortKey, fi.Help, fi.Default, fi.Kind)
	}
	snc.Exclusive(SyncExcl...)

	var RootItems = []FlagItem{
		FlagItem{
			"packages",
			"P",
			"work with packages",
			"",
			KindOptional,

			pkg,
		},
		FlagItem{
			"database",
			"D",
			"work with database",
			"",
			KindOptional,

			dbs,
		},
		FlagItem{
			"sync",
			"S",
			"package sync",
			"",
			KindOptional,

			snc,
		},
	}

	flag := New()
	for _, fi := range RootItems {
		flag.Sub(fi.Key, fi.ShortKey, fi.Help, fi.Sub)
	}

	type TestItem struct {
		Args        string
		ExpectedErr error
	}

	var TestItems = []TestItem{
		TestItem{"", ErrParams},
		TestItem{"-P", ErrSub},
		TestItem{"-P -l", nil},
		TestItem{"-P -e", nil},
		TestItem{"-P -l -c", nil},
		TestItem{"-P -e -c", nil},
		TestItem{"-P -l -e", ErrExcl},
		TestItem{"-D -c", nil},
		TestItem{"-D -b", nil},
		TestItem{"-D -c -b", ErrExcl},
		TestItem{"-S -i", ErrRequired},
		TestItem{"-S -t -i", ErrReqVal},
		TestItem{"-S -t target -m improved -u", nil},
		TestItem{"-S -t target -m new -i", nil},
		TestItem{"-S -t target -b", ErrRequired},
		TestItem{"-S -t target -i -u", ErrExcl},
		TestItem{"-S -t target -i -b", ErrExcl},
		TestItem{"-S -t target -u -b", ErrExcl},
		TestItem{"-S -t target -i -v extra", ErrSwitch},
		TestItem{"-S -i -v -t", ErrReqVal},
		TestItem{"-S -? -v", ErrNotFound},
		TestItem{"-S -v -?", ErrSwitch},
		TestItem{"-S -? -!", ErrNotFound},
		TestItem{"-S", ErrSub},
		TestItem{"-?", ErrNotFound},
		TestItem{"-S -t target -m new -v -v", ErrDupKey},
		TestItem{"-S", ErrSub},
		TestItem{"-S -S", ErrNotFound},
		TestItem{"-S -S -S", ErrNotFound},
		TestItem{"-? -?", ErrNotFound},
		TestItem{"-S -!", ErrNotFound},
		TestItem{"-S -t -v", ErrReqVal},
		TestItem{"-S -m mode -v", ErrRequired},
		TestItem{"-S -t -m mode -v", ErrReqVal},
		TestItem{"-S -t target -v", ErrRequired},
		TestItem{"-S -t target -m -v", ErrReqVal},
		TestItem{"-S -t target -m mode -v", nil},
		TestItem{"-S --target target -m mode -v", nil},
		TestItem{"-S -t target --mode mode -v", nil},
		TestItem{"-S --target target --mode mode -v", nil},
		TestItem{"-S --target target --mode mode -v extra", ErrSwitch},
		TestItem{"-S --target target --mode mode -v -v", ErrDupKey},
		TestItem{"-S --target target --mode mode -v --target target", ErrDupKey},
		TestItem{"-S --target target --mode mode -v --mode mode", ErrDupKey},
		TestItem{"-Svi --mode mode --target target", nil},
		TestItem{"-Plc", nil},
		TestItem{"-Dcb", ErrExcl},
		TestItem{"-Svt target -m mode", nil},
		TestItem{"-Svt target -m", ErrReqVal},
		TestItem{"-Svm mode --target", ErrReqVal},
		TestItem{"-Svm --?", ErrRequired},
		TestItem{"-Sv --mode any --target best", nil},
		TestItem{"-Sv --mode --? --target --!", nil},
		TestItem{"-Sib", ErrExcl},
		TestItem{"-S -i -b", ErrExcl},
	}

	for i := 0; i < len(TestItems); i++ {
		err := flag.Parse(strings.Split(TestItems[i].Args, " "))
		if !errors.Is(err, TestItems[i].ExpectedErr) {
			log.Fatalf("'%s': expected '%v', got '%v'", TestItems[i].Args, TestItems[i].ExpectedErr, err)
		}
	}
}
