package main

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"strings"
	"syscall/js"

	tycho "github.com/snowpackjs/astro/internal"
	"github.com/snowpackjs/astro/internal/printer"
	"github.com/snowpackjs/astro/internal/transform"
	"github.com/snowpackjs/astro/internal/xxhash"
)

func main() {
	js.Global().Set("__astro_transform", js.FuncOf(Transform))
	<-make(chan bool)
}

func jsString(j js.Value) string {
	if j.IsUndefined() || j.IsNull() {
		return ""
	}
	return j.String()
}

func Transform(this js.Value, args []js.Value) interface{} {
	source := jsString(args[0])
	// options := js.Value(args[1])
	doc, _ := tycho.Parse(strings.NewReader(source))
	hash := hashFromSource(source)

	transform.Transform(doc, transform.TransformOptions{
		Scope: hash,
	})

	result := printer.PrintToJS(source, doc)
	content, _ := json.Marshal(source)
	sourcemap := `{ "version": 3, "sources": ["file.astro"], "names": [], "mappings": "` + string(result.SourceMapChunk.Buffer) + `", "sourcesContent": [` + string(content) + `] }`
	b64 := base64.StdEncoding.EncodeToString([]byte(sourcemap))
	output := string(result.Output) + string('\n') + `//# sourceMappingURL=data:application/json;base64,` + b64 + string('\n')

	// internalURL := jsString(options.Get("internalURL"))
	// if internalURL == "" {
	// 	internalURL = "@astrojs/compiler/internal"
	// }

	return output
}

func hashFromSource(source string) string {
	h := xxhash.New()
	h.Write([]byte(source))
	hashBytes := h.Sum(nil)
	return base32.StdEncoding.EncodeToString(hashBytes)[:8]
}