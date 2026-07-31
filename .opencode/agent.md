# **Port WireViz → Go (wgo) — Direct Porting Instructions**

### **Reference**
Original Python repository:  
[https://github.com/wireviz/WireViz.git](https://github.com/wireviz/WireViz.git)

Your task:  
**Re‑implement the same functionality in Go**, preserving behavior, structure, and output semantics where possible.

No new features.  
No architectural expansion.  
No redesign.  
Just a clean, idiomatic Go port.

---

## **1. Identify WireViz Components (Python)**

From the repository, port these modules:

- `wireviz.py` — CLI entry point  
- `parser.py` — YAML parsing  
- `graph.py` — Graphviz DOT generation  
- `markdown.py` — Markdown table generation  
- `utils.py` — helpers  
- `schema/` — YAML schema definitions  
- `examples/` — reference inputs  

Your Go version must replicate these behaviors.

---

## **2. Create Equivalent Go Module Structure**

```
wgo/
  cmd/wgo/        # CLI equivalent to wireviz.py
  pkg/
    parser/       # YAML parser (parser.py)
    model/        # Data structures (schema)
    graph/        # DOT generation (graph.py)
    markdown/     # Markdown generation (markdown.py)
    utils/        # helpers (utils.py)
```

This mirrors the Python repo one‑to‑one.

---

## **3. Port YAML Schema → Go Structs**

Translate WireViz YAML schema exactly:

- connectors  
- cables  
- wires  
- signals  
- metadata  
- mapping tables  

Use:

```go
gopkg.in/yaml.v3
```

Constraints:

- same field names  
- same nesting  
- same validation rules  
- reject unknown fields (WireViz does this implicitly)

---

## **4. Port Python Logic → Go Logic**

### **4.1 parser.py → parser package**
Implement:

- YAML load  
- schema validation  
- normalization  
- default handling  

Match Python behavior exactly.

### **4.2 graph.py → graph package**
Implement:

- DOT node generation  
- DOT edge generation  
- label formatting  
- color/style mapping  
- layout hints  

Output must match WireViz’s DOT output as closely as possible.

### **4.3 markdown.py → markdown package**
Implement:

- connector tables  
- cable tables  
- mapping tables  
- diagram embedding  

Markdown output must be identical in structure.

### **4.4 utils.py → utils package**
Port:

- string helpers  
- sorting rules  
- formatting helpers  

---

## **5. Port CLI Behavior**

`wireviz.py` supports:

- input YAML file  
- output directory  
- diagram generation  
- markdown generation  

Your Go CLI must replicate:

```
wgo input.yaml
```

and produce:

- `diagram.dot`  
- `diagram.svg` (if Graphviz installed)  
- `output.md`  

Same filenames, same flow.

---

## **6. Rendering Backend**

WireViz uses Graphviz via CLI:

```
dot -Tsvg
```

Your Go port must:

- generate DOT text exactly like WireViz  
- optionally call Graphviz if available  
- otherwise skip SVG generation (same behavior)

No new renderers.  
No D2.  
No Mermaid.  
No ASCII.

Just the WireViz DOT → SVG pipeline.

---

## **7. Output Equivalence Requirements**

Your Go port must match WireViz output:

- same DOT structure  
- same Markdown tables  
- same ordering  
- same formatting  
- same filenames  

If WireViz sorts connectors alphabetically, wgo must do the same.  
If WireViz prints mapping lines in insertion order, wgo must do the same.

---

## **8. Testing Against WireViz**

For each example in `examples/`:

1. Run WireViz → produce DOT + MD  
2. Run wgo → produce DOT + MD  
3. Compare outputs (string diff)

Your port is correct when:

- DOT matches  
- Markdown matches  
- SVG matches (if Graphviz installed)

---

## **9. No New Requirements**

You must **not**:

- add new renderers  
- add new schema fields  
- add new CLI flags  
- add plugin systems  
- add feature flags  
- add new markdown formats  
- add new diagram formats  
- add new pipeline logic  

This is a **direct port**, not an extended tool.
