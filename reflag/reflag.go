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
	ErrNotFound = ErrReflag.WrapFormat("no flags defined for field '%s', tag '%s'")
	// ErrParam is returned when an invalid parameter is encountered.
	ErrParam = ErrReflag.Wrap("invalid parameter")
)

func parsetag(tag string) string {
	a := strings.Split(tag, ",")
	if len(a) > 0 {
		return a[0]
	}
	return ""
}

// flagsFromStruct creates Flags from struct v.
// Parsing is multilevel, root and v are initial flags and a struct.
func flagsFromStruct(root *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {

	v = reflect.Indirect(v)

	for i := 0; i < v.NumField(); i++ {
		fldname := strings.ToLower(v.Type().Field(i).Name)
		fldtag := parsetag(v.Type().Field(i).Tag.Get("json"))
		fldval := reflect.Indirect(v.Field(i))
		paramhelp := strings.ToLower(v.Field(i).Type().Name())
		flaghelp := v.Type().Field(i).Name
		key := fldname
		if fldtag != "" {
			key = fldtag
		}
		flagshort := string(key[0])
		if _, ok := root.Short(flagshort); ok {
			flagshort = ""
		}

		// fmt.Printf("Field: Name:'%s', Tag:'%s', Val:'%v', Short:'%s', Paramhelp:'%s', flaghelp:'%s', Key:'%s'\n",
		// 	fldname, fldtag, fldval, flagshort, paramhelp, flaghelp, key)

		_, ok := (fldval.Interface()).(encoding.TextMarshaler)
		if ok {
			if err := root.Def(key, flagshort, flaghelp, paramhelp, "", flagex.KindOptional); err != nil {
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
			if err := root.Sub(key, flagshort, flaghelp, new); err != nil {
				return nil, err
			}
		case reflect.Bool:
			if err := root.Switch(key, flagshort, flaghelp); err != nil {
				return nil, err
			}
		default:
			if err := root.Def(key, flagshort, flaghelp, paramhelp, "", flagex.KindOptional); err != nil {
				return nil, err
			}
		}
	}

	return root, nil
}

// structApplyFlags applies pared values in flags to struct v.
func structApplyFlags(flags *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {

	var flag *flagex.Flag
	for i := 0; i < v.NumField(); i++ {

		ok := false
		fv := reflect.Indirect(v.Field(i))
		fldname := strings.ToLower(v.Type().Field(i).Name)
		fldtag := parsetag(v.Type().Field(i).Tag.Get("json"))

		if fldtag != "" {
			flag, ok = flags.Key(fldtag)
		}
		if !ok {
			flag, ok = flags.Key(fldname)
		}
		if !ok {
			return nil, ErrNotFound.WithArgs(fldname, fldtag)
		}
		if !flag.Parsed() {
			continue
		}

		// fmt.Printf("ApplyField: Name'%s', Tag:'%s', FKey:'%s', Parsed:'%t', Value:'%s'\n ",
		// 	fldname, fldtag, flag.Key(), flag.Parsed(), flag.Value())

		if fv.Kind() == reflect.Ptr {
			fmt.Println("don't panic")
		}

		intf, ok := (fv.Addr().Interface()).(encoding.TextUnmarshaler)
		if ok {
			if err := intf.UnmarshalText([]byte(flag.Value())); err != nil {
				return nil, ErrUnmarshal.CauseArgs(err, fldname, flag.Value(), fv.Type().Name())
			}
			continue
		}

		if fv.Kind() == reflect.Struct {
			_, err := structApplyFlags(flag.Sub(), fv)
			if err != nil {
				return nil, err
			}
		}
		val := reflect.Indirect(reflect.New(fv.Type()))
		var err error
		if flag.Sub() == nil {
			if flag.Kind() == flagex.KindSwitch {
				err = reflectex.StringToValue("true", val)
			} else {
				err = reflectex.StringToValue(flag.Value(), val)
			}
			if err != nil {
				return nil, ErrConvert.CauseArgs(err, fldname, flag.Value(), fv.Type().Name())
			}
			fv.Set(val)
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
