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
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	harnesspkg "github.com/dmundt/wgo/pkg/harness"
	"github.com/dmundt/wgo/pkg/model"
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
	if v, err := strconv.ParseFloat(clean, 64); err == nil {
		return v, true
	}
	return nil, false
}

// Run parses a YAML harness file and produces the requested outputs.
func Run(yamlInput, yamlFilePath, outputDir, outputName string, formats []string, imagePaths []string) error {
	root, err := LoadYAML([]byte(yamlInput))
	if err != nil {
		return err
	}
	om, ok := root.(*utils.OrderedMap)
	if !ok {
		tname := "NoneType"
		switch root.(type) {
		case []any:
			tname = "list"
		case string:
			tname = "str"
		case int:
			tname = "int"
		case float64:
			tname = "float"
		case bool:
			tname = "bool"
		}
		return fmt.Errorf("Expected a dict as top-level YAML input, but got: %s", tname)
	}

	if yamlFilePath != "" {
		abs, _ := filepath.Abs(yamlFilePath)
		defImg := filepath.Dir(abs)
		found := false
		for _, p := range imagePaths {
			ap, _ := filepath.Abs(p)
			if ap == defImg {
				found = true
				break
			}
		}
		if !found {
			imagePaths = append(imagePaths, defImg)
		}
	}

	var metadata *utils.OrderedMap
	if v, ok := om.Get("metadata"); ok {
		if m, ok := v.(*utils.OrderedMap); ok {
			metadata = m
		}
	}
	if metadata == nil {
		metadata = utils.NewOrderedMap()
	}

	var opts *model.Options
	if v, ok := om.Get("options"); ok {
		if m, ok := v.(*utils.OrderedMap); ok {
			opts = model.NewOptions(m)
		}
	}
	if opts == nil {
		opts = model.NewOptions(nil)
	}

	var tweak *model.Tweak
	if v, ok := om.Get("tweak"); ok {
		if m, ok := v.(*utils.OrderedMap); ok {
			tweak = model.NewTweak(m)
		}
	}
	if tweak == nil {
		tweak = model.NewTweak(nil)
	}

	h := harnesspkg.NewHarness(metadata, opts, tweak)

	if !metadata.Has("title") {
		title := outputName
		if title == "" {
			title = fmt.Sprintf("%s diagram and BOM", model.AppName)
		}
		metadata.Set("title", title)
	}

	templateConnectors := utils.NewOrderedMap()
	templateCables := utils.NewOrderedMap()

	sections := []string{"connectors", "cables", "connections"}
	types := []string{"dict", "dict", "list"}
	for idx, sec := range sections {
		v, exists := om.Get(sec)
		rightType := false
		if types[idx] == "dict" {
			_, rightType = v.(*utils.OrderedMap)
		} else {
			_, rightType = v.([]any)
		}
		if exists && rightType {
			if types[idx] == "dict" {
				m := v.(*utils.OrderedMap)
				if m.Len() > 0 {
					for _, kv := range m.Items() {
						attribs, ok := kv.Value.(*utils.OrderedMap)
						if !ok {
							continue
						}
						if imgVal, ok := attribs.Get("image"); ok {
							if imgMap, ok := imgVal.(*utils.OrderedMap); ok {
								srcVal, _ := imgMap.Get("src")
								if src, ok := srcVal.(string); ok && src != "" && !filepath.IsAbs(src) {
									resolved, err := utils.SmartFileResolve(src, imagePaths)
									if err != nil {
										return err
									}
									imgMap.Set("src", resolved)
								}
							}
						}
						if sec == "connectors" {
							templateConnectors.Set(kv.Key, attribs)
						} else {
							templateCables.Set(kv.Key, attribs)
						}
					}
				}
			}
		} else {
			if types[idx] == "dict" {
				om.Set(sec, utils.NewOrderedMap())
			} else {
				om.Set(sec, []any{})
			}
		}
	}

	connectionSetsVal, _ := om.Get("connections")
	connectionSets, _ := connectionSetsVal.([]any)

	templateSeparator := opts.TemplateSeparator
	designatorsAndTemplates := map[string]string{}
	autogeneratedDesignators := map[string]int{}
	expectedType := ""

	resolveDesignator := func(inp, sep string) (string, string) {
		if strings.Contains(inp, sep) {
			if strings.Count(inp, sep) > 1 {
				panic(fmt.Errorf("%s - Found more than one separator (%s)", inp, sep))
			}
			parts := strings.SplitN(inp, sep, 2)
			template, designator := parts[0], parts[1]
			if designator == "" {
				autogeneratedDesignators[template]++
				designator = fmt.Sprintf("__%s_%d", template, autogeneratedDesignators[template])
			}
			if t, ok := designatorsAndTemplates[designator]; ok {
				if t != template {
					panic(fmt.Errorf("Trying to redefine %s from %s to %s", designator, t, template))
				}
			} else {
				designatorsAndTemplates[designator] = template
			}
			return template, designator
		}
		template, designator := inp, inp
		if _, ok := designatorsAndTemplates[designator]; !ok {
			designatorsAndTemplates[designator] = template
		}
		return template, designator
	}

	checkType := func(designator, template, actualType string) {
		if expectedType == "" {
			expectedType = actualType
		}
		if actualType != expectedType {
			panic(fmt.Errorf("Expected %s, but \"%s\" (\"%s\") is %s", expectedType, designator, template, actualType))
		}
	}

	alternateType := func() {
		if expectedType == "connector" {
			expectedType = "cable/arrow"
		} else {
			expectedType = "connector"
		}
	}

	for _, connectionSetVal := range connectionSets {
		connectionSet, ok := connectionSetVal.([]any)
		if !ok {
			continue
		}

		var connectioncount []int
		for _, entry := range connectionSet {
			switch e := entry.(type) {
			case []any:
				connectioncount = append(connectioncount, len(e))
			case *utils.OrderedMap:
				connectioncount = append(connectioncount, len(utils.Expand(e.First().Value)))
			}
		}
		hasCount := false
		for _, c := range connectioncount {
			if c != 0 {
				hasCount = true
				break
			}
		}
		if !hasCount {
			connectioncount = []int{1}
		}
		first := connectioncount[0]
		for _, c := range connectioncount {
			if c != first {
				panic(fmt.Errorf("All items in connection set must reference the same number of connections"))
			}
		}
		count := first

		for index, entry := range connectionSet {
			if s, ok := entry.(string); ok {
				expanded := make([]any, count)
				for i := 0; i < count; i++ {
					expanded[i] = s
				}
				connectionSet[index] = expanded
			}
		}

		for index, entry := range connectionSet {
			switch e := entry.(type) {
			case []any:
				for subindex, item := range e {
					_, des := resolveDesignator(utils.PyStr(item), templateSeparator)
					e[subindex] = des
				}
			case *utils.OrderedMap:
				kv := e.First()
				_, des := resolveDesignator(kv.Key, templateSeparator)
				nm := utils.NewOrderedMap()
				nm.Set(des, kv.Value)
				connectionSet[index] = nm
			}
		}

		for index, entry := range connectionSet {
			switch e := entry.(type) {
			case []any:
				var newList []any
				for _, des := range e {
					nm := utils.NewOrderedMap()
					nm.Set(utils.PyStr(des), 1)
					newList = append(newList, nm)
				}
				connectionSet[index] = newList
			case *utils.OrderedMap:
				kv := e.First()
				pinlist := utils.Expand(kv.Value)
				var newList []any
				for _, pin := range pinlist {
					nm := utils.NewOrderedMap()
					nm.Set(kv.Key, pin)
					newList = append(newList, nm)
				}
				connectionSet[index] = newList
			}
		}

		expectedType = ""
		for _, entry := range connectionSet {
			for _, item := range entry.([]any) {
				m := item.(*utils.OrderedMap)
				designator := m.First().Key
				template := designatorsAndTemplates[designator]
				if _, ok := h.ConnectorByName(designator); ok {
					checkType(designator, template, "connector")
				} else if t, ok := templateConnectors.Get(template); ok {
					checkType(designator, template, "connector")
					h.AddConnector(designator, t.(*utils.OrderedMap))
				} else if _, ok := h.CableByName(designator); ok {
					checkType(designator, template, "cable/arrow")
				} else if t, ok := templateCables.Get(template); ok {
					checkType(designator, template, "cable/arrow")
					h.AddCable(designator, t.(*utils.OrderedMap))
				} else if utils.IsArrow(designator) {
					checkType(designator, template, "cable/arrow")
				} else {
					panic(fmt.Errorf("%s is an unknown template/designator/arrow.", template))
				}
			}
			alternateType()
		}

		transposed := make([][]any, count)
		for _, entry := range connectionSet {
			inner := entry.([]any)
			for pinIdx, item := range inner {
				transposed[pinIdx] = append(transposed[pinIdx], item)
			}
		}

		for indexEntry, entry := range transposed {
			for indexItem, item := range entry {
				m := item.(*utils.OrderedMap)
				designator := m.First().Key
				if _, ok := h.CableByName(designator); ok {
					var fromName string
					var fromPin any
					if indexItem == 0 {
						fromName, fromPin = "", nil
					} else {
						kv := entry[indexItem-1].(*utils.OrderedMap).First()
						fromName, fromPin = kv.Key, kv.Value
					}
					viaName := designator
					viaPin := m.First().Value
					var toName string
					var toPin any
					if indexItem == len(entry)-1 {
						toName, toPin = "", nil
					} else {
						kv := entry[indexItem+1].(*utils.OrderedMap).First()
						toName, toPin = kv.Key, kv.Value
					}
					h.Connect(fromName, fromPin, viaName, viaPin, toName, toPin)
				} else if utils.IsArrow(designator) {
					if indexItem == 0 {
						panic(fmt.Errorf("An arrow cannot be at the start of a connection set"))
					}
					if indexItem == len(entry)-1 {
						panic(fmt.Errorf("An arrow cannot be at the end of a connection set"))
					}
					kvFrom := entry[indexItem-1].(*utils.OrderedMap).First()
					kvTo := entry[indexItem+1].(*utils.OrderedMap).First()
					fromName, fromPin := kvFrom.Key, kvFrom.Value
					toName, toPin := kvTo.Key, kvTo.Value
					if strings.Contains(designator, "-") {
						h.AddMatePin(fromName, fromPin, toName, toPin, designator)
					} else if strings.Contains(designator, "=") && indexEntry == 0 {
						h.AddMateComponent(fromName, toName, designator)
					}
				}
			}
		}
	}

	var proposed []string
	proposed = append(proposed, templateConnectors.Keys()...)
	proposed = append(proposed, templateCables.Keys()...)
	used := map[string]bool{}
	for _, t := range designatorsAndTemplates {
		used[t] = true
	}
	var forgotten []string
	for _, c := range proposed {
		if !used[c] {
			forgotten = append(forgotten, c)
		}
	}
	if len(forgotten) > 0 {
		fmt.Println("Warning: The following components are not referenced in any connection set:")
		fmt.Println(strings.Join(forgotten, ", "))
	}

	if v, ok := om.Get("additional_bom_items"); ok {
		if l, ok := v.([]any); ok {
			for _, line := range l {
				if m, ok := line.(*utils.OrderedMap); ok {
					h.AddBomItem(harnesspkg.NewBOMEntryFromMap(m))
				}
			}
		}
	}

	if len(formats) > 0 {
		outputFile := filepath.Join(outputDir, outputName)
		return h.Output(outputFile, formats)
	}
	return nil
}
