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
	"reflect"
	"strings"
	"testing"

	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

func om(pairs ...any) *utils.OrderedMap {
	m := utils.NewOrderedMap()
	for i := 0; i+1 < len(pairs); i += 2 {
		m.Set(pairs[i].(string), pairs[i+1])
	}
	return m
}

func newTestHarness() *Harness {
	return NewHarness(utils.NewOrderedMap(), model.NewOptions(nil), model.NewTweak(nil))
}

func TestPnInfoString(t *testing.T) {
	cases := []struct {
		header, name, number string
		want                 any
	}{
		{PnHeader, "", "ABC", "P/N: ABC"},
		{MpnHeader, "Molex", "08500030", "Molex: 08500030"},
		{SpnHeader, "Digimouse", "", "Digimouse"},
		{PnHeader, "", "", nil},
		{PnHeader, "", "  X  ", "P/N: X"},
	}
	for _, c := range cases {
		got := pnInfoString(c.header, c.name, c.number)
		if got != c.want {
			t.Errorf("pnInfoString(%q,%q,%q) = %v, want %v", c.header, c.name, c.number, got, c.want)
		}
	}
}

func TestComponentTableEntry(t *testing.T) {
	cases := []struct {
		args []any
		want string
	}{
		{[]any{"Crimp, Molex KK 254, 22-30 AWG", 4, "", "", "", "Molex", "08500030", "", ""},
			`<table border="0" cellspacing="0" cellpadding="3" cellborder="1"><tr>` + "\n" +
				`   <td align="left" balign="left">4 x Crimp, Molex KK 254, 22-30 AWG<br/>Molex: 08500030</td>` + "\n" +
				`  </tr></table>`},
		{[]any{"Test", 1.5, "m", "", "ABC", "Molex", "45454", "Mousikey", "9999"},
			`<table border="0" cellspacing="0" cellpadding="3" cellborder="1"><tr>` + "\n" +
				`   <td align="left" balign="left">1.5 m x Test<br/>P/N: ABC, Molex: 45454, Mousikey: 9999</td>` + "\n" +
				`  </tr></table>`},
		{[]any{"#4 (Crimp)", 4, "m", "YE", "", "", "", "", ""},
			`<table border="0" cellspacing="0" cellpadding="3" cellborder="1" bgcolor="#FFFF00"><tr>` + "\n" +
				`   <td align="left" balign="left">4 m x #4 (Crimp)</td>` + "\n" +
				`  </tr></table>`},
	}
	for i, c := range cases {
		got := componentTableEntry(c.args[0].(string), c.args[1], c.args[2].(string), c.args[3].(string),
			c.args[4].(string), c.args[5].(string), c.args[6].(string), c.args[7].(string), c.args[8].(string))
		if got != c.want {
			t.Errorf("case %d:\n got %q\nwant %q", i, got, c.want)
		}
	}
}

func TestGenerateBom(t *testing.T) {
	h := newTestHarness()
	h.AddConnector("X1", om("type", "Molex KK 254", "subtype", "female",
		"pinlabels", []any{"GND", "VCC", "RX", "TX"}))
	h.AddConnector("X2", om("type", "Molex KK 254", "subtype", "female",
		"pinlabels", []any{"GND", "VCC", "RX", "TX"}))
	h.AddCable("W1", om("wirecount", 4, "color_code", "IEC", "gauge", "0.25 mm2",
		"length", 0.2, "shield", true))

	bom := GenerateBom(h)
	if len(bom) != 2 {
		t.Fatalf("expected 2 BOM entries, got %d: %+v", len(bom), bom)
	}
	// Cable entry first (sorted by key).
	c0 := bom[0]
	if c0.Description != "Cable, 4 x 0.25 mm² shielded" {
		t.Errorf("cable description = %q", c0.Description)
	}
	if c0.Qty != 0.2 {
		t.Errorf("cable qty = %v", c0.Qty)
	}
	if c0.Unit != "m" {
		t.Errorf("cable unit = %q", c0.Unit)
	}
	if !reflect.DeepEqual(c0.Designators, []string{"W1"}) {
		t.Errorf("cable designators = %v", c0.Designators)
	}
	if c0.ID != 1 {
		t.Errorf("cable id = %d", c0.ID)
	}
	// Connector entry: qty summed and deduped.
	c1 := bom[1]
	if c1.Description != "Connector, Molex KK 254, female, 4 pins" {
		t.Errorf("connector description = %q", c1.Description)
	}
	if c1.Qty != 2 {
		t.Errorf("connector qty = %v", c1.Qty)
	}
	if !reflect.DeepEqual(c1.Designators, []string{"X1", "X2"}) {
		t.Errorf("connector designators = %v", c1.Designators)
	}
}

