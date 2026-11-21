package studioassets_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	studioassets "github.com/abeta-dev/layout-engine/internal/handler/studio_assets"
)

func TestEmbeddedStudioAssets_All17SnippetTypes(t *testing.T) {
	t.Parallel()

	htmlBytes, err := fs.ReadFile(studioassets.Files, "index.html")
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	jsBytes, err := fs.ReadFile(studioassets.Files, "app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	cssBytes, err := fs.ReadFile(studioassets.Files, "app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}

	html := string(htmlBytes)
	js := string(jsBytes)
	css := string(cssBytes)

	all17 := []string{
		snippets.SnippetTypeAccordion,
		snippets.SnippetTypeActionRow,
		snippets.SnippetTypeBanner,
		snippets.SnippetTypeCard,
		snippets.SnippetTypeCarousel,
		snippets.SnippetTypeChipGroup,
		snippets.SnippetTypeDivider,
		snippets.SnippetTypeEmptyState,
		snippets.SnippetTypeForm,
		snippets.SnippetTypeGrid,
		snippets.SnippetTypeHeader,
		snippets.SnippetTypeMetricCard,
		snippets.SnippetTypeProductCard,
		snippets.SnippetTypeSearchBar,
		snippets.SnippetTypeSpacer,
		snippets.SnippetTypeSwipeDeck,
		snippets.SnippetTypeTabs,
	}

	for _, st := range all17 {
		if !strings.Contains(html, `value="`+st+`"`) {
			t.Errorf("index.html missing option value %q", st)
		}
		if !strings.Contains(js, `'`+st+`'`) {
			t.Errorf("app.js missing snippet type %q", st)
		}
	}

	for _, selector := range []string{".kmp-swipe-deck", ".kmp-search-bar", ".kmp-chip-group"} {
		if !strings.Contains(css, selector) {
			t.Errorf("app.css missing preview style %q", selector)
		}
	}
}
