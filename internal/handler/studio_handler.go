package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/abeta-dev/go-libs/apperror"
	"github.com/abeta-dev/go-libs/ginmw"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/internal/usecase"
	"github.com/gin-gonic/gin"
)

type studioHandler struct {
	studio  *usecase.StudioService
	resolve *usecase.ResolveService
}

func (h *studioHandler) actor(c *gin.Context) string {
	if v := strings.TrimSpace(c.GetHeader("X-Studio-Actor")); v != "" {
		return v
	}
	return "studio-admin"
}

func (h *studioHandler) listTenants(c *gin.Context) {
	items, err := h.studio.ListTenants(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"tenants": items})
}

func (h *studioHandler) upsertTenant(c *gin.Context) {
	var body domain.Tenant
	if err := c.ShouldBindJSON(&body); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid tenant payload", err.Error())
		return
	}
	saved, err := h.studio.UpsertTenant(c.Request.Context(), body, h.actor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, saved)
}

func (h *studioHandler) listTemplates(c *gin.Context) {
	items, err := h.studio.ListTemplates(c.Request.Context(), extractTenantSlug(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"templates": items})
}

func (h *studioHandler) listWidgets(c *gin.Context) {
	items, err := h.studio.ListWidgets(c.Request.Context(), extractTenantSlug(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"widgets": items})
}

func (h *studioHandler) upsertWidget(c *gin.Context) {
	var body domain.Widget
	if err := c.ShouldBindJSON(&body); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid widget payload", err.Error())
		return
	}
	saved, err := h.studio.UpsertWidget(c.Request.Context(), extractTenantSlug(c), body, h.actor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, saved)
}

func (h *studioHandler) deleteWidget(c *gin.Context) {
	if err := h.studio.DeleteWidget(c.Request.Context(), extractTenantSlug(c), c.Param("id"), h.actor(c)); err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"deleted": c.Param("id")})
}

func (h *studioHandler) listLayouts(c *gin.Context) {
	items, err := h.studio.ListLayouts(c.Request.Context(), extractTenantSlug(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"layouts": items})
}

func (h *studioHandler) getLayout(c *gin.Context) {
	item, err := h.studio.GetLayout(c.Request.Context(), extractTenantSlug(c), c.Param("slug"))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, item)
}

func (h *studioHandler) upsertLayout(c *gin.Context) {
	var body domain.Layout
	if err := c.ShouldBindJSON(&body); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid layout payload", err.Error())
		return
	}
	saved, err := h.studio.UpsertLayout(c.Request.Context(), extractTenantSlug(c), c.Param("slug"), body, h.actor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, saved)
}

func (h *studioHandler) validateLayout(c *gin.Context) {
	issues, err := h.studio.ValidateLayout(c.Request.Context(), extractTenantSlug(c), c.Param("slug"))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"valid": len(issues) == 0, "issues": issues})
}

type publishRequest struct {
	Actor   string `json:"actor"`
	Comment string `json:"comment"`
}

func (h *studioHandler) publishLayout(c *gin.Context) {
	var req publishRequest
	_ = c.ShouldBindJSON(&req)
	actor := req.Actor
	if strings.TrimSpace(actor) == "" {
		actor = h.actor(c)
	}
	res, err := h.studio.PublishLayout(c.Request.Context(), extractTenantSlug(c), c.Param("slug"), actor, req.Comment)
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, res)
}

func (h *studioHandler) listVersions(c *gin.Context) {
	snaps, err := h.studio.ListSnapshots(c.Request.Context(), extractTenantSlug(c), c.Param("slug"))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"snapshots": snaps})
}

type rollbackRequest struct {
	Version int    `json:"version"`
	Actor   string `json:"actor"`
	Comment string `json:"comment"`
}

func (h *studioHandler) rollbackLayout(c *gin.Context) {
	var req rollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid rollback payload", err.Error())
		return
	}
	actor := req.Actor
	if strings.TrimSpace(actor) == "" {
		actor = h.actor(c)
	}
	res, err := h.studio.RollbackLayout(c.Request.Context(), extractTenantSlug(c), c.Param("slug"), req.Version, actor, req.Comment)
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, res)
}

func (h *studioHandler) promoteFallback(c *gin.Context) {
	var req publishRequest
	_ = c.ShouldBindJSON(&req)
	actor := req.Actor
	if strings.TrimSpace(actor) == "" {
		actor = h.actor(c)
	}
	fb, err := h.studio.PromoteFallback(c.Request.Context(), extractTenantSlug(c), c.Param("slug"), actor)
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, fb)
}

type previewContextPayload struct {
	Platform    string            `json:"platform"`
	AppVersion  string            `json:"app_version"`
	City        string            `json:"city"`
	Zone        string            `json:"zone"`
	Segment     string            `json:"segment"`
	UserID      string            `json:"user_id"`
	WindowClass string            `json:"window_class"`
	DarkMode    bool              `json:"dark_mode"`
	At          string            `json:"at"`
	Attributes  map[string]string `json:"attributes"`
}

