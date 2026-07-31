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
	"reflect"
	"testing"

	"github.com/dmundt/wgo/pkg/utils"
)

func om(pairs ...any) *utils.OrderedMap {
	m := utils.NewOrderedMap()
	for i := 0; i+1 < len(pairs); i += 2 {
		m.Set(pairs[i].(string), pairs[i+1])
	}
	return m
}

func strList(vals ...string) []any {
	out := make([]any, len(vals))
	for i, v := range vals {
		out[i] = v
	}
	return out
}

func TestGetColorHex(t *testing.T) {
	cases := []struct {
		in   string
		pad  bool
		want []string
	}{
		{"BN", false, []string{"#895956"}},
		{"BUWH", false, []string{"#0066ff", "#ffffff", "#0066ff"}},
		{"WH", true, []string{"#ffffff", "#ffffff", "#ffffff"}},
		{"WH", false, []string{"#ffffff"}},
		{"#ff0000:#00ff00", false, []string{"#ff0000", "#00ff00", "#ff0000"}},
		{"", false, []string{"#ffffff"}},
		{"XX", false, []string{"#ffffff"}},
		{"#zz0000", false, []string{"#ffffff"}},
	}
	for _, c := range cases {
		got := GetColorHex(c.in, c.pad)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("GetColorHex(%q, %v) = %v, want %v", c.in, c.pad, got, c.want)
		}
	}
}

func TestTranslateColor(t *testing.T) {
	cases := []struct {
		in   string
		mode string
		want string
	}{
		{"BN", "SHORT", "BN"},
		{"BN", "short", "bn"},
		{"BN", "FULL", "BROWN"},
		{"BN", "full", "brown"},
		{"BN", "GER", "BR"},
		{"BN", "ger", "br"},
		{"BN", "HEX", "#895956"},
		{"BN", "hex", "#895956"},
		{"BUWH", "HEX", "#0066FF:#FFFFFF:#0066FF"},
		{"XX", "short", "xx"},
		{"", "short", ""},
	}
	for _, c := range cases {
		got := TranslateColor(c.in, c.mode)
		if got != c.want {
			t.Errorf("TranslateColor(%q, %q) = %q, want %q", c.in, c.mode, got, c.want)
		}
	}
}

func TestTranslateColorBadMode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for mixed-case color mode")
		}
	}()
	TranslateColor("BN", "HeX")
}

func TestColorHexLookup(t *testing.T) {
	if v, ok := ColorHexLookup("BK"); !ok || v != "#000000" {
		t.Errorf("ColorHexLookup(BK) = %q, %v", v, ok)
	}
	if _, ok := ColorHexLookup("XX"); ok {
		t.Errorf("ColorHexLookup(XX) should be absent")
	}
}

func TestOptionsDefaults(t *testing.T) {
	o := NewOptions(nil)
	if o.Fontname != "arial" || o.Bgcolor != "WH" || o.BgcolorNode != "WH" ||
		o.BgcolorConnector != "WH" || o.BgcolorCable != "WH" || o.BgcolorBundle != "WH" ||
		o.ColorMode != "SHORT" || !o.MiniBomMode || o.TemplateSeparator != "." {
		t.Errorf("unexpected defaults: %+v", o)
	}
}

func TestOptionsCustom(t *testing.T) {
	o := NewOptions(om("bgcolor", "BK", "mini_bom_mode", false, "fontname", "helvetica"))
	// bgcolor_node/connector/cable/bundle default to "WH" and are kept unless
	// explicitly null, matching the Python Options.post_init behavior.
	if o.Bgcolor != "BK" || o.BgcolorNode != "WH" || o.BgcolorConnector != "WH" ||
		o.BgcolorCable != "WH" || o.BgcolorBundle != "WH" {
		t.Errorf("bgcolor chain not applied: %+v", o)
	}
	if o.MiniBomMode {
		t.Errorf("mini_bom_mode should be false")
	}
	if o.Fontname != "helvetica" {
		t.Errorf("fontname = %q", o.Fontname)
	}
}

