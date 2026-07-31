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

// Package harness ports WireViz's Harness.py: the wiring graph model, DOT
// generation, and file output.
package harness

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dmundt/wgo/pkg/graph"
	"github.com/dmundt/wgo/pkg/html"
	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

// Harness mirrors the WireViz Harness dataclass.
type Harness struct {
	Metadata *utils.OrderedMap
	Options  *model.Options
	Tweak    *model.Tweak

	Connectors   []*model.Connector
	connectorIdx map[string]*model.Connector
	Cables       []*model.Cable
	cableIdx     map[string]*model.Cable
	Mates        []model.Mate

	bomCache           []BOMEntry
	AdditionalBomItems []BOMEntry
}

// NewHarness creates an empty Harness.
func NewHarness(metadata *utils.OrderedMap, options *model.Options, tweak *model.Tweak) *Harness {
	return &Harness{
		Metadata:     metadata,
		Options:      options,
		Tweak:        tweak,
		connectorIdx: map[string]*model.Connector{},
		cableIdx:     map[string]*model.Cable{},
	}
}

// AddConnector adds a connector instance.
func (h *Harness) AddConnector(name string, attrs *utils.OrderedMap) {
	c := model.NewConnector(name, attrs)
	h.Connectors = append(h.Connectors, c)
	h.connectorIdx[name] = c
}

// AddCable adds a cable instance.
func (h *Harness) AddCable(name string, attrs *utils.OrderedMap) {
	c := model.NewCable(name, attrs)
	h.Cables = append(h.Cables, c)
	h.cableIdx[name] = c
}

// ConnectorByName returns a connector instance by designator.
func (h *Harness) ConnectorByName(name string) (*model.Connector, bool) {
	c, ok := h.connectorIdx[name]
	return c, ok
}

// CableByName returns a cable instance by designator.
func (h *Harness) CableByName(name string) (*model.Cable, bool) {
	c, ok := h.cableIdx[name]
	return c, ok
}

// AddMatePin records a pin-to-pin mate.
func (h *Harness) AddMatePin(fromName string, fromPin any, toName string, toPin any, shape string) {
	h.Mates = append(h.Mates, model.Mate{Kind: model.MatePinKind, FromName: fromName, FromPin: fromPin, ToName: toName, ToPin: toPin, Shape: shape})
	h.connectorIdx[fromName].ActivatePin(fromPin, model.SideRight)
	h.connectorIdx[toName].ActivatePin(toPin, model.SideLeft)
}

// AddMateComponent records a component-level mate.
func (h *Harness) AddMateComponent(fromName, toName, shape string) {
	h.Mates = append(h.Mates, model.Mate{Kind: model.MateComponentKind, FromName: fromName, ToName: toName, Shape: shape})
}

// AddBomItem adds an additional BOM item.
func (h *Harness) AddBomItem(item BOMEntry) {
	h.AdditionalBomItems = append(h.AdditionalBomItems, item)
}

