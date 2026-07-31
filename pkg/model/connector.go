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

// Connector mirrors the WireViz Connector dataclass.
type Connector struct {
	Name                 string
	Bgcolor              string
	BgcolorTitle         string
	Manufacturer         string
	Mpn                  string
	Supplier             string
	Spn                  string
	Pn                   string
	Style                string
	Category             string
	Type                 any
	Subtype              any
	Pincount             int
	Image                *Image
	Notes                string
	Pins                 []any
	Pinlabels            []any
	Pincolors            []string
	Color                string
	ShowName             bool
	ShowPincount         bool
	HideDisconnectedPins bool
	Loops                [][]any
	IgnoreInBom          bool
	AdditionalComponents []AdditionalComponent
	PortsLeft            bool
	PortsRight           bool
	VisiblePins          map[any]bool

	showNameSet     bool
	showPincountSet bool
}

var connectorFields = map[string]bool{
	"bgcolor": true, "bgcolor_title": true, "manufacturer": true, "mpn": true,
	"supplier": true, "spn": true, "pn": true, "style": true, "category": true,
	"type": true, "subtype": true, "pincount": true, "image": true, "notes": true,
	"pins": true, "pinlabels": true, "pincolors": true, "color": true,
	"show_name": true, "show_pincount": true, "hide_disconnected_pins": true,
	"loops": true, "ignore_in_bom": true, "additional_components": true,
}

// NewConnector builds a Connector from a YAML dict and runs the post-init logic.
func NewConnector(name string, attrs *utils.OrderedMap) (*Connector, error) {
	if err := CheckOld(fmt.Sprintf("Connector '%s'", name), attrs); err != nil {
		return nil, err
	}
	for _, k := range attrs.Keys() {
		if !connectorFields[k] {
			return nil, fmt.Errorf("Connector.__init__() got an unexpected keyword argument '%s'", k)
		}
	}
	c := &Connector{
		Name:        name,
		VisiblePins: map[any]bool{},
	}
	if attrs == nil {
		return c, c.postInit()
	}
	c.Bgcolor = attrStr(attrs, "bgcolor")
	c.BgcolorTitle = attrStr(attrs, "bgcolor_title")
	c.Manufacturer = attrStr(attrs, "manufacturer")
	c.Mpn = attrStr(attrs, "mpn")
	c.Supplier = attrStr(attrs, "supplier")
	c.Spn = attrStr(attrs, "spn")
	c.Pn = attrStr(attrs, "pn")
	c.Style = attrStr(attrs, "style")
	c.Category = attrStr(attrs, "category")
	c.Type = attrAny(attrs, "type")
	c.Subtype = attrAny(attrs, "subtype")
	c.Notes = attrStr(attrs, "notes")
	c.Color = attrStr(attrs, "color")
	if v, ok := attrs.Get("pincount"); ok && v != nil {
		c.Pincount = toInt(v)
	}
	if v, ok := attrs.Get("image"); ok && v != nil {
		if m, ok := v.(*utils.OrderedMap); ok {
			c.Image = NewImage(m)
		}
	}
	if v, ok := attrs.Get("pins"); ok && v != nil {
		if l, ok := v.([]any); ok {
			c.Pins = l
		}
	}
	if v, ok := attrs.Get("pinlabels"); ok && v != nil {
		if l, ok := v.([]any); ok {
			c.Pinlabels = l
		}
	}
	if v, ok := attrs.Get("pincolors"); ok && v != nil {
		if l, ok := v.([]any); ok {
			for _, e := range l {
				c.Pincolors = append(c.Pincolors, utils.PyStr(e))
			}
		}
	}
	if v, ok := attrs.Get("show_name"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.ShowName = b
			c.showNameSet = true
		}
	}
	if v, ok := attrs.Get("show_pincount"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.ShowPincount = b
			c.showPincountSet = true
		}
	}
	if v, ok := attrs.Get("hide_disconnected_pins"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.HideDisconnectedPins = b
		}
	}
	if v, ok := attrs.Get("ignore_in_bom"); ok && v != nil {
		if b, ok := v.(bool); ok {
			c.IgnoreInBom = b
		}
	}
	if v, ok := attrs.Get("loops"); ok && v != nil {
		if l, ok := v.([]any); ok {
			for _, e := range l {
				if inner, ok := e.([]any); ok {
					c.Loops = append(c.Loops, inner)
				}
			}
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

func (c *Connector) postInit() error {
	c.PortsLeft = false
	c.PortsRight = false
	c.VisiblePins = map[any]bool{}

	if c.Style == "simple" {
		if c.Pincount > 1 {
			return fmt.Errorf("Connectors with style set to simple may only have one pin")
		}
		c.Pincount = 1
	}

	if c.Pincount == 0 {
		c.Pincount = max(len(c.Pins), len(c.Pinlabels), len(c.Pincolors))
		if c.Pincount == 0 {
			return fmt.Errorf("You need to specify at least one, pincount, pins, pinlabels, or pincolors")
		}
	}

	if len(c.Pins) == 0 {
		for i := 1; i <= c.Pincount; i++ {
			c.Pins = append(c.Pins, i)
		}
	}

	if len(c.Pins) != len(uniqueSet(c.Pins)) {
		return fmt.Errorf("Pins are not unique")
	}

	if !c.showNameSet {
		c.ShowName = c.Style != "simple" && !hasAutoPrefix(c.Name)
	}
	if !c.showPincountSet {
		c.ShowPincount = c.Style != "simple"
	}

	for _, loop := range c.Loops {
		if len(loop) != 2 {
			return fmt.Errorf("Loops must be between exactly two pins!")
		}
		for _, pin := range loop {
			if !containsAny(c.Pins, pin) {
				return fmt.Errorf("Unknown loop pin %q for connector %q!", utils.PyStr(pin), c.Name)
			}
			c.ActivatePin(pin, SideNone)
		}
	}
	return nil
}

func (c *Connector) ActivatePin(pin any, side Side) {
	c.VisiblePins[pin] = true
	switch side {
	case SideLeft:
		c.PortsLeft = true
	case SideRight:
		c.PortsRight = true
	}
}

// GetQtyMultiplier returns the multiplier for the given qty_multiplier name.
func (c *Connector) GetQtyMultiplier(qtyMultiplier string) (any, error) {
	switch qtyMultiplier {
	case "":
		return 1, nil
	case "pincount":
		return c.Pincount, nil
	case "populated":
		n := 0
		for _, v := range c.VisiblePins {
			if v {
				n++
			}
		}
		return n, nil
	case "unpopulated":
		n := 0
		for _, v := range c.VisiblePins {
			if v {
				n++
			}
		}
		if c.Pincount-n < 0 {
			return 0, nil
		}
		return c.Pincount - n, nil
	default:
		return nil, fmt.Errorf("invalid qty multiplier parameter for connector %s", qtyMultiplier)
	}
}
