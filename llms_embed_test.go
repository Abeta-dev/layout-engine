package layoutengine_test

import (
	"bytes"
	"testing"

	layoutengine "github.com/abeta-dev/layout-engine"
)

func TestEmbeddedLLMsDocuments(t *testing.T) {
	t.Parallel()

	const wantPrefix = "# Abeta Dynamic Layout Engine"

	if len(layoutengine.LLMsTxt) == 0 {
		t.Fatal("LLMsTxt must not be empty")
	}
	if !bytes.HasPrefix(layoutengine.LLMsTxt, []byte(wantPrefix)) {
		t.Errorf("LLMsTxt must start with %q", wantPrefix)
	}

	if len(layoutengine.LLMsFullTxt) == 0 {
		t.Fatal("LLMsFullTxt must not be empty")
	}
	if !bytes.HasPrefix(layoutengine.LLMsFullTxt, []byte(wantPrefix)) {
		t.Errorf("LLMsFullTxt must start with %q", wantPrefix)
	}
}
