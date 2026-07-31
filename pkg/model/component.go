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

// CheckOld raises for outdated connector attributes.
func CheckOld(node string, attrs *utils.OrderedMap) {
	old := map[string]string{
		"pinout":       "was renamed to 'pinlabels' in v0.2",
		"pinnumbers":   "was renamed to 'pins' in v0.2",
		"autogenerate": "is replaced with new syntax in v0.4",
	}
	for attr, descr := range old {
		if attrs.Has(attr) {
			panic(fmt.Errorf("'%s' in %s: '%s' %s", attr, node, attr, descr))
		}
	}
}

// Connection mirrors the WireViz Connection dataclass.
type Connection struct {
	FromName string
	FromPin  any
	ViaPort  any
	ToName   string
	ToPin    any
}

// MateKind distinguishes pin mates from component mates.
type MateKind int

const (
	MatePinKind MateKind = iota
	MateComponentKind
)

// Mate mirrors either MatePin or MateComponent.
type Mate struct {
	Kind     MateKind
	FromName string
	FromPin  any
	ToName   string
	ToPin    any
	Shape    string
}

// AdditionalComponent mirrors the WireViz AdditionalComponent dataclass.
type AdditionalComponent struct {
	Type          string
	Subtype       string
	Manufacturer  string
	Mpn           string
	Supplier      string
	Spn           string
	Pn            string
	Qty           any // int or float64
	Unit          string
	QtyMultiplier string
	Bgcolor       string
}

// Description returns "type, subtype" trimmed of trailing whitespace.
func (c AdditionalComponent) Description() string {
	t := trimRstrip(c.Type)
	if c.Subtype != "" {
		t += ", " + trimRstrip(c.Subtype)
	}
	return t
}

func trimRstrip(s string) string {
	return rstripSpace(s)
}

