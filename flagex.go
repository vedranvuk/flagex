// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package flaxeg implements a command line parser.
// Not thread-safe.
package flagex

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/vedranvuk/errorex"
)

var (
	// ErrFlag is the base flagex error.
	ErrFlag = errorex.New("flagex")
	// ErrInvalid is returned when an invalid flag key is specified.
	ErrInvalid = ErrFlag.WrapFormat("invalid key")
	// ErrNotFound is returned when a non existent key is requested.
	ErrNotFound = ErrFlag.WrapFormat("key '%s' not found")
	// ErrDupKey is returned when a flag with a duplicate key is being registered.
	ErrDupKey = ErrFlag.WrapFormat("duplicate key '%s'")
	// ErrDupShort is returned when a flag with a duplicate shortkey is being registered.
	ErrDupShort = ErrFlag.WrapFormat("duplicate shortkey '%s'")
	// ErrExcl is returned when a more than flag from an exclusive set is parsed.
	ErrExcl = ErrFlag.WrapFormat("'%s' is exclusive to '%s'")
	// ErrRequired is returned when a required flag was not parsed.
	ErrRequired = ErrFlag.WrapFormat("required key '%s' not specified")
	// ErrReqVal is returned when no value was passed to a key that requires one.
	ErrReqVal = ErrFlag.WrapFormat("arg '%s' requires a param.")
	// ErrSwitch is returned when a switch was passed a param.
	ErrSwitch = ErrFlag.WrapFormat("switch '%s' takes no params")
	// ErrParams is returned when Parse is called with empty params.
	ErrParams = ErrFlag.Wrap("parse invoken with no params")
	// ErrSub is returned when a sub switsh was parsed with no args following it.
	ErrSub = ErrFlag.WrapFormat("sub '%s' invoken with no params")
)

// FlagKind specifies Flag kind.
type FlagKind byte

const (
	// KindOptional marks a flag as optional as well as its params.
	KindOptional FlagKind = iota
	// KindRequired marks a flag as required as well as is its params.
	KindRequired
	// KindSwitch marks a flag as optional that takes no params.
	// Flags with sub flags take no flags and are marked as KindSwitch.
	KindSwitch
	// KindSub marks a flag as a Flags subcategory prefix.
	KindSub
)

// String implements Stringer interface on FlagKind.
func (fk FlagKind) String() string {
	switch fk {
	case KindOptional:
		return "optional"
	case KindRequired:
		return "required"
	case KindSwitch:
		return "switch"
	case KindSub:
		return "sub"
	}
	return ""
}

// Flag represents a defined flag.
type Flag struct {
	key, shortkey, help, paramhelp, defval string

	kind      FlagKind
	sub       *Flags
	excl      bool
	parsed    bool
	parsedval bool
	value     string
}

// Key returns Flag key.
func (f *Flag) Key() string { return f.key }

// Shortkey returns Flag shortkey.
func (f *Flag) Shortkey() string { return f.shortkey }

// Help returns Flag help text.
func (f *Flag) Help() string { return f.help }

// ParamHelp returns Flag param help text.
func (f *Flag) ParamHelp() string { return f.paramhelp }

// Default is returned by Value if no value was parsed for this Flag.
func (f *Flag) Default() string { return f.defval }

// Kind returns Flag kind.
func (f *Flag) Kind() FlagKind { return f.kind }

// Excl returns if this flag is exclusive in Flags.
func (f *Flag) Excl() bool { return f.excl }

// Parsed returns if this Flag was parsed.
func (f *Flag) Parsed() bool { return f.parsed }

// ParsedVal returns if FLag value was parsed.
func (f *Flag) ParsedVal() bool { return f.parsedval }

// Kind returns current FLag value.
func (f *Flag) Value() string {
	if !f.parsed || !f.parsedval {
		return f.defval
	}
	return f.value
}

// Flags defines a set of unique flags.
type Flags struct {
	keys  map[string]*Flag
	short map[string]string
	last  string
}

// New creates a new Flags instance.
func New() *Flags {
	return &Flags{
		keys:  make(map[string]*Flag),
		short: make(map[string]string),
	}
}

// def defines a flag then returns it or an error.
func (f *Flags) def(key, shortkey, help, paramhelp, defval string, typ FlagKind) (*Flag, error) {
	if key == "" {
		return nil, ErrInvalid
	}
	if _, ok := f.keys[key]; ok {
		return nil, ErrDupKey.WithArgs(key)
	}
	if _, ok := f.short[shortkey]; shortkey != "" && ok {
		return nil, ErrDupShort.WithArgs(shortkey)
	}
	flag := &Flag{key, shortkey, help, paramhelp, defval, typ, nil, false, false, false, ""}
	f.keys[key] = flag
	if shortkey != "" {
		f.short[shortkey] = key
	}
	return flag, nil
}

