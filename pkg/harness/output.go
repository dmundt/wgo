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

// Package harness ports WireViz's Harness.py: the wiring graph model, DOT
// generation, and file output.
package harness

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dmundt/wgo/pkg/html"
	"github.com/dmundt/wgo/pkg/utils"
)

// Output writes the requested output files.
func (h *Harness) Output(filename string, fmts []string) error {
	if dir := filepath.Dir(filename); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	g, err := h.CreateGraph()
	if err != nil {
		return err
	}
	source := g.Source()

	for _, f := range fmts {
		switch f {
		case "png", "svg", "html":
			if f == "html" {
				f = "svg"
			}
			outFile := filename
			if f == "svg" {
				outFile = filename + ".tmp"
			}
			if err := renderDOT(source, f, outFile); err != nil {
				return err
			}
		}
	}

	if contains(fmts, "svg") || contains(fmts, "html") {
		if err := html.EmbedSVGImagesFile(filename + ".tmp.svg"); err != nil {
			return err
		}
	}

	if contains(fmts, "gv") {
		if err := utils.FileWriteText(filename+".gv", source); err != nil {
			return err
		}
	}

	bom, err := h.Bom()
	if err != nil {
		return err
	}
	bomlist := bomList(bom)
	if contains(fmts, "tsv") {
		if err := utils.FileWriteText(filename+".bom.tsv", utils.Tuplelist2tsv(bomlist, nil)); err != nil {
			return err
		}
	}
	if contains(fmts, "csv") {
		fmt.Println("CSV output is not yet supported")
	}
	if contains(fmts, "html") {
		if err := html.GenerateHTMLOutput(filename, bomlist, h.Metadata, h.Options); err != nil {
			return err
		}
	}
	if contains(fmts, "pdf") {
		fmt.Println("PDF output is not yet supported")
	}

	if contains(fmts, "html") && !contains(fmts, "svg") {
		_ = removeFile(filename + ".tmp.svg")
	} else if contains(fmts, "svg") {
		if err := renameFile(filename+".tmp.svg", filename+".svg"); err != nil {
			return err
		}
	}
	return nil
}

func renderDOT(source, format, outFile string) error {
	dotPath, err := exec.LookPath("dot")
	if err != nil {
		return fmt.Errorf("Graphviz (dot) not found in PATH")
	}
	cmd := exec.Command(dotPath, "-T"+format, "-o", outFile+"."+format)
	cmd.Stdin = strings.NewReader(source)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dot failed: %v\n%s", err, string(out))
	}
	return nil
}

func removeFile(path string) error {
	return utils.RemoveFile(path)
}

func renameFile(old, new string) error {
	return utils.RenameFile(old, new)
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}
