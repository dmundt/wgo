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

// Package parser loads a WireViz harness YAML document with PyYAML-compatible
// semantics (ordered mappings, merge keys, scalar resolution) and drives the
// parsing pipeline defined in wireviz.py.
package parser

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/dmundt/wgo/pkg/utils"
)

// LoadYAML parses YAML text into Python-like values (OrderedMap, []any,
// int, float64, string, bool, nil), resolving anchors and merge keys.
func LoadYAML(data []byte) (any, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}
	anchors := map[string]*yaml.Node{}
	collectAnchors(&doc, anchors)
	return nodeValue(doc.Content[0], anchors)
}

func collectAnchors(n *yaml.Node, anchors map[string]*yaml.Node) {
	if n == nil {
		return
	}
	if n.Anchor != "" {
		anchors[n.Anchor] = n
	}
	for _, c := range n.Content {
		collectAnchors(c, anchors)
	}
}

func nodeValue(n *yaml.Node, anchors map[string]*yaml.Node) (any, error) {
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) == 0 {
			return nil, nil
		}
		return nodeValue(n.Content[0], anchors)
	case yaml.SequenceNode:
		var out []any
		for _, c := range n.Content {
			v, err := nodeValue(c, anchors)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case yaml.MappingNode:
		m := utils.NewOrderedMap()
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, v := n.Content[i], n.Content[i+1]
			if k.Tag == "!!merge" || k.Value == "<<" {
				for _, target := range mergeTargets(v, anchors) {
					merged, err := nodeValue(target, anchors)
					if err != nil {
						return nil, err
					}
					if mm, ok := merged.(*utils.OrderedMap); ok {
						for _, kv := range mm.Items() {
							m.Set(kv.Key, kv.Value)
						}
					}
				}
				continue
			}
			val, err := nodeValue(v, anchors)
			if err != nil {
				return nil, err
			}
			m.Set(scalarKey(k), val)
		}
		return m, nil
	case yaml.ScalarNode:
		return scalarValue(n), nil
	case yaml.AliasNode:
		if target, ok := anchors[n.Value]; ok {
			return nodeValue(target, anchors)
		}
		return nil, nil
	}
	return nil, nil
}

func mergeTargets(v *yaml.Node, anchors map[string]*yaml.Node) []*yaml.Node {
	switch v.Kind {
	case yaml.AliasNode:
		if t, ok := anchors[v.Value]; ok {
			return []*yaml.Node{t}
		}
	case yaml.MappingNode:
		return []*yaml.Node{v}
	case yaml.SequenceNode:
		var out []*yaml.Node
		for _, c := range v.Content {
			out = append(out, mergeTargets(c, anchors)...)
		}
		return out
	}
	return nil
}

func scalarKey(n *yaml.Node) string {
	if n.Kind == yaml.AliasNode {
		return n.Value
	}
	return n.Value
}

func scalarValue(n *yaml.Node) any {
	// Quoted and block scalars are always strings in pyyaml.
	switch n.Style {
	case yaml.SingleQuotedStyle, yaml.DoubleQuotedStyle, yaml.LiteralStyle, yaml.FoldedStyle:
		return n.Value
	}
	v := n.Value
	if pyNullRe.MatchString(v) {
		return nil
	}
	if pyBoolRe.MatchString(v) {
		return boolVal(v)
	}
	if iv, ok := pyInt(v); ok {
		return iv
	}
	if fv, ok := pyFloat(v); ok {
		return fv
	}
	return v
}

var (
	pyNullRe  = regexp.MustCompile(`^(?:~|null|Null|NULL| )$`)
	pyBoolRe  = regexp.MustCompile(`^(?:yes|Yes|YES|no|No|NO|true|True|TRUE|false|False|FALSE|on|On|ON|off|Off|OFF)$`)
	pyIntRe   = regexp.MustCompile(`^(?:[-+]?0b[0-1_]+|[-+]?0[0-7_]+|[-+]?(?:0|[1-9][0-9_]*)|[-+]?0x[0-9a-fA-F_]+|[-+]?[1-9][0-9_]*(?::[0-5]?[0-9])+)$`)
	pyFloatRe = regexp.MustCompile(`^(?:[-+]?(?:[0-9][0-9_]*)\.[0-9_]*(?:[eE][-+][0-9]+)?|\.[0-9][0-9_]*(?:[eE][-+][0-9]+)?|[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\.[0-9_]*|[-+]?\.(?:inf|Inf|INF)|\.(?:nan|NaN|NAN))$`)
)

func boolVal(v string) bool {
	switch strings.ToLower(v) {
	case "yes", "true", "on":
		return true
	}
	return false
}

func pyInt(s string) (any, bool) {
	if !pyIntRe.MatchString(s) {
		return nil, false
	}
	sign := ""
	body := strings.ReplaceAll(s, "_", "")
	if strings.HasPrefix(body, "+") || strings.HasPrefix(body, "-") {
		sign = body[:1]
		body = body[1:]
	}
	switch {
	case strings.HasPrefix(body, "0b"):
		if v, err := strconv.ParseInt(sign+body[2:], 2, 64); err == nil {
			return int(v), true
		}
	case strings.HasPrefix(body, "0x"):
		if v, err := strconv.ParseInt(sign+body[2:], 16, 64); err == nil {
			return int(v), true
		}
	case len(body) > 1 && body[0] == '0':
		if v, err := strconv.ParseInt(sign+body[1:], 8, 64); err == nil {
			return int(v), true
		}
	case strings.Contains(body, ":"):
		return sexagesimalInt(sign + body)
	default:
		if v, err := strconv.ParseInt(sign+body, 10, 64); err == nil {
			return int(v), true
		}
	}
	return nil, false
}

func sexagesimalInt(s string) (any, bool) {
	sign := 1
	if strings.HasPrefix(s, "-") {
		sign = -1
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	parts := strings.Split(s, ":")
	total := int64(0)
	for _, p := range parts {
		p = strings.ReplaceAll(p, "_", "")
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, false
		}
		total = total*60 + v
	}
	return int(sign) * int(total), true
}

func pyFloat(s string) (any, bool) {
	if !pyFloatRe.MatchString(s) {
		return nil, false
	}
	clean := strings.ReplaceAll(s, "_", "")
	lower := strings.ToLower(clean)
	switch {
	case strings.HasSuffix(lower, "inf"):
		if strings.HasPrefix(clean, "-") {
			return math.Inf(-1), true
		}
		return math.Inf(1), true
	case strings.HasSuffix(lower, "nan"):
		return math.NaN(), true
	}
	if strings.Contains(clean, ":") {
		return sexagesimalFloat(clean)
	}
	if v, err := strconv.ParseFloat(clean, 64); err == nil {
		return v, true
	}
	return nil, false
}

// sexagesimalFloat converts "1:30.5" to 90.5 like PyYAML's
// construct_yaml_float: sign is stripped first, then the parts are summed
// base-60 from the right (e.g. "12:34:56.78" = 12*60^2 + 34*60 + 56.78).
func sexagesimalFloat(s string) (any, bool) {
	sign := 1.0
	if strings.HasPrefix(s, "-") {
		sign = -1
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	parts := strings.Split(s, ":")
	var total float64
	base := 1.0
	for i := len(parts) - 1; i >= 0; i-- {
		v, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return nil, false
		}
		total += v * base
		base *= 60
	}
	return sign * total, true
}
