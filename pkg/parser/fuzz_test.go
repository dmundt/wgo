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
	"testing"
)

// FuzzLoadYAML verifies the PyYAML-compatible loader never panics on
// arbitrary input. Run with: go test ./pkg/parser -fuzz=FuzzLoadYAML
func FuzzLoadYAML(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("a: 1\n"))
	f.Add([]byte("connectors:\n  X1:\n    pincount: 4\n"))
	f.Add([]byte("connections:\n  - [A, B]\n"))
	f.Add([]byte("<<: {a: 1}\nc: 3\n"))
	f.Add([]byte("08500030: 0123\n"))
	f.Add([]byte("k: yes\n"))
	f.Add([]byte("\x00\xff"))
	f.Add([]byte("[\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = LoadYAML(data)
	})
}

// FuzzRun exercises the full parsing pipeline (templates, connection sets,
// designator resolution, wiring) on arbitrary input. It must never panic or
// hang. Outputs are skipped (empty formats) so nothing touches the filesystem.
// Run with: go test ./pkg/parser -fuzz=FuzzRun
func FuzzRun(f *testing.F) {
	f.Add([]byte("connectors:\n  X1:\n    pincount: 2\n  X2:\n    pincount: 2\ncables:\n  W1:\n    wirecount: 2\nconnections:\n  - [X1, W1, X2]\n"))
	f.Add([]byte("connections:\n  - [BOGUS, BOGUS]\n"))
	f.Add([]byte("connectors:\n  X1:\n    style: simple\nconnections:\n  - [X1, X1]\n"))
	f.Add([]byte("metadata: {}\noptions:\n  template_separator: ':'\nconnections:\n  - [A:1, B:2]\n"))
	f.Add([]byte("tweak:\n  override:\n    graph:\n      rankdir: TB\n"))
	f.Add([]byte("connectors:\n  C:\n    pincount: 4\n    loops:\n      - [1, 2]\nconnections:\n  - [C, C]\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		_ = Run(string(data), "fuzz.yml", dir, "fuzz", nil, []string{dir})
	})
}
