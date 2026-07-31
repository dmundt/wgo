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

import "github.com/dmundt/wgo/pkg/utils"

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
