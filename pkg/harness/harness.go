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
func (h *Harness) AddConnector(name string, attrs *utils.OrderedMap) error {
	c, err := model.NewConnector(name, attrs)
	if err != nil {
		return err
	}
	h.Connectors = append(h.Connectors, c)
	h.connectorIdx[name] = c
	return nil
}

// AddCable adds a cable instance.
func (h *Harness) AddCable(name string, attrs *utils.OrderedMap) error {
	c, err := model.NewCable(name, attrs)
	if err != nil {
		return err
	}
	h.Cables = append(h.Cables, c)
	h.cableIdx[name] = c
	return nil
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
func (h *Harness) Connect(fromName string, fromPin any, viaName string, viaWire any, toName string, toPin any) error {
	for _, pair := range [][2]any{{fromName, fromPin}, {toName, toPin}} {
		name := pair[0].(string)
		pin := pair[1]
		if name == "" {
			continue
		}
		if connector, ok := h.connectorIdx[name]; ok {
			if containsAny(connector.Pins, pin) && containsAny(connector.Pinlabels, pin) {
				if indexOfAny(connector.Pins, pin) != indexOfAny(connector.Pinlabels, pin) {
					return fmt.Errorf("%s:%s is defined both in pinlabels and pins, for different pins.",
						name, utils.PyStr(pin))
				}
			}
			if containsAny(connector.Pinlabels, pin) {
				if countAny(connector.Pinlabels, pin) > 1 {
					return fmt.Errorf("%s:%s is defined more than once.", name, utils.PyStr(pin))
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
				return fmt.Errorf("%s:%s not found.", name, utils.PyStr(pin))
			}
		}
	}

	if cable, ok := h.cableIdx[viaName]; ok {
		if ws, isStr := viaWire.(string); isStr {
			if indexOfStrings(cable.Colors, ws) >= 0 {
				if countStrings(cable.Colors, ws) > 1 {
					return fmt.Errorf("%s:%s is used for more than one wire.", viaName, ws)
				}
				viaWire = indexOfStrings(cable.Colors, ws) + 1
			} else if containsAny(cable.Wirelabels, viaWire) {
				if countAny(cable.Wirelabels, viaWire) > 1 {
					return fmt.Errorf("%s:%s is used for more than one wire.", viaName, ws)
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
	return nil
}

// Bom returns the generated BOM, cached.
func (h *Harness) Bom() ([]BOMEntry, error) {
	if h.bomCache == nil {
		bom, err := GenerateBom(h)
		if err != nil {
			return nil, err
		}
		h.bomCache = bom
	}
	return h.bomCache, nil
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
