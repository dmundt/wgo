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
	"sort"
	"strings"

	"github.com/dmundt/wgo/pkg/graph"
	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

//go:embed templates/*.html
var templatesFS embed.FS

var svgHeaderRe = regexp.MustCompile(`^<\?xml [^?>]*\?>[^<]*<!DOCTYPE [^>]*>`)

// GenerateHTMLOutput writes the HTML output file.
func GenerateHTMLOutput(filename string, bomList [][]any, metadata *utils.OrderedMap, options *model.Options) error {
	var templateText string
	var err error
	templatename := ""
	if metadata != nil {
		if tmpl, ok := metadata.Get("template"); ok {
			if tm, ok := tmpl.(*utils.OrderedMap); ok {
				if v, ok := tm.Get("name"); ok && v != nil {
					templatename = utils.PyStr(v)
				}
			}
		}
	}
	if templatename != "" {
		// check output dir first, then bundled templates
		resolved, rerr := utils.SmartFileResolve(templatename+".html", []string{filepath.Dir(filename)})
		if rerr == nil {
			templateText, err = utils.FileReadText(resolved)
		} else {
			var b []byte
			b, err = templatesFS.ReadFile("templates/" + templatename + ".html")
			templateText = string(b)
		}
	} else {
		var b []byte
		b, err = templatesFS.ReadFile("templates/simple.html")
		templateText = string(b)
	}
	if err != nil {
		return err
	}
	templateText = strings.ReplaceAll(templateText, "\r\n", "\n")

	svgdata := func() string {
		text, err := utils.FileReadText(filename + ".tmp.svg")
		if err != nil {
			return ""
		}
		text = strings.ReplaceAll(text, "\r\n", "\n")
		return svgHeaderRe.ReplaceAllString(text, "<!-- XML and DOCTYPE declarations from SVG file removed -->")
	}

	bom := utils.Flatten2d(bomList)

	bomHeaderHTML := "  <tr>\n"
	for _, item := range bom[0] {
		thClass := fmt.Sprintf("bom_col_%s", strings.ToLower(item))
		bomHeaderHTML += fmt.Sprintf(`    <th class="%s">%s</th>`+"\n", thClass, item)
	}
	bomHeaderHTML += "  </tr>\n"

	var bomContents []string
	for _, row := range bom[1:] {
		rowHTML := "  <tr>\n"
		for i, item := range row {
			tdClass := fmt.Sprintf("bom_col_%s", strings.ToLower(bom[0][i]))
			rowHTML += fmt.Sprintf(`    <td class="%s">%s</td>`+"\n", tdClass, item)
		}
		rowHTML += "  </tr>\n"
		bomContents = append(bomContents, rowHTML)
	}

	bomHTML := `<table class="bom">` + "\n" + bomHeaderHTML + strings.Join(bomContents, "") + "</table>\n"
	var reversedContents []string
	for i := len(bomContents) - 1; i >= 0; i-- {
		reversedContents = append(reversedContents, bomContents[i])
	}
	bomHTMLReversed := `<table class="bom">` + "\n" + strings.Join(reversedContents, "") + bomHeaderHTML + "</table>\n"

	replacements := map[string]string{
		"<!-- %generator% -->":          fmt.Sprintf("%s %s - %s", model.AppName, model.Version, model.AppURL),
		"<!-- %fontname% -->":           options.Fontname,
		"<!-- %bgcolor% -->":            model.TranslateColor(options.Bgcolor, "hex"),
		"<!-- %filename% -->":           filename,
		"<!-- %filename_stem% -->":      stem(filepath.Base(filename)),
		"<!-- %bom% -->":                bomHTML,
		"<!-- %bom_reversed% -->":       bomHTMLReversed,
		"<!-- %sheet_current% -->":      "1",
		"<!-- %sheet_total% -->":        "1",
		"<!-- %template_sheetsize% -->": templateSheetsize(metadata),
	}

	if strings.Contains(templateText, "<!-- %diagram% -->") {
		replacements["<!-- %diagram% -->"] = svgdata()
	}
	if strings.Contains(templateText, "<!-- %diagram_png_b64% -->") {
		replacements["<!-- %diagram_png_b64% -->"] = DataURIBase64(filename+".png", "image")
	}

	if metadata != nil {
		for _, item := range metadata.Items() {
			switch contents := item.Value.(type) {
			case string, int, float64:
				replacements[fmt.Sprintf("<!-- %%%s%% -->", item.Key)] = graph.HTMLLineBreaksStr(utils.PyStr(contents))
			case *utils.OrderedMap:
				for index, entry := range contents.Items() {
					if em, ok := entry.Value.(*utils.OrderedMap); ok {
						replacements[fmt.Sprintf("<!-- %%%s_%d%% -->", item.Key, index+1)] = utils.PyStr(entry.Key)
						for _, ekv := range em.Items() {
							replacements[fmt.Sprintf("<!-- %%%s_%d_%s%% -->", item.Key, index+1, ekv.Key)] =
								graph.HTMLLineBreaksStr(utils.PyStr(ekv.Value))
						}
					}
				}
			}
		}
	}

	keys := make([]string, 0, len(replacements))
	for k := range replacements {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	var alternatives []string
	for _, k := range keys {
		alternatives = append(alternatives, regexp.QuoteMeta(k))
	}
	pattern := regexp.MustCompile(strings.Join(alternatives, "|"))
	templateText = pattern.ReplaceAllStringFunc(templateText, func(m string) string {
		return replacements[m]
	})

	return utils.FileWriteText(filename+".html", templateText)
}

func templateSheetsize(metadata *utils.OrderedMap) string {
	if metadata == nil {
		return ""
	}
	if tmpl, ok := metadata.Get("template"); ok {
		if tm, ok := tmpl.(*utils.OrderedMap); ok {
			if v, ok := tm.Get("sheetsize"); ok && v != nil {
				return utils.PyStr(v)
			}
		}
	}
	return ""
}

func stem(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return strings.TrimSuffix(name, ext)
}
