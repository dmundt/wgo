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

// Package model ports WireViz's DataClasses.py and wv_colors.py: the data
// structures (Options, Tweak, Image, Connector, Cable, ...) and the color
// code tables and translation functions.
package model

import (
	"fmt"
	"strconv"
	"strings"
)

// ColorCodes maps scheme name to ordered color name list.
var ColorCodes = map[string][]string{
	"DIN": {
		"WH", "BN", "GN", "YE", "GY", "PK", "BU", "RD", "BK", "VT", "GYPK", "RDBU",
		"WHGN", "BNGN", "WHYE", "YEBN", "WHGY", "GYBN", "WHPK", "PKBN", "WHBU", "BNBU",
		"WHRD", "BNRD", "WHBK", "BNBK", "GYGN", "YEGY", "PKGN", "YEPK", "GNBU", "YEBU",
		"GNRD", "YERD", "GNBK", "YEBK", "GYBU", "PKBU", "GYRD", "PKRD", "GYBK", "PKBK",
		"BUBK", "RDBK", "WHBNBK", "YEGNBK", "GYPKBK", "RDBUBK", "WHGNBK", "BNGNBK",
		"WHYEBK", "YEBNBK", "WHGYBK", "GYBNBK", "WHPKBK", "PKBNBK", "WHBUBK",
		"BNBUBK", "WHRDBK", "BNRDBK",
	},
	"IEC":    {"BN", "RD", "OG", "YE", "GN", "BU", "VT", "GY", "WH", "BK"},
	"BW":     {"BK", "WH"},
	"TEL":    {"BUWH", "WHBU", "OGWH", "WHOG", "GNWH", "WHGN", "BNWH", "WHBN", "SLWH", "WHSL", "BURD", "RDBU", "OGRD", "RDOG", "GNRD", "RDGN", "BNRD", "RDBN", "SLRD", "RDSL", "BUBK", "BKBU", "OGBK", "BKOG", "GNBK", "BKGN", "BNBK", "BKBN", "SLBK", "BKSL", "BUYE", "YEBU", "OGYE", "YEOG", "GNYE", "YEGN", "BNYE", "YEBN", "SLYE", "YESL", "BUVT", "VTBU", "OGVT", "VTOG", "GNVT", "VTGN", "BNVT", "VTBN", "SLVT", "VTSL"},
	"TELALT": {"WHBU", "BU", "WHOG", "OG", "WHGN", "GN", "WHBN", "BN", "WHSL", "SL", "RDBU", "BURD", "RDOG", "OGRD", "RDGN", "GNRD", "RDBN", "BNRD", "RDSL", "SLRD", "BKBU", "BUBK", "BKOG", "OGBK", "BKGN", "GNBK", "BKBN", "BNBK", "BKSL", "SLBK", "YEBU", "BUYE", "YEOG", "OGYE", "YEGN", "GNYE", "YEBN", "BNYE", "YESL", "SLYE", "VTBU", "BUVT", "VTOG", "OGVT", "VTGN", "GNVT", "VTBN", "BNVT", "VTSL", "SLVT"},
	"T568A":  {"WHGN", "GN", "WHOG", "BU", "WHBU", "OG", "WHBN", "BN"},
	"T568B":  {"WHOG", "OG", "WHGN", "BU", "WHBU", "GN", "WHBN", "BN"},
}

var colorHex = map[string]string{
	"BK": "#000000",
	"WH": "#ffffff",
	"GY": "#999999",
	"PK": "#ff66cc",
	"RD": "#ff0000",
	"OG": "#ff8000",
	"YE": "#ffff00",
	"OL": "#708000",
	"GN": "#00ff00",
	"TQ": "#00ffff",
	"LB": "#a0dfff",
	"BU": "#0066ff",
	"VT": "#8000ff",
	"BN": "#895956",
	"BG": "#ceb673",
	"IV": "#f5f0d0",
	"SL": "#708090",
	"CU": "#d6775e",
	"SN": "#aaaaaa",
	"SR": "#84878c",
	"GD": "#ffcf80",
}

var colorFull = map[string]string{
	"BK": "black",
	"WH": "white",
	"GY": "grey",
	"PK": "pink",
	"RD": "red",
	"OG": "orange",
	"YE": "yellow",
	"OL": "olive green",
	"GN": "green",
	"TQ": "turquoise",
	"LB": "light blue",
	"BU": "blue",
	"VT": "violet",
	"BN": "brown",
	"BG": "beige",
	"IV": "ivory",
	"SL": "slate",
	"CU": "copper",
	"SN": "tin",
	"SR": "silver",
	"GD": "gold",
}