func TestOptionsNullChain(t *testing.T) {
	o := NewOptions(om("bgcolor", "BK", "bgcolor_node", nil))
	if o.BgcolorNode != "BK" || o.BgcolorConnector != "BK" || o.BgcolorCable != "BK" || o.BgcolorBundle != "BK" {
		t.Errorf("null bgcolor_node should chain: %+v", o)
	}
	o2 := NewOptions(om("bgcolor_cable", "RD"))
	if o2.BgcolorCable != "RD" || o2.BgcolorBundle != "RD" {
		t.Errorf("cable->bundle chain: %+v", o2)
	}
}

func TestImageDefaults(t *testing.T) {
	img := NewImage(om("src", "x.png"))
	if img.Scale != "false" || img.Fixedsize {
		t.Errorf("scale=%q fixedsize=%v", img.Scale, img.Fixedsize)
	}
}

func TestImageFixedsizeWidth(t *testing.T) {
	// aspect_ratio fails (file missing) and returns 1.0, so width = 70 * 1.
	img := NewImage(om("src", "missing.png", "height", 70))
	if img.Scale != "true" || !img.Fixedsize {
		t.Errorf("scale=%q fixedsize=%v", img.Scale, img.Fixedsize)
	}
	if got := utils.PyStr(img.Width); got != "70.0" {
		t.Errorf("width = %s, want 70.0", got)
	}
}

func TestImageFixedsizeBoth(t *testing.T) {
	img := NewImage(om("src", "x.png", "width", 100, "height", 50))
	if img.Scale != "both" || !img.Fixedsize {
		t.Errorf("scale=%q fixedsize=%v", img.Scale, img.Fixedsize)
	}
}

func TestImageExplicitFixedsizeFalse(t *testing.T) {
	img := NewImage(om("src", "x.png", "width", 100, "fixedsize", false))
	if img.Fixedsize {
		t.Errorf("fixedsize should be false")
	}
	if img.Scale != "true" {
		t.Errorf("scale = %q, want true", img.Scale)
	}
}

func TestConnectorBasics(t *testing.T) {
	c := NewConnector("X1", om("pins", strList("T", "R", "S"), "pinlabels", strList("Dot", "Dash", "Ground")))
	if c.Pincount != 3 {
		t.Errorf("pincount = %d", c.Pincount)
	}
	if !reflect.DeepEqual(c.Pins, []any{"T", "R", "S"}) {
		t.Errorf("pins = %v", c.Pins)
	}
	if !c.ShowName || !c.ShowPincount {
		t.Errorf("show_name=%v show_pincount=%v", c.ShowName, c.ShowPincount)
	}
	if c.PortsLeft || c.PortsRight {
		t.Errorf("ports should be unset initially")
	}
}

func TestConnectorImplicitPincount(t *testing.T) {
	c := NewConnector("X1", om("pinlabels", strList("A", "B", "C", "D")))
	if c.Pincount != 4 {
		t.Errorf("pincount = %d", c.Pincount)
	}
	if !reflect.DeepEqual(c.Pins, []any{1, 2, 3, 4}) {
		t.Errorf("pins = %v", c.Pins)
	}
}

func TestConnectorSimple(t *testing.T) {
	c := NewConnector("F1", om("style", "simple", "type", "ferrule"))
	if c.Pincount != 1 {
		t.Errorf("simple pincount = %d", c.Pincount)
	}
	if c.ShowName || c.ShowPincount {
		t.Errorf("simple connector should hide name and pincount")
	}
}

func TestConnectorSimpleError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for simple connector with pincount > 1")
		}
	}()
	NewConnector("X", om("style", "simple", "pincount", 4))
}

func TestConnectorNoPinsError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for connector without pin info")
		}
	}()
	NewConnector("X", om("type", "unknown"))
}

func TestConnectorShowNameAuto(t *testing.T) {
	c := NewConnector("__X_1", om("pincount", 2))
	if c.ShowName {
		t.Errorf("auto-generated connector should hide name")
	}
	if !c.ShowPincount {
		t.Errorf("pincount should be shown")
	}
}

func TestConnectorDuplicatePins(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for duplicate pins")
		}
	}()
	NewConnector("X", om("pins", strList("A", "A")))
}

