package main

import (
	"bytes"
	"fmt"

	"golang.org/x/net/html"
)

type descriptionListWriter struct {
	htmlWriter *bytes.Buffer
}

func writeDescriptionList(htmlWriter *bytes.Buffer, f func(dl *descriptionListWriter)) {
	htmlWriter.WriteString(`<dl class="grid-1/3-2/3 grid-column-gap-1 grid-row-gap-1">`)
	dl := descriptionListWriter{htmlWriter}
	f(&dl)
	htmlWriter.WriteString(`</dl>`)
}

func (dl *descriptionListWriter) key(key string) {
	dl.htmlWriter.WriteString(fmt.Sprintf(`<dt class="font-bold">%s</dt>`, html.EscapeString(key)))
}

func (dl *descriptionListWriter) value(value string) {
	dl.htmlWriter.WriteString(fmt.Sprintf(`<dd class="mb-2">%s</dd>`, html.EscapeString(value)))
}