var colorGer = map[string]string{
	"BK": "sw",
	"WH": "ws",
	"GY": "gr",
	"PK": "rs",
	"RD": "rt",
	"OG": "or",
	"YE": "ge",
	"OL": "ol",
	"GN": "gn",
	"TQ": "tk",
	"LB": "hb",
	"BU": "bl",
	"VT": "vi",
	"BN": "br",
	"BG": "bg",
	"IV": "eb",
	"SL": "si",
	"CU": "ku",
	"SN": "vz",
	"SR": "ag",
	"GD": "au",
}

// ColorDefault is the fallback hex color.
const ColorDefault = "#ffffff"

var hexDigits = map[byte]bool{
	'0': true, '1': true, '2': true, '3': true, '4': true, '5': true, '6': true, '7': true,
	'8': true, '9': true, 'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true,
	'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true,
}

// GetColorHex returns a list of hex colors from color names or :-separated hex colors.
func GetColorHex(input string, pad bool) []string {
	var output []string
	if input == "" {
		return []string{ColorDefault}
	} else if strings.HasPrefix(input, "#") {
		parts := strings.Split(input, ":")
		for i, c := range parts {
			if !strings.HasPrefix(c, "#") || !allHex(c[1:]) {
				if c != input {
					c += " in input: " + input
				}
				fmt.Println("Invalid hex color: " + c)
				parts[i] = ColorDefault
			}
		}
		output = parts
	} else {
		for i := 0; i < len(input); i += 2 {
			c := input[i : i+2]
			v, ok := colorHex[c]
			if !ok {
				if c != input {
					c += " in input: " + input
				}
				fmt.Println("Unknown color name: " + c)
				v = ColorDefault
			}
			output = append(output, v)
		}
	}
	if len(output) == 2 {
		output = append(output, output[0])
	} else if pad && len(output) == 1 {
		output = append(output, output[0], output[0])
	}
	return output
}

func allHex(s string) bool {
	for i := 0; i < len(s); i++ {
		if !hexDigits[s[i]] {
			return false
		}
	}
	return true
}

func getColorTranslation(translate map[string]string, input string) []string {
	fromHex := func(hexInput string) string {
		for color, hx := range colorHex {
			if hx == hexInput {
				return translate[color]
			}
		}
		var parts []string
		for i := 1; i <= 5; i += 2 {
			v, _ := strconv.ParseInt(hexInput[i:i+2], 16, 64)
			parts = append(parts, strconv.Itoa(int(v)))
		}
		return "(" + strings.Join(parts, ",") + ")"
	}
	if strings.HasPrefix(input, "#") {
		var out []string
		for _, h := range strings.Split(strings.ToLower(input), ":") {
			out = append(out, fromHex(h))
		}
		return out
	}
	var out []string
	for i := 0; i < len(input); i += 2 {
		c := input[i : i+2]
		v, ok := translate[c]
		if !ok {
			v = "??"
		}
		out = append(out, v)
	}
	return out
}

// TranslateColor renders a color in the given mode (full/hex/ger/short, case-preserving).
func TranslateColor(input string, colorMode string) string {
	if input == "" {
		return ""
	}
	upper := isUpper(colorMode)
	if !isUpper(colorMode) && !isLower(colorMode) {
		panic("Unknown color mode capitalization")
	}
	lower := strings.ToLower(colorMode)
	var output string
	switch lower {
	case "full":
		output = strings.Join(getColorTranslation(colorFull, input), "/")
	case "hex":
		output = strings.Join(GetColorHex(input, false), ":")
	case "ger":
		output = strings.Join(getColorTranslation(colorGer, input), "")
	case "short":
		output = input
	default:
		panic("Unknown color mode")
	}
	if upper {
		return strings.ToUpper(output)
	}
	return strings.ToLower(output)
}

func isUpper(s string) bool {
	hasCased := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			hasCased = true
		}
	}
	return hasCased
}

func isLower(s string) bool {
	hasCased := false
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return false
		}
		if r >= 'a' && r <= 'z' {
			hasCased = true
		}
	}
	return hasCased
}
