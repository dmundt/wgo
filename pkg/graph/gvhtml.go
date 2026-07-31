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

package graph

import (
	"fmt"
	"strings"

	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

// NestedHTMLTable mirrors wv_gv_html.nested_html_table, returning a list of lines.
func NestedHTMLTable(rows []any, tableAttrs string) []string {
	var html []string
	html = append(html, `<table border="0" cellspacing="0" cellpadding="0"`+tableAttrs+`>`)
	numRows := 0
	for _, row := range rows {
		if rowList, ok := row.([]any); ok {
			if len(rowList) > 0 && anyTruthy(rowList) {
				html = append(html, " <tr><td>")
				html = append(html, `  <table border="0" cellspacing="0" cellpadding="3" cellborder="1"><tr>`)
				for _, cell := range rowList {
					if cell != nil {
						line := fmt.Sprintf(`   <td balign="left">%s</td>`, utils.PyStr(cell))
						html = append(html, strings.ReplaceAll(line, "><tdX", ""))
					}
				}
				html = append(html, "  </tr></table>")
				html = append(html, " </td></tr>")
				numRows++
			}
		} else if row != nil {
			html = append(html, " <tr><td>")
			html = append(html, fmt.Sprintf("  %v", row))
			html = append(html, " </td></tr>")
			numRows++
		}
	}
	if numRows == 0 {
		html = append(html, "<tr><td></td></tr>")
	}
	html = append(html, "</table>")
	return html
}

// HTMLBgcolorAttr mirrors wv_gv_html.html_bgcolor_attr.
func HTMLBgcolorAttr(color string) string {
	if color == "" {
		return ""
	}
	return fmt.Sprintf(` bgcolor="%s"`, model.TranslateColor(color, "HEX"))
}

// HTMLBgcolor mirrors wv_gv_html.html_bgcolor.
func HTMLBgcolor(color string, extraAttr string) string {
	if color == "" {
		return ""
	}
	return fmt.Sprintf("<tdX%s%s>", HTMLBgcolorAttr(color), extraAttr)
}

// HTMLColorbar mirrors wv_gv_html.html_colorbar.
func HTMLColorbar(color string) any {
	if color == "" {
		return nil
	}
	return HTMLBgcolor(color, ` width="4"`)
}

// HTMLImage mirrors wv_gv_html.html_image.
func HTMLImage(image *model.Image) any {
	if image == nil {
		return nil
	}
	html := HTMLSizeAttr(image) + `><img scale="` + image.Scale + `" src="` + image.Src + `"/>`
	if image.Fixedsize {
		html = `>
    <table border="0" cellspacing="0" cellborder="0"><tr>
     <td` + html + `</td>
    </tr></table>
   `
	}
	side := ""
	if image.Caption != "" {
		side = ` sides="TLR"`
	}
	return `<tdX` + side + HTMLBgcolorAttr(image.Bgcolor) + html
}

// HTMLCaption mirrors wv_gv_html.html_caption.
func HTMLCaption(image *model.Image) any {
	if image == nil || image.Caption == "" {
		return nil
	}
	return fmt.Sprintf(`<tdX sides="BLR"%s>%s`, HTMLBgcolorAttr(image.Bgcolor), HTMLLineBreaks(image.Caption))
}

// HTMLSizeAttr mirrors wv_gv_html.html_size_attr.
func HTMLSizeAttr(image *model.Image) string {
	if image == nil {
		return ""
	}
	var sb strings.Builder
	if image.Width != nil {
		fmt.Fprintf(&sb, ` width="%s"`, utils.PyStr(image.Width))
	}
	if image.Height != nil {
		fmt.Fprintf(&sb, ` height="%s"`, utils.PyStr(image.Height))
	}
	if image.Fixedsize {
		sb.WriteString(` fixedsize="true"`)
	}
	return sb.String()
}

// HTMLLineBreaks mirrors wv_gv_html.html_line_breaks.
func HTMLLineBreaks(inp any) any {
	if s, ok := inp.(string); ok {
		return strings.ReplaceAll(utils.RemoveLinksStr(s), "\n", "<br />")
	}
	return inp
}

// HTMLLineBreaksStr is the string-typed variant of HTMLLineBreaks.
func HTMLLineBreaksStr(s string) string {
	return strings.ReplaceAll(utils.RemoveLinksStr(s), "\n", "<br />")
}

func anyTruthy(vals []any) bool {
	for _, v := range vals {
		if truthy(v) {
			return true
		}
	}
	return false
}

func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case int:
		return x != 0
	case float64:
		return x != 0
	default:
		return true
	}
}
