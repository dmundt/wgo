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
	"math"
	"sort"
	"strings"

	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

// BOM column headers.
const (
	PnHeader  = "P/N"
	MpnHeader = "MPN"
	SpnHeader = "SPN"
)

var bomColumnsAlways = []string{"id", "description", "qty", "unit", "designators"}
var bomColumnsOptional = []string{"pn", "manufacturer", "mpn", "supplier", "spn"}

// BOMEntry mirrors the WireViz BOM entry dict.
type BOMEntry struct {
	ID           int
	Description  string
	Qty          any
	Unit         string
	Designators  any // string, []string, or nil
	Pn           any
	Manufacturer any
	Mpn          any
	Supplier     any
	Spn          any
}

// NewBOMEntryFromMap builds a BOM entry from a YAML dict (additional_bom_items).
func NewBOMEntryFromMap(m *utils.OrderedMap) BOMEntry {
	e := BOMEntry{}
	if v, ok := m.Get("description"); ok && v != nil {
		e.Description = utils.PyStr(v)
	}
	if v, ok := m.Get("qty"); ok && v != nil {
		e.Qty = v
	}
	if v, ok := m.Get("unit"); ok && v != nil {
		e.Unit = utils.PyStr(v)
	}
	if v, ok := m.Get("designators"); ok && v != nil {
		e.Designators = v
	}
	if v, ok := m.Get("pn"); ok && v != nil {
		e.Pn = v
	}
	if v, ok := m.Get("manufacturer"); ok && v != nil {
		e.Manufacturer = v
	}
	if v, ok := m.Get("mpn"); ok && v != nil {
		e.Mpn = v
	}
	if v, ok := m.Get("supplier"); ok && v != nil {
		e.Supplier = v
	}
	if v, ok := m.Get("spn"); ok && v != nil {
		e.Spn = v
	}
	return e
}

// GenerateBom mirrors wv_bom.generate_bom.
func GenerateBom(h *Harness) []BOMEntry {
	var entries []BOMEntry
	for _, connector := range h.Connectors {
		if !connector.IgnoreInBom {
			description := "Connector"
			if truthyVal(connector.Type) {
				description += ", " + utils.PyStr(connector.Type)
			}
			if truthyVal(connector.Subtype) {
				description += ", " + utils.PyStr(connector.Subtype)
			}
			if connector.ShowPincount {
				description += fmt.Sprintf(", %d pins", connector.Pincount)
			}
			if connector.Color != "" {
				description += ", " + model.TranslateColor(connector.Color, h.Options.ColorMode)
			}
			var designators any
			if connector.ShowName {
				designators = connector.Name
			}
			entries = append(entries, BOMEntry{
				Description:  description,
				Designators:  designators,
				Pn:           strOrNil(connector.Pn),
				Manufacturer: strOrNil(connector.Manufacturer),
				Mpn:          strOrNil(connector.Mpn),
				Supplier:     strOrNil(connector.Supplier),
				Spn:          strOrNil(connector.Spn),
			})
		}
		entries = append(entries, getAdditionalComponentBom(connector)...)
	}

	for _, cable := range h.Cables {
		if !cable.IgnoreInBom {
			if cable.Category != "bundle" {
				description := "Cable"
				if truthyVal(cable.Type) {
					description += ", " + utils.PyStr(cable.Type)
				}
				description += fmt.Sprintf(", %d", cable.Wirecount)
				if cable.Gauge != nil {
					description += fmt.Sprintf(" x %s %s", utils.PyStr(cable.Gauge), cable.GaugeUnit)
				} else {
					description += " wires"
				}
				if shieldTruthy(cable.Shield) {
					description += " shielded"
				}
				if cable.Color != "" {
					description += ", " + model.TranslateColor(cable.Color, h.Options.ColorMode)
				}
				var designators any
				if cable.ShowName {
					designators = cable.Name
				}
				entries = append(entries, BOMEntry{
					Description:  description,
					Qty:          cable.Length,
					Unit:         cable.LengthUnit,
					Designators:  designators,
					Pn:           cable.Pn,
					Manufacturer: cable.Manufacturer,
					Mpn:          cable.Mpn,
					Supplier:     cable.Supplier,
					Spn:          cable.Spn,
				})
			} else {
				for index, color := range cable.Colors {
					description := "Wire"
					if truthyVal(cable.Type) {
						description += ", " + utils.PyStr(cable.Type)
					}
					if cable.Gauge != nil {
						description += fmt.Sprintf(", %s %s", utils.PyStr(cable.Gauge), cable.GaugeUnit)
					}
					if color != "" {
						description += ", " + model.TranslateColor(color, h.Options.ColorMode)
					}
					var designators any
					if cable.ShowName {
						designators = cable.Name
					}
					entries = append(entries, BOMEntry{
						Description:  description,
						Qty:          cable.Length,
						Unit:         cable.LengthUnit,
						Designators:  designators,
						Pn:           indexIfList(cable.Pn, index),
						Manufacturer: indexIfList(cable.Manufacturer, index),
						Mpn:          indexIfList(cable.Mpn, index),
						Supplier:     indexIfList(cable.Supplier, index),
						Spn:          indexIfList(cable.Spn, index),
					})
				}
			}
		}
		entries = append(entries, getAdditionalComponentBom(cable)...)
	}

	entries = append(entries, h.AdditionalBomItems...)

	for i := range entries {
		entries[i].Description = utils.CleanWhitespaceStr(entries[i].Description)
		entries[i].Unit = utils.CleanWhitespaceStr(entries[i].Unit)
		entries[i].Designators = cleanAny(entries[i].Designators)
		entries[i].Pn = cleanAny(entries[i].Pn)
		entries[i].Manufacturer = cleanAny(entries[i].Manufacturer)
		entries[i].Mpn = cleanAny(entries[i].Mpn)
		entries[i].Supplier = cleanAny(entries[i].Supplier)
		entries[i].Spn = cleanAny(entries[i].Spn)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return compareKeys(bomEntryKey(entries[i]), bomEntryKey(entries[j])) < 0
	})

	var bom []BOMEntry
	i := 0
	for i < len(entries) {
		j := i
		for j < len(entries) && keysEqual(bomEntryKey(entries[j]), bomEntryKey(entries[i])) {
			j++
		}
		group := entries[i:j]
		var designators []string
		totalQty := 0.0
		for _, e := range group {
			designators = append(designators, makeList(e.Designators)...)
			q := e.Qty
			if q == nil {
				q = 1
			}
			totalQty += toFloat(q)
		}
		var qty any
		if math.Mod(totalQty, 1) == 0 {
			qty = int(totalQty)
		} else {
			qty = math.Round(totalQty*1000) / 1000
		}
		entry0 := group[0]
		entry0.Qty = qty
		entry0.Designators = sortedSet(designators)
		bom = append(bom, entry0)
		i = j
	}

	for idx := range bom {
		bom[idx].ID = idx + 1
	}
	return bom
}

