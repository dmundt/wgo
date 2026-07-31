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

package graph

import (
	"reflect"
	"testing"

	"github.com/dmundt/wgo/pkg/model"
)

func TestQuote(t *testing.T) {
	cases := map[string]string{
		"X1":          "X1",
		"LR":          "LR",
		"0.33":        "0.33",
		"-4.2":        "-4.2",
		"#FFFFFF":     `"#FFFFFF"`,
		"a b":         `"a b"`,
		"node":        `"node"`,
		"Node":        `"Node"`,
		"":            `""`,
		"<a>html</a>": "<a>html</a>",
	}
	for in, want := range cases {
		if got := Quote(in); got != want {
			t.Errorf("Quote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestQuoteEdge(t *testing.T) {
	cases := map[string]string{
		"X1:p1r:e": "X1:p1r:e",
		"W1:w1:w":  "W1:w1:w",
		"a b:c":    `"a b":c`,
		"x":        "x",
		"a:b":      "a:b",
		"a":        "a",
	}
	for in, want := range cases {
		if got := QuoteEdge(in); got != want {
			t.Errorf("QuoteEdge(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAttrList(t *testing.T) {
	cases := []struct {
		label string
		attrs map[string]string
		want  string
	}{
		{"", nil, ""},
		{"", map[string]string{"color": "#000000"}, ` [color="#000000"]`},
		{"spam spam", nil, ` [label="spam spam"]`},
		{"l", map[string]string{"b": "2", "a": "1"}, " [label=l a=1 b=2]"},
		{"", map[string]string{"rankdir": "LR", "ranksep": "2"}, " [rankdir=LR ranksep=2]"},
	}
	for _, c := range cases {
		if got := AttrList(c.label, c.attrs); got != c.want {
			t.Errorf("AttrList(%q, %v) = %q, want %q", c.label, c.attrs, got, c.want)
		}
	}
}

func TestGraphSource(t *testing.T) {
	g := NewGraph()
	if got := g.Source(); got != "graph {\n}\n" {
		t.Errorf("empty source = %q", got)
	}
	g.Attr("graph", map[string]string{"rankdir": "LR"})
	g.Node("X1", "<b>x</b>", map[string]string{"shape": "box", "fillcolor": "#FFFFFF"})
	g.Edge("X1:p1r:e", "W1:w1:w", "", nil)
	want := "graph {\n" +
		"\tgraph [rankdir=LR]\n" +
		"\tX1 [label=<b>x</b> fillcolor=\"#FFFFFF\" shape=box]\n" +
		"\tX1:p1r:e -- W1:w1:w\n" +
		"}\n"
	if got := g.Source(); got != want {
		t.Errorf("source = %q, want %q", got, want)
	}
}

func TestNestedHTMLTable(t *testing.T) {
	rows := []any{
		[]any{"X1"},
		[]any{"Molex KK 254", "female", "4-pin"},
		"<!-- connector table -->",
		[]any{"cell"},
		[]any{nil},
		nil,
	}
	got := NestedHTMLTable(rows, "")
	want := []string{
		`<table border="0" cellspacing="0" cellpadding="0">`,
		` <tr><td>`,
		`  <table border="0" cellspacing="0" cellpadding="3" cellborder="1"><tr>`,
		`   <td balign="left">X1</td>`,
		`  </tr></table>`,
		` </td></tr>`,
		` <tr><td>`,
		`  <table border="0" cellspacing="0" cellpadding="3" cellborder="1"><tr>`,
		`   <td balign="left">Molex KK 254</td>`,
		`   <td balign="left">female</td>`,
		`   <td balign="left">4-pin</td>`,
		`  </tr></table>`,
		` </td></tr>`,
		` <tr><td>`,
		`  <!-- connector table -->`,
		` </td></tr>`,
		` <tr><td>`,
		`  <table border="0" cellspacing="0" cellpadding="3" cellborder="1"><tr>`,
		`   <td balign="left">cell</td>`,
		`  </tr></table>`,
		` </td></tr>`,
		`</table>`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NestedHTMLTable:\n got %q\nwant %q", got, want)
	}
}

func TestNestedHTMLTableEmpty(t *testing.T) {
	got := NestedHTMLTable([]any{[]any{nil}}, "")
	if got[len(got)-2] != "<tr><td></td></tr>" {
		t.Errorf("empty table should include placeholder, got %v", got)
	}
}

func TestNestedHTMLTableTdX(t *testing.T) {
	rows := []any{[]any{"<tdX bgcolor=\"#FFFF00\" width=\"4\">"}}
	got := NestedHTMLTable(rows, "")
	if got[3] != `   <td balign="left" bgcolor="#FFFF00" width="4"></td>` {
		t.Errorf("tdX injection failed: %q", got[3])
	}
}

func TestHTMLBgcolorAttr(t *testing.T) {
	if got := HTMLBgcolorAttr("YE"); got != ` bgcolor="#FFFF00"` {
		t.Errorf("HTMLBgcolorAttr(YE) = %q", got)
	}
	if got := HTMLBgcolorAttr(""); got != "" {
		t.Errorf("HTMLBgcolorAttr() = %q", got)
	}
}

func TestHTMLColorbar(t *testing.T) {
	if got := HTMLColorbar("YE"); got != `<tdX bgcolor="#FFFF00" width="4">` {
		t.Errorf("HTMLColorbar(YE) = %v", got)
	}
	if got := HTMLColorbar(""); got != nil {
		t.Errorf("HTMLColorbar() = %v", got)
	}
}

func TestHTMLLineBreaks(t *testing.T) {
	if got := HTMLLineBreaks("a\nb"); got != "a<br />b" {
		t.Errorf("HTMLLineBreaks = %v", got)
	}
	if got := HTMLLineBreaks(`<a href="x">T</a>` + "\nb"); got != "T<br />b" {
		t.Errorf("HTMLLineBreaks link = %v", got)
	}
	if got := HTMLLineBreaks(nil); got != nil {
		t.Errorf("HTMLLineBreaks(nil) = %v", got)
	}
}

func TestHTMLImage(t *testing.T) {
	i1 := &model.Image{Src: "x.png", Scale: "false"}
	if got := HTMLImage(i1); got != `<tdX><img scale="false" src="x.png"/>` {
		t.Errorf("HTMLImage i1 = %v", got)
	}
	i2 := &model.Image{Src: "x.png", Scale: "both", Width: 70.0, Height: 70, Fixedsize: true}
	want2 := "<tdX>\n    <table border=\"0\" cellspacing=\"0\" cellborder=\"0\"><tr>\n     <td width=\"70.0\" height=\"70\" fixedsize=\"true\"><img scale=\"both\" src=\"x.png\"/></td>\n    </tr></table>\n   "
	if got := HTMLImage(i2); got != want2 {
		t.Errorf("HTMLImage i2:\n got %q\nwant %q", got, want2)
	}
	i3 := &model.Image{Src: "x.png", Scale: "false", Caption: "Cap"}
	if got := HTMLImage(i3); got != `<tdX sides="TLR"><img scale="false" src="x.png"/>` {
		t.Errorf("HTMLImage i3 = %v", got)
	}
	if got := HTMLImage(nil); got != nil {
		t.Errorf("HTMLImage(nil) = %v", got)
	}
}

func TestHTMLCaption(t *testing.T) {
	if got := HTMLCaption(&model.Image{Src: "x", Caption: "Cap"}); got != `<tdX sides="BLR">Cap` {
		t.Errorf("HTMLCaption = %v", got)
	}
	if got := HTMLCaption(&model.Image{Src: "x"}); got != nil {
		t.Errorf("HTMLCaption no caption = %v", got)
	}
}

func TestHTMLSizeAttr(t *testing.T) {
	if got := HTMLSizeAttr(&model.Image{Src: "x", Width: 70.0, Height: 70, Fixedsize: true}); got != ` width="70.0" height="70" fixedsize="true"` {
		t.Errorf("HTMLSizeAttr = %q", got)
	}
	if got := HTMLSizeAttr(&model.Image{Src: "x"}); got != "" {
		t.Errorf("HTMLSizeAttr empty = %q", got)
	}
}