// Connect wires a cable between two connectors (either side may be empty).
func (h *Harness) Connect(fromName string, fromPin any, viaName string, viaWire any, toName string, toPin any) {
	for _, pair := range [][2]any{{fromName, fromPin}, {toName, toPin}} {
		name := pair[0].(string)
		pin := pair[1]
		if name == "" {
			continue
		}
		if connector, ok := h.connectorIdx[name]; ok {
			if containsAny(connector.Pins, pin) && containsAny(connector.Pinlabels, pin) {
				if indexOfAny(connector.Pins, pin) != indexOfAny(connector.Pinlabels, pin) {
					panic(fmt.Errorf("%s:%s is defined both in pinlabels and pins, for different pins.",
						name, utils.PyStr(pin)))
				}
			}
			if containsAny(connector.Pinlabels, pin) {
				if countAny(connector.Pinlabels, pin) > 1 {
					panic(fmt.Errorf("%s:%s is defined more than once.", name, utils.PyStr(pin)))
				}
				index := indexOfAny(connector.Pinlabels, pin)
				pin = connector.Pins[index]
				if name == fromName {
					fromPin = pin
				}
				if name == toName {
					toPin = pin
				}
			}
			if !containsAny(connector.Pins, pin) {
				panic(fmt.Errorf("%s:%s not found.", name, utils.PyStr(pin)))
			}
		}
	}

	if cable, ok := h.cableIdx[viaName]; ok {
		if ws, isStr := viaWire.(string); isStr {
			if indexOfStrings(cable.Colors, ws) >= 0 {
				if countStrings(cable.Colors, ws) > 1 {
					panic(fmt.Errorf("%s:%s is used for more than one wire.", viaName, ws))
				}
				viaWire = indexOfStrings(cable.Colors, ws) + 1
			} else if containsAny(cable.Wirelabels, viaWire) {
				if countAny(cable.Wirelabels, viaWire) > 1 {
					panic(fmt.Errorf("%s:%s is used for more than one wire.", viaName, ws))
				}
				viaWire = indexOfAny(cable.Wirelabels, viaWire) + 1
			}
		}
	}

	cable := h.cableIdx[viaName]
	cable.Connect(fromName, fromPin, viaWire, toName, toPin)
	if fromName != "" {
		if c, ok := h.connectorIdx[fromName]; ok {
			c.ActivatePin(fromPin, model.SideRight)
		}
	}
	if toName != "" {
		if c, ok := h.connectorIdx[toName]; ok {
			c.ActivatePin(toPin, model.SideLeft)
		}
	}
}

// Bom returns the generated BOM, cached.
func (h *Harness) Bom() []BOMEntry {
	if h.bomCache == nil {
		h.bomCache = GenerateBom(h)
	}
	return h.bomCache
}