func strOrNil(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func cleanAny(v any) any {
	switch x := v.(type) {
	case string:
		return utils.CleanWhitespaceStr(x)
	case []string:
		out := make([]string, len(x))
		for i, e := range x {
			out[i] = utils.CleanWhitespaceStr(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = utils.CleanWhitespace(e)
		}
		return out
	}
	return v
}

func indexIfList(v any, index int) any {
	if l, ok := v.([]any); ok {
		if index < len(l) {
			return l[index]
		}
		return nil
	}
	if l, ok := v.([]string); ok {
		if index < len(l) {
			return l[index]
		}
		return nil
	}
	return v
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case float64:
		return x
	}
	return 0
}

func sortedSet(list []string) []string {
	set := map[string]bool{}
	var out []string
	for _, e := range list {
		if !set[e] {
			set[e] = true
			out = append(out, e)
		}
	}
	sort.Strings(out)
	return out
}

func makeList(v any) []string {
	switch x := v.(type) {
	case nil:
		return nil
	case []string:
		return x
	case []any:
		var out []string
		for _, e := range x {
			out = append(out, utils.PyStr(e))
		}
		return out
	default:
		return []string{utils.PyStr(v)}
	}
}

func makeStr(v any) string {
	parts := makeList(v)
	return strings.Join(parts, ", ")
}

// bomEntryKey returns the sorting/grouping key of a BOM entry.
func bomEntryKey(entry BOMEntry) []string {
	return []string{
		utils.CleanWhitespaceStr(makeStr(entry.Description)),
		utils.CleanWhitespaceStr(makeStr(entry.Unit)),
		utils.CleanWhitespaceStr(makeStr(entry.Pn)),
		utils.CleanWhitespaceStr(makeStr(entry.Manufacturer)),
		utils.CleanWhitespaceStr(makeStr(entry.Mpn)),
		utils.CleanWhitespaceStr(makeStr(entry.Supplier)),
		utils.CleanWhitespaceStr(makeStr(entry.Spn)),
	}
}

func compareKeys(a, b []string) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func keysEqual(a, b []string) bool {
	return compareKeys(a, b) == 0
}

// getBomIndex returns the id of a BOM entry matching target, panicking if absent.
func getBomIndex(bom []BOMEntry, target []string) int {
	for _, e := range bom {
		if keysEqual(bomEntryKey(e), target) {
			return e.ID
		}
	}
	panic(fmt.Errorf("%s", "Internal error: No BOM entry found matching: "+strings.Join(target, "|")))
}
