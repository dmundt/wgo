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

package html

import (
	"embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

//go:embed templates/*.html
var templatesFS embed.FS

var svgHeaderRe = regexp.MustCompile(`^<\?xml [^?>]*\?>[^<]*<!DOCTYPE [^>]*>`)

// GenerateHTMLOutput writes the HTML output file.
func GenerateHTMLOutput(filename string, bomList [][]any, metadata *utils.OrderedMap, options *model.Options) error {
	text, err := loadTemplate(filename, metadata)
	if err != nil {
		return err
	}
	tmpl := newHTMLTemplate(text)

	values := map[string]string{
		"<!-- %generator% -->":          fmt.Sprintf("%s %s - %s", model.AppName, model.Version, model.AppURL),
		"<!-- %fontname% -->":           options.Fontname,
		"<!-- %bgcolor% -->":            model.TranslateColor(options.Bgcolor, "hex"),
		"<!-- %filename% -->":           filename,
		"<!-- %filename_stem% -->":      stem(filepath.Base(filename)),
		"<!-- %sheet_current% -->":      "1",
		"<!-- %sheet_total% -->":        "1",
		"<!-- %template_sheetsize% -->": templateSheetsize(metadata),
	}

	bomHTML, bomHTMLReversed := renderBOM(bomList)
	values["<!-- %bom% -->"] = bomHTML
	values["<!-- %bom_reversed% -->"] = bomHTMLReversed

	// Only read the auxiliary outputs when the template uses them, matching
	// the reference implementation.
	if tmpl.Has("<!-- %diagram% -->") {
		values["<!-- %diagram% -->"] = svgData(filename + ".tmp.svg")
	}
	if tmpl.Has("<!-- %diagram_png_b64% -->") {
		values["<!-- %diagram_png_b64% -->"] = DataURIBase64(filename+".png", "image")
	}

	addMetadataPlaceholders(values, metadata)

	return utils.FileWriteText(filename+".html", tmpl.Render(values))
}

func stem(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return strings.TrimSuffix(name, ext)
}