type previewRequest struct {
	Draft   bool                  `json:"draft"`
	Context previewContextPayload `json:"context"`
}

func (h *studioHandler) previewLayout(c *gin.Context) {
	var req previewRequest
	_ = c.ShouldBindJSON(&req)
	reqCtx := extractRequestContext(c)
	if v := strings.TrimSpace(req.Context.Platform); v != "" {
		reqCtx.Platform = strings.ToLower(v)
	}
	if v := strings.TrimSpace(req.Context.AppVersion); v != "" {
		reqCtx.AppVersion = v
	}
	if v := strings.TrimSpace(req.Context.City); v != "" {
		reqCtx.City = v
	}
	if v := strings.TrimSpace(req.Context.Zone); v != "" {
		reqCtx.Zone = v
	}
	if v := strings.TrimSpace(req.Context.Segment); v != "" {
		reqCtx.Segment = strings.ToLower(v)
	}
	if v := strings.TrimSpace(req.Context.UserID); v != "" {
		reqCtx.UserID = v
	}
	if v := strings.TrimSpace(req.Context.WindowClass); v != "" {
		reqCtx.WindowClass = strings.ToLower(v)
	}
	reqCtx.DarkMode = req.Context.DarkMode
	if len(req.Context.Attributes) > 0 {
		reqCtx.Attributes = req.Context.Attributes
	}
	if v := strings.TrimSpace(req.Context.At); v != "" {
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			reqCtx.Now = parsed.UTC()
		}
	}
	res, err := h.resolve.PreviewPage(c.Request.Context(), c.Param("slug"), reqCtx, req.Draft)
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, res)
}

func (h *studioHandler) listRules(c *gin.Context) {
	items, err := h.studio.ListRules(c.Request.Context(), extractTenantSlug(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"rules": items})
}

func (h *studioHandler) upsertRule(c *gin.Context) {
	var body domain.TargetingRule
	if err := c.ShouldBindJSON(&body); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid rule payload", err.Error())
		return
	}
	saved, err := h.studio.UpsertRule(c.Request.Context(), extractTenantSlug(c), body, h.actor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, saved)
}

func (h *studioHandler) deleteRule(c *gin.Context) {
	if err := h.studio.DeleteRule(c.Request.Context(), extractTenantSlug(c), c.Param("id"), h.actor(c)); err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"deleted": c.Param("id")})
}

func (h *studioHandler) listSlotting(c *gin.Context) {
	items, err := h.studio.ListSlottingRules(c.Request.Context(), extractTenantSlug(c), c.Query("layout_slug"))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"slotting_rules": items})
}

func (h *studioHandler) upsertSlotting(c *gin.Context) {
	var body domain.SlottingRule
	if err := c.ShouldBindJSON(&body); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid slotting rule payload", err.Error())
		return
	}
	saved, err := h.studio.UpsertSlottingRule(c.Request.Context(), extractTenantSlug(c), body, h.actor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, saved)
}

func (h *studioHandler) deleteSlotting(c *gin.Context) {
	if err := h.studio.DeleteSlottingRule(c.Request.Context(), extractTenantSlug(c), c.Param("id"), h.actor(c)); err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"deleted": c.Param("id")})
}

func (h *studioHandler) listSplash(c *gin.Context) {
	items, err := h.studio.ListSplashConfigs(c.Request.Context(), extractTenantSlug(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"splash_configs": items})
}

func (h *studioHandler) upsertSplash(c *gin.Context) {
	var body domain.SplashConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid splash config payload", err.Error())
		return
	}
	saved, err := h.studio.UpsertSplashConfig(c.Request.Context(), extractTenantSlug(c), body, h.actor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, saved)
}

func (h *studioHandler) listTokens(c *gin.Context) {
	items, err := h.studio.ListTokenOverrides(c.Request.Context(), extractTenantSlug(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"token_overrides": items})
}

func (h *studioHandler) upsertTokens(c *gin.Context) {
	var body domain.TokenOverride
	if err := c.ShouldBindJSON(&body); err != nil {
		ginmw.RespondErrorWithDetail(c, http.StatusBadRequest, string(apperror.CodeBadRequest), "invalid token override payload", err.Error())
		return
	}
	saved, err := h.studio.UpsertTokenOverride(c.Request.Context(), extractTenantSlug(c), body, h.actor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, saved)
}

func (h *studioHandler) listChangeLog(c *gin.Context) {
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	entries, err := h.studio.ListChangeLogs(c.Request.Context(), extractTenantSlug(c), limit)
	if err != nil {
		respondError(c, err)
		return
	}
	ginmw.RespondSuccess(c, gin.H{"entries": entries})
}
