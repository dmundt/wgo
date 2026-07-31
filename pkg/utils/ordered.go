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

package utils

// KV is a single key/value pair in insertion order.
type KV struct {
	Key   string
	Value any
}

// OrderedMap preserves insertion order of keys, mirroring Python dict semantics.
type OrderedMap struct {
	keys []string
	idx  map[string]int
	vals []any
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{idx: map[string]int{}}
}

// Set inserts or replaces a key. Replacing keeps the original position.
func (m *OrderedMap) Set(k string, v any) {
	if i, ok := m.idx[k]; ok {
		m.vals[i] = v
		return
	}
	m.idx[k] = len(m.keys)
	m.keys = append(m.keys, k)
	m.vals = append(m.vals, v)
}

// SetGet inserts a key and returns the map for chaining.
func (m *OrderedMap) SetGet(k string, v any) *OrderedMap {
	m.Set(k, v)
	return m
}

func (m *OrderedMap) Get(k string) (any, bool) {
	if i, ok := m.idx[k]; ok {
		return m.vals[i], true
	}
	return nil, false
}

func (m *OrderedMap) Has(k string) bool {
	_, ok := m.idx[k]
	return ok
}

func (m *OrderedMap) Len() int {
	return len(m.keys)
}

func (m *OrderedMap) Keys() []string {
	return append([]string(nil), m.keys...)
}

// Items returns key/value pairs in insertion order.
func (m *OrderedMap) Items() []KV {
	out := make([]KV, 0, len(m.keys))
	for i, k := range m.keys {
		out = append(out, KV{Key: k, Value: m.vals[i]})
	}
	return out
}

// First returns the first key/value pair. The map is assumed non-empty.
func (m *OrderedMap) First() KV {
	return KV{Key: m.keys[0], Value: m.vals[0]}
}