// CreateGraph builds the DOT graph.
func (h *Harness) CreateGraph() *graph.Graph {
	dot := graph.NewGraph()
	dot.Body = append(dot.Body, fmt.Sprintf("// Graph generated by %s %s\n", model.AppName, model.Version))
	dot.Body = append(dot.Body, fmt.Sprintf("// %s\n", model.AppURL))
	dot.Attr("graph", map[string]string{
		"rankdir":  "LR",
		"ranksep":  "2",
		"bgcolor":  model.TranslateColor(h.Options.Bgcolor, "HEX"),
		"nodesep":  "0.33",
		"fontname": h.Options.Fontname,
	})
	dot.Attr("node", map[string]string{
		"shape":     "none",
		"width":     "0",
		"height":    "0",
		"margin":    "0",
		"style":     "filled",
		"fillcolor": model.TranslateColor(h.Options.BgcolorNode, "HEX"),
		"fontname":  h.Options.Fontname,
	})
	dot.Attr("edge", map[string]string{"style": "bold", "fontname": h.Options.Fontname})

	for _, connector := range h.Connectors {
		if !connector.PortsLeft && !connector.PortsRight {
			connector.PortsLeft = true
		}

		var rows []any
		var nameCell any
		if connector.ShowName {
			nameCell = graph.HTMLBgcolor(connector.BgcolorTitle, "") + utils.RemoveLinksStr(connector.Name)
		}
		rows = append(rows, []any{nameCell})
		pn := pnInfoString(PnHeader, "", utils.RemoveLinksStr(connector.Pn))
		mpn := graph.HTMLLineBreaks(pnInfoString(MpnHeader, connector.Manufacturer, connector.Mpn))
		spn := graph.HTMLLineBreaks(pnInfoString(SpnHeader, connector.Supplier, connector.Spn))
		rows = append(rows, []any{pn, mpn, spn})
		var pincountCell any
		if connector.ShowPincount {
			pincountCell = fmt.Sprintf("%d-pin", connector.Pincount)
		}
		var colorCell any
		if connector.Color != "" {
			colorCell = model.TranslateColor(connector.Color, h.Options.ColorMode)
		}
		rows = append(rows, []any{
			graph.HTMLLineBreaks(connector.Type),
			graph.HTMLLineBreaks(connector.Subtype),
			pincountCell,
			colorCell,
			graph.HTMLColorbar(connector.Color),
		})
		if connector.Style != "simple" {
			rows = append(rows, "<!-- connector table -->")
		} else {
			rows = append(rows, nil)
		}
		rows = append(rows, []any{graph.HTMLImage(connector.Image)})
		rows = append(rows, []any{graph.HTMLCaption(connector.Image)})
		rows = append(rows, getAdditionalComponentTable(h, connector)...)
		rows = append(rows, []any{graph.HTMLLineBreaks(connector.Notes)})
		htmlLines := graph.NestedHTMLTable(rows, graph.HTMLBgcolorAttr(connector.Bgcolor))

		if connector.Style != "simple" {
			var pinhtml []string
			pinhtml = append(pinhtml, `<table border="0" cellspacing="0" cellpadding="3" cellborder="1">`)
			n := max3(len(connector.Pins), len(connector.Pinlabels), len(connector.Pincolors))
			for pinindex := 0; pinindex < n; pinindex++ {
				var pinname, pinlabel any
				var pincolor string
				if pinindex < len(connector.Pins) {
					pinname = connector.Pins[pinindex]
				}
				if pinindex < len(connector.Pinlabels) {
					pinlabel = connector.Pinlabels[pinindex]
				}
				if pinindex < len(connector.Pincolors) {
					pincolor = connector.Pincolors[pinindex]
				}
				if connector.HideDisconnectedPins && !connector.VisiblePins[pinname] {
					continue
				}
				pinhtml = append(pinhtml, "   <tr>")
				if connector.PortsLeft {
					pinhtml = append(pinhtml, fmt.Sprintf(`    <td port="p%dl">%s</td>`, pinindex+1, utils.PyStr(pinname)))
				}
				if truthyVal(pinlabel) {
					pinhtml = append(pinhtml, fmt.Sprintf("    <td>%s</td>", utils.PyStr(pinlabel)))
				}
				if len(connector.Pincolors) > 0 {
					if _, ok := model.ColorHexLookup(pincolor); ok {
						pinhtml = append(pinhtml, fmt.Sprintf(`    <td sides="tbl">%s</td>`, model.TranslateColor(pincolor, h.Options.ColorMode)))
						pinhtml = append(pinhtml, `    <td sides="tbr">`)
						pinhtml = append(pinhtml, `     <table border="0" cellborder="1"><tr>`)
						pinhtml = append(pinhtml, fmt.Sprintf(`      <td bgcolor="%s" width="8" height="8" fixedsize="true"></td>`, model.TranslateColor(pincolor, "HEX")))
						pinhtml = append(pinhtml, `     </tr></table>`)
						pinhtml = append(pinhtml, `    </td>`)
					} else {
						pinhtml = append(pinhtml, `    <td colspan="2"></td>`)
					}
				}
				if connector.PortsRight {
					pinhtml = append(pinhtml, fmt.Sprintf(`    <td port="p%dr">%s</td>`, pinindex+1, utils.PyStr(pinname)))
				}
				pinhtml = append(pinhtml, "   </tr>")
			}
			pinhtml = append(pinhtml, "  </table>")
			if len(pinhtml) == 2 {
				pinhtml = []string{"<!-- all pins hidden -->"}
			}
			for i, row := range htmlLines {
				htmlLines[i] = strings.ReplaceAll(row, "<!-- connector table -->", strings.Join(pinhtml, "\n"))
			}
		}

		htmlJoined := strings.Join(htmlLines, "\n")
		dot.Node(connector.Name, fmt.Sprintf("<\n%s\n>", htmlJoined), map[string]string{
			"shape":     "box",
			"style":     "filled",
			"fillcolor": model.TranslateColor(h.Options.BgcolorConnector, "HEX"),
		})

		if len(connector.Loops) > 0 {
			dot.Attr("edge", map[string]string{"color": "#000000:#ffffff:#000000"})
			var loopSide, loopDir string
			if connector.PortsLeft {
				loopSide, loopDir = "l", "w"
			} else if connector.PortsRight {
				loopSide, loopDir = "r", "e"
			} else {
				panic(fmt.Errorf("No side for loops"))
			}
			for _, loop := range connector.Loops {
				dot.Edge(
					fmt.Sprintf("%s:p%v%v:%v", connector.Name, loop[0], loopSide, loopDir),
					fmt.Sprintf("%s:p%v%v:%v", connector.Name, loop[1], loopSide, loopDir),
					" ",
					nil,
				)
			}
		}
	}

	pad := false
	for _, cable := range h.Cables {
		for _, colorstr := range cable.Colors {
			if len(colorstr) > 2 {
				pad = true
			}
		}
	}

	for _, cable := range h.Cables {
		var rows []any
		var nameCell any
		if cable.ShowName {
			nameCell = graph.HTMLBgcolor(cable.BgcolorTitle, "") + utils.RemoveLinksStr(cable.Name)
		}
		rows = append(rows, []any{nameCell})
		var pnCell any
		if !isListValue(cable.Pn) {
			pnCell = pnInfoString(PnHeader, "", utils.RemoveLinksStr(strVal(cable.Pn)))
		}
		mfrName := strValIfString(cable.Manufacturer)
		mpnVal := strValIfString(cable.Mpn)
		supName := strValIfString(cable.Supplier)
		spnVal := strValIfString(cable.Spn)
		rows = append(rows, []any{
			pnCell,
			graph.HTMLLineBreaks(pnInfoString(MpnHeader, mfrName, mpnVal)),
			graph.HTMLLineBreaks(pnInfoString(SpnHeader, supName, spnVal)),
		})
		var wirecountCell any
		if cable.ShowWirecount {
			wirecountCell = fmt.Sprintf("%dx", cable.Wirecount)
		}
		var gaugeCell any
		if cable.Gauge != nil {
			gaugeCell = gaugeDisplay(cable)
		}
		var shieldCell any
		if shieldTruthy(cable.Shield) {
			shieldCell = "+ S"
		}
		var lengthCell any
		if numPositive(cable.Length) {
			lengthCell = fmt.Sprintf("%s %s", utils.PyStr(cable.Length), cable.LengthUnit)
		}
		var colorCell any
		if cable.Color != "" {
			colorCell = model.TranslateColor(cable.Color, h.Options.ColorMode)
		}
		rows = append(rows, []any{
			graph.HTMLLineBreaks(cable.Type),
			wirecountCell,
			gaugeCell,
			shieldCell,
			lengthCell,
			colorCell,
			graph.HTMLColorbar(cable.Color),
		})
		rows = append(rows, "<!-- wire table -->")
		rows = append(rows, []any{graph.HTMLImage(cable.Image)})
		rows = append(rows, []any{graph.HTMLCaption(cable.Image)})
		rows = append(rows, getAdditionalComponentTable(h, cable)...)
		rows = append(rows, []any{graph.HTMLLineBreaks(cable.Notes)})
		htmlLines := graph.NestedHTMLTable(rows, graph.HTMLBgcolorAttr(cable.Bgcolor))

		var wirehtml []string
		wirehtml = append(wirehtml, `<table border="0" cellspacing="0" cellborder="0">`)
		wirehtml = append(wirehtml, `   <tr><td>&nbsp;</td></tr>`)

		n := max(len(cable.Colors), len(cable.Wirelabels))
		for i := 1; i <= n; i++ {
			var connectionColor string
			if i-1 < len(cable.Colors) {
				connectionColor = cable.Colors[i-1]
			}
			var wirelabel any
			if i-1 < len(cable.Wirelabels) {
				wirelabel = cable.Wirelabels[i-1]
			}
			wirehtml = append(wirehtml, "   <tr>")
			wirehtml = append(wirehtml, fmt.Sprintf("    <td><!-- %d_in --></td>", i))
			wirehtml = append(wirehtml, "    <td>")
			var wireinfo []string
			if cable.ShowWirenumbers {
				wireinfo = append(wireinfo, strconv.Itoa(i))
			}
			colorstr := model.TranslateColor(connectionColor, h.Options.ColorMode)
			if colorstr != "" {
				wireinfo = append(wireinfo, colorstr)
			}
			if len(cable.Wirelabels) > 0 {
				if wirelabel != nil {
					wireinfo = append(wireinfo, utils.PyStr(wirelabel))
				} else {
					wireinfo = append(wireinfo, "")
				}
			}
			wirehtml = append(wirehtml, "     "+strings.Join(wireinfo, ":"))
			wirehtml = append(wirehtml, "    </td>")
			wirehtml = append(wirehtml, fmt.Sprintf("    <td><!-- %d_out --></td>", i))
			wirehtml = append(wirehtml, "   </tr>")

			bgcolors := model.GetColorHex(connectionColor, pad)
			bgcolors = append([]string{"#000000"}, bgcolors...)
			bgcolors = append(bgcolors, "#000000")
			wirehtml = append(wirehtml, "   <tr>")
			wirehtml = append(wirehtml, fmt.Sprintf(`    <td colspan="3" border="0" cellspacing="0" cellpadding="0" port="w%d" height="%d">`, i, 2*len(bgcolors)))
			wirehtml = append(wirehtml, `     <table cellspacing="0" cellborder="0" border="0">`)
			for j := len(bgcolors) - 1; j >= 0; j-- {
				bgc := bgcolors[j]
				if bgc == "" {
					bgc = model.ColorDefault
				}
				wirehtml = append(wirehtml, fmt.Sprintf(`      <tr><td colspan="3" cellpadding="0" height="2" bgcolor="%s" border="0"></td></tr>`, bgc))
			}
			wirehtml = append(wirehtml, "     </table>")
			wirehtml = append(wirehtml, "    </td>")
			wirehtml = append(wirehtml, "   </tr>")

			if cable.Category == "bundle" {
				var wireidentification []any
				if l, ok := cable.Pn.([]any); ok {
					wireidentification = append(wireidentification, pnInfoString(PnHeader, "", utils.RemoveLinksStr(listIdxStr(l, i-1))))
				}
				manufacturerInfo := pnInfoString(MpnHeader, listIdxOrEmpty(cable.Manufacturer, i-1), listIdxOrEmpty(cable.Mpn, i-1))
				supplierInfo := pnInfoString(SpnHeader, listIdxOrEmpty(cable.Supplier, i-1), listIdxOrEmpty(cable.Spn, i-1))
				if manufacturerInfo != nil {
					wireidentification = append(wireidentification, graph.HTMLLineBreaks(manufacturerInfo))
				}
				if supplierInfo != nil {
					wireidentification = append(wireidentification, graph.HTMLLineBreaks(supplierInfo))
				}
				if len(wireidentification) > 0 {
					wirehtml = append(wirehtml, `   <tr><td colspan="3">`)
					wirehtml = append(wirehtml, `    <table border="0" cellspacing="0" cellborder="0"><tr>`)
					for _, attrib := range wireidentification {
						wirehtml = append(wirehtml, fmt.Sprintf("     <td>%s</td>", utils.PyStr(attrib)))
					}
					wirehtml = append(wirehtml, "    </tr></table>")
					wirehtml = append(wirehtml, "   </td></tr>")
				}
			}
		}

		shieldColorHex := ""
		if shieldTruthy(cable.Shield) {
			wirehtml = append(wirehtml, "   <tr><td>&nbsp;</td></tr>")
			wirehtml = append(wirehtml, "   <tr>")
			wirehtml = append(wirehtml, "    <td><!-- s_in --></td>")
			wirehtml = append(wirehtml, "    <td>Shield</td>")
			wirehtml = append(wirehtml, "    <td><!-- s_out --></td>")
			wirehtml = append(wirehtml, "   </tr>")
			var attributes string
			if sh, ok := cable.Shield.(string); ok {
				shieldColorHex = model.GetColorHex(sh, false)[0]
				attributes = fmt.Sprintf(`height="6" bgcolor="%s" border="2" sides="tb"`, shieldColorHex)
			} else {
				attributes = `height="2" bgcolor="#000000" border="0"`
			}
			wirehtml = append(wirehtml, fmt.Sprintf(`   <tr><td colspan="3" cellpadding="0" %s port="ws"></td></tr>`, attributes))
		}
		wirehtml = append(wirehtml, "   <tr><td>&nbsp;</td></tr>")
		wirehtml = append(wirehtml, "  </table>")

		for i, row := range htmlLines {
			htmlLines[i] = strings.ReplaceAll(row, "<!-- wire table -->", strings.Join(wirehtml, "\n"))
		}

		for _, connection := range cable.Connections {
			if viaPort, ok := connection.ViaPort.(int); ok {
				hexes := model.GetColorHex(cable.Colors[viaPort-1], pad)
				color := "#000000:" + strings.Join(hexes, ":") + ":#000000"
				dot.Attr("edge", map[string]string{"color": color})
			} else {
				if _, isStr := cable.Shield.(string); isStr {
					dot.Attr("edge", map[string]string{"color": "#000000:" + shieldColorHex + ":#000000"})
				} else {
					dot.Attr("edge", map[string]string{"color": "#000000"})
				}
			}
			if connection.FromPin != nil {
				fromConnector := h.connectorIdx[connection.FromName]
				fromPinIndex := indexOfAny(fromConnector.Pins, connection.FromPin)
				fromPortStr := ""
				if fromConnector.Style != "simple" {
					fromPortStr = fmt.Sprintf(":p%dr", fromPinIndex+1)
				}
				codeLeft1 := connection.FromName + fromPortStr + ":e"
				codeLeft2 := fmt.Sprintf("%s:w%v:w", cable.Name, connection.ViaPort)
				dot.Edge(codeLeft1, codeLeft2, "", nil)
				fromString := ""
				if fromConnector.ShowName {
					fromInfo := []string{connection.FromName, utils.PyStr(connection.FromPin)}
					if len(fromConnector.Pinlabels) > 0 {
						pinlabel := utils.PyStr(fromConnector.Pinlabels[fromPinIndex])
						if pinlabel != "" {
							fromInfo = append(fromInfo, pinlabel)
						}
					}
					fromString = strings.Join(fromInfo, ":")
				}
				marker := fmt.Sprintf("<!-- %v_in -->", connection.ViaPort)
				for i, row := range htmlLines {
					htmlLines[i] = strings.ReplaceAll(row, marker, fromString)
				}
			}
			if connection.ToPin != nil {
				toConnector := h.connectorIdx[connection.ToName]
				toPinIndex := indexOfAny(toConnector.Pins, connection.ToPin)
				toPortStr := ""
				if toConnector.Style != "simple" {
					toPortStr = fmt.Sprintf(":p%dl", toPinIndex+1)
				}
				codeRight1 := fmt.Sprintf("%s:w%v:e", cable.Name, connection.ViaPort)
				codeRight2 := connection.ToName + toPortStr + ":w"
				dot.Edge(codeRight1, codeRight2, "", nil)
				toString := ""
				if toConnector.ShowName {
					toInfo := []string{connection.ToName, utils.PyStr(connection.ToPin)}
					if len(toConnector.Pinlabels) > 0 {
						pinlabel := utils.PyStr(toConnector.Pinlabels[toPinIndex])
						if pinlabel != "" {
							toInfo = append(toInfo, pinlabel)
						}
					}
					toString = strings.Join(toInfo, ":")
				}
				marker := fmt.Sprintf("<!-- %v_out -->", connection.ViaPort)
				for i, row := range htmlLines {
					htmlLines[i] = strings.ReplaceAll(row, marker, toString)
				}
			}
		}

		style := "filled,dashed"
		bgcolor := h.Options.BgcolorBundle
		if cable.Category != "bundle" {
			style = "filled"
			bgcolor = h.Options.BgcolorCable
		}
		htmlJoined := strings.Join(htmlLines, "\n")
		dot.Node(cable.Name, fmt.Sprintf("<\n%s\n>", htmlJoined), map[string]string{
			"shape":     "box",
			"style":     style,
			"fillcolor": model.TranslateColor(bgcolor, "HEX"),
		})
	}

	for _, mate := range h.Mates {
		var dir string
		if strings.HasSuffix(mate.Shape, ">") {
			if strings.HasPrefix(mate.Shape, "<") {
				dir = "both"
			} else {
				dir = "forward"
			}
		} else {
			if strings.HasPrefix(mate.Shape, "<") {
				dir = "back"
			} else {
				dir = "none"
			}
		}
		color := "#000000"
		if mate.Kind == model.MateComponentKind {
			color = "#000000:#000000"
		}
		fromConnector := h.connectorIdx[mate.FromName]
		toConnector := h.connectorIdx[mate.ToName]
		fromPortStr := ""
		if mate.Kind == model.MatePinKind && fromConnector.Style != "simple" {
			fromPinIndex := indexOfAny(fromConnector.Pins, mate.FromPin)
			fromPortStr = fmt.Sprintf(":p%dr", fromPinIndex+1)
		}
		toPortStr := ""
		if mate.Kind == model.MatePinKind && toConnector.Style != "simple" {
			toPinIndex := indexOfAny(toConnector.Pins, mate.ToPin)
			toPortStr = fmt.Sprintf(":p%dl", toPinIndex+1)
		}
		codeFrom := mate.FromName + fromPortStr + ":e"
		codeTo := mate.ToName + toPortStr + ":w"
		dot.Attr("edge", map[string]string{"color": color, "style": "dashed", "dir": dir})
		dot.Edge(codeFrom, codeTo, "", nil)
	}

	h.applyTweak(dot)

	return dot
}

