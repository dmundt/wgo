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

// Package harness ports WireViz's Harness.py: the wiring graph model, DOT
// generation, and file output.
package harness

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dmundt/wgo/pkg/graph"
	"github.com/dmundt/wgo/pkg/utils"
)

func (h *Harness) applyTweak(dot *graph.Graph) error {
	if h.Tweak == nil || h.Tweak.Override == nil {
		goto appendBlock
	}
	for i, entry := range dot.Body {
		keyword := bodyKeyword(entry)
		if keyword == "" {
			continue
		}
		d, ok := h.Tweak.Override.Get(keyword)
		if !ok {
			continue
		}
		om, ok := d.(*utils.OrderedMap)
		if !ok {
			return fmt.Errorf("Unexpected value type of tweak.override.%s: Expected dict, got %T", keyword, d)
		}
		for _, av := range om.Items() {
			a := av.Key
			value := av.Value
			if value == nil {
				entry, n := removeAttr(entry, a)
				if n < 1 {
					fmt.Printf("Harness.create_graph() warning: %s not found in %s!\n", a, keyword)
				} else if n > 1 {
					fmt.Printf("Harness.create_graph() warning: %s removed %d times in %s!\n", a, n, keyword)
				}
				dot.Body[i] = entry
				continue
			}
			vs := utils.PyStr(value)
			if len(vs) == 0 || strings.Contains(vs, " ") {
				vs = strings.ReplaceAll(vs, `"`, `\"`)
				vs = `"` + vs + `"`
			}
			entry, n := overrideAttr(entry, a, vs)
			if n < 1 {
				entry = appendAttr(entry, a, vs)
			} else if n > 1 {
				fmt.Printf("Harness.create_graph() warning: %s overridden %d times in %s!\n", a, n, keyword)
			}
			dot.Body[i] = entry
		}
	}

appendBlock:
	if h.Tweak != nil && h.Tweak.Append != nil {
		switch app := h.Tweak.Append.(type) {
		case []any:
			for _, e := range app {
				dot.Body = append(dot.Body, utils.PyStr(e))
			}
		default:
			dot.Body = append(dot.Body, utils.PyStr(app))
		}
	}
	return nil
}

var bodyKeywordRe = regexp.MustCompile(`(?s)^\t*(?:"([^"]+)"|([^ "]+)) \[.*\]$`)

func bodyKeyword(entry string) string {
	m := bodyKeywordRe.FindStringSubmatch(entry)
	if m == nil {
		return ""
	}
	if m[1] != "" {
		return m[1]
	}
	return m[2]
}

func removeAttr(entry, attr string) (string, int) {
	// Build a pattern equivalent to Python's
	// ( +)?{attr}=("[^"]*"|[^] ]*)(?(1)| *)
	pattern := regexp.MustCompile(`( +)` + regexp.QuoteMeta(attr) + `=("[^"]*"|[^] ]*)|` +
		regexp.QuoteMeta(attr) + `=("[^"]*"|[^] ]*)( +)?`)
	count := 0
	out := pattern.ReplaceAllStringFunc(entry, func(m string) string {
		count++
		return ""
	})
	return out, count
}

func overrideAttr(entry, attr, value string) (string, int) {
	pattern := regexp.MustCompile(regexp.QuoteMeta(attr) + `=("[^"]*"|[^] ]*)`)
	count := 0
	out := pattern.ReplaceAllStringFunc(entry, func(m string) string {
		count++
		return attr + "=" + value
	})
	return out, count
}

func appendAttr(entry, attr, value string) string {
	idx := len(entry)
	if strings.HasSuffix(entry, "\n") {
		idx--
	}
	if idx > 0 && entry[idx-1] == ']' {
		return entry[:idx-1] + " " + attr + "=" + value + "]" + entry[idx:]
	}
	return entry
}
