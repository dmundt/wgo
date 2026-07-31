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
	"fmt"

	"github.com/dmundt/wgo/pkg/utils"
)

// Cable mirrors the WireViz Cable dataclass.
type Cable struct {
	Name                 string
	Bgcolor              string
	BgcolorTitle         string
	Manufacturer         any // string or []string
	Mpn                  any
	Supplier             any
	Spn                  any
	Pn                   any
	Category             string
	Type                 any
	Gauge                any // int, float64, or string
	GaugeUnit            string
	ShowEquiv            bool
	Length               any // int or float64
	LengthUnit           string
	Color                string
	Wirecount            int
	Shield               any // bool or string
	Image                *Image
	Notes                string
	Colors               []string
	Wirelabels           []any
	ColorCode            string
	ShowName             bool
	ShowWirecount        bool
	ShowWirenumbers      bool
	IgnoreInBom          bool
	AdditionalComponents []AdditionalComponent
	Connections          []Connection

	showNameSet        bool
	showWirecountSet   bool
	showWirenumbersSet bool
}

var cableFields = map[string]bool{
	"bgcolor": true, "bgcolor_title": true, "manufacturer": true, "mpn": true,
	"supplier": true, "spn": true, "pn": true, "category": true, "type": true,
	"gauge": true, "gauge_unit": true, "show_equiv": true, "length": true,
	"length_unit": true, "color": true, "wirecount": true, "shield": true,
	"image": true, "notes": true, "colors": true, "wirelabels": true,
	"color_code": true, "show_name": true, "show_wirecount": true,
	"show_wirenumbers": true, "ignore_in_bom": true, "additional_components": true,
}

// NewCable builds a Cable from a YAML dict and runs the post-init logic.
func NewCable(name string, attrs *utils.OrderedMap) (*Cable, error) {
	for _, k := range attrs.Keys() {
		if !cableFields[k] {
			return nil, fmt.Errorf("Cable.__init__() got an unexpected keyword argument '%s'", k)
		}
	}
	c := &Cable{
		Name:          name,
		ShowWirecount: true,
	}
	c.Bgcolor = attrStr(attrs, "bgcolor")
	c.BgcolorTitle = attrStr(attrs, "bgcolor_title")
	c.Manufacturer = attrAny(attrs, "manufacturer")
	c.Mpn = attrAny(attrs, "mpn")
	c.Supplier = attrAny(attrs, "supplier")
	c.Spn = attrAny(attrs, "spn")
	c.Pn = attrAny(attrs, "pn")
	c.Category = attrStr(attrs, "category")
	c.Type = attrAny(attrs, "type")
	c.Gauge = attrAny(attrs, "gauge")
	c.GaugeUnit = attrStr(attrs, "gauge_unit")
	c.Length = attrAny(attrs, "length")
	if c.Length == nil {
		c.Length = 0
	}
	c.LengthUnit = attrStr(attrs, "length_unit")
	c.Color = attrStr(attrs, "color")
	c.Notes = attrStr(attrs, "notes")
	c.ColorCode = attrStr(attrs, "color_code")
	if v, ok := attrs.Get("show_equiv"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.ShowEquiv = b
		}
	}
	if v, ok := attrs.Get("wirecount"); ok && v != nil {
		c.Wirecount = toInt(v)
	}
	c.Shield = attrAny(attrs, "shield")
	if v, ok := attrs.Get("image"); ok && v != nil {
		if m, ok := v.(*utils.OrderedMap); ok {
			c.Image = NewImage(m)
		}
	}
	if v, ok := attrs.Get("colors"); ok && v != nil {
		if l, ok := v.([]any); ok {
			for _, e := range l {
				c.Colors = append(c.Colors, utils.PyStr(e))
			}
		}
	}
	if v, ok := attrs.Get("wirelabels"); ok && v != nil {
		if l, ok := v.([]any); ok {
			c.Wirelabels = l
		}
	}
	if v, ok := attrs.Get("show_name"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.ShowName = b
			c.showNameSet = true
		}
	}
	if v, ok := attrs.Get("show_wirecount"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.ShowWirecount = b
			c.showWirecountSet = true
		}
	}
	if v, ok := attrs.Get("show_wirenumbers"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.ShowWirenumbers = b
			c.showWirenumbersSet = true
		}
	}
	if v, ok := attrs.Get("ignore_in_bom"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.IgnoreInBom = b
		}
	}
	if v, ok := attrs.Get("additional_components"); ok && v != nil {
		if l, ok := v.([]any); ok {
			for _, e := range l {
				if m, ok := e.(*utils.OrderedMap); ok {
					c.AdditionalComponents = append(c.AdditionalComponents, NewAdditionalComponent(m))
				}
			}
		}
	}
	return c, c.postInit()
}

