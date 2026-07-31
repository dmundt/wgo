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

// Package html ports WireViz's wv_html.py and svgembed.py, plus the bundled
// HTML templates.
package html

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dmundt/wgo/pkg/utils"
)

var mimeSubtypeReplacements = map[string]string{
	"jpg": "jpeg",
	"tif": "tiff",
}

// GetMimeSubtype returns the lowercased file extension mime subtype.
func GetMimeSubtype(filename string) string {
	mimeSubtype := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if r, ok := mimeSubtypeReplacements[mimeSubtype]; ok {
		mimeSubtype = r
	}
	return mimeSubtype
}

// DataURIBase64 returns a base64 data URI of a file.
func DataURIBase64(file string, media string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	b64 := base64.StdEncoding.EncodeToString(b)
	uri := "data:" + media + "/" + GetMimeSubtype(file) + ";base64, " + b64
	if len(uri) > 65535 {
		println("data_URI_base64(): Warning: Browsers might have different URI length limitations")
	}
	return uri
}

var imageTagRe = regexp.MustCompile(`(?i)<image( [^>]*?)? xlink:href="([^"]*?)"( [^>]*?)?>`)

// EmbedSVGImages returns svg with image hrefs replaced by base64 data URIs.
func EmbedSVGImages(svgIn string, basePath string) string {
	imagesB64 := map[string]string{}
	repl := func(match string) string {
		m := imageTagRe.FindStringSubmatch(match)
		if m == nil {
			return match
		}
		pre, url, post := m[1], m[2], m[3]
		b64, ok := imagesB64[url]
		if !ok {
			imgurlAbs := url
			if !filepath.IsAbs(url) {
				imgurlAbs = filepath.Join(basePath, url)
			}
			if a, err := filepath.Abs(imgurlAbs); err == nil {
				imgurlAbs = a
			}
			if data, err := os.ReadFile(imgurlAbs); err == nil {
				b64 = base64.StdEncoding.EncodeToString(data)
			}
			imagesB64[url] = b64
		}
		return "<image" + pre + ` xlink:href="data:image/` + GetMimeSubtype(url) + ";base64, " + b64 + `"` + post + ">"
	}
	return imageTagRe.ReplaceAllStringFunc(svgIn, repl)
}

// EmbedSVGImagesFile rewrites an SVG file, embedding its images in place.
func EmbedSVGImagesFile(filenameIn string) error {
	content, err := os.ReadFile(filenameIn)
	if err != nil {
		return err
	}
	filenameAbs, _ := filepath.Abs(filenameIn)
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	out := EmbedSVGImages(text, filepath.Dir(filenameAbs))
	return utils.FileWriteText(filenameIn, out)
}
