package engine

import (
	"fmt"
	"strings"

	"github.com/abeta-dev/kmp/bff/pkg/sdui/atoms"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/snippets"
	"github.com/abeta-dev/layout-engine/internal/domain"
)

func (v *DefaultLayoutValidator) validateSwipeCardsProp(
	path string,
	raw any,
	issues *[]domain.ValidationIssue,
) {
	switch cards := raw.(type) {
	case []snippets.SwipeCardItem:
		for i, card := range cards {
			validateTypedSwipeCard(fmt.Sprintf("%s/%d", path, i), card, issues)
		}
	case []map[string]any:
		for i, card := range cards {
			validateMapSwipeCard(fmt.Sprintf("%s/%d", path, i), card, issues)
		}
	case []any:
		for i, elem := range cards {
			cardPath := fmt.Sprintf("%s/%d", path, i)
			switch typed := elem.(type) {
			case snippets.SwipeCardItem:
				validateTypedSwipeCard(cardPath, typed, issues)
			case map[string]any:
				validateMapSwipeCard(cardPath, typed, issues)
			default:
				*issues = append(*issues, domain.ValidationIssue{
					Path:    cardPath,
					Code:    "invalid_swipe_card",
					Message: "swipe card entry must be an object",
				})
			}
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_swipe_cards",
			Message: "swipe_deck cards must be a list of card objects",
		})
	}
}

func validateTypedSwipeCard(path string, card snippets.SwipeCardItem, issues *[]domain.ValidationIssue) {
	if strings.TrimSpace(card.CardID) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/card_id",
			Code:    "required",
			Message: "swipe card requires card_id",
		})
	}
	if strings.TrimSpace(card.Title.Text) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/title",
			Code:    "required",
			Message: "swipe card requires title",
		})
	}
}

func validateMapSwipeCard(path string, m map[string]any, issues *[]domain.ValidationIssue) {
	cardID, _ := m["card_id"].(string)
	if strings.TrimSpace(cardID) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/card_id",
			Code:    "required",
			Message: "swipe card requires card_id",
		})
	}
	if extractAtomText(m["title"]) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/title",
			Code:    "required",
			Message: "swipe card requires title",
		})
	}
}

func (v *DefaultLayoutValidator) validateChipGroupProp(
	path string,
	raw any,
	issues *[]domain.ValidationIssue,
) {
	switch chips := raw.(type) {
	case []snippets.FilterChipItem:
		for i, chip := range chips {
			validateTypedFilterChip(fmt.Sprintf("%s/%d", path, i), chip, issues)
		}
	case []map[string]any:
		for i, chip := range chips {
			validateMapFilterChip(fmt.Sprintf("%s/%d", path, i), chip, issues)
		}
	case []any:
		for i, elem := range chips {
			chipPath := fmt.Sprintf("%s/%d", path, i)
			switch typed := elem.(type) {
			case snippets.FilterChipItem:
				validateTypedFilterChip(chipPath, typed, issues)
			case map[string]any:
				validateMapFilterChip(chipPath, typed, issues)
			default:
				*issues = append(*issues, domain.ValidationIssue{
					Path:    chipPath,
					Code:    "invalid_chip",
					Message: "chip entry must be an object",
				})
			}
		}
	default:
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path,
			Code:    "invalid_chips",
			Message: "chip_group chips must be a list of chip objects",
		})
	}
}

func validateTypedFilterChip(path string, chip snippets.FilterChipItem, issues *[]domain.ValidationIssue) {
	if strings.TrimSpace(chip.ID) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/id",
			Code:    "required",
			Message: "chip item requires id",
		})
	}
	if strings.TrimSpace(chip.Label) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/label",
			Code:    "required",
			Message: "chip item requires label",
		})
	}
}

func validateMapFilterChip(path string, m map[string]any, issues *[]domain.ValidationIssue) {
	id, _ := m["id"].(string)
	label, _ := m["label"].(string)
	if strings.TrimSpace(id) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/id",
			Code:    "required",
			Message: "chip item requires id",
		})
	}
	if strings.TrimSpace(label) == "" {
		*issues = append(*issues, domain.ValidationIssue{
			Path:    path + "/label",
			Code:    "required",
			Message: "chip item requires label",
		})
	}
}

func extractAtomText(raw any) string {
	switch val := raw.(type) {
	case string:
		return strings.TrimSpace(val)
	case map[string]any:
		s, _ := val["text"].(string)
		return strings.TrimSpace(s)
	case atoms.TextAtom:
		return strings.TrimSpace(val.Text)
	case *atoms.TextAtom:
		if val != nil {
			return strings.TrimSpace(val.Text)
		}
	}
	return ""
}
