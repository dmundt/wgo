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
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/utils"
)

func TestGetMimeSubtype(t *testing.T) {
	cases := map[string]string{
		"x.jpg": "jpeg",
		"x.JPG": "jpeg",
		"x.png": "png",
		"x.tif": "tiff",
		"x.svg": "svg",
		"x.PNG": "png",
		"noext": "",
	}
	for in, want := range cases {
		if got := GetMimeSubtype(in); got != want {
			t.Errorf("GetMimeSubtype(%q) = %q, want %q", in, got, want)
		}
	}
}

func writePNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.White)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestDataURIBase64(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "a.png")
	writePNG(t, img)
	uri := DataURIBase64(img, "image")
	if !strings.HasPrefix(uri, "data:image/png;base64, ") {
		t.Errorf("uri = %q", uri)
	}
}

func TestEmbedSVGImages(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "a.png")
	writePNG(t, img)
	svg := `<svg xmlns="http://www.w3.org/2000/svg"><image xlink:href="a.png" width="10" height="10"/></svg>`
	out := EmbedSVGImages(svg, dir)
	if !strings.Contains(out, `xlink:href="data:image/png;base64, `) {
		t.Errorf("image not embedded: %q", out)
	}
	if strings.Contains(out, `xlink:href="a.png"`) {
		t.Errorf("original href still present: %q", out)
	}
	// Images are cached: embedding the same svg again yields the same bytes.
	out2 := EmbedSVGImages(svg, dir)
	if out != out2 {
		t.Errorf("embedding not deterministic")
	}
}

func TestEmbedSVGImagesAbsolute(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "a.png")
	writePNG(t, img)
	svg := `<svg><image xlink:href="` + img + `" width="10" height="10"/></svg>`
	out := EmbedSVGImages(svg, t.TempDir())
	if !strings.Contains(out, "data:image/png;base64, ") {
		t.Errorf("absolute path image not embedded: %q", out)
	}
}

func TestEmbedSVGImagesFile(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "a.png")
	writePNG(t, img)
	svg := filepath.Join(dir, "out.svg")
	if err := os.WriteFile(svg, []byte(`<svg><image xlink:href="a.png" width="10" height="10"/></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EmbedSVGImagesFile(svg); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(svg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "data:image/png;base64, ") {
		t.Errorf("file not embedded: %q", string(content))
	}
}

func TestGenerateHTMLOutputSimple(t *testing.T) {
	dir := t.TempDir()
	outBase := filepath.Join(dir, "out")
	if err := os.WriteFile(outBase+".tmp.svg", []byte(
		`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			`<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">`+"\n"+
			`<svg xmlns="http://www.w3.org/2000/svg"><rect width="10" height="10"/></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	metadata := utils.NewOrderedMap()
	metadata.Set("title", "My Doc")
	metadata.Set("notes", "line1\nline2")
	bom := [][]any{{"Id", "Description", "Qty"}, {"1", "Cable, x", "2"}}

	options := model.NewOptions(nil)
	if err := GenerateHTMLOutput(outBase, bom, metadata, options); err != nil {
		t.Fatal(err)
	}
	htmlOut, err := os.ReadFile(outBase + ".html")
	if err != nil {
		t.Fatal(err)
	}
	content := string(htmlOut)
	for _, want := range []string{
		"<title>My Doc</title>",
		"<!-- XML and DOCTYPE declarations from SVG file removed -->",
		"line1<br />line2",
		`<th class="bom_col_id">Id</th>`,
		`<td class="bom_col_description">Cable, x</td>`,
		`<meta name="generator" content="WireViz 0.4.1 - https://github.com/wireviz/WireViz">`,
		`<h1>My Doc</h1>`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("html missing %q", want)
		}
	}
	if strings.Contains(content, "<?xml") {
		t.Errorf("xml declaration not removed")
	}
	if strings.Contains(content, "<!DOCTYPE svg") {
		t.Errorf("svg doctype not removed")
	}
}

func TestGenerateHTMLOutputDin6771(t *testing.T) {
	dir := t.TempDir()
	outBase := filepath.Join(dir, "demo")
	if err := os.WriteFile(outBase+".tmp.svg", []byte(`<svg></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}
	metadata := utils.NewOrderedMap()
	metadata.Set("title", "Demo 2")
	metadata.Set("pn", "WV-DEMO-02")
	metadata.Set("template", utils.NewOrderedMap().SetGet("name", "din-6771").SetGet("sheetsize", "A3"))
	authors := utils.NewOrderedMap()
	authors.Set("Created", utils.NewOrderedMap().SetGet("name", "D. Rojas").SetGet("date", "2020-05-20"))
	metadata.Set("authors", authors)
	revisions := utils.NewOrderedMap()
	revisions.Set("A", utils.NewOrderedMap().SetGet("name", "D. Rojas").SetGet("date", "2020-10-17").SetGet("changelog", "v0.2"))
	metadata.Set("revisions", revisions)

	if err := GenerateHTMLOutput(outBase, [][]any{{"Id"}, {"1"}}, metadata, model.NewOptions(nil)); err != nil {
		t.Fatal(err)
	}
	htmlOut, err := os.ReadFile(outBase + ".html")
	if err != nil {
		t.Fatal(err)
	}
	content := string(htmlOut)
	for _, want := range []string{
		`class="A3"`,
		"WV-DEMO-02",
		"D. Rojas",
		"2020-05-20",
		"v0.2",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("din-6771 html missing %q", want)
		}
	}
	if !strings.Contains(content, "<!-- %revisions_8% -->") {
		t.Errorf("unused revision slot should remain a placeholder")
	}
}

func TestDataURIBase64Large(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big.png")
	// Create a file large enough to trigger the URI length warning (> 65535).
	data := make([]byte, 90000)
	for i := range data {
		data[i] = byte(i % 251)
	}
	if err := os.WriteFile(big, data, 0o644); err != nil {
		t.Fatal(err)
	}
	uri := DataURIBase64(big, "image")
	if !strings.HasPrefix(uri, "data:image/png;base64, ") {
		t.Errorf("uri prefix = %q", uri)
	}
	if len(uri) < 100000 {
		t.Errorf("uri too short: %d", len(uri))
	}
}
