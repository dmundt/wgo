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

// Package utils ports WireViz's wv_helper.py and provides supporting helpers
// (ordered maps, Python-style string/number formatting, file I/O).
package utils

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var awgEquivTable = map[string]string{
	"0.09": "28",
	"0.14": "26",
	"0.25": "24",
	"0.34": "22",
	"0.5":  "21",
	"0.75": "20",
	"1":    "18",
	"1.5":  "16",
	"2.5":  "14",
	"4":    "12",
	"6":    "10",
	"10":   "8",
	"16":   "6",
	"25":   "4",
	"35":   "2",
	"50":   "1",
}

var mm2EquivTable = map[string]string{}

func init() {
	for k, v := range awgEquivTable {
		mm2EquivTable[v] = k
	}
}

// AWGEquiv returns the AWG equivalent of an mm2 gauge, or "Unknown".
func AWGEquiv(mm2 any) string {
	if v, ok := awgEquivTable[PyStr(mm2)]; ok {
		return v
	}
	return "Unknown"
}

// MM2Equiv returns the mm2 equivalent of an AWG gauge, or "Unknown".
func MM2Equiv(awg any) string {
	if v, ok := mm2EquivTable[PyStr(awg)]; ok {
		return v
	}
	return "Unknown"
}

// Expand turns a pin spec (scalar, list, or '#-#' range) into a list.
func Expand(yamlData any) []any {
	var seq []any
	if l, ok := yamlData.([]any); ok {
		seq = l
	} else {
		seq = []any{yamlData}
	}
	output := []any{}
	for _, e := range seq {
		s := PyStr(e)
		if strings.Contains(s, "-") {
			parts := strings.SplitN(s, "-", 2)
			a, errA := strconv.Atoi(parts[0])
			b, errB := strconv.Atoi(parts[1])
			if errA == nil && errB == nil {
				if a < b {
					for x := a; x <= b; x++ {
						output = append(output, x)
					}
				} else if a > b {
					for x := a; x >= b; x-- {
						output = append(output, x)
					}
				} else {
					output = append(output, a)
				}
			} else {
				output = append(output, s)
			}
		} else {
			if x, err := strconv.Atoi(s); err == nil {
				output = append(output, x)
			} else {
				output = append(output, s)
			}
		}
	}
	return output
}

// GetSingleKeyAndValue returns the first key and value of a single-key dict.
func GetSingleKeyAndValue(m *OrderedMap) (string, any) {
	kv := m.First()
	return kv.Key, kv.Value
}

var linkRe = regexp.MustCompile(`<[aA] [^>]*>([^<]*)</[aA]>`)

// RemoveLinks removes HTML anchor tags, returning the inner text.
func RemoveLinks(inp any) any {
	if s, ok := inp.(string); ok {
		return RemoveLinksStr(s)
	}
	return inp
}

// RemoveLinksStr removes HTML anchor tags from a string.
func RemoveLinksStr(s string) string {
	return linkRe.ReplaceAllString(s, "${1}")
}

// CleanWhitespace collapses whitespace runs to single spaces.
func CleanWhitespace(inp any) any {
	if s, ok := inp.(string); ok {
		return CleanWhitespaceStr(s)
	}
	return inp
}

// CleanWhitespaceStr collapses whitespace runs to single spaces.
func CleanWhitespaceStr(s string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(s), " "), " ,", ",")
}

// Flatten2d converts a 2D row/col list to strings, joining inner lists with ", ".
func Flatten2d(inp [][]any) [][]string {
	var out [][]string
	for _, row := range inp {
		var nr []string
		for _, item := range row {
			if l, ok := item.([]any); ok {
				var parts []string
				for _, e := range l {
					parts = append(parts, PyStr(e))
				}
				nr = append(nr, strings.Join(parts, ", "))
			} else if l, ok := item.([]string); ok {
				nr = append(nr, strings.Join(l, ", "))
			} else {
				nr = append(nr, PyStr(item))
			}
		}
		out = append(out, nr)
	}
	return out
}

// Tuplelist2tsv converts a list of rows to a TSV string with optional header.
func Tuplelist2tsv(inp [][]any, header []string) string {
	if header != nil {
		row := make([]any, len(header))
		for i, h := range header {
			row[i] = h
		}
		inp = append([][]any{row}, inp...)
	}
	flat := Flatten2d(inp)
	var sb strings.Builder
	for _, row := range flat {
		var cells []string
		for _, item := range row {
			cells = append(cells, RemoveLinksStr(item))
		}
		sb.WriteString(strings.Join(cells, "\t"))
		sb.WriteString("\n")
	}
	return sb.String()
}

// IsArrow reports whether s is an arrow like <--, ==>, <=>.
func IsArrow(s string) bool {
	return arrowRe.MatchString(s)
}

var arrowRe = regexp.MustCompile(`^\s*<?(-+|=+)>?\s*$`)

// AspectRatio returns width/height of an image file, defaulting to 1 on error.
func AspectRatio(imageSrc string) float64 {
	f, err := os.Open(imageSrc)
	if err != nil {
		fmt.Printf("aspect_ratio(): %s: %v\n", pyErrType(err), err)
		return 1
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		fmt.Printf("aspect_ratio(): %s: %v\n", pyErrType(err), err)
		return 1
	}
	if cfg.Width > 0 && cfg.Height > 0 {
		return float64(cfg.Width) / float64(cfg.Height)
	}
	fmt.Printf("aspect_ratio(): Invalid image size %d x %d\n", cfg.Width, cfg.Height)
	return 1
}

func pyErrType(err error) string {
	if os.IsNotExist(err) {
		return "FileNotFoundError"
	}
	return "OSError"
}

// SmartFileResolve finds filename in the list of possible base paths.
func SmartFileResolve(filename string, possiblePaths []string) (string, error) {
	if filepath.IsAbs(filename) {
		if fileExists(filename) {
			return filename, nil
		}
		return "", fmt.Errorf("%s does not exist.", filename)
	}
	var resolved []string
	for _, p := range possiblePaths {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		resolved = append(resolved, abs)
		rp := filepath.Join(abs, filename)
		if fileExists(rp) {
			return rp, nil
		}
	}
	return "", fmt.Errorf("%s was not found in any of the following locations: \n%s",
		filename, strings.Join(resolved, "\n"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileReadText reads a UTF-8 text file.
func FileReadText(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

// FileWriteText writes a UTF-8 text file, translating newlines like Python's
// text-mode file writes (os.linesep on the current platform).
func FileWriteText(path string, text string) error {
	return os.WriteFile(path, []byte(LineEndings(text)), 0o644)
}

// LineEndings translates "\n" to os.linesep as Python text mode does.
func LineEndings(text string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(text, "\n", "\r\n")
	}
	return text
}

// FileWriteBytes writes raw bytes.
func FileWriteBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

// FileExists reports whether a path exists.
func FileExists(path string) bool {
	return fileExists(path)
}

// Exists is an alias for FileExists.
func Exists(path string) bool {
	return fileExists(path)
}

// RemoveFile removes a file, ignoring "not found" errors.
func RemoveFile(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// RenameFile renames a file.
func RenameFile(old, new string) error {
	return os.Rename(old, new)
}
