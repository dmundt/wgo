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

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestStem(t *testing.T) {
	cases := map[string]string{
		"ex01.yml": "ex01",
		"demo01":   "demo01",
		"a.b.c":    "a.b",
		"x":        "x",
	}
	for in, want := range cases {
		if got := stem(in); got != want {
			t.Errorf("stem(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUniqueStrings(t *testing.T) {
	in := []string{"gv", "tsv", "gv", "png"}
	if got := uniqueStrings(in); !reflect.DeepEqual(got, []string{"gv", "tsv", "png"}) {
		t.Errorf("uniqueStrings = %v", got)
	}
}

func TestSplitArgs(t *testing.T) {
	flags, pos := splitArgs([]string{"-f", "gt", "a.yml", "-o", "out", "b.yml"})
	if !reflect.DeepEqual(flags, []string{"-f", "gt", "-o", "out"}) {
		t.Errorf("flags = %v", flags)
	}
	if !reflect.DeepEqual(pos, []string{"a.yml", "b.yml"}) {
		t.Errorf("pos = %v", pos)
	}
	flags, pos = splitArgs([]string{"--format=hpst", "x.yml"})
	if !reflect.DeepEqual(flags, []string{"--format=hpst"}) {
		t.Errorf("flags = %v", flags)
	}
	if !reflect.DeepEqual(pos, []string{"x.yml"}) {
		t.Errorf("pos = %v", pos)
	}
}

// TestCLIEndToEnd builds the binary and runs it on a testdata example,
// verifying the output files are created and the DOT output matches the
// golden file.
func TestCLIEndToEnd(t *testing.T) {
	root := filepath.Join("..", "..")
	dataDir := filepath.Join(root, "testdata")
	ex01 := filepath.Join(dataDir, "ex01.yml")
	if _, err := os.Stat(ex01); err != nil {
		t.Skip("testdata not present")
	}

	bin := filepath.Join(t.TempDir(), "wgo"+exeSuffix())
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	outDir := filepath.Join(t.TempDir(), "out")
	cmd := exec.Command(bin, "-f", "gt", ex01, "-o", outDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wgo failed: %v\n%s", err, output)
	}

	want, err := os.ReadFile(filepath.Join(dataDir, "golden", "ex01.gv"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(outDir, "ex01.gv"))
	if err != nil {
		t.Fatalf("ex01.gv missing: %v", err)
	}
	// Line endings are platform-dependent; compare with universal newlines.
	want = bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n"))
	got = bytes.ReplaceAll(got, []byte("\r\n"), []byte("\n"))
	if !bytes.Equal(want, got) {
		t.Errorf("ex01.gv differs from golden")
	}
	if _, err := os.Stat(filepath.Join(outDir, "ex01.bom.tsv")); err != nil {
		t.Errorf("ex01.bom.tsv missing: %v", err)
	}

	// -V prints the version banner and nothing else.
	vout, err := exec.Command(bin, "-V").CombinedOutput()
	if err != nil {
		t.Fatalf("-V failed: %v", err)
	}
	if !strings.Contains(string(vout), "WireViz 0.4.1") {
		t.Errorf("-V output = %q", string(vout))
	}
}

func exeSuffix() string {
	if len(os.Getenv("GOOS")) > 0 && os.Getenv("GOOS") != "windows" {
		return ""
	}
	return ".exe"
}