func TestGenerateBomBundle(t *testing.T) {
	h := newTestHarness()
	h.AddCable("W2", om("category", "bundle", "colors", []any{"YE", "BK", "BK", "RD"},
		"gauge", "0.25 mm2", "length", 1,
		"pn", []any{"WIRE1", "WIRE2", "WIRE2", "WIRE3"}))
	bom := GenerateBom(h)
	if len(bom) != 3 {
		t.Fatalf("expected 3 bundle wire entries, got %d", len(bom))
	}
	// Keys sort by description: BK, RD, YE.
	if bom[0].Description != "Wire, 0.25 mm², BK" {
		t.Errorf("bom[0] = %q", bom[0].Description)
	}
	if bom[0].Qty != 2 {
		t.Errorf("BK qty = %v", bom[0].Qty)
	}
	if bom[0].Pn != "WIRE2" {
		t.Errorf("BK pn = %v", bom[0].Pn)
	}
	if bom[1].Pn != "WIRE3" || bom[2].Pn != "WIRE1" {
		t.Errorf("pn order wrong: %v, %v", bom[1].Pn, bom[2].Pn)
	}
}

func TestGenerateBomIgnoreInBom(t *testing.T) {
	h := newTestHarness()
	h.AddConnector("X1", om("pincount", 2, "ignore_in_bom", true))
	bom := GenerateBom(h)
	if len(bom) != 0 {
		t.Errorf("expected no BOM entries, got %d", len(bom))
	}
}

func TestBomList(t *testing.T) {
	h := newTestHarness()
	h.AddConnector("X1", om("type", "Molex KK 254", "subtype", "female", "pincount", 4))
	h.AddCable("W1", om("wirecount", 2, "length", 0.5))
	rows := bomList(GenerateBom(h))
	wantHeader := []any{"Id", "Description", "Qty", "Unit", "Designators"}
	if !reflect.DeepEqual(rows[0], wantHeader) {
		t.Errorf("header = %v", rows[0])
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[1][0] != "1" || rows[1][1] != "Cable, 2 wires" || rows[1][2] != "0.5" || rows[1][3] != "m" || rows[1][4] != "W1" {
		t.Errorf("cable row = %v", rows[1])
	}
	if rows[2][2] != "1" || rows[2][4] != "X1" {
		t.Errorf("connector row = %v", rows[2])
	}
}

func TestGetBomIndex(t *testing.T) {
	h := newTestHarness()
	h.AddConnector("X1", om("pincount", 2, "additional_components", []any{
		om("type", "Crimp", "subtype", "x", "qty_multiplier", "pincount"),
	}))
	bom := GenerateBom(h)
	part := h.Connectors[0].AdditionalComponents[0]
	id := getBomIndex(bom, partKey(part))
	found := false
	for _, e := range bom {
		if e.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("getBomIndex = %d not in bom", id)
	}
}

func TestBomEntryKey(t *testing.T) {
	e := BOMEntry{Description: "Connector, Molex, female", Unit: "", Pn: "CON4"}
	k := bomEntryKey(e)
	if k[0] != "Connector, Molex, female" {
		t.Errorf("key[0] = %q", k[0])
	}
	e2 := BOMEntry{Description: "same", Pn: `<a href="x">Molex</a>`}
	e3 := BOMEntry{Description: "same", Pn: `<a href="x">Molex</a>`}
	if compareKeys(bomEntryKey(e2), bomEntryKey(e3)) != 0 {
		t.Errorf("identical entries should have equal keys")
	}
	e4 := BOMEntry{Description: "different", Pn: `<a href="x">Molex</a>`}
	if compareKeys(bomEntryKey(e2), bomEntryKey(e4)) == 0 {
		t.Errorf("entries differing in description should have different keys")
	}
	e5 := BOMEntry{Description: "a   b  , c"}
	if bomEntryKey(e5)[0] != "a b, c" {
		t.Errorf("description not cleaned: %q", bomEntryKey(e5)[0])
	}
}

func TestConnectErrors(t *testing.T) {
	assertPanicMsg := func(fn func(), contains string) {
		t.Helper()
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("expected panic containing %q", contains)
				return
			}
			if !strings.Contains(strings.ToLower(fmt.Sprintf("%v", r)), strings.ToLower(contains)) {
				t.Errorf("panic %v does not contain %q", r, contains)
			}
		}()
		fn()
	}

	h := newTestHarness()
	h.AddConnector("X1", om("pincount", 2))
	h.AddCable("W1", om("wirecount", 1))
	assertPanicMsg(func() { h.Connect("X1", 9, "W1", 1, "", nil) }, "not found")

	// Pin defined in both pins and pinlabels at different positions is ambiguous.
	h2 := newTestHarness()
	h2.AddConnector("X1", om("pins", []any{1, 2}, "pinlabels", []any{2, 1}))
	h2.AddCable("W1", om("wirecount", 1))
	assertPanicMsg(func() { h2.Connect("X1", 2, "W1", 1, "", nil) }, "defined both in pinlabels and pins")

	// Pin label defined more than once.
	h3 := newTestHarness()
	h3.AddConnector("X1", om("pins", []any{1, 2, 3}, "pinlabels", []any{"A", "A", "B"}))
	h3.AddCable("W1", om("wirecount", 1))
	assertPanicMsg(func() { h3.Connect("X1", "A", "W1", 1, "", nil) }, "defined more than once")

	// Pin not found at all.
	h4 := newTestHarness()
	h4.AddConnector("X1", om("pins", []any{1, 2}, "pinlabels", []any{"A", "B"}))
	h4.AddCable("W1", om("wirecount", 1))
	assertPanicMsg(func() { h4.Connect("X1", "C", "W1", 1, "", nil) }, "not found")

	// Cable color used for more than one wire.
	h5 := newTestHarness()
	h5.AddConnector("X1", om("pincount", 4))
	h5.AddCable("W1", om("wirecount", 4, "colors", []any{"BN", "BN", "GN", "YE"}))
	assertPanicMsg(func() { h5.Connect("X1", 1, "W1", "BN", "", nil) }, "used for more than one wire")

	// Valid connection must not panic.
	h6 := newTestHarness()
	h6.AddConnector("X1", om("pincount", 2))
	h6.AddConnector("X2", om("pincount", 2))
	h6.AddCable("W1", om("wirecount", 2))
	h6.Connect("X1", 1, "W1", 1, "X2", 1)
	h6.Connect("X1", 2, "W1", 2, "X2", 2)
	c, _ := h6.CableByName("W1")
	if len(c.Connections) != 2 {
		t.Errorf("connections = %d", len(c.Connections))
	}
}