func (h *Harness) applyTweak(dot *graph.Graph) {
	if h.Tweak == nil || h.Tweak.Override == nil {
		goto appendBlock
	}
	for i, entry := range dot.Body {
		keyword := bodyKeyword(entry)
		if keyword == "" {
			continue
		}
		d, ok := h.Tweak.Override.Get(keyword)
		if !ok {
			continue
		}
		om, ok := d.(*utils.OrderedMap)
		if !ok {
			panic(fmt.Errorf("Unexpected value type of tweak.override.%s: Expected dict, got %T", keyword, d))
		}
		for _, av := range om.Items() {
			a := av.Key
			value := av.Value
			if value == nil {
				entry, n := removeAttr(entry, a)
				if n < 1 {
					fmt.Printf("Harness.create_graph() warning: %s not found in %s!\n", a, keyword)
				} else if n > 1 {
					fmt.Printf("Harness.create_graph() warning: %s removed %d times in %s!\n", a, n, keyword)
				}
				dot.Body[i] = entry
				continue
			}
			vs := utils.PyStr(value)
			if len(vs) == 0 || strings.Contains(vs, " ") {
				vs = strings.ReplaceAll(vs, `"`, `\"`)
				vs = `"` + vs + `"`
			}
			entry, n := overrideAttr(entry, a, vs)
			if n < 1 {
				entry = appendAttr(entry, a, vs)
			} else if n > 1 {
				fmt.Printf("Harness.create_graph() warning: %s overridden %d times in %s!\n", a, n, keyword)
			}
			dot.Body[i] = entry
		}
	}

appendBlock:
	if h.Tweak != nil && h.Tweak.Append != nil {
		switch app := h.Tweak.Append.(type) {
		case []any:
			for _, e := range app {
				dot.Body = append(dot.Body, utils.PyStr(e))
			}
		default:
			dot.Body = append(dot.Body, utils.PyStr(app))
		}
	}
}

