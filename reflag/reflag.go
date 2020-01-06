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
)

// flagsFromStruct creates Flags from struct v.
// Parsing is multilevel, root and v are initial flags and a struct.
func flagsFromStruct(root *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {
	v = reflect.Indirect(v)

	for i := 0; i < v.NumField(); i++ {
		key := strings.ToLower(v.Type().Field(i).Name)
		fv := reflect.Indirect(v.Field(i))
		short := string(key[0])
		paramhelp := strings.ToLower(v.Field(i).Type().Name())
		help := v.Type().Field(i).Name
		_, ok := (fv.Interface()).(encoding.TextMarshaler)
		if ok {
			if err := root.Def(key, short, help, paramhelp, "", flagex.KindOptional); err != nil {
				return nil, err
			}
			continue
		}
		switch fv.Kind() {
		case reflect.Struct:
			new, err := flagsFromStruct(flagex.New(), v.Field(i))
			if err != nil {
				return nil, err
			}
			if err := root.Sub(key, short, help, new); err != nil {
				return nil, err
			}
		case reflect.Bool:
			if err := root.Switch(key, short, help); err != nil {
				return nil, err
			}
		default:
			if err := root.Def(key, short, help, paramhelp, "", flagex.KindOptional); err != nil {
				return nil, err
			}
		}
	}

	return root, nil
}

// structApplyFlags applies pared values in flags to struct v.
func structApplyFlags(flags *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {
	for i := 0; i < v.NumField(); i++ {
		fv := reflect.Indirect(v.Field(i))
		key := strings.ToLower(v.Type().Field(i).Name)
		flag, ok := flags.Key(key)
		if !ok {
			panic("no key found")
		}
		if !flag.Parsed() {
			continue
		}
		if fv.Kind() == reflect.Ptr {
			fmt.Println("empty ptr")
		}
		intf, ok := (fv.Addr().Interface()).(encoding.TextUnmarshaler)
		if ok {
			if err := intf.UnmarshalText([]byte(flag.Value())); err != nil {
				return nil, ErrUnmarshal.CauseArgs(err, key, flag.Value(), fv.Type().Name())
			}
			continue
		}
		if fv.Kind() == reflect.Struct {
			_, err := structApplyFlags(flag.Sub(), fv)
			if err != nil {
				return nil, err
			}
		}
		var val reflect.Value
		var err error
		if flag.Sub() == nil {
			if flag.Kind() == flagex.KindSwitch {
				val, err = reflectex.StringToValue("true", fv)
			} else {
				val, err = reflectex.StringToValue(flag.Value(), fv)
			}
			if err != nil {
				return nil, ErrConvert.CauseArgs(err, key, flag.Value(), fv.Type().Name())
			}
			fv.Set(val)
		}
	}
	return flags, nil
}

// FromStruct creates Flags from struct v and returns it or an error.
func FromStruct(v interface{}, args []string) (*flagex.Flags, error) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.IsValid() {
		return nil, flagex.ErrParam
	}
	if rv.Kind() != reflect.Struct {
		return nil, flagex.ErrParam
	}
	flags, err := flagsFromStruct(flagex.New(), rv)
	if err != nil {
		return nil, err
	}
	return flags, nil
}

// ToStruct applies Flags (prefferably gotten with FromStruct) to a
// struct (preferrably one used with FromStruct).
// Flags should be parsed.
func ToStruct(v interface{}, flags *flagex.Flags) error {
	_, err := structApplyFlags(flags, reflect.Indirect(reflect.ValueOf(v)))
	if err != nil {
		return err
	}
	return nil
}

// Struct takes a struct v, creates flags from v's exported fields,
// parses flags, sets v fields to parsed values then returns flags
// or an error if one occured.
func Struct(v interface{}, args []string) (*flagex.Flags, error) {
	flags, err := FromStruct(v, args)
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
