package studioassets_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	studioassets "github.com/abeta-dev/layout-engine/internal/handler/studio_assets"
)

func TestEmbeddedStudioAssets_All21SnippetTypes(t *testing.T) {
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

	all21 := []string{
		snippets.SnippetTypeAccordion,
		snippets.SnippetTypeActionRow,
		snippets.SnippetTypeBanner,
		snippets.SnippetTypeCard,
		snippets.SnippetTypeCarousel,
		snippets.SnippetTypeChipGroup,
		snippets.SnippetTypeComposedBlock,
		snippets.SnippetTypeDivider,
		snippets.SnippetTypeEmptyState,
		snippets.SnippetTypeForm,
		snippets.SnippetTypeGrid,
		snippets.SnippetTypeHeader,
		snippets.SnippetTypeMarqueeStrip,
		snippets.SnippetTypeMetricCard,
		snippets.SnippetTypeNudgeBar,
		snippets.SnippetTypeOrderTracker,
		snippets.SnippetTypeProductCard,
		snippets.SnippetTypeSearchBar,
		snippets.SnippetTypeSpacer,
		snippets.SnippetTypeSwipeDeck,
		snippets.SnippetTypeTabs,
	}

	for _, st := range all21 {
		if !strings.Contains(html, `value="`+st+`"`) {
			t.Errorf("index.html missing option value %q", st)
		}
		if !strings.Contains(js, `'`+st+`'`) {
			t.Errorf("app.js missing snippet type %q", st)
		}
	}

	for _, selector := range []string{
		".kmp-swipe-deck",
		".kmp-search-bar",
		".kmp-chip-group",
		".kmp-nudge-bar",
		".kmp-marquee-strip",
		".kmp-order-tracker",
		".kmp-composed-block",
		".kmp-cart-nudge",
		".kmp-radio-group",
		".kmp-select",
		".kmp-otp-input",
		".kmp-chip",
		".kmp-segmented",
		".kmp-rating",
		".kmp-price-atom",
		".kmp-badge-atom",
		".kmp-alert-banner",
		".kmp-stat-tile",
		".kmp-kv-row",
		".kmp-qr-code",
		".kmp-composed-scroll-row",
		".kmp-composed-grid",
		".kmp-composed-textfield",
		".kmp-composed-switch",
		".kmp-composed-checkbox",
		".kmp-composed-slider",
		".kmp-media-pager",
		".kmp-sticky-wrap",
	} {
		if !strings.Contains(css, selector) {
			t.Errorf("app.css missing preview style %q", selector)
		}
	}

	all36ComposedKinds := []string{
		snippets.ComposedNodeColumn,
		snippets.ComposedNodeRow,
		snippets.ComposedNodeBox,
		snippets.ComposedNodeCard,
		snippets.ComposedNodeFlowRow,
		snippets.ComposedNodeScrollRow,
		snippets.ComposedNodeGrid,
		snippets.ComposedNodeText,
		snippets.ComposedNodeButton,
		snippets.ComposedNodeTag,
		snippets.ComposedNodeIcon,
		snippets.ComposedNodeImage,
		snippets.ComposedNodeAvatar,
		snippets.ComposedNodeProgress,
		snippets.ComposedNodeDivider,
		snippets.ComposedNodeSpacer,
		snippets.ComposedNodeStepper,
		snippets.ComposedNodeCountdown,
		snippets.ComposedNodeRollingCounter,
		snippets.ComposedNodeRadioGroup,
		snippets.ComposedNodeSelect,
		snippets.ComposedNodeOtpInput,
		snippets.ComposedNodeChip,
		snippets.ComposedNodeSegmentedControl,
		snippets.ComposedNodeRating,
		snippets.ComposedNodePrice,
		snippets.ComposedNodeBadge,
		snippets.ComposedNodeAlertBanner,
		snippets.ComposedNodeStatTile,
		snippets.ComposedNodeKeyValueRow,
		snippets.ComposedNodeQrCode,
		snippets.ComposedNodeTextField,
		snippets.ComposedNodeSwitch,
		snippets.ComposedNodeCheckbox,
		snippets.ComposedNodeSlider,
		snippets.ComposedNodeMediaPager,
	}
	for _, kind := range all36ComposedKinds {
		if !strings.Contains(html, `value="`+kind+`"`) {
			t.Errorf("index.html composed molder picker missing node kind %q", kind)
		}
		if !strings.Contains(js, `'`+kind+`'`) {
			t.Errorf("app.js missing composed node kind %q", kind)
		}
	}

	all9BoxAlignments := []string{
		snippets.BoxAlignTopStart,
		snippets.BoxAlignTopCenter,
		snippets.BoxAlignTopEnd,
		snippets.BoxAlignCenterStart,
		snippets.BoxAlignCenter,
		snippets.BoxAlignCenterEnd,
		snippets.BoxAlignBottomStart,
		snippets.BoxAlignBottomCenter,
		snippets.BoxAlignBottomEnd,
	}
	for _, align := range all9BoxAlignments {
		if !strings.Contains(js, `'`+align+`'`) {
			t.Errorf("app.js missing box_alignment %q", align)
		}
		if !strings.Contains(css, ".kmp-box-align-"+align) {
			t.Errorf("app.css missing box_alignment class .kmp-box-align-%s", align)
		}
	}
}