var bodyKeywordRe = regexp.MustCompile(`(?s)^\t*(?:"([^"]+)"|([^ "]+)) \[.*\]$`)

func bodyKeyword(entry string) string {
	m := bodyKeywordRe.FindStringSubmatch(entry)
	if m == nil {
		return ""
	}
	if m[1] != "" {
		return m[1]
	}
	return m[2]
}

func removeAttr(entry, attr string) (string, int) {
	// Build a pattern equivalent to Python's
	// ( +)?{attr}=("[^"]*"|[^] ]*)(?(1)| *)
	pattern := regexp.MustCompile(`( +)` + regexp.QuoteMeta(attr) + `=("[^"]*"|[^] ]*)|` +
		regexp.QuoteMeta(attr) + `=("[^"]*"|[^] ]*)( +)?`)
	count := 0
	out := pattern.ReplaceAllStringFunc(entry, func(m string) string {
		count++
		return ""
	})
	return out, count
}

func overrideAttr(entry, attr, value string) (string, int) {
	pattern := regexp.MustCompile(regexp.QuoteMeta(attr) + `=("[^"]*"|[^] ]*)`)
	count := 0
	out := pattern.ReplaceAllStringFunc(entry, func(m string) string {
		count++
		return attr + "=" + value
	})
	return out, count
}

