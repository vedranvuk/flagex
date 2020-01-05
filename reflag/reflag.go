// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package reflag provides reflect based helpers for flagex.
package reflag

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vedranvuk/flagex"
)

func makeflags(flags *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {

	for i := 0; i < v.NumField(); i++ {
		fv := reflect.Indirect(v.Field(i))
		key := strings.ToLower(v.Type().Field(i).Name)
		short := string(key[0])
		paramhelp := strings.ToLower(v.Field(i).Type().Name())
		if fv.Kind() == reflect.Struct {
			new, err := makeflags(flagex.New(), v.Field(i).Elem())
			if err != nil {
				return nil, err
			}
			if err := flags.Sub(key, short, fmt.Sprintf("Submenu '%s'", key), new); err != nil {
				return nil, err
			}
		}
		kind := flagex.KindOptional
		if fv.Kind() == reflect.String {
			if err := flags.Def(key, short, fmt.Sprintf("Field '%s' (%s)", key, paramhelp), paramhelp, "", kind); err != nil {
				return nil, err
			}
		}
	}

	return flags, nil
}

func applyflags(flags *flagex.Flags, v reflect.Value) (*flagex.Flags, error) {
	return flags, nil
}

// ToStruct parses args to a struct.
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
