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

package utils

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestPyStr(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, "None"},
		{true, "True"},
		{false, "False"},
		{42, "42"},
		{-7, "-7"},
		{"text", "text"},
		{0.2, "0.2"},
		{1.0, "1.0"},
		{2.0, "2.0"},
		{70.0, "70.0"},
		{1.5, "1.5"},
		{[]any{1, 2}, "[1, 2]"},
		{[]string{"a", "b"}, "['a', 'b']"},
	}
	for _, c := range cases {
		if got := PyStr(c.in); got != c.want {
			t.Errorf("PyStr(%#v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestExpand(t *testing.T) {
	eq := func(got, want []any) bool {
		return reflect.DeepEqual(got, want)
	}
	cases := []struct {
		in   any
		want []any
	}{
		{5, []any{5}},
		{"s", []any{"s"}},
		{[]any{1, 2, 3, 4, 5, 6}, []any{1, 2, 3, 4, 5, 6}},
		{"1-4", []any{1, 2, 3, 4}},
		{"4-1", []any{4, 3, 2, 1}},
		{"3-3", []any{3}},
		{"S-1", []any{"S-1"}},
		{"1-2-3", []any{"1-2-3"}},
		{[]any{"1-2", 6}, []any{1, 2, 6}},
		{"-", []any{"-"}},
	}
	for _, c := range cases {
		if got := Expand(c.in); !eq(got, c.want) {
			t.Errorf("Expand(%#v) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestIsArrow(t *testing.T) {
	cases := map[string]bool{
		"<=>": true,
		"=>":  true,
		"-":   true,
		"=->": false,
		"<==": true,
		"<--": true,
		"--":  true,
		"==":  true,
		"<=":  true,
		"X1":  false,
		"-=":  false,
		"":    false,
	}
	for in, want := range cases {
		if got := IsArrow(in); got != want {
			t.Errorf("IsArrow(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestRemoveLinks(t *testing.T) {
	if got := RemoveLinksStr(`<a href="https://x.example/">Text</a>`); got != "Text" {
		t.Errorf("RemoveLinksStr = %q, want %q", got, "Text")
	}
	if got := RemoveLinksStr("plain"); got != "plain" {
		t.Errorf("RemoveLinksStr plain = %q", got)
	}
	if got := RemoveLinks(42); got != 42 {
		t.Errorf("RemoveLinks(42) = %#v", got)
	}
}

func TestCleanWhitespace(t *testing.T) {
	cases := map[string]string{
		"a   b":   "a b",
		"x , y":   "x, y",
		"a\nb\tc": "a b c",
		"  sp  ":  "sp",
	}
	for in, want := range cases {
		if got := CleanWhitespaceStr(in); got != want {
			t.Errorf("CleanWhitespaceStr(%q) = %q, want %q", in, got, want)
		}
	}
	if got := CleanWhitespace(5); got != 5 {
		t.Errorf("CleanWhitespace(5) = %#v", got)
	}
}

func TestAWGEquiv(t *testing.T) {
	if got := AWGEquiv(0.25); got != "24" {
		t.Errorf("AWGEquiv(0.25) = %q", got)
	}
	if got := AWGEquiv("1"); got != "18" {
		t.Errorf("AWGEquiv(1) = %q", got)
	}
	if got := AWGEquiv(999); got != "Unknown" {
		t.Errorf("AWGEquiv(999) = %q", got)
	}
	if got := MM2Equiv("24"); got != "0.25" {
		t.Errorf("MM2Equiv(24) = %q", got)
	}
	if got := MM2Equiv("999"); got != "Unknown" {
		t.Errorf("MM2Equiv(999) = %q", got)
	}
}

func TestFlatten2d(t *testing.T) {
	in := [][]any{
		{[]any{"x", "y"}, "z"},
		{1, 2.0},
	}
	want := [][]string{
		{"x, y", "z"},
		{"1", "2.0"},
	}
	got := Flatten2d(in)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Flatten2d = %#v, want %#v", got, want)
	}
}

func TestTuplelist2tsv(t *testing.T) {
	in := [][]any{{"a", "b"}, {"1", 2}}
	got := Tuplelist2tsv(in, []string{"Col1", "Col2"})
	want := "Col1\tCol2\na\tb\n1\t2\n"
	if got != want {
		t.Errorf("Tuplelist2tsv = %q, want %q", got, want)
	}
	got = Tuplelist2tsv([][]any{{`<a href="x">T</a>`, "b"}}, nil)
	want = "T\tb\n"
	if got != want {
		t.Errorf("Tuplelist2tsv links = %q, want %q", got, want)
	}
}

func TestOrderedMap(t *testing.T) {
	m := NewOrderedMap()
	m.Set("b", 2)
	m.Set("a", 1)
	m.Set("c", 3)
	if !reflect.DeepEqual(m.Keys(), []string{"b", "a", "c"}) {
		t.Errorf("Keys = %v", m.Keys())
	}
	m.Set("b", 20) // overwrite keeps position
	if !reflect.DeepEqual(m.Keys(), []string{"b", "a", "c"}) {
		t.Errorf("Keys after overwrite = %v", m.Keys())
	}
	if v, _ := m.Get("b"); v != 20 {
		t.Errorf("Get(b) = %v", v)
	}
	if _, ok := m.Get("x"); ok {
		t.Errorf("Get(x) should be absent")
	}
	if !m.Has("a") {
		t.Errorf("Has(a) should be true")
	}
	if m.Len() != 3 {
		t.Errorf("Len = %d", m.Len())
	}
	items := m.Items()
	if items[0].Key != "b" || items[1].Key != "a" || items[2].Key != "c" {
		t.Errorf("Items order = %v", items)
	}
	f := m.First()
	if f.Key != "b" || f.Value != 20 {
		t.Errorf("First = %v", f)
	}
}

func TestSmartFileResolve(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(sub, "x.png")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SmartFileResolve("x.png", []string{sub})
	if err != nil {
		t.Fatalf("SmartFileResolve: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("result not absolute: %q", got)
	}
	if _, err := SmartFileResolve("missing.png", []string{dir}); err == nil {
		t.Errorf("expected error for missing file")
	}
	if _, err := SmartFileResolve(file, []string{dir}); err != nil {
		t.Errorf("absolute existing should resolve: %v", err)
	}
	if _, err := SmartFileResolve(filepath.Join(dir, "nope.png"), []string{dir}); err == nil {
		t.Errorf("expected error for absolute missing file")
	}
}

func TestLineEndings(t *testing.T) {
	got := LineEndings("a\nb")
	want := "a\nb"
	if runtime.GOOS == "windows" {
		want = "a\r\nb"
	}
	if got != want {
		t.Errorf("LineEndings = %q, want %q", got, want)
	}
}

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.White)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestAspectRatio(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "img.png")
	writePNG(t, img, 100, 50)
	if got := AspectRatio(img); got != 2.0 {
		t.Errorf("AspectRatio = %v, want 2", got)
	}
	if got := AspectRatio(filepath.Join(dir, "missing.png")); got != 1.0 {
		t.Errorf("AspectRatio missing = %v, want 1", got)
	}
}

func TestFileWriteText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	if err := FileWriteText(path, "a\nb\n"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		if string(got) != "a\r\nb\r\n" {
			t.Errorf("file content = %q", string(got))
		}
	} else if string(got) != "a\nb\n" {
		t.Errorf("file content = %q", string(got))
	}
}
