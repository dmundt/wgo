// wgo - Go port of WireViz.
//
// Copyright (C) 2026 the wgo contributors.
// Based on WireViz (https://github.com/wireviz/WireViz), copyright its
// respective contributors.
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"strconv"
	"strings"

	"github.com/dmundt/wgo/pkg/utils"
)

func attrStr(attrs *utils.OrderedMap, key string) string {
	v, ok := attrs.Get(key)
	if !ok || v == nil {
		return ""
	}
	return utils.PyStr(v)
}

func attrAny(attrs *utils.OrderedMap, key string) any {
	v, ok := attrs.Get(key)
	if !ok || v == nil {
		return nil
	}
	return v
}

func toInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case float64:
		return int(x)
	default:
		return 0
	}
}

func hasAutoPrefix(name string) bool {
	return len(name) >= 2 && name[0:2] == "__"
}

func uniqueSet(seq []any) []any {
	var out []any
	for _, e := range seq {
		if !containsAny(out, e) {
			out = append(out, e)
		}
	}
	return out
}

func containsAny(seq []any, v any) bool {
	return indexOfAny(seq, v) >= 0
}

func indexOfAny(seq []any, v any) int {
	for i, e := range seq {
		if anyEqual(e, v) {
			return i
		}
	}
	return -1
}

func anyEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}
	switch x := a.(type) {
	case int:
		if y, ok := b.(int); ok {
			return x == y
		}
	case string:
		if y, ok := b.(string); ok {
			return x == y
		}
	}
	return false
}

func splitSpace(s string) []string {
	return strings.Split(s, " ")
}

func stringsUpper(s string) string {
	return strings.ToUpper(s)
}

func replaceMM2(s string) string {
	return strings.ReplaceAll(s, "mm2", "mm²")
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func rstripSpace(s string) string {
	return strings.TrimRight(s, " \t\n\r\v\f\u00a0\u2028\u2029")
}
