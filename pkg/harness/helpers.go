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

package harness

import (
	"fmt"
	"strings"

	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

func indexOfAny(seq []any, v any) int {
	for i, e := range seq {
		if anyEqual(e, v) {
			return i
		}
	}
	return -1
}

func containsAny(seq []any, v any) bool {
	return indexOfAny(seq, v) >= 0
}

func countAny(seq []any, v any) int {
	n := 0
	for _, e := range seq {
		if anyEqual(e, v) {
			n++
		}
	}
	return n
}

func indexOfStrings(seq []string, v string) int {
	for i, e := range seq {
		if e == v {
			return i
		}
	}
	return -1
}

func countStrings(seq []string, v string) int {
	n := 0
	for _, e := range seq {
		if e == v {
			n++
		}
	}
	return n
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

func truthyVal(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case int:
		return x != 0
	case float64:
		return x != 0
	default:
		return true
	}
}

func isListValue(v any) bool {
	_, ok1 := v.([]any)
	_, ok2 := v.([]string)
	return ok1 || ok2
}

func strVal(v any) string {
	if v == nil {
		return ""
	}
	return utils.PyStr(v)
}

func strValIfString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func listIdxStr(l []any, i int) string {
	if i < len(l) && l[i] != nil {
		return utils.PyStr(l[i])
	}
	return ""
}

func listIdxOrEmpty(v any, i int) string {
	if l, ok := v.([]any); ok {
		return listIdxStr(l, i)
	}
	return ""
}

func gaugeDisplay(cable *model.Cable) string {
	awgFmt := ""
	if cable.ShowEquiv {
		if cable.GaugeUnit == "mm²" {
			awgFmt = fmt.Sprintf(" (%s AWG)", utils.AWGEquiv(cable.Gauge))
		} else if strings.ToUpper(cable.GaugeUnit) == "AWG" {
			awgFmt = fmt.Sprintf(" (%s mm²)", utils.MM2Equiv(cable.Gauge))
		}
	}
	return fmt.Sprintf("%s %s%s", utils.PyStr(cable.Gauge), cable.GaugeUnit, awgFmt)
}

func shieldTruthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	default:
		return true
	}
}

func numPositive(v any) bool {
	switch x := v.(type) {
	case int:
		return x > 0
	case float64:
		return x > 0
	}
	return false
}

func rstripSpace(s string) string {
	return strings.TrimRight(s, " \t\n\r\v\f\u00a0\u2028\u2029")
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
