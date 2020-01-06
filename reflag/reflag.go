// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package reflag provides reflect based helpers for flagex.
package reflag

import (
	"reflect"
	"strings"

	"github.com/vedranvuk/errorex"
	"github.com/vedranvuk/flagex"
	"github.com/vedranvuk/reflectex"
)

var (
	ErrReflag  = errorex.New("reflag")
	ErrConvert = ErrReflag.WrapFormat("error converting arg '%s' value '%s' to '%s'")
)

func makeflags(flags *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {

	for i := 0; i < v.NumField(); i++ {
		fv := reflect.Indirect(v.Field(i))
		key := strings.ToLower(v.Type().Field(i).Name)
		short := string(key[0])
		paramhelp := strings.ToLower(v.Field(i).Type().Name())
		help := v.Type().Field(i).Name
		if fv.Kind() == reflect.Struct {
			new, err := makeflags(flagex.New(), v.Field(i).Elem())
			if err != nil {
				return nil, err
			}
			if err := flags.Sub(key, short, help, new); err != nil {
				return nil, err
			}
		}
		kind := flagex.KindOptional
		if fv.Kind() == reflect.String {
			if err := flags.Def(key, short, help, paramhelp, "", kind); err != nil {
				return nil, err
			}
		}
		if fv.Kind() == reflect.Int {
			if err := flags.Def(key, short, help, paramhelp, "", kind); err != nil {
				return nil, err
			}
		}
		if fv.Kind() == reflect.Bool {
			if err := flags.Switch(key, short, help); err != nil {
				return nil, err
			}
		}
	}

	return flags, nil
}

func applyflags(flags *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {
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
		if fv.Kind() == reflect.Struct {
			_, err := applyflags(flag.Sub(), fv)
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

// ToStruct parses args to a struct v.
func FromStruct(v interface{}, args []string) (*flagex.Flags, error) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.IsValid() {
		return nil, flagex.ErrParam
	}
	if rv.Kind() != reflect.Struct {
		return nil, flagex.ErrParam
	}
	flags, err := makeflags(flagex.New(), rv)
	if err != nil {
		return nil, err
	}
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	flags, err = applyflags(flags, rv)
	if err != nil {
		return nil, err
	}
	return flags, nil
}
