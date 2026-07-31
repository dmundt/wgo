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

// Package graph ports the graphviz Python library's DOT source generation
// (version 0.21) used by WireViz, plus the Graphviz HTML-label helpers.
package graph

import (
	"regexp"
	"sort"
	"strings"
)

// Graph mirrors graphviz.Graph from the Python library.
type Graph struct {
	Body []string
}

// NewGraph creates an empty undirected graph.
func NewGraph() *Graph {
	return &Graph{}
}

// Source renders the DOT source, matching graphviz 0.21's ”.join(self).
func (g *Graph) Source() string {
	var sb strings.Builder
	sb.WriteString("graph {\n")
	for _, l := range g.Body {
		sb.WriteString(l)
	}
	sb.WriteString("}\n")
	return sb.String()
}

// Attr appends a graph/node/edge attribute statement line.
func (g *Graph) Attr(kw string, attrs map[string]string) {
	g.Body = append(g.Body, "\t"+kw+AttrList("", attrs)+"\n")
}

// Node appends a node statement line.
func (g *Graph) Node(name, label string, attrs map[string]string) {
	g.Body = append(g.Body, "\t"+Quote(name)+AttrList(label, attrs)+"\n")
}

// Edge appends an undirected edge statement line.
func (g *Graph) Edge(tail, head, label string, attrs map[string]string) {
	g.Body = append(g.Body, "\t"+QuoteEdge(tail)+" -- "+QuoteEdge(head)+AttrList(label, attrs)+"\n")
}

var idRe = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*|-?(\.[0-9]+|[0-9]+(\.[0-9]*)?))$`)

var dotKeywords = map[string]bool{
	"node": true, "edge": true, "graph": true, "digraph": true, "subgraph": true, "strict": true,
}

// Quote mirrors graphviz.quoting.quote.
func Quote(identifier string) string {
	if strings.HasPrefix(identifier, "<") && strings.HasSuffix(identifier, ">") {
		return identifier
	}
	if !idRe.MatchString(identifier) || dotKeywords[strings.ToLower(identifier)] {
		return `"` + escapeQuotes(identifier) + `"`
	}
	return identifier
}

// QuoteEdge mirrors graphviz.quoting.quote_edge.
func QuoteEdge(identifier string) string {
	node, rest, _ := strings.Cut(identifier, ":")
	parts := []string{Quote(node)}
	if rest != "" {
		port, compass, _ := strings.Cut(rest, ":")
		parts = append(parts, Quote(port))
		if compass != "" {
			parts = append(parts, compass)
		}
	}
	return strings.Join(parts, ":")
}

// AttrList assembles a DOT attribute list string like graphviz.quoting.attr_list.
func AttrList(label string, attrs map[string]string) string {
	var parts []string
	if label != "" {
		parts = append(parts, "label="+Quote(label))
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, k+"="+Quote(attrs[k]))
	}
	if len(parts) == 0 {
		return ""
	}
	return " [" + strings.Join(parts, " ") + "]"
}

// escapeQuotes escapes quotes that are not already escaped.
func escapeQuotes(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			j := i - 1
			n := 0
			for j >= 0 && s[j] == '\\' {
				n++
				j--
			}
			if n%2 == 0 {
				sb.WriteByte('\\')
			}
			sb.WriteByte('"')
		} else {
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}