func appendAttr(entry, attr, value string) string {
	idx := len(entry)
	if strings.HasSuffix(entry, "\n") {
		idx--
	}
	if idx > 0 && entry[idx-1] == ']' {
		return entry[:idx-1] + " " + attr + "=" + value + "]" + entry[idx:]
	}
	return entry
}

// Output writes the requested output files.
func (h *Harness) Output(filename string, fmts []string) error {
	if dir := filepath.Dir(filename); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	g := h.CreateGraph()
	source := g.Source()

	for _, f := range fmts {
		switch f {
		case "png", "svg", "html":
			if f == "html" {
				f = "svg"
			}
			outFile := filename
			if f == "svg" {
				outFile = filename + ".tmp"
			}
			if err := renderDOT(source, f, outFile); err != nil {
				return err
			}
		}
	}

	if contains(fmts, "svg") || contains(fmts, "html") {
		if err := html.EmbedSVGImagesFile(filename + ".tmp.svg"); err != nil {
			return err
		}
	}

	if contains(fmts, "gv") {
		if err := utils.FileWriteText(filename+".gv", source); err != nil {
			return err
		}
	}

	bomlist := bomList(h.Bom())
	if contains(fmts, "tsv") {
		if err := utils.FileWriteText(filename+".bom.tsv", utils.Tuplelist2tsv(bomlist, nil)); err != nil {
			return err
		}
	}
	if contains(fmts, "csv") {
		fmt.Println("CSV output is not yet supported")
	}
	if contains(fmts, "html") {
		if err := html.GenerateHTMLOutput(filename, bomlist, h.Metadata, h.Options); err != nil {
			return err
		}
	}
	if contains(fmts, "pdf") {
		fmt.Println("PDF output is not yet supported")
	}

	if contains(fmts, "html") && !contains(fmts, "svg") {
		_ = removeFile(filename + ".tmp.svg")
	} else if contains(fmts, "svg") {
		if err := renameFile(filename+".tmp.svg", filename+".svg"); err != nil {
			return err
		}
	}
	return nil
}

func renderDOT(source, format, outFile string) error {
	dotPath, err := exec.LookPath("dot")
	if err != nil {
		return fmt.Errorf("Graphviz (dot) not found in PATH")
	}
	cmd := exec.Command(dotPath, "-T"+format, "-o", outFile+"."+format)
	cmd.Stdin = strings.NewReader(source)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dot failed: %v\n%s", err, string(out))
	}
	return nil
}

func removeFile(path string) error {
	return utils.RemoveFile(path)
}

func renameFile(old, new string) error {
	return utils.RenameFile(old, new)
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

func max3(a, b, c int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