func (c *Cable) postInit() error {
	if s, ok := c.Gauge.(string); ok {
		parts := splitSpace(s)
		if len(parts) != 2 {
			return fmt.Errorf("Cable %s gauge=%s - Gauge must be a number, or number and unit separated by a space", c.Name, utils.PyStr(c.Gauge))
		}
		g, u := parts[0], parts[1]
		c.Gauge = g
		if c.GaugeUnit != "" {
			fmt.Printf("Warning: Cable %s gauge_unit=%s is ignored because its gauge contains %s\n", c.Name, c.GaugeUnit, u)
		}
		if stringsUpper(u) == "AWG" {
			c.GaugeUnit = stringsUpper(u)
		} else {
			c.GaugeUnit = replaceMM2(u)
		}
	} else if c.Gauge != nil {
		if c.GaugeUnit == "" {
			c.GaugeUnit = "mm²"
		}
	}

	if s, ok := c.Length.(string); ok {
		parts := splitSpace(s)
		if len(parts) != 2 {
			return fmt.Errorf("Cable %s length=%s - Length must be a number, or number and unit separated by a space", c.Name, utils.PyStr(c.Length))
		}
		var L float64
		var err error
		L, err = parseFloat(parts[0])
		if err != nil {
			return fmt.Errorf("Cable %s length=%s - Length must be a number, or number and unit separated by a space", c.Name, utils.PyStr(c.Length))
		}
		c.Length = L
		if c.LengthUnit != "" {
			fmt.Printf("Warning: Cable %s length_unit=%s is ignored because its length contains %s\n", c.Name, c.LengthUnit, parts[1])
		}
		c.LengthUnit = parts[1]
	} else if c.Length != nil {
		switch c.Length.(type) {
		case int, float64:
		default:
			return fmt.Errorf("Cable %s length has a non-numeric value", c.Name)
		}
		if c.LengthUnit == "" {
			c.LengthUnit = "m"
		}
	}

	c.Connections = []Connection{}

	if c.Wirecount != 0 {
		if len(c.Colors) > 0 {
			// use custom color palette
		} else if c.ColorCode != "" {
			colors, ok := ColorCodes[c.ColorCode]
			if !ok {
				return fmt.Errorf("Unknown color code")
			}
			c.Colors = append([]string(nil), colors...)
		} else {
			for i := 0; i < c.Wirecount; i++ {
				c.Colors = append(c.Colors, "")
			}
		}
		if c.Wirecount > len(c.Colors) {
			m := c.Wirecount/len(c.Colors) + 1
			var looped []string
			for i := 0; i < m; i++ {
				looped = append(looped, c.Colors...)
			}
			c.Colors = looped
		}
		c.Colors = c.Colors[:c.Wirecount]
	} else {
		if len(c.Colors) == 0 {
			return fmt.Errorf("Unknown number of wires. Must specify wirecount or colors (implicit length)")
		}
		c.Wirecount = len(c.Colors)
	}

	if len(c.Wirelabels) > 0 && c.Shield != nil && c.Shield != false {
		for _, wl := range c.Wirelabels {
			if utils.PyStr(wl) == "s" {
				return fmt.Errorf("%q may not be used as a wire label for a shielded cable.", "s")
			}
		}
	}

	for _, idfield := range []any{c.Manufacturer, c.Mpn, c.Supplier, c.Spn, c.Pn} {
		if l, ok := idfield.([]any); ok {
			if c.Category == "bundle" {
				if len(l) != c.Wirecount {
					return fmt.Errorf("lists of part data must match wirecount")
				}
			} else {
				return fmt.Errorf("lists of part data are only supported for bundles")
			}
		}
	}

	if !c.showNameSet {
		c.ShowName = !hasAutoPrefix(c.Name)
	}
	if !c.showWirecountSet {
		c.ShowWirecount = true
	}
	if !c.showWirenumbersSet {
		c.ShowWirenumbers = c.Category != "bundle"
	}
	return nil
}

// Connect appends a Connection to the cable.
func (c *Cable) Connect(fromName string, fromPin any, viaWire any, toName string, toPin any) {
	c.Connections = append(c.Connections, Connection{
		FromName: fromName,
		FromPin:  fromPin,
		ViaPort:  viaWire,
		ToName:   toName,
		ToPin:    toPin,
	})
}

// GetQtyMultiplier returns the multiplier for the given qty_multiplier name.
func (c *Cable) GetQtyMultiplier(qtyMultiplier string) (any, error) {
	switch qtyMultiplier {
	case "":
		return 1, nil
	case "wirecount":
		return c.Wirecount, nil
	case "terminations":
		return len(c.Connections), nil
	case "length":
		return c.Length, nil
	case "total_length":
		return MulNum(c.Length, c.Wirecount), nil
	default:
		return nil, fmt.Errorf("invalid qty multiplier parameter for cable %s", qtyMultiplier)
	}
}

// MulNum multiplies two int/float values with Python semantics.
func MulNum(a, b any) any {
	af, aInt := toFloatInt(a)
	bf, bInt := toFloatInt(b)
	if aInt && bInt {
		return int(af) * int(bf)
	}
	return af * bf
}

func toFloatInt(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case float64:
		return x, false
	default:
		return 0, false
	}
}