// Switch defines an optional switch without a param.
func (f *Flags) Switch(key, shortkey, help string) (err error) {
	_, err = f.def(key, shortkey, help, "", "", KindSwitch)
	return
}

// Opt defines an optional flag with a required param.
func (f *Flags) Opt(key, shortkey, help, paramhelp, defval string) (err error) {
	_, err = f.def(key, shortkey, help, paramhelp, defval, KindOptional)
	return
}

// Req defines a required flag with a required param.
func (f *Flags) Req(key, shortkey, help, paramhelp, defval string) (err error) {
	_, err = f.def(key, shortkey, help, paramhelp, defval, KindRequired)
	return
}

// Def defines a flag under specified key and optional
// longkey with specified help and default value defval.
// key and shortkey must be unique in Flags, shortkey is optional.
// If a non-nil error is returned flag was not defined.
func (f *Flags) Def(key, shortkey, help, paramhelp, defval string, typ FlagKind) (err error) {
	_, err = f.def(key, shortkey, help, paramhelp, defval, typ)
	return
}

// Sub defines child Flags under specified key and optional shortkey which
// must be unique in these Flags. When invoken rest of params are passed to it.
// help defines the flag help. If a non-nil error is returned flag was not defined.
func (f *Flags) Sub(key, shortkey, help string, sub *Flags) error {
	flag, err := f.def(key, shortkey, help, "", "", KindSub)
	if err != nil {
		return err
	}
	flag.sub = sub
	return nil
}

// Exclusive sets specified keys as mutually exclusive in Flags.
// If more than one key from exclusive group are parsed, parse will error.
// Keys must already be defined.
// Subsequent calls redefine exclusivity.
func (f *Flags) Exclusive(keys ...string) error {
	for _, flag := range f.keys {
		flag.excl = false
	}
	for _, key := range keys {
		flag, ok := f.Key(key)
		if !ok {
			return ErrNotFound.WithArgs(key)
		}
		flag.excl = true
	}
	return nil
}

// Key returns Flag if under specified key and a truth if it exists.
func (f *Flags) Key(key string) (flag *Flag, truth bool) {
	flag, truth = f.keys[key]
	return
}

// Short returns Flag under specified shortkey and a truth if it exists.
func (f *Flags) Short(shortkey string) (flag *Flag, truth bool) {
	return f.Key(f.short[shortkey])
}

// reset resets values and parsed states to self.
func (f *Flags) reset() {
	for _, flag := range f.keys {
		flag.parsed = false
		flag.parsedval = false
		flag.value = ""
		if flag.sub != nil {
			flag.sub.reset()
		}
	}
}

// matchcombined matches a possibly multilevel combined key against defined Flags.
func (f *Flags) matchcombined(arg string) bool {
	var flag *Flag
	var ok bool
	for i := 0; i < len(arg); i++ {
		_, ok = f.Short(string(arg[i]))
		if !ok && flag != nil && flag.sub != nil {
			if len(arg[i:]) == 0 {
				return false
			}
			return flag.sub.matchcombined(arg[i:])
		}
		flag, _ = f.Short(string(arg[i]))
		return true
	}
	return false
}

// findflag finds a flag by key or shortkey from arg and
// returns it if found and truth if exists.
func (f *Flags) findflag(arg string) (*Flag, bool) {
	var flag *Flag
	var ok bool
	if strings.HasPrefix(arg, "-") {
		key := strings.TrimPrefix(arg, "-")
		if f.matchcombined(key) {
			flag, ok = f.Short(string(key[0]))
		} else {
			flag, ok = f.Short(key)
		}
		if ok {
			return flag, ok
		}

		if strings.HasPrefix(key, "-") {
			flag, ok = f.Key(strings.TrimPrefix(key, "-"))
		}
	}
	if !ok {
		return nil, false
	}
	return flag, true
}

// consume marks a flag as parsed and sets its value if not empty.
func (f *Flags) consume(key, value string) error {

	flag, ok := f.keys[key]
	if !ok {
		return ErrNotFound.WithArgs(key)
	}
	if flag.Parsed() {
		return ErrDupKey.WithArgs(key)
	}
	if flag.Excl() {
		for _, v := range f.keys {
			if v.Parsed() && v.Excl() {
				return ErrExcl.WithArgs(v.Key(), key)
			}
		}
	}
	flag.parsed = true
	if value != "" {
		flag.value = value
		flag.parsedval = true
	}
	return nil
}

