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
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dmundt/wgo/pkg/graph"
	"github.com/dmundt/wgo/pkg/utils"
)

// htmlTemplate renders WireViz-compatible HTML templates. Placeholders are
// HTML comments of the form <!-- %name% -->; this syntax is part of the
// WireViz contract (users write their own templates with it) and must not
// change.
type htmlTemplate struct {
	text string
}

// newHTMLTemplate normalizes the template text to LF line endings.
func newHTMLTemplate(text string) *htmlTemplate {
	return &htmlTemplate{text: strings.ReplaceAll(text, "\r\n", "\n")}
}

// Has reports whether the template contains the given placeholder marker.
func (t *htmlTemplate) Has(marker string) bool {
	return strings.Contains(t.text, marker)
}

// Render substitutes placeholders from values, keyed by full marker strings
// (e.g. "<!-- %generator% -->"). Longer markers are matched first so a more
// specific placeholder wins over a shorter one that is its substring.
func (t *htmlTemplate) Render(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	var alternatives []string
	for _, k := range keys {
		alternatives = append(alternatives, regexp.QuoteMeta(k))
	}
	pattern := regexp.MustCompile(strings.Join(alternatives, "|"))
	return pattern.ReplaceAllStringFunc(t.text, func(m string) string {
		return values[m]
	})
}

// loadTemplate reads the user-selected template (from the output dir first,
// then the bundled templates) or falls back to the bundled simple.html.
func loadTemplate(filename string, metadata *utils.OrderedMap) (string, error) {
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
		resolved, rerr := utils.SmartFileResolve(templatename+".html", []string{filepath.Dir(filename)})
		if rerr == nil {
			return utils.FileReadText(resolved)
		}
		b, err := templatesFS.ReadFile("templates/" + templatename + ".html")
		return string(b), err
	}
	b, err := templatesFS.ReadFile("templates/simple.html")
	return string(b), err
}

// renderBOM builds the forward and reversed BOM tables.
func renderBOM(bomList [][]any) (string, string) {
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
	return bomHTML, bomHTMLReversed
}

// svgData returns the embedded SVG body with the XML/DOCTYPE header stripped.
func svgData(filename string) string {
	text, err := utils.FileReadText(filename)
	if err != nil {
		return ""
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return svgHeaderRe.ReplaceAllString(text, "<!-- XML and DOCTYPE declarations from SVG file removed -->")
}

// addMetadataPlaceholders registers <!-- %key% --> and <!-- %key_N_field% -->
// placeholders derived from the harness metadata.
func addMetadataPlaceholders(values map[string]string, metadata *utils.OrderedMap) {
	if metadata == nil {
		return
	}
	for _, item := range metadata.Items() {
		switch contents := item.Value.(type) {
		case string, int, float64:
			values[fmt.Sprintf("<!-- %%%s%% -->", item.Key)] = graph.HTMLLineBreaksStr(utils.PyStr(contents))
		case *utils.OrderedMap:
			for index, entry := range contents.Items() {
				if em, ok := entry.Value.(*utils.OrderedMap); ok {
					values[fmt.Sprintf("<!-- %%%s_%d%% -->", item.Key, index+1)] = utils.PyStr(entry.Key)
					for _, ekv := range em.Items() {
						values[fmt.Sprintf("<!-- %%%s_%d_%s%% -->", item.Key, index+1, ekv.Key)] =
							graph.HTMLLineBreaksStr(utils.PyStr(ekv.Value))
					}
				}
			}
		}
	}
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
