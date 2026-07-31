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
	"github.com/dmundt/wgo/pkg/utils"
)

// Image mirrors the WireViz Image dataclass.
type Image struct {
	Src       string
	Scale     string
	Width     any // int or float64
	Height    any
	Fixedsize bool
	Bgcolor   string
	Caption   string
}

// NewImage builds an Image from a YAML dict, running the post-init logic.
func NewImage(attrs *utils.OrderedMap) *Image {
	img := &Image{}
	if attrs == nil {
		return img
	}
	if v, ok := attrs.Get("src"); ok && v != nil {
		img.Src = utils.PyStr(v)
	}
	if v, ok := attrs.Get("scale"); ok && v != nil {
		img.Scale = utils.PyStr(v)
	}
	if v, ok := attrs.Get("width"); ok && v != nil {
		img.Width = v
	}
	if v, ok := attrs.Get("height"); ok && v != nil {
		img.Height = v
	}
	if v, ok := attrs.Get("bgcolor"); ok && v != nil {
		img.Bgcolor = utils.PyStr(v)
	}
	if v, ok := attrs.Get("caption"); ok && v != nil {
		img.Caption = utils.PyStr(v)
	}
	var fixedsize *bool
	if v, ok := attrs.Get("fixedsize"); ok && v != nil {
		if b, ok := v.(bool); ok {
			fixedsize = &b
		}
	}
	img.postInit(fixedsize)
	return img
}

func (img *Image) postInit(fixedsize *bool) {
	wTruthy := img.Width != nil && numTruthy(img.Width)
	hTruthy := img.Height != nil && numTruthy(img.Height)

	if fixedsize == nil {
		img.Fixedsize = (wTruthy || hTruthy) && img.Scale == ""
	} else {
		img.Fixedsize = *fixedsize
	}

	if img.Scale == "" {
		switch {
		case !wTruthy && !hTruthy:
			img.Scale = "false"
		case wTruthy && hTruthy:
			img.Scale = "both"
		default:
			img.Scale = "true"
		}
	}

	if img.Fixedsize {
		if hTruthy {
			if !wTruthy {
				img.Width = numToFloat(img.Height) * utils.AspectRatio(img.Src)
			}
		} else {
			if wTruthy {
				img.Height = numToFloat(img.Width) / utils.AspectRatio(img.Src)
			}
		}
	}
}

func numTruthy(v any) bool {
	switch x := v.(type) {
	case int:
		return x != 0
	case float64:
		return x != 0
	default:
		return true
	}
}

func numToFloat(v any) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case float64:
		return x
	default:
		return 0
	}
}