// Parse parses specified args.
func (f *Flags) Parse(args []string) error {
	f.last = strings.Join(args, " ")
	f.reset()
	var flag *Flag
	var ok bool
	var saved string
	var arg string
	for i := 0; i < len(args); i++ {
		arg = strings.TrimSpace(args[i])
		if arg == "" {
			continue
		}
		flag, ok = f.findflag(arg)

		if !ok {
			if saved == "" {
				saved = arg
				continue
			}
			flag, ok = f.findflag(saved)
			if !ok {
				return ErrNotFound.WithArgs(saved)
			}
			if flag.Kind() == KindSwitch {
				return ErrSwitch.WithArgs(flag.Key())
			}
			if err := f.consume(flag.Key(), arg); err != nil {
				return err
			}
			saved = ""
			continue
		}

		if saved == "" {
			if flag.sub != nil {
				comb := f.matchcombined(arg)
				if i == len(args)-1 {
					if comb && len(arg[1:]) == 1 {
						return ErrSub.WithArgs(flag.Key())
					}
				}
				flag.parsed = true
				if comb {
					a := strings.Split(arg[1:], "")
					for i, v := range a {
						a[i] = "-" + v
					}
					return flag.sub.Parse(append(a[1:], args[i+1:]...))
				}
				return flag.sub.Parse(args[i+1:])
			}
			saved = args[i]
			continue
		}

		flag, ok = f.findflag(saved)
		if !ok {
			return ErrNotFound.WithArgs(saved)
		}
		if flag.Kind() == KindRequired {
			return ErrReqVal.WithArgs(saved)
		}
		if err := f.consume(flag.Key(), ""); err != nil {
			return err
		}
		saved = arg
	}
	if saved != "" {
		flag, ok := f.findflag(saved)
		if !ok {
			return ErrNotFound.WithArgs(saved)
		}
		if flag.Kind() == KindRequired {
			return ErrReqVal.WithArgs(saved)
		}
		if err := f.consume(flag.Key(), ""); err != nil {
			return err
		}
	}
	noparse := true
	for _, flag = range f.keys {
		if flag.Kind() == KindRequired && !flag.Parsed() {
			return ErrRequired.WithArgs(flag.Key())
		}
		if flag.Parsed() {
			noparse = false
		}
	}
	if noparse {
		return ErrParams
	}
	return nil
}

// printindent prints flags to w indented with indent.
func (f *Flags) printindent(w io.Writer, indent string) {
	fmt.Fprintf(w, "%s[Short]\t[Key]\t[Kind]\t[Help]\t\n", indent)
	for _, flag := range f.keys {
		val := flag.Key()
		if flag.paramhelp != "" {
			val = fmt.Sprintf("%s <%s>", val, flag.paramhelp)
		}
		if flag.Shortkey() == "" {
			fmt.Fprintf(w, "%s%s\t--%s\t%s\t%s\t\n", indent, "", val, flag.Kind(), flag.Help())
		} else {
			fmt.Fprintf(w, "%s-%s\t--%s\t%s\t%s\t\n", indent, flag.Shortkey(), val, flag.Kind(), flag.Help())
		}
		if flag.sub != nil {
			flag.sub.printindent(w, indent+"\t")
		}
	}
}

// Print returns a printable string of Flags.
func (f *Flags) Print() string {
	buf := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', 0)
	f.printindent(w, "")
	w.Flush()
	return string(buf.Bytes())
}

// Parsed returns a map of parsed Flag key:value pairs.
// Sub will return a map, Flags may return a string if parsed or
// nil if not parsed. Parsed returns whichever args were parsed
// at last Parse. Parsed is as valid as what Parse returned.
func (f *Flags) Parsed() map[interface{}]interface{} {
	ret := make(map[interface{}]interface{})
	for kk, kv := range f.keys {
		if kv.Parsed() {
			if kv.sub != nil {
				_, ok := ret[kk]
				if !ok {
					ret[kk] = make(map[interface{}]interface{})
				}
				ret[kk] = kv.sub.Parsed()
				continue
			}
			if kv.ParsedVal() {
				ret[kk] = kv.Value()
			} else {
				ret[kk] = nil
			}
		}
	}
	return ret
}
