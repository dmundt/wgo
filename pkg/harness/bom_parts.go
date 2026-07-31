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
	"strings"

	"github.com/dmundt/wgo/pkg/graph"
	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

// partKey mirrors bom_entry_key({**asdict(part), "description": part.description}).
func partKey(part model.AdditionalComponent) []string {
	return []string{
		utils.CleanWhitespaceStr(makeStr(part.Description())),
		utils.CleanWhitespaceStr(makeStr(part.Unit)),
		utils.CleanWhitespaceStr(makeStr(strOrNil(part.Pn))),
		utils.CleanWhitespaceStr(makeStr(strOrNil(part.Manufacturer))),
		utils.CleanWhitespaceStr(makeStr(strOrNil(part.Mpn))),
		utils.CleanWhitespaceStr(makeStr(strOrNil(part.Supplier))),
		utils.CleanWhitespaceStr(makeStr(strOrNil(part.Spn))),
	}
}

// getAdditionalComponentTable returns diagram node row strings for additional components.
func getAdditionalComponentTable(h *Harness, component any) ([]any, error) {
	var rows []any
	parts := compAdditionalComponents(component)
	if len(parts) > 0 {
		rows = append(rows, []any{"Additional components"})
		for _, part := range parts {
			mult, err := compQtyMultiplier(component, part.QtyMultiplier)
			if err != nil {
				return nil, err
			}
			if !truthyVal(mult) {
				continue
			}
			qty := model.MulNum(part.Qty, mult)
			if h.Options.MiniBomMode {
				bom, err := h.Bom()
				if err != nil {
					return nil, err
				}
				id, err := getBomIndex(bom, partKey(part))
				if err != nil {
					return nil, err
				}
				rows = append(rows, componentTableEntry(
					fmt.Sprintf("#%d (%s)", id, rstripSpace(part.Type)),
					qty, part.Unit, part.Bgcolor, "", "", "", "", "",
				))
			} else {
				rows = append(rows, componentTableEntry(
					part.Description(), qty, part.Unit, part.Bgcolor,
					part.Pn, part.Manufacturer, part.Mpn, part.Supplier, part.Spn,
				))
			}
		}
	}
	return rows, nil
}

// getAdditionalComponentBom returns BOM entries for additional components.
func getAdditionalComponentBom(component any) ([]BOMEntry, error) {
	var entries []BOMEntry
	parts := compAdditionalComponents(component)
	name := compName(component)
	showName := compShowName(component)
	for _, part := range parts {
		mult, err := compQtyMultiplier(component, part.QtyMultiplier)
		if err != nil {
			return nil, err
		}
		if !truthyVal(mult) {
			continue
		}
		var designators any
		if showName {
			designators = name
		}
		entries = append(entries, BOMEntry{
			Description:  part.Description(),
			Qty:          model.MulNum(part.Qty, mult),
			Unit:         part.Unit,
			Designators:  designators,
			Pn:           strOrNil(part.Pn),
			Manufacturer: strOrNil(part.Manufacturer),
			Mpn:          strOrNil(part.Mpn),
			Supplier:     strOrNil(part.Supplier),
			Spn:          strOrNil(part.Spn),
		})
	}
	return entries, nil
}

func compAdditionalComponents(component any) []model.AdditionalComponent {
	switch c := component.(type) {
	case *model.Connector:
		return c.AdditionalComponents
	case *model.Cable:
		return c.AdditionalComponents
	}
	return nil
}

func compQtyMultiplier(component any, qm string) (any, error) {
	switch c := component.(type) {
	case *model.Connector:
		return c.GetQtyMultiplier(qm)
	case *model.Cable:
		return c.GetQtyMultiplier(qm)
	}
	return 1, nil
}

func compName(component any) string {
	switch c := component.(type) {
	case *model.Connector:
		return c.Name
	case *model.Cable:
		return c.Name
	}
	return ""
}

func compShowName(component any) bool {
	switch c := component.(type) {
	case *model.Connector:
		return c.ShowName
	case *model.Cable:
		return c.ShowName
	}
	return false
}

func componentTableEntry(typeStr string, qty any, unit string, bgcolor string,
	pn, manufacturer, mpn, supplier, spn string) string {
	partNumberList := []any{
		pnInfoString(PnHeader, "", pn),
		pnInfoString(MpnHeader, manufacturer, mpn),
		pnInfoString(SpnHeader, supplier, spn),
	}
	output := utils.PyStr(qty)
	if unit != "" {
		output += " " + unit
	}
	output += " x " + typeStr
	if anyTruthyAny(partNumberList) {
		output += "<br/>"
	}
	var pns []string
	for _, p := range partNumberList {
		if p != nil {
			pns = append(pns, utils.PyStr(p))
		}
	}
	if len(pns) > 0 {
		output += strings.Join(pns, ", ")
	}
	return `<table border="0" cellspacing="0" cellpadding="3" cellborder="1"` +
		graph.HTMLBgcolorAttr(bgcolor) + `><tr>
   <td align="left" balign="left">` + graph.HTMLLineBreaksStr(output) + `</td>
  </tr></table>`
}

func pnInfoString(header, name, number string) any {
	number = strings.TrimSpace(number)
	if name != "" || number != "" {
		prefix := name
		if prefix == "" {
			prefix = header
		}
		if number != "" {
			prefix += ": " + number
		}
		return prefix
	}
	return nil
}

func anyTruthyAny(vals []any) bool {
	for _, v := range vals {
		if v != nil {
			return true
		}
	}
	return false
}

// BomList mirrors wv_bom.bom_list.
func bomList(bom []BOMEntry) [][]any {
	keys := append([]string(nil), bomColumnsAlways...)
	for _, fieldname := range bomColumnsOptional {
		if anyEntryField(bom, fieldname) {
			keys = append(keys, fieldname)
		}
	}
	headings := map[string]string{"pn": PnHeader, "mpn": MpnHeader, "spn": SpnHeader}
	var header []any
	for _, k := range keys {
		if h, ok := headings[k]; ok {
			header = append(header, h)
		} else {
			header = append(header, capitalize(k))
		}
	}
	rows := [][]any{header}
	for _, e := range bom {
		var row []any
		for _, k := range keys {
			row = append(row, makeStr(entryField(e, k)))
		}
		rows = append(rows, row)
	}
	return rows
}

func anyEntryField(bom []BOMEntry, field string) bool {
	for _, e := range bom {
		if truthyVal(entryField(e, field)) {
			return true
		}
	}
	return false
}

func entryField(e BOMEntry, k string) any {
	switch k {
	case "id":
		return e.ID
	case "description":
		return e.Description
	case "qty":
		return e.Qty
	case "unit":
		return e.Unit
	case "designators":
		return e.Designators
	case "pn":
		return e.Pn
	case "manufacturer":
		return e.Manufacturer
	case "mpn":
		return e.Mpn
	case "supplier":
		return e.Supplier
	case "spn":
		return e.Spn
	}
	return nil
}
