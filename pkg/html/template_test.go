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
	"strings"
	"testing"

	"github.com/dmundt/wgo/pkg/utils"
)

func TestHTMLTemplateRender(t *testing.T) {
	tmpl := newHTMLTemplate("a<!-- %x% -->b<!-- %y% -->c")
	got := tmpl.Render(map[string]string{
		"<!-- %x% -->": "X",
		"<!-- %y% -->": "Y",
	})
	if got != "aXbYc" {
		t.Errorf("Render = %q, want aXbYc", got)
	}
}

func TestHTMLTemplateRenderMissingValue(t *testing.T) {
	// A placeholder with no value must be left untouched.
	tmpl := newHTMLTemplate("<!-- %x% -->")
	got := tmpl.Render(map[string]string{"<!-- %other% -->": "O"})
	if got != "<!-- %x% -->" {
		t.Errorf("Render = %q, want placeholder untouched", got)
	}
}

func TestHTMLTemplateRenderLongestFirst(t *testing.T) {
	// A longer marker that begins with a shorter one must win at the same
	// position, so a specific placeholder is not swallowed by a prefix match.
	tmpl := newHTMLTemplate("<!-- %a% -->-suffix")
	got := tmpl.Render(map[string]string{
		"<!-- %a% -->":        "A",
		"<!-- %a% -->-suffix": "FULL",
	})
	if got != "FULL" {
		t.Errorf("Render = %q, want FULL (longest marker wins)", got)
	}
}

func TestHTMLTemplateHas(t *testing.T) {
	tmpl := newHTMLTemplate("x<!-- %diagram% -->y")
	if !tmpl.Has("<!-- %diagram% -->") {
		t.Errorf("Has(diagram) = false, want true")
	}
	if tmpl.Has("<!-- %diagram_png_b64% -->") {
		t.Errorf("Has(diagram_png_b64) = true, want false")
	}
}

func TestHTMLTemplateNormalizesLineEndings(t *testing.T) {
	tmpl := newHTMLTemplate("a\r\nb\r\nc")
	if tmpl.text != "a\nb\nc" {
		t.Errorf("text = %q, want a\\nb\\nc", tmpl.text)
	}
}

func TestRenderBOM(t *testing.T) {
	bomList := [][]any{
		{"Id", "Description"},
		{"1", "Connector"},
		{"2", "Cable"},
	}
	fwd, rev := renderBOM(bomList)
	if !strings.Contains(fwd, `<table class="bom">`) {
		t.Errorf("forward missing table tag")
	}
	if strings.Index(fwd, "Connector") > strings.Index(fwd, "Cable") {
		t.Errorf("forward should keep row order")
	}
	if strings.Index(rev, "Connector") < strings.Index(rev, "Cable") {
		t.Errorf("reversed should invert row order")
	}
}

func TestAddMetadataPlaceholders(t *testing.T) {
	metadata := utils.NewOrderedMap()
	metadata.Set("title", "Demo harness")
	people := utils.NewOrderedMap()
	person := utils.NewOrderedMap()
	person.Set("name", "A. Author")
	person.Set("date", "2026-07-31")
	people.Set("A. Author", person)
	metadata.Set("people", people)

	values := map[string]string{}
	addMetadataPlaceholders(values, metadata)

	if got, ok := values["<!-- %title% -->"]; !ok || got != "Demo harness" {
		t.Errorf("title = %q, %v", got, ok)
	}
	if got, ok := values["<!-- %people_1% -->"]; !ok || got != "A. Author" {
		t.Errorf("people_1 = %q, %v", got, ok)
	}
	if got, ok := values["<!-- %people_1_name% -->"]; !ok || got != "A. Author" {
		t.Errorf("people_1_name = %q, %v", got, ok)
	}
	if _, ok := values["<!-- %people_1_date% -->"]; !ok {
		t.Errorf("people_1_date missing")
	}
}