func TestAddMateActivatesPins(t *testing.T) {
	h := newTestHarness()
	h.AddConnector("X1", om("pincount", 2))
	h.AddConnector("X2", om("pincount", 2))
	h.AddMatePin("X1", 1, "X2", 2, "<->")
	if len(h.Mates) != 1 {
		t.Fatalf("mates = %d", len(h.Mates))
	}
	c1, _ := h.ConnectorByName("X1")
	c2, _ := h.ConnectorByName("X2")
	if !c1.PortsRight || !c2.PortsLeft {
		t.Errorf("ports not activated: X1.right=%v X2.left=%v", c1.PortsRight, c2.PortsLeft)
	}
}

func TestCreateGraphSmoke(t *testing.T) {
	h := newTestHarness()
	h.AddConnector("X1", om("type", "Molex KK 254", "subtype", "female", "pincount", 2))
	h.AddConnector("X2", om("type", "Molex KK 254", "subtype", "female", "pincount", 2))
	h.AddCable("W1", om("wirecount", 2, "color_code", "IEC", "gauge", "0.25 mm2", "length", 0.2))
	h.Connect("X1", 1, "W1", 1, "X2", 1)
	h.Connect("X1", 2, "W1", 2, "X2", 2)
	g := h.CreateGraph()
	src := g.Source()
	for _, want := range []string{
		"// Graph generated by WireViz " + model.Version,
		"// https://github.com/wireviz/WireViz",
		"\tgraph [bgcolor=\"#FFFFFF\" fontname=arial nodesep=0.33 rankdir=LR ranksep=2]",
		"\tX1 [label=<",
		"X1:p1r:e -- W1:w1:w",
		"W1:w2:e -- X2:p2l:w",
		"\tW1 [label=<",
		"}\n",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("source missing %q", want)
		}
	}
}
