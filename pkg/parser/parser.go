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
	"path/filepath"

	harnesspkg "github.com/dmundt/wgo/pkg/harness"
	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

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

	if err := processConnections(connectionSets, opts.TemplateSeparator, templateConnectors, templateCables, h); err != nil {
		return err
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
