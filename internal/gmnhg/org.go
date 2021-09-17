package gmnhg

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/niklasfasching/go-org/org"
)

// for strings "true", "false", and int/float numbers will try to
// convert them to Go values; will simply return value on fail or any
// other kind of input
func parseValue(value interface{}) interface{} {
	switch value := value.(type) {
	case string:
		num, err := strconv.Atoi(value)
		if err == nil {
			return num
		}
		float, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return float
		}
		boolean, err := strconv.ParseBool(value)
		if err == nil {
			return boolean
		}
	}
	return value
}

// for key "key" will set either map key "key" or struct field tagged
// `tag:"key"` with value; expects a pointer
func reflectSetKey(mapOrStruct interface{}, tag, key string, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("recovered from panic: %v", e)
		}
	}()
	v := reflect.ValueOf(mapOrStruct).Elem()
	switch kind := v.Kind(); kind {
	case reflect.Map:
		v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	case reflect.Struct:
		var fieldName string
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			v, ok := field.Tag.Lookup(tag)
			if !ok {
				continue
			}
			if v != key {
				continue
			}
			fieldName = field.Name
		}
		if fieldName == "" {
			return fmt.Errorf("cannot find tag %v with key %v in struct", tag, key)
		}
		v.FieldByName(fieldName).Set(reflect.ValueOf(parseValue(value)))
	default:
		return fmt.Errorf("cannot set key of %v", kind.String())
	}
	return nil
}

func unmarshalORG(data []byte, p interface{}) (err error) {
	parser := org.New()
	document := parser.Parse(bytes.NewReader(data), "")
	if document.Error != nil {
		return document.Error
	}
	for k, v := range document.BufferSettings {
		var (
			key   string      = k
			value interface{} = v
		)
		if strings.HasSuffix(k, "[]") {
			key = k[:len(k)-2]
			value = strings.Fields(v)
		} else if k == "tags" || k == "categories" || k == "aliases" {
			value = strings.Fields(v)
		} else if k == "date" {
			value = parseORGDate(v)
		}
		if err := reflectSetKey(p, "org", strings.ToLower(key), value); err != nil {
			return err
		}
	}
	return nil
}

// Some Org parsing code below was originally taken from Hugo and was
// tweaked for gmnhg purposes.
//
// Copyright 2018 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you
// may not use this file except in compliance with the License. You may
// obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

var orgDateRegex = regexp.MustCompile(`[<\[](\d{4}-\d{2}-\d{2}) .*[>\]]`)

func parseORGDate(s string) string {
	if m := orgDateRegex.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return s
}
