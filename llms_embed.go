// Package layoutengine exposes embedded repository-root LLM reference documents
// so the HTTP server serves the exact same llms.txt and llms-full.txt tracked in Git.
package layoutengine

import _ "embed"

// LLMsTxt contains the embedded repository-root llms.txt document.
//
//go:embed llms.txt
var LLMsTxt []byte

// LLMsFullTxt contains the embedded repository-root llms-full.txt document.
//
//go:embed llms-full.txt
var LLMsFullTxt []byte