func TestConnectorLoops(t *testing.T) {
	c := NewConnector("X", om("pincount", 4, "loops", []any{[]any{1, 2}}))
	if len(c.Loops) != 1 {
		t.Fatalf("loops = %v", c.Loops)
	}
	if !c.VisiblePins[1] || !c.VisiblePins[2] {
		t.Errorf("loop pins should be visible")
	}
}

func TestConnectorLoopErrors(t *testing.T) {
	assertPanic := func(name string, attrs *utils.OrderedMap) {
		t.Helper()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s: expected panic", name)
			}
		}()
		NewConnector(name, attrs)
	}
	assertPanic("badlen", om("pincount", 4, "loops", []any{[]any{1}}))
	assertPanic("badpin", om("pincount", 4, "loops", []any{[]any{1, 9}}))
}

func TestConnectorQtyMultiplier(t *testing.T) {
	c := NewConnector("X", om("pincount", 4))
	if got := c.GetQtyMultiplier(""); got != 1 {
		t.Errorf("default = %v", got)
	}
	if got := c.GetQtyMultiplier("pincount"); got != 4 {
		t.Errorf("pincount = %v", got)
	}
	if got := c.GetQtyMultiplier("populated"); got != 0 {
		t.Errorf("populated = %v", got)
	}
	if got := c.GetQtyMultiplier("unpopulated"); got != 4 {
		t.Errorf("unpopulated = %v", got)
	}
	c.ActivatePin(1, SideLeft)
	c.ActivatePin(2, SideRight)
	if got := c.GetQtyMultiplier("populated"); got != 2 {
		t.Errorf("populated = %v", got)
	}
	if got := c.GetQtyMultiplier("unpopulated"); got != 2 {
		t.Errorf("unpopulated = %v", got)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid multiplier")
		}
	}()
	c.GetQtyMultiplier("bogus")
}

func TestConnectorUnknownField(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unknown field")
		}
	}()
	NewConnector("X", om("bogus_field", "x"))
}

func TestConnectorOldField(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for old field")
		}
	}()
	NewConnector("X", om("pinout", []any{"A"}))
}

func TestCableColorCode(t *testing.T) {
	c := NewCable("W1", om("wirecount", 4, "color_code", "IEC"))
	want := []string{"BN", "RD", "OG", "YE"}
	if !reflect.DeepEqual(c.Colors, want) {
		t.Errorf("colors = %v, want %v", c.Colors, want)
	}
	if c.Wirecount != 4 {
		t.Errorf("wirecount = %d", c.Wirecount)
	}
}

func TestCableDummyColors(t *testing.T) {
	c := NewCable("W1", om("wirecount", 3))
	if !reflect.DeepEqual(c.Colors, []string{"", "", ""}) {
		t.Errorf("colors = %v", c.Colors)
	}
}

func TestCableColorLoop(t *testing.T) {
	c := NewCable("W1", om("wirecount", 5, "colors", strList("RD", "BU")))
	want := []string{"RD", "BU", "RD", "BU", "RD"}
	if !reflect.DeepEqual(c.Colors, want) {
		t.Errorf("colors = %v, want %v", c.Colors, want)
	}
}

func TestCableImplicitWirecount(t *testing.T) {
	c := NewCable("W1", om("colors", strList("RD", "BU", "GN")))
	if c.Wirecount != 3 {
		t.Errorf("wirecount = %d", c.Wirecount)
	}
}

func TestCableUnknownColorCode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unknown color code")
		}
	}()
	NewCable("W1", om("wirecount", 2, "color_code", "BOGUS"))
}

func TestCableGaugeParse(t *testing.T) {
	c := NewCable("W1", om("gauge", "0.25 mm2", "wirecount", 1))
	if c.Gauge != "0.25" || c.GaugeUnit != "mm²" {
		t.Errorf("gauge=%v unit=%q", c.Gauge, c.GaugeUnit)
	}
	c2 := NewCable("W1", om("gauge", "24 awg", "wirecount", 1))
	if c2.GaugeUnit != "AWG" {
		t.Errorf("gauge_unit = %q, want AWG", c2.GaugeUnit)
	}
	c3 := NewCable("W1", om("gauge", 0.25, "wirecount", 1))
	if c3.GaugeUnit != "mm²" {
		t.Errorf("gauge_unit = %q, want mm²", c3.GaugeUnit)
	}
}