// NewAdditionalComponent builds an AdditionalComponent from a YAML dict.
func NewAdditionalComponent(attrs *utils.OrderedMap) AdditionalComponent {
	c := AdditionalComponent{Qty: 1}
	if attrs == nil {
		return c
	}
	if v, ok := attrs.Get("type"); ok && v != nil {
		c.Type = utils.PyStr(v)
	}
	if v, ok := attrs.Get("subtype"); ok && v != nil {
		c.Subtype = utils.PyStr(v)
	}
	if v, ok := attrs.Get("manufacturer"); ok && v != nil {
		c.Manufacturer = utils.PyStr(v)
	}
	if v, ok := attrs.Get("mpn"); ok && v != nil {
		c.Mpn = utils.PyStr(v)
	}
	if v, ok := attrs.Get("supplier"); ok && v != nil {
		c.Supplier = utils.PyStr(v)
	}
	if v, ok := attrs.Get("spn"); ok && v != nil {
		c.Spn = utils.PyStr(v)
	}
	if v, ok := attrs.Get("pn"); ok && v != nil {
		c.Pn = utils.PyStr(v)
	}
	if v, ok := attrs.Get("qty"); ok && v != nil {
		c.Qty = v
	}
	if v, ok := attrs.Get("unit"); ok && v != nil {
		c.Unit = utils.PyStr(v)
	}
	if v, ok := attrs.Get("qty_multiplier"); ok && v != nil {
		c.QtyMultiplier = utils.PyStr(v)
	}
	if v, ok := attrs.Get("bgcolor"); ok && v != nil {
		c.Bgcolor = utils.PyStr(v)
	}
	return c
}

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
func NewConnector(name string, attrs *utils.OrderedMap) *Connector {
	CheckOld(fmt.Sprintf("Connector '%s'", name), attrs)
	for _, k := range attrs.Keys() {
		if !connectorFields[k] {
			panic(fmt.Errorf("Connector.__init__() got an unexpected keyword argument '%s'", k))
		}
	}
	c := &Connector{
		Name:        name,
		VisiblePins: map[any]bool{},
	}
	if attrs == nil {
		c.postInit()
		return c
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
	c.postInit()
	return c
}

func (c *Connector) postInit() {
	c.PortsLeft = false
	c.PortsRight = false
	c.VisiblePins = map[any]bool{}

	if c.Style == "simple" {
		if c.Pincount > 1 {
			panic(fmt.Errorf("Connectors with style set to simple may only have one pin"))
		}
		c.Pincount = 1
	}

	if c.Pincount == 0 {
		c.Pincount = max(len(c.Pins), len(c.Pinlabels), len(c.Pincolors))
		if c.Pincount == 0 {
			panic(fmt.Errorf("You need to specify at least one, pincount, pins, pinlabels, or pincolors"))
		}
	}

	if len(c.Pins) == 0 {
		for i := 1; i <= c.Pincount; i++ {
			c.Pins = append(c.Pins, i)
		}
	}

	if len(c.Pins) != len(uniqueSet(c.Pins)) {
		panic(fmt.Errorf("Pins are not unique"))
	}

	if !c.showNameSet {
		c.ShowName = c.Style != "simple" && !hasAutoPrefix(c.Name)
	}
	if !c.showPincountSet {
		c.ShowPincount = c.Style != "simple"
	}

	for _, loop := range c.Loops {
		if len(loop) != 2 {
			panic(fmt.Errorf("Loops must be between exactly two pins!"))
		}
		for _, pin := range loop {
			if !containsAny(c.Pins, pin) {
				panic(fmt.Errorf("Unknown loop pin %q for connector %q!", utils.PyStr(pin), c.Name))
			}
			c.ActivatePin(pin, SideNone)
		}
	}
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
func (c *Connector) GetQtyMultiplier(qtyMultiplier string) any {
	switch qtyMultiplier {
	case "":
		return 1
	case "pincount":
		return c.Pincount
	case "populated":
		n := 0
		for _, v := range c.VisiblePins {
			if v {
				n++
			}
		}
		return n
	case "unpopulated":
		n := 0
		for _, v := range c.VisiblePins {
			if v {
				n++
			}
		}
		if c.Pincount-n < 0 {
			return 0
		}
		return c.Pincount - n
	default:
		panic(fmt.Errorf("invalid qty multiplier parameter for connector %s", qtyMultiplier))
	}
}

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
func NewCable(name string, attrs *utils.OrderedMap) *Cable {
	for _, k := range attrs.Keys() {
		if !cableFields[k] {
			panic(fmt.Errorf("Cable.__init__() got an unexpected keyword argument '%s'", k))
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
	c.postInit()
	return c
}

func (c *Cable) postInit() {
	if s, ok := c.Gauge.(string); ok {
		parts := splitSpace(s)
		if len(parts) != 2 {
			panic(fmt.Errorf("Cable %s gauge=%s - Gauge must be a number, or number and unit separated by a space", c.Name, utils.PyStr(c.Gauge)))
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
			panic(fmt.Errorf("Cable %s length=%s - Length must be a number, or number and unit separated by a space", c.Name, utils.PyStr(c.Length)))
		}
		var L float64
		var err error
		L, err = parseFloat(parts[0])
		if err != nil {
			panic(fmt.Errorf("Cable %s length=%s - Length must be a number, or number and unit separated by a space", c.Name, utils.PyStr(c.Length)))
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
			panic(fmt.Errorf("Cable %s length has a non-numeric value", c.Name))
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
				panic(fmt.Errorf("Unknown color code"))
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
			panic(fmt.Errorf("Unknown number of wires. Must specify wirecount or colors (implicit length)"))
		}
		c.Wirecount = len(c.Colors)
	}

	if len(c.Wirelabels) > 0 && c.Shield != nil && c.Shield != false {
		for _, wl := range c.Wirelabels {
			if utils.PyStr(wl) == "s" {
				panic(fmt.Errorf("%q may not be used as a wire label for a shielded cable.", "s"))
			}
		}
	}

	for _, idfield := range []any{c.Manufacturer, c.Mpn, c.Supplier, c.Spn, c.Pn} {
		if l, ok := idfield.([]any); ok {
			if c.Category == "bundle" {
				if len(l) != c.Wirecount {
					panic(fmt.Errorf("lists of part data must match wirecount"))
				}
			} else {
				panic(fmt.Errorf("lists of part data are only supported for bundles"))
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
func (c *Cable) GetQtyMultiplier(qtyMultiplier string) any {
	switch qtyMultiplier {
	case "":
		return 1
	case "wirecount":
		return c.Wirecount
	case "terminations":
		return len(c.Connections)
	case "length":
		return c.Length
	case "total_length":
		return MulNum(c.Length, c.Wirecount)
	default:
		panic(fmt.Errorf("invalid qty multiplier parameter for cable %s", qtyMultiplier))
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
