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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dmundt/wgo/pkg/model"
	"github.com/dmundt/wgo/pkg/parser"
	"github.com/dmundt/wgo/pkg/utils"
)

var formatCodes = map[string]string{
	"g": "gv",
	"h": "html",
	"p": "png",
	"s": "svg",
	"t": "tsv",
}

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", r)
			os.Exit(1)
		}
	}()

	var format string
	var outputDir string
	var outputName string
	var version bool
	var prepend stringSlice

	fs := flag.NewFlagSet("wgo", flag.ContinueOnError)
	fs.StringVar(&format, "f", "hpst", "Output formats (g=gv h=html p=png s=svg t=tsv)")
	fs.StringVar(&format, "format", "hpst", "Output formats (g=gv h=html p=png s=svg t=tsv)")
	fs.StringVar(&outputDir, "o", "", "Directory to use for output files, if different from input file directory.")
	fs.StringVar(&outputDir, "output-dir", "", "Directory to use for output files, if different from input file directory.")
	fs.StringVar(&outputName, "O", "", "File name (without extension) to use for output files, if different from input file name.")
	fs.StringVar(&outputName, "output-name", "", "File name (without extension) to use for output files, if different from input file name.")
	fs.BoolVar(&version, "V", false, "Output WireViz version and exit.")
	fs.BoolVar(&version, "version", false, "Output WireViz version and exit.")
	fs.Var(&prepend, "p", "YAML file to prepend to the input file (optional).")
	fs.Var(&prepend, "prepend", "YAML file to prepend to the input file (optional).")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: wgo [options] FILE...\n\n")
		fmt.Fprintf(os.Stderr, "Parses the provided FILE and generates the specified outputs.\n\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nThe -f or --format option accepts a string containing one or more of the\n")
		fmt.Fprintf(os.Stderr, "following characters to specify which file types to output:\n")
		fmt.Fprintf(os.Stderr, "g (GV), h (HTML), p (PNG), s (SVG), t (TSV)\n")
	}

	flagArgs, positional := splitArgs(os.Args[1:])
	if err := fs.Parse(flagArgs); err != nil {
		os.Exit(1)
	}
	files := positional

	fmt.Println()
	fmt.Printf("%s %s\n", model.AppName, model.Version)
	if version {
		return
	}

	if len(files) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	var outputFormats []string
	for _, code := range format {
		if f, ok := formatCodes[string(code)]; ok {
			outputFormats = append(outputFormats, f)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown output format: %s\n", string(code))
			os.Exit(1)
		}
	}
	sort.Strings(outputFormats)
	outputFormats = uniqueStrings(outputFormats)
	outputFormatsStr := ""
	if len(outputFormats) > 1 {
		outputFormatsStr = "[" + strings.Join(outputFormats, "|") + "]"
	} else if len(outputFormats) == 1 {
		outputFormatsStr = outputFormats[0]
	}

	prependInput := ""
	for _, pf := range prepend {
		if !utils.FileExists(pf) {
			fmt.Fprintf(os.Stderr, "File does not exist:\n%s\n", pf)
			os.Exit(1)
		}
		fmt.Println("Prepend file:", pf)
		t, err := utils.FileReadText(pf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		prependInput += t + "\n"
	}

	for _, file := range files {
		if !utils.FileExists(file) {
			fmt.Fprintf(os.Stderr, "File does not exist:\n%s\n", file)
			os.Exit(1)
		}
		_outputDir := filepath.Dir(file)
		if outputDir != "" {
			_outputDir = outputDir
		}
		_outputName := stem(filepath.Base(file))
		if outputName != "" {
			_outputName = outputName
		}

		fmt.Println("Input file:  ", file)
		fmt.Printf("Output file: %s.%s\n", filepath.Join(_outputDir, _outputName), outputFormatsStr)

		yamlInput, err := utils.FileReadText(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		yamlInput = prependInput + yamlInput
		fileDir := filepath.Dir(file)
		imagePaths := []string{fileDir}
		for _, p := range prepend {
			imagePaths = append(imagePaths, filepath.Dir(p))
		}

		if err := parser.Run(yamlInput, file, _outputDir, _outputName, outputFormats, imagePaths); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println()
}

func stem(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return strings.TrimSuffix(name, ext)
}

func uniqueStrings(list []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range list {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// splitArgs separates flags (and their values) from positional file arguments,
// so flags may appear before or after files like Python's click allows.
func splitArgs(args []string) (flagArgs, positional []string) {
	knownVal := map[string]bool{
		"-f": true, "--format": true,
		"-o": true, "--output-dir": true,
		"-O": true, "--output-name": true,
		"-p": true, "--prepend": true,
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") && a != "-" {
			flagArgs = append(flagArgs, a)
			if knownVal[a] {
				i++
				if i < len(args) {
					flagArgs = append(flagArgs, args[i])
				}
			}
		} else {
			positional = append(positional, a)
		}
	}
	return flagArgs, positional
}