func TestCableGaugeBad(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for bad gauge")
		}
	}()
	NewCable("W1", om("gauge", "not a gauge", "wirecount", 1))
}

func TestCableLengthParse(t *testing.T) {
	c := NewCable("W1", om("length", "1 m", "wirecount", 1))
	if c.Length != 1.0 || c.LengthUnit != "m" {
		t.Errorf("length=%v unit=%q", c.Length, c.LengthUnit)
	}
	c2 := NewCable("W1", om("length", 1, "wirecount", 1))
	if c2.Length != 1 || c2.LengthUnit != "m" {
		t.Errorf("length=%v unit=%q", c2.Length, c2.LengthUnit)
	}
	c3 := NewCable("W1", om("length", 0.2, "wirecount", 1))
	if c3.Length != 0.2 || c3.LengthUnit != "m" {
		t.Errorf("length=%v unit=%q", c3.Length, c3.LengthUnit)
	}
}

func TestCableShieldWirelabelError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for shield + s wirelabel")
		}
	}()
	NewCable("W1", om("wirecount", 1, "shield", true, "wirelabels", strList("s")))
}

func TestCableBundlePartData(t *testing.T) {
	c := NewCable("W1", om("category", "bundle", "colors", strList("RD", "BU"), "pn", []any{"A", "B"}))
	if c.Pn == nil {
		t.Errorf("pn should be set")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for list part data on non-bundle")
		}
	}()
	NewCable("W2", om("wirecount", 2, "pn", []any{"A", "B"}))
}

func TestCableShowWirenumbers(t *testing.T) {
	c := NewCable("W1", om("wirecount", 2))
	if !c.ShowWirenumbers {
		t.Errorf("cable should show wire numbers")
	}
	b := NewCable("W1", om("category", "bundle", "wirecount", 2))
	if b.ShowWirenumbers {
		t.Errorf("bundle should hide wire numbers")
	}
}

func TestCableQtyMultiplier(t *testing.T) {
	c := NewCable("W1", om("wirecount", 4, "length", 0.5))
	if got := c.GetQtyMultiplier(""); got != 1 {
		t.Errorf("default = %v", got)
	}
	if got := c.GetQtyMultiplier("wirecount"); got != 4 {
		t.Errorf("wirecount = %v", got)
	}
	c.Connect("X1", 1, 1, "X2", 1)
	c.Connect("X1", 2, 2, "X2", 2)
	if got := c.GetQtyMultiplier("terminations"); got != 2 {
		t.Errorf("terminations = %v", got)
	}
	if got := c.GetQtyMultiplier("length"); got != 0.5 {
		t.Errorf("length = %v", got)
	}
	if got := c.GetQtyMultiplier("total_length"); got != 2.0 {
		t.Errorf("total_length = %v", got)
	}
}

func TestMulNum(t *testing.T) {
	if got := MulNum(1, 4); got != 4 {
		t.Errorf("int*int = %v (%T)", got, got)
	}
	got := MulNum(1.0, 4)
	if got != 4.0 {
		t.Errorf("float*int = %v (%T)", got, got)
	}
}

func TestAdditionalComponentDescription(t *testing.T) {
	c := NewAdditionalComponent(om("type", "Crimp ", "subtype", "Molex KK 254 "))
	if d := c.Description(); d != "Crimp, Molex KK 254" {
		t.Errorf("description = %q", d)
	}
	c2 := NewAdditionalComponent(om("type", "Test"))
	if d := c2.Description(); d != "Test" {
		t.Errorf("description = %q", d)
	}
	if got := c2.Qty; got != 1 {
		t.Errorf("default qty = %v", got)
	}
}

func TestCableConnect(t *testing.T) {
	c := NewCable("W1", om("wirecount", 2))
	c.Connect("X1", 1, 1, "X2", 1)
	if len(c.Connections) != 1 {
		t.Fatalf("connections = %d", len(c.Connections))
	}
	conn := c.Connections[0]
	if conn.FromName != "X1" || conn.FromPin != 1 || conn.ViaPort != 1 || conn.ToName != "X2" || conn.ToPin != 1 {
		t.Errorf("connection = %+v", conn)
	}
}
