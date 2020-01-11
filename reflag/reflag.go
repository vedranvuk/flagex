// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package reflag provides reflect based helpers for flagex.
package reflag

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"

	"github.com/vedranvuk/errorex"
	"github.com/vedranvuk/flagex"
	"github.com/vedranvuk/reflectex"
)

var (
	// ErrReflag is the base error from reflag package.
	ErrReflag = errorex.New("reflag")
	// ErrConvert is returned when an arg was not convertable to a value.
	ErrConvert = ErrReflag.WrapFormat("error converting arg '%s' value '%s' to '%s'")
	// ErrUnmarshal is returned when an arg was not unmarshalable to a value.
	ErrUnmarshal = ErrReflag.WrapFormat("error unmarshaling arg '%s' value '%s' to '%s'")
	// ErrNotFound is returned when applying to to a struct with a field not defined in flags.
	ErrNotFound = ErrReflag.WrapFormat("no flags defined for field '%s'")
	// ErrParam is returned when an invalid parameter is encountered.
	ErrParam = ErrReflag.Wrap("invalid parameter")
)

const (
	// KeyJSON identifies a JSON tag in a struct field.
	KeyJSON = "json"
	// KeyReflag identifies a reflag tag in a struct field.
	KeyReflag = "reflag"
	// KeyKey identifies a key for a flag in a struct field.
	KeyKey = "key"
	// KeyShort identifies a shortkey for a flag in a struct field.
	KeyShort = "short"
	// KeyHelp identifies help for a flag in a struct field.
	KeyHelp = "help"
	// KeyParamHelp identifies parameter help for a flag in a struct field.
	KeyParamHelp = "paramhelp"
)

// namefromjsontag retrieves the name from a json tag.
// If no name found, result is an empty string.
func namefromjsontag(tag string) string {
	a := strings.Split(tag, ",")
	if len(a) > 0 {
		return a[0]
	}
	return ""
}

// reflagtagtomap parses out comma-sepparated key/value pairs from a reflag tag.
// If a key has no value it is added as a key with an empty value.
func reflagtagtomap(tag string) (m map[string]string) {
	m = make(map[string]string)
	for _, pair := range strings.Split(tag, ",") {
		switch a := strings.Split(pair, "="); len(a) {
		case 1:
			m[a[0]] = ""
		case 2:
			if a[0] == "" || a[1] == "" {
				return nil
			}
			m[a[0]] = a[1]
		}
	}
	return
}

// flagParamsFromField returns parameters for defining a flag from a StructField f.
func flagParamsFromField(f reflect.StructField) (key, shortkey, help, paramhelp string) {

	if rftag, ok := f.Tag.Lookup(KeyReflag); ok {
		m := reflagtagtomap(rftag)
		key = m[KeyKey]
		shortkey = m[KeyShort]
		help = m[KeyHelp]
		paramhelp = m[KeyParamHelp]
	}
	if key == "" {
		if jstag, ok := f.Tag.Lookup(KeyJSON); ok {
			key = namefromjsontag(jstag)
		}
	}
	if key == "" {
		key = strings.ToLower(f.Name)
	}
	if shortkey == "" {
		shortkey = string(key[0])
	}
	return
}

