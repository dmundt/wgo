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

package parser

import (
	"math"
	"reflect"
	"testing"

	"github.com/dmundt/wgo/pkg/utils"
)

func resolveKey(t *testing.T, doc string) any {
	t.Helper()
	v, err := LoadYAML([]byte(doc))
	if err != nil {
		t.Fatalf("LoadYAML(%q): %v", doc, err)
	}
	om, ok := v.(*utils.OrderedMap)
	if !ok {
		t.Fatalf("LoadYAML(%q): expected mapping, got %T", doc, v)
	}
	val, _ := om.Get("k")
	return val
}

// TestScalarResolution pins the PyYAML (YAML 1.1) scalar-resolution rules that
// wgo must match byte-for-byte.
func TestScalarResolution(t *testing.T) {
	cases := []struct {
		in   string
		want any
	}{
		{"k: 08500030", "08500030"}, // leading-zero decimal stays a string
		{"k: 0123", 83},             // leading zero is octal
		{"k: 0x2A", 42},
		{"k: 0b1010", 10},
		{"k: 1_000", 1000},
		{"k: yes", true},
		{"k: No", false},
		{"k: on", true},
		{"k: OFF", false},
		{"k: 1e3", "1e3"}, // no decimal point: stays a string
		{"k: 1.5", 1.5},
		{"k: -0.5", -0.5},
		{"k: 1:30", 90},     // sexagesimal int
		{"k: 1:30.5", 90.5}, // sexagesimal float
		{"k: -1:30.5", -90.5},
		{"k: 0:30.5", 30.5},
		{"k: 12:34:56.78", 45296.78},
		{"k: 1:30.5e2", "1:30.5e2"},   // no exponent in sexagesimal floats
		{"k: '08500030'", "08500030"}, // quoted: always string
		{`k: "yes"`, "yes"},
		{"k: |\n  yes\n", "yes\n"}, // block literal: string
		{"k: ~", nil},
		{"k: null", nil},
		{"k: NULL", nil},
	}
	for _, c := range cases {
		got := resolveKey(t, c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s = %#v (%T), want %#v (%T)", c.in, got, got, c.want, c.want)
		}
	}
}

func TestScalarResolutionInf(t *testing.T) {
	got := resolveKey(t, "k: .inf")
	if f, ok := got.(float64); !ok || !math.IsInf(f, 1) {
		t.Errorf(".inf = %#v, want +Inf", got)
	}
	got = resolveKey(t, "k: -.Inf")
	if f, ok := got.(float64); !ok || !math.IsInf(f, -1) {
		t.Errorf("-.Inf = %#v, want -Inf", got)
	}
}

// TestBlockScalarChomping pins the chomping behavior at end of file, which
// PyYAML and yaml.v3 must agree on. A clip-chomped block scalar keeps its
// trailing newline only when the document ends with one.
func TestBlockScalarChomping(t *testing.T) {
	cases := []struct {
		in   string
		want any
	}{
		{"k: |\n  yes", "yes"},      // clip, EOF without newline
		{"k: |\n  yes\n", "yes\n"},  // clip, EOF with newline
		{"k: |-\n  yes\n", "yes"},   // strip
		{"k: |+\n  yes\n", "yes\n"}, // keep
		{"k: >\n  a\n  b", "a b"},   // folded, EOF without newline
		{"k: >\n  a\n  b\n", "a b\n"},
	}
	for _, c := range cases {
		got := resolveKey(t, c.in)
		if got != c.want {
			t.Errorf("%q = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestScalarResolutionTopLevelTypes(t *testing.T) {
	// Non-mapping documents produce the raw value, not an error.
	for _, c := range []struct {
		in   string
		want any
	}{
		{"42", 42},
		{"[1, 2]", []any{1, 2}},
		{"'str'", "str"},
		{"~", nil},
	} {
		v, err := LoadYAML([]byte(c.in))
		if err != nil {
			t.Fatalf("LoadYAML(%q): %v", c.in, err)
		}
		if !reflect.DeepEqual(v, c.want) {
			t.Errorf("LoadYAML(%q) = %#v, want %#v", c.in, v, c.want)
		}
	}
}

// TestLoadYAMLPreservesOrder verifies mapping keys keep insertion order like
// Python dicts, which the DOT/BOM output depends on.
func TestLoadYAMLPreservesOrder(t *testing.T) {
	v, err := LoadYAML([]byte("z: 1\na: 2\nm: 3\n"))
	if err != nil {
		t.Fatal(err)
	}
	om := v.(*utils.OrderedMap)
	if got, want := om.Keys(), []string{"z", "a", "m"}; !reflect.DeepEqual(got, want) {
		t.Errorf("keys = %v, want %v", got, want)
	}
}

// TestLoadYAMLMergeKeys verifies YAML merge keys (<<) with anchors, including
// that an explicit key overrides the merged value but keeps merge order.
func TestLoadYAMLMergeKeys(t *testing.T) {
	doc := "base: &b\n  x: 1\n  y: 2\nderived:\n  <<: *b\n  y: 3\n  z: 4\n"
	v, err := LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	om := v.(*utils.OrderedMap)
	dv, _ := om.Get("derived")
	derived := dv.(*utils.OrderedMap)
	if got, want := derived.Keys(), []string{"x", "y", "z"}; !reflect.DeepEqual(got, want) {
		t.Errorf("derived keys = %v, want %v", got, want)
	}
	if got, _ := derived.Get("x"); got != 1 {
		t.Errorf("x = %v, want 1", got)
	}
	if got, _ := derived.Get("y"); got != 3 {
		t.Errorf("y = %v, want 3 (own key overrides merge)", got)
	}
	if got, _ := derived.Get("z"); got != 4 {
		t.Errorf("z = %v, want 4", got)
	}
}

// TestLoadYAMLAnchors verifies alias resolution across the document.
func TestLoadYAMLAnchors(t *testing.T) {
	doc := "a: &x\n  p: 1\nb: *x\n"
	v, err := LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	om := v.(*utils.OrderedMap)
	av, _ := om.Get("b")
	b := av.(*utils.OrderedMap)
	if got, _ := b.Get("p"); got != 1 {
		t.Errorf("b.p = %v, want 1", got)
	}
}
