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

// Application constants mirroring wireviz/__init__.py.
const (
	AppName = "WireViz"
	AppURL  = "https://github.com/wireviz/WireViz"
	Version = "0.4.1"
)

// ColorHexLookup returns the hex value for a color name.
func ColorHexLookup(name string) (string, bool) {
	v, ok := colorHex[name]
	return v, ok
}
