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
