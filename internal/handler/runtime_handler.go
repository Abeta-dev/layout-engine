package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/abeta-dev/go-libs/apperror"
	"github.com/abeta-dev/go-libs/ginmw"
	"github.com/abeta-dev/kmp/bff/pkg/sdui/actions"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/internal/usecase"
	"github.com/gin-gonic/gin"
)

type runtimeHandler struct {
	resolve *usecase.ResolveService
	actions *usecase.ActionService
}

func (h *runtimeHandler) listScreens(c *gin.Context) {
	tenant := extractTenantSlug(c)
	screens, err := h.resolve.ListPublishedScreens(c.Request.Context(), tenant)
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"screens": screens, "tenant": tenant})
}

// getScreen serves KMP-client drop-in compatible PageResponse at GET /api/v1/screens/:screenId.
func (h *runtimeHandler) getScreen(c *gin.Context) {
	reqCtx := extractRequestContext(c)
	res, err := h.resolve.ResolvePage(c.Request.Context(), c.Param("screenId"), reqCtx)
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("X-Layout-Version", strconv.Itoa(res.Trace.SnapshotVersion))
	if res.Trace.UsedFallback {
		c.Header("X-Layout-Fallback", "true")
	}
	ginmw.RespondSuccess(c, res.Page)
}

// getPage serves GET /api/v1/pages/:slug (supports ?include_trace=true for resolution trace debugging).
func (h *runtimeHandler) getPage(c *gin.Context) {
	reqCtx := extractRequestContext(c)
	res, err := h.resolve.ResolvePage(c.Request.Context(), c.Param("slug"), reqCtx)
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("X-Layout-Version", strconv.Itoa(res.Trace.SnapshotVersion))
	if res.Trace.UsedFallback {
		c.Header("X-Layout-Fallback", "true")
	}
	if strings.EqualFold(c.Query("include_trace"), "true") {
		ginmw.RespondSuccess(c, res)
		return
	}
	ginmw.RespondSuccess(c, res.Page)
}

func (h *runtimeHandler) getSplash(c *gin.Context) {
	cfg, err := h.resolve.ResolveSplash(c.Request.Context(), extractRequestContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, cfg)
}

func (h *runtimeHandler) getTokenCatalog(c *gin.Context) {
	cat, err := h.resolve.ResolveTokenCatalog(c.Request.Context(), extractRequestContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, cat)
}

func (h *runtimeHandler) executeAction(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid action payload", err.Error())
		return
	}
	var input usecase.ActionExecutionInput
	if err := json.Unmarshal(rawBody, &input); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid action JSON", err.Error())
		return
	}
	if input.Trigger.Type == "" {
		var rawMap map[string]any
		if json.Unmarshal(rawBody, &rawMap) == nil {
			if formID, ok := rawMap["form_id"].(string); ok && strings.TrimSpace(formID) != "" {
				input = usecase.ActionExecutionInput{
					Trigger: actions.NewCallAPIAction("submit_"+strings.TrimSpace(formID), actions.CallAPIPayload{
						Endpoint: "/api/v1/actions/execute",
						Method:   http.MethodPost,
						Body:     rawMap,
					}),
					ScreenID: "form-demo",
				}
			}
		}
	}
	res, err := h.actions.Execute(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, res)
}

func (h *runtimeHandler) placeExpressOrder(c *gin.Context) {
	var body map[string]any
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
			ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid express order payload", err.Error())
			return
		}
	}
	res, err := h.actions.Execute(c.Request.Context(), usecase.ActionExecutionInput{
		Trigger: actions.NewCallAPIAction("place_express_order", actions.CallAPIPayload{
			Endpoint: "/api/v1/orders/express",
			Method:   http.MethodPost,
			Body:     body,
		}),
		ScreenID: "store-home",
	})
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, res)
}

func extractTenantSlug(c *gin.Context) string {
	for _, v := range []string{c.Query("tenant"), c.GetHeader("X-Tenant-ID")} {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return usecase.DefaultTenantSlug
}

func extractRequestContext(c *gin.Context) domain.RequestContext {
	pick := func(queryKey, headerKey, fallback string) string {
		if v := strings.TrimSpace(c.Query(queryKey)); v != "" {
			return v
		}
		if v := strings.TrimSpace(c.GetHeader(headerKey)); v != "" {
			return v
		}
		return fallback
	}

	now := time.Now().UTC()
	if rawAt := strings.TrimSpace(c.Query("at")); rawAt != "" {
		if parsed, err := time.Parse(time.RFC3339, rawAt); err == nil {
			now = parsed.UTC()
		}
	}

	darkMode := strings.EqualFold(pick("dark_mode", "X-Dark-Mode", "false"), "true")
	attrs := make(map[string]string)
	for k, vals := range c.Request.URL.Query() {
		if strings.HasPrefix(k, "attr_") && len(vals) > 0 {
			attrs[strings.TrimPrefix(k, "attr_")] = vals[0]
		}
	}

	return domain.RequestContext{
		TenantSlug:  extractTenantSlug(c),
		Platform:    strings.ToLower(pick("platform", "X-Platform", "desktop")),
		AppVersion:  pick("app_version", "X-App-Version", "1.0.0"),
		City:        pick("city", "X-City", "Bengaluru"),
		Zone:        pick("zone", "X-Zone", "koramangala-01"),
		Segment:     strings.ToLower(pick("segment", "X-User-Segment", "standard")),
		UserID:      pick("user_id", "X-User-ID", "guest"),
		WindowClass: strings.ToLower(pick("window_class", "X-Window-Class", "compact")),
		DarkMode:    darkMode,
		Attributes:  attrs,
		Now:         now,
	}
}

func respondError(c *gin.Context, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		if appErr.Details != nil {
			if raw, marshalErr := json.Marshal(appErr.Details); marshalErr == nil {
				ginmw.RespondErrorWithDetail(c, appErr.GetStatus(), appErr.GetCode(), appErr.Message, string(raw))
				return
			}
		}
		ginmw.RespondError(c, appErr.GetStatus(), appErr.GetCode(), appErr.Message)
		return
	}
	ginmw.RespondError(c, http.StatusInternalServerError, string(apperror.CodeInternal), "internal server error")
}
