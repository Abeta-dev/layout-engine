package usecase

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/abeta-dev/go-libs/apperror"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
)

// ActionExecutionInput mirrors the KMP client action execution contract for drop-in compatibility.
type ActionExecutionInput struct {
	Trigger    actions.ActionTrigger `json:"trigger"`
	FormValues map[string]any        `json:"form_values,omitempty"`
	ScreenID   string                `json:"screen_id,omitempty"`
	Metadata   map[string]any        `json:"metadata,omitempty"`
}

// ActionExecutionResult mirrors the KMP client action execution response.
type ActionExecutionResult struct {
	ActionID        string                  `json:"action_id"`
	ActionType      string                  `json:"action_type"`
	Status          string                  `json:"status"`
	Message         string                  `json:"message,omitempty"`
	UpdatedState    map[string]any          `json:"updated_state,omitempty"`
	FollowUpActions []actions.ActionTrigger `json:"follow_up_actions,omitempty"`
}

// ActionService handles KMP SDUI action callbacks (/api/v1/actions/execute & /api/v1/orders/express)
// so any KMP app pointed at layout-engine works end-to-end with zero extra BFF required.
type ActionService struct {
	mu    sync.RWMutex
	state map[string]any
}

// NewActionService creates an ActionService.
func NewActionService() *ActionService {
	return &ActionService{state: make(map[string]any)}
}

// Execute processes all 18 KMP v1 action types cleanly.
func (s *ActionService) Execute(ctx context.Context, input ActionExecutionInput) (ActionExecutionResult, error) {
	if err := ctx.Err(); err != nil {
		return ActionExecutionResult{}, apperror.Wrap(err, apperror.CodeUnavailable, "request cancelled")
	}
	t := input.Trigger
	if strings.TrimSpace(t.ID) == "" || strings.TrimSpace(string(t.Type)) == "" {
		return ActionExecutionResult{}, apperror.BadRequest("trigger.id and trigger.type are required")
	}
	res := ActionExecutionResult{
		ActionID:   t.ID,
		ActionType: string(t.Type),
		Status:     "executed",
	}

	switch t.Type {
	case actions.ActionCallAPI:
		p, err := actions.DecodePayload[actions.CallAPIPayload](t)
		if err != nil {
			return res, apperror.BadRequest(err.Error())
		}
		res.Message = fmt.Sprintf("executed %s %s", p.Method, p.Endpoint)
		if p.Endpoint == "/api/v1/orders/express" {
			res.FollowUpActions = append(res.FollowUpActions,
				actions.NewTriggerFeedbackAction("express_order_celebration", actions.TriggerFeedbackPayload{
					Effect:     actions.FeedbackCelebration,
					Title:      "Order Confirmed!",
					Subtitle:   "Abeta Fresh Express rider arriving in 8 minutes",
					DurationMs: 2400,
				}),
				actions.NewShowToastAction("express_order_toast", actions.ShowToastPayload{
					Message:    "Express order placed via Layout Engine!",
					DurationMs: 2500,
					Variant:    "success",
				}),
			)
		}
		if p.OnSuccess != nil {
			res.FollowUpActions = append(res.FollowUpActions, *p.OnSuccess)
		}
		if len(input.FormValues) > 0 {
			res.UpdatedState = map[string]any{"last_submitted_form": input.FormValues}
		}
	case actions.ActionSaveState:
		p, err := actions.DecodePayload[actions.SaveStatePayload](t)
		if err != nil {
			return res, apperror.BadRequest(err.Error())
		}
		s.mu.Lock()
		s.state[p.Key] = p.Value
		s.mu.Unlock()
		res.UpdatedState = map[string]any{p.Key: p.Value}
		res.Message = fmt.Sprintf("saved state %s", p.Key)
	case actions.ActionDeleteState:
		p, err := actions.DecodePayload[actions.DeleteStatePayload](t)
		if err != nil {
			return res, apperror.BadRequest(err.Error())
		}
		s.mu.Lock()
		delete(s.state, p.Key)
		s.mu.Unlock()
		res.UpdatedState = map[string]any{p.Key: nil}
		res.Message = fmt.Sprintf("deleted state %s", p.Key)
	case actions.ActionNavigate,
		actions.ActionBackPress,
		actions.ActionOpenDialog,
		actions.ActionDismissOverlay,
		actions.ActionShowToast,
		actions.ActionTrackEvent,
		actions.ActionRefreshPage,
		actions.ActionCopyToClipboard,
		actions.ActionUpdateSnippets,
		actions.ActionBatch,
		actions.ActionTriggerFeedback,
		actions.ActionOpenURL,
		actions.ActionShare,
		actions.ActionRequestPermission,
		actions.ActionBiometricAuth:
		res.Message = fmt.Sprintf("acknowledged %s", t.Type)
	default:
		return res, apperror.BadRequest(fmt.Sprintf("unsupported action type: %s", t.Type))
	}

	if len(t.SecondaryActions) > 0 {
		res.FollowUpActions = append(res.FollowUpActions, t.SecondaryActions...)
	}
	return res, nil
}