// flagsFromStruct creates Flags from struct v.
// Parsing is multilevel, root and v are initial flags and a struct.
func flagsFromStruct(root *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {

	v = reflect.Indirect(v)

	for i := 0; i < v.NumField(); i++ {

		if !v.Field(i).CanSet() {
			continue
		}

		key, shortkey, help, paramhelp := flagParamsFromField(v.Type().Field(i))
		if _, ok := root.Short(shortkey); ok {
			shortkey = ""
		}
		fldval := reflect.Indirect(v.Field(i))

		_, ok := (fldval.Interface()).(encoding.TextMarshaler)
		if ok {
			if err := root.Def(key, shortkey, help, paramhelp, "", flagex.KindOptional); err != nil {
				return nil, err
			}
			continue
		}

		switch fldval.Kind() {
		case reflect.Struct:
			new, err := flagsFromStruct(flagex.New(), v.Field(i))
			if err != nil {
				return nil, err
			}
			if err := root.Sub(key, shortkey, help, new); err != nil {
				return nil, err
			}
		case reflect.Bool:
			if err := root.Switch(key, shortkey, help); err != nil {
				return nil, err
			}
		default:
			if err := root.Def(key, shortkey, help, paramhelp, "", flagex.KindOptional); err != nil {
				return nil, err
			}
		}
	}

	return root, nil
}

// structApplyFlags applies pared values in flags to struct v.
func structApplyFlags(flags *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {

	for i := 0; i < v.NumField(); i++ {

		if !v.Field(i).CanSet() {
			continue
		}

		key, _, _, _ := flagParamsFromField(v.Type().Field(i))
		flag, ok := flags.Key(key)
		if !ok {
			return nil, ErrNotFound.WithArgs(key)
		}
		if !flag.Parsed() {
			continue
		}

		fldval := reflect.Indirect(v.Field(i))
		if fldval.Kind() == reflect.Ptr {
			fmt.Println("don't panic")
		}

		intf, ok := (fldval.Addr().Interface()).(encoding.TextUnmarshaler)
		if ok {
			if err := intf.UnmarshalText([]byte(flag.Value())); err != nil {
				return nil, ErrUnmarshal.CauseArgs(err, key, flag.Value(), fldval.Type().Name())
			}
			continue
		}

		if fldval.Kind() == reflect.Struct {
			_, err := structApplyFlags(flag.Sub(), fldval)
			if err != nil {
				return nil, err
			}
		}
		val := reflect.Indirect(reflect.New(fldval.Type()))
		var err error
		if flag.Sub() == nil {
			if flag.Kind() == flagex.KindSwitch {
				err = reflectex.StringToValue("true", val)
			} else {
				err = reflectex.StringToValue(flag.Value(), val)
			}
			if err != nil {
				return nil, ErrConvert.CauseArgs(err, key, flag.Value(), fldval.Type().Name())
			}
			fldval.Set(val)
		}
	}
	return flags, nil
}

// FromStruct creates Flags from struct v and returns it or an error.
// See Struct for more details.
func FromStruct(v interface{}) (*flagex.Flags, error) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.IsValid() {
		return nil, ErrParam
	}
	if rv.Kind() != reflect.Struct {
		return nil, ErrParam
	}
	flags, err := flagsFromStruct(flagex.New(), rv)
	if err != nil {
		return nil, err
	}
	return flags, nil
}

// ToStruct applies Flags (prefferably gotten with FromStruct) to a
// struct (preferrably one used with FromStruct).
// Flags should be Parse()'d with no error. See Struct for more details.
func ToStruct(v interface{}, flags *flagex.Flags) error {
	_, err := structApplyFlags(flags, reflect.Indirect(reflect.ValueOf(v)))
	if err != nil {
		return err
	}
	return nil
}

// Struct takes a struct v, creates flags from v's exported fields,
// parses flags from args, sets v fields values to parsed values then
// returns flags or an error if one occured.
//
// Struct v can have nested structs but any pointer types inside at any
// level must have been allocated prior to this call. It makes sense to
// allocate v fully, pass it to Struct() to have any field values changed
// by parsing args then to check v's values. To check if a field was
// changed use flags's Parsed() function.
//
// Struct prefers json tags from v's Fields to specify defined flag names.
// If none, field names are lowercased and defined in flags as such.
// First letter of derived key is used for shortkey, until uniques are
// exhausted.
// Shortkey definitions can differ in Struct calls if there would be
// duplicates due to key names because of the random nature of go's map
// iteration.
// All unparsed fields are always ommitted when setting field values.
//
// Syntax for specifying array, slice or map values is described by
// StringToInterface function from reflectex which this function uses to
// convert string values to v's field values.
func Struct(v interface{}, args []string) (*flagex.Flags, error) {
	flags, err := FromStruct(v)
	if err != nil {
		return nil, err
	}
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if err := ToStruct(v, flags); err != nil {
		return nil, err
	}
	return flags, nil
}
