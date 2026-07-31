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

import "github.com/dmundt/wgo/pkg/utils"

// Options mirrors the WireViz Options dataclass.
type Options struct {
	Fontname          string
	Bgcolor           string
	BgcolorNode       string
	BgcolorConnector  string
	BgcolorCable      string
	BgcolorBundle     string
	ColorMode         string
	MiniBomMode       bool
	TemplateSeparator string
}

// NewOptions builds Options from a YAML dict, applying defaults.
func NewOptions(attrs *utils.OrderedMap) *Options {
	o := &Options{
		Fontname:          "arial",
		Bgcolor:           "WH",
		BgcolorNode:       "WH",
		BgcolorConnector:  "",
		BgcolorCable:      "",
		BgcolorBundle:     "",
		ColorMode:         "SHORT",
		MiniBomMode:       true,
		TemplateSeparator: ".",
	}
	if attrs == nil {
		return o.postInit()
	}
	if v, ok := attrs.Get("fontname"); ok && v != nil {
		o.Fontname = utils.PyStr(v)
	}
	if v, ok := attrs.Get("bgcolor"); ok && v != nil {
		o.Bgcolor = utils.PyStr(v)
	}
	// An explicit null ("None") must clear the default so the post-init
	// chain replaces it, matching the Python dataclass behavior.
	if v, ok := attrs.Get("bgcolor_node"); ok {
		if v != nil {
			o.BgcolorNode = utils.PyStr(v)
		} else {
			o.BgcolorNode = ""
		}
	}
	if v, ok := attrs.Get("bgcolor_connector"); ok {
		if v != nil {
			o.BgcolorConnector = utils.PyStr(v)
		} else {
			o.BgcolorConnector = ""
		}
	}
	if v, ok := attrs.Get("bgcolor_cable"); ok {
		if v != nil {
			o.BgcolorCable = utils.PyStr(v)
		} else {
			o.BgcolorCable = ""
		}
	}
	if v, ok := attrs.Get("bgcolor_bundle"); ok {
		if v != nil {
			o.BgcolorBundle = utils.PyStr(v)
		} else {
			o.BgcolorBundle = ""
		}
	}
	if v, ok := attrs.Get("color_mode"); ok && v != nil {
		o.ColorMode = utils.PyStr(v)
	}
	if v, ok := attrs.Get("mini_bom_mode"); ok && v != nil {
		if b, ok := v.(bool); ok {
			o.MiniBomMode = b
		}
	}
	if v, ok := attrs.Get("template_separator"); ok && v != nil {
		o.TemplateSeparator = utils.PyStr(v)
	}
	return o.postInit()
}

func (o *Options) postInit() *Options {
	if o.BgcolorNode == "" {
		o.BgcolorNode = o.Bgcolor
	}
	if o.BgcolorConnector == "" {
		o.BgcolorConnector = o.BgcolorNode
	}
	if o.BgcolorCable == "" {
		o.BgcolorCable = o.BgcolorNode
	}
	if o.BgcolorBundle == "" {
		o.BgcolorBundle = o.BgcolorCable
	}
	return o
}

// Tweak mirrors the WireViz Tweak dataclass.
type Tweak struct {
	Override *utils.OrderedMap // designator -> OrderedMap(attr -> string or nil)
	Append   any               // string, []any, or nil
}

// NewTweak builds Tweak from a YAML dict.
func NewTweak(attrs *utils.OrderedMap) *Tweak {
	t := &Tweak{}
	if attrs == nil {
		return t
	}
	if v, ok := attrs.Get("override"); ok && v != nil {
		if m, ok := v.(*utils.OrderedMap); ok {
			t.Override = m
		}
	}
	if v, ok := attrs.Get("append"); ok && v != nil {
		t.Append = v
	}
	return t
}
