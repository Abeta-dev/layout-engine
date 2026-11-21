(() => {
  'use strict';

  const state = {
    tenant: 'abeta-fresh',
    tenants: [],
    templates: [],
    widgets: [],
    layouts: [],
    rules: [],
    slotting: [],
    splash: [],
    tokens: [],
    changelog: [],
    activeSlug: 'store-home',
    activeLayout: null,
    dirty: false,
    centerTab: 'composer',
    leftTab: 'pages',
    device: 'phone',
    previewDraft: true,
    previewResult: null,
    cartCounts: {},
    dragIndex: null,
    previewTimer: null,
  };

  const $ = (id) => document.getElementById(id);

  function esc(val) {
    return String(val == null ? '' : val)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function showLoader(on) {
    const el = $('global-loader');
    if (el) el.hidden = !on;
  }

  function toast(msg, kind = 'info') {
    const host = $('toast-container');
    if (!host) return;
    const node = document.createElement('div');
    node.className = `toast toast-${kind}`;
    node.textContent = msg;
    host.appendChild(node);
    setTimeout(() => {
      node.style.opacity = '0';
      setTimeout(() => node.remove(), 220);
    }, 3200);
  }

  async function api(method, path, body) {
    showLoader(true);
    try {
      const headers = { 'X-Tenant-ID': state.tenant, 'X-Studio-Actor': 'studio-admin' };
      const init = { method, headers };
      if (body !== undefined) {
        headers['Content-Type'] = 'application/json';
        init.body = JSON.stringify(body);
      }
      const res = await fetch(path, init);
      const raw = await res.text();
      let env = null;
      try { env = raw ? JSON.parse(raw) : null; } catch { env = null; }
      if (!res.ok || (env && env.success === false)) {
        const errMsg = env?.error?.message || `HTTP ${res.status}`;
        const detail = env?.error?.detail ? `: ${env.error.detail}` : '';
        throw new Error(errMsg + detail);
      }
      return env && 'data' in env ? env.data : env;
    } finally {
      showLoader(false);
    }
  }

  function snippetBadgeClass(type) {
    switch (type) {
      case 'header_v1': case 'banner_v1': return 'badge-emerald';
      case 'product_card': case 'metric_card': return 'badge-amber';
      case 'grid': case 'carousel': case 'tabs': case 'accordion': return 'badge-sky';
      case 'form_v1': case 'action_row': return 'badge-violet';
      case 'card_v1': return 'badge-teal';
      case 'empty_state': return 'badge-rose';
      default: return 'badge-neutral';
    }
  }

  function widgetById(id) {
    return state.widgets.find((w) => w.id === id) || null;
  }

  function ruleById(id) {
    return state.rules.find((r) => r.id === id || r.key === id) || null;
  }

  function setDirty(flag) {
    state.dirty = flag;
    const ind = $('unsaved-indicator');
    if (ind) ind.hidden = !flag;
    const bar = $('preview-dirty-bar');
    if (bar) bar.hidden = !(flag && state.previewDraft);
  }

  function openModal(id) {
    const el = $(id);
    if (el) el.hidden = false;
  }

  function closeModal(id) {
    const el = $(id);
    if (el) el.hidden = true;
  }

  function currentContext() {
    const ctx = {
      platform: $('ctx-platform')?.value || 'ios',
      app_version: $('ctx-app-version')?.value?.trim() || '1.2.0',
      city: $('ctx-city')?.value?.trim() || 'Bengaluru',
      zone: $('ctx-zone')?.value?.trim() || 'koramangala-01',
      segment: $('ctx-segment')?.value || 'standard',
      window_class: $('ctx-window-class')?.value || 'compact',
      user_id: $('ctx-user-id')?.value?.trim() || 'user-42',
      dark_mode: Boolean($('ctx-dark-mode')?.checked),
    };
    if ($('ctx-hour-enabled')?.checked) {
      const hr = String(Number($('ctx-utc-hour')?.value || 0)).padStart(2, '0');
      const now = new Date();
      const yyyy = now.getUTCFullYear();
      const mm = String(now.getUTCMonth() + 1).padStart(2, '0');
      const dd = String(now.getUTCDate()).padStart(2, '0');
      ctx.at = `${yyyy}-${mm}-${dd}T${hr}:30:00Z`;
    }
    return ctx;
  }

  // ── Data Bootstrap ──────────────────────────────────────────────────────────

  async function loadAll() {
    try {
      const tData = await api('GET', '/api/v1/studio/tenants');
      state.tenants = tData?.tenants || [];
      renderTenants();

      const [tpls, wids, lays, rules, slots, chg, spl, tok] = await Promise.all([
        api('GET', '/api/v1/studio/templates'),
        api('GET', '/api/v1/studio/widgets'),
        api('GET', '/api/v1/studio/layouts'),
        api('GET', '/api/v1/studio/rules'),
        api('GET', '/api/v1/studio/slotting'),
        api('GET', '/api/v1/studio/changelog?limit=50'),
        api('GET', '/api/v1/studio/splash'),
        api('GET', '/api/v1/studio/tokens'),
      ]);

      state.templates = tpls?.templates || [];
      state.widgets = wids?.widgets || [];
      state.layouts = lays?.layouts || [];
      state.rules = rules?.rules || [];
      state.slotting = slots?.slotting_rules || [];
      state.changelog = chg?.entries || [];
      state.splash = spl?.splash_configs || [];
      state.tokens = tok?.token_overrides || [];

      if (!state.layouts.some((l) => l.slug === state.activeSlug) && state.layouts.length > 0) {
        state.activeSlug = state.layouts[0].slug;
      }
      if (state.activeSlug) {
        const active = state.layouts.find((l) => l.slug === state.activeSlug);
        state.activeLayout = active ? structuredClone(active) : null;
      }
      setDirty(false);
      populateSelects();
      renderAll();
      if (state.activeLayout) {
        await refreshPreview();
      }
    } catch (err) {
      toast(`Failed to load studio workspace: ${err.message}`, 'error');
    }
  }

  function renderTenants() {
    const sel = $('tenant-select');
    if (!sel) return;
    sel.innerHTML = state.tenants
      .map((t) => `<option value="${esc(t.slug)}" ${t.slug === state.tenant ? 'selected' : ''}>${esc(t.name)} (${esc(t.slug)})</option>`)
      .join('');
  }

  function populateSelects() {
    const typeFilter = $('widget-type-filter');
    const types = [...new Set(state.templates.map((t) => t.snippet_type).concat(state.widgets.map((w) => w.snippet_type)))].sort();
    if (typeFilter) {
      const cur = typeFilter.value;
      typeFilter.innerHTML = `<option value="">All 14 snippet types</option>` +
        types.map((t) => `<option value="${esc(t)}" ${t === cur ? 'selected' : ''}>${esc(t)}</option>`).join('');
    }
    const wfTpl = $('wf-template');
    if (wfTpl) {
      wfTpl.innerHTML = `<option value="">Custom (manual props)</option>` +
        state.templates.map((t) => `<option value="${esc(t.key)}">${esc(t.name)} (${esc(t.snippet_type)})</option>`).join('');
    }
    const wfType = $('wf-snippet-type');
    if (wfType) {
      wfType.innerHTML = types.map((t) => `<option value="${esc(t)}">${esc(t)}</option>`).join('');
    }

    const ruleOpts = `<option value="">No targeting rule (always match)</option>` +
      state.rules.map((r) => `<option value="${esc(r.id)}">${esc(r.name)} (${esc(r.key)})</option>`).join('');
    ['sf-rule-id', 'ns-rule-id', 'sp-rule-id', 'tk-rule-id'].forEach((id) => {
      const el = $(id);
      if (el) el.innerHTML = ruleOpts;
    });

    const widSel = $('ns-widget-id');
    if (widSel) {
      widSel.innerHTML = state.widgets.map((w) => `<option value="${esc(w.id)}">${esc(w.name)} [${esc(w.snippet_type)}]</option>`).join('');
    }
  }

  function renderAll() {
    renderCounts();
    renderHeader();
    renderPagesList();
    renderWidgetsList();
    renderRulesList();
    renderSlottingList();
    renderAuditList();
    renderComposer();
    renderBrandConfig();
  }

  function renderCounts() {
    if ($('count-pages')) $('count-pages').textContent = String(state.layouts.length);
    if ($('count-widgets')) $('count-widgets').textContent = String(state.widgets.length);
    if ($('count-rules')) $('count-rules').textContent = String(state.rules.length);
    if ($('count-slotting')) $('count-slotting').textContent = String(state.slotting.length);
    if ($('count-audit')) $('count-audit').textContent = String(state.changelog.length);
  }

  function renderHeader() {
    const l = state.activeLayout;
    if ($('active-layout-slug')) $('active-layout-slug').textContent = l ? l.slug : '—';
    const verEl = $('active-layout-version');
    if (verEl) {
      if (!l) {
        verEl.textContent = 'No layout';
        verEl.className = 'status-badge badge-neutral';
      } else if (l.published_version > 0) {
        verEl.textContent = `v${l.published_version} Published • Draft v${l.draft_version}`;
        verEl.className = 'status-badge badge-emerald';
      } else {
        verEl.textContent = `Draft v${l.draft_version} (Unpublished)`;
        verEl.className = 'status-badge badge-amber';
      }
    }
    const meta = $('stage-layout-meta');
    if (meta && l) {
      const mode = l.page_config?.adaptive_layout?.layout_mode || 'single_column';
      const cart = l.page_config?.floating_cart ? ' • Floating Cart' : '';
      meta.textContent = `${l.title} • ${mode}${cart}`;
    }
    if ($('slotting-layout-label')) $('slotting-layout-label').textContent = l ? l.slug : 'all';
    if ($('ns-layout-slug') && l && !$('ns-layout-slug').value) $('ns-layout-slug').value = l.slug;
  }

  // ── Left Rail Renderers ─────────────────────────────────────────────────────

  function renderPagesList() {
    const host = $('page-list');
    if (!host) return;
    if (state.layouts.length === 0) {
      host.innerHTML = `<div class="empty-state">No dynamic pages yet. Create one above.</div>`;
      return;
    }
    host.innerHTML = state.layouts.map((l) => {
      const sel = l.slug === state.activeSlug ? 'selected' : '';
      const badge = l.published_version > 0
        ? `<span class="status-badge badge-emerald">v${l.published_version}</span>`
        : `<span class="status-badge badge-amber">draft</span>`;
      return `
        <div class="resource-card clickable ${sel}" data-select-layout="${esc(l.slug)}">
          <div class="card-row">
            <span class="card-title">${esc(l.slug)}</span>
            ${badge}
          </div>
          <div class="card-sub">${esc(l.title)} • ${(l.slots || []).length} slots</div>
        </div>`;
    }).join('');
  }

  function renderWidgetsList() {
    const host = $('widget-list');
    if (!host) return;
    const q = ($('widget-search')?.value || '').trim().toLowerCase();
    const tFilter = $('widget-type-filter')?.value || '';
    const filtered = state.widgets.filter((w) => {
      if (tFilter && w.snippet_type !== tFilter) return false;
      if (!q) return true;
      return `${w.name} ${w.slug} ${w.id} ${w.snippet_type}`.toLowerCase().includes(q);
    });
    if (filtered.length === 0) {
      host.innerHTML = `<div class="empty-state">No widgets match your filter.</div>`;
      return;
    }
    host.innerHTML = filtered.map((w) => `
      <div class="resource-card">
        <div class="card-row">
          <span class="card-title">${esc(w.name)}</span>
          <span class="status-badge ${snippetBadgeClass(w.snippet_type)}">${esc(w.snippet_type)}</span>
        </div>
        <div class="card-sub mono">${esc(w.id)}</div>
        <div class="card-actions">
          <button type="button" class="chip-btn" data-add-widget="${esc(w.id)}">+ Add to Layout</button>
          <button type="button" class="chip-btn" data-edit-widget="${esc(w.id)}">Edit</button>
          <button type="button" class="chip-btn danger" data-del-widget="${esc(w.id)}">Del</button>
        </div>
      </div>`).join('');
  }

  function renderRulesList() {
    const host = $('rule-list');
    if (!host) return;
    if (state.rules.length === 0) {
      host.innerHTML = `<div class="empty-state">No targeting rules yet.</div>`;
      return;
    }
    host.innerHTML = state.rules.map((r) => {
      const conds = (r.conditions || []).map((c) => `${c.field} ${c.operator} ${JSON.stringify(c.value)}`).join(` ${r.match_mode || 'all'} `);
      const sched = r.schedule && r.schedule.start_hour_utc != null
        ? `UTC ${r.schedule.start_hour_utc}:00–${r.schedule.end_hour_utc}:00`
        : '';
      return `
        <div class="resource-card">
          <div class="card-row">
            <span class="card-title">${esc(r.name)}</span>
            <span class="status-badge ${r.enabled ? 'badge-emerald' : 'badge-neutral'}">P${r.priority}</span>
          </div>
          <div class="card-sub mono">${esc(r.key)}</div>
          <div class="card-sub">${esc(conds || sched || 'Always match')}</div>
          ${sched && conds ? `<div class="card-sub">${esc(sched)}</div>` : ''}
          <div class="card-actions">
            <button type="button" class="chip-btn danger" data-del-rule="${esc(r.id)}">Delete</button>
          </div>
        </div>`;
    }).join('');
  }

  function renderSlottingList() {
    const host = $('slotting-list');
    if (!host) return;
    const filtered = state.slotting.filter((s) => !state.activeSlug || s.layout_slug === state.activeSlug);
    if (filtered.length === 0) {
      host.innerHTML = `<div class="empty-state">No dynamic slotting rules for ${esc(state.activeSlug || 'this layout')}.</div>`;
      return;
    }
    host.innerHTML = filtered.map((s) => {
      const w = widgetById(s.widget_id);
      const r = ruleById(s.rule_id);
      return `
        <div class="resource-card">
          <div class="card-row">
            <span class="card-title">${esc(s.name)}</span>
            <span class="status-badge badge-amber">${esc(s.action)} #${s.target_position} • P${s.priority}</span>
          </div>
          <div class="card-sub">Widget: <strong>${esc(w ? w.name : s.widget_id)}</strong></div>
          <div class="card-sub">Rule: ${esc(r ? r.name : (s.rule_id || 'Always'))}</div>
          <div class="card-actions">
            <button type="button" class="chip-btn danger" data-del-slotting="${esc(s.id)}">Delete</button>
          </div>
        </div>`;
    }).join('');
  }

  function renderAuditList() {
    const host = $('audit-list');
    if (!host) return;
    if (state.changelog.length === 0) {
      host.innerHTML = `<div class="empty-state">No changelog entries recorded yet.</div>`;
      return;
    }
    host.innerHTML = state.changelog.map((e) => {
      const diffs = (e.diff || []).slice(0, 5).map((d) =>
        `<div class="diff-item diff-${esc(d.op)}"><span class="mono">${esc(d.op)}</span> ${esc(d.path)}</div>`
      ).join('');
      const ts = e.created_at ? new Date(e.created_at).toLocaleTimeString() : '';
      return `
        <div class="audit-entry">
          <div class="card-row">
            <span class="status-badge badge-sky">${esc(e.action)}</span>
            <span class="card-sub mono">${esc(ts)}</span>
          </div>
          <div class="card-title">${esc(e.summary)}</div>
          <div class="card-sub">${esc(e.entity_type)}/${esc(e.entity_id)} • by ${esc(e.actor)}</div>
          ${diffs ? `<div class="diff-list">${diffs}</div>` : ''}
        </div>`;
    }).join('');
  }

  // ── Center Canvas — Slot Composer ───────────────────────────────────────────

  function renderComposer() {
    const host = $('slot-list');
    const l = state.activeLayout;
    if (!host || !l) return;
    const slots = l.slots || [];
    if ($('composer-slot-count')) $('composer-slot-count').textContent = `${slots.length} slot${slots.length === 1 ? '' : 's'}`;
    if (slots.length === 0) {
      host.innerHTML = `<div class="empty-state">This dynamic layout has no slots yet. Add widgets from the left palette or click "+ Add Widget".</div>`;
      return;
    }
    host.innerHTML = slots.map((slot, idx) => {
      const w = widgetById(slot.widget_id);
      const r = slot.rule_id ? ruleById(slot.rule_id) : null;
      const cls = [
        'slot-card',
        !slot.enabled ? 'slot-disabled' : '',
        slot.pinned ? 'slot-pinned' : '',
      ].filter(Boolean).join(' ');
      const titlePreview = w?.props?.title?.text || w?.props?.label?.text || w?.props?.headline || w?.slug || '';
      const chips = [];
      if (r) chips.push(`<span class="mini-chip badge-sky">Rule: ${esc(r.name)}</span>`);
      if (slot.schedule && slot.schedule.start_hour_utc != null) {
        chips.push(`<span class="mini-chip badge-amber">UTC ${slot.schedule.start_hour_utc}–${slot.schedule.end_hour_utc}h</span>`);
      }
      if (slot.pinned) chips.push(`<span class="mini-chip badge-violet">Pinned</span>`);
      if (slot.overrides && Object.keys(slot.overrides).length > 0) {
        chips.push(`<span class="mini-chip badge-teal">Overrides</span>`);
      }
      return `
        <div class="${cls}" draggable="true" data-slot-idx="${idx}">
          <div class="slot-pos-col">
            <span class="slot-pos-pill">#${idx}</span>
          </div>
          <div class="slot-main">
            <div class="slot-head">
              <span class="slot-widget-name">${esc(w ? w.name : slot.widget_id)}</span>
              <span class="status-badge ${snippetBadgeClass(w?.snippet_type)}">${esc(w?.snippet_type || 'unknown')}</span>
              <span class="card-sub mono">${esc(slot.slot_id)}</span>
            </div>
            ${titlePreview ? `<div class="slot-preview-line">${esc(titlePreview)}</div>` : ''}
            ${chips.length ? `<div class="chip-wrap">${chips.join('')}</div>` : ''}
          </div>
          <div class="slot-controls">
            <button type="button" class="btn-icon-xs" title="Move Up" data-slot-up="${idx}" ${idx === 0 ? 'disabled' : ''}>↑</button>
            <button type="button" class="btn-icon-xs" title="Move Down" data-slot-down="${idx}" ${idx === slots.length - 1 ? 'disabled' : ''}>↓</button>
            <button type="button" class="btn-icon-xs ${slot.pinned ? 'active-pin' : ''}" title="Pin Slot" data-slot-pin="${idx}">📌</button>
            <button type="button" class="btn-icon-xs ${slot.enabled ? 'active-toggle' : ''}" title="Toggle Enabled" data-slot-toggle="${idx}">${slot.enabled ? 'ON' : 'OFF'}</button>
            <button type="button" class="btn-icon-xs" title="Edit Rule & Overrides" data-slot-edit="${idx}">⚙</button>
            <button type="button" class="btn-icon-xs danger" title="Remove Slot" data-slot-remove="${idx}">✕</button>
          </div>
        </div>`;
    }).join('');
    bindSlotDragDrop();
  }

  function bindSlotDragDrop() {
    document.querySelectorAll('.slot-card').forEach((card) => {
      card.addEventListener('dragstart', () => {
        state.dragIndex = Number(card.dataset.slotIdx);
        card.classList.add('dragging');
      });
      card.addEventListener('dragend', () => {
        card.classList.remove('dragging');
        document.querySelectorAll('.slot-card.drag-over').forEach((c) => c.classList.remove('drag-over'));
      });
      card.addEventListener('dragover', (e) => {
        e.preventDefault();
        card.classList.add('drag-over');
      });
      card.addEventListener('dragleave', () => card.classList.remove('drag-over'));
      card.addEventListener('drop', (e) => {
        e.preventDefault();
        card.classList.remove('drag-over');
        const targetIdx = Number(card.dataset.slotIdx);
        if (state.dragIndex == null || state.dragIndex === targetIdx || !state.activeLayout) return;
        const arr = state.activeLayout.slots;
        const [moved] = arr.splice(state.dragIndex, 1);
        arr.splice(targetIdx, 0, moved);
        arr.forEach((s, i) => { s.position = i; });
        state.dragIndex = null;
        setDirty(true);
        renderComposer();
      });
    });
  }

  // ── Center Canvas — Live Device Preview (All 14 KMP v1 Snippets) ────────────

  async function refreshPreview() {
    if (!state.activeSlug) return;
    const spinner = $('preview-spinner');
    if (spinner) spinner.hidden = false;
    try {
      const res = await api('POST', `/api/v1/studio/layouts/${encodeURIComponent(state.activeSlug)}/preview`, {
        draft: state.previewDraft,
        context: currentContext(),
      });
      state.previewResult = res;
      renderTrace(res.trace, res.page);
      renderDevicePreview(res.page);
      renderJsonView(res);
    } catch (err) {
      toast(`Preview failed: ${err.message}`, 'warning');
    } finally {
      if (spinner) spinner.hidden = true;
    }
  }

  function schedulePreview() {
    clearTimeout(state.previewTimer);
    state.previewTimer = setTimeout(refreshPreview, 180);
  }

  function textOf(atom) {
    if (!atom) return '';
    if (typeof atom === 'string') return atom;
    return atom.text || '';
  }

  function renderSnippet(env) {
    const d = env.data || {};
    switch (env.snippet_type) {
      case 'header_v1':
        return `
          <div class="kmp-header-snip">
            <div class="kmp-header-left">
              ${d.avatar?.initials ? `<span class="kmp-avatar">${esc(d.avatar.initials)}</span>` : ''}
              <div>
                <div class="kmp-h-title">${esc(textOf(d.title))}</div>
                ${d.subtitle ? `<div class="kmp-h-sub">${esc(textOf(d.subtitle))}</div>` : ''}
              </div>
            </div>
            ${d.trailing_tag ? `<span class="kmp-tag">${esc(textOf(d.trailing_tag.label || d.trailing_tag))}</span>` : ''}
          </div>`;

      case 'banner_v1':
        return `
          <div class="kmp-banner">
            <div class="kmp-banner-row">
              <div>
                ${d.tag ? `<span class="kmp-tag">${esc(d.tag.label || textOf(d.tag))}</span>` : ''}
                <div class="kmp-banner-title">${esc(textOf(d.title))}</div>
                ${d.body ? `<div class="kmp-banner-body">${esc(textOf(d.body))}</div>` : ''}
                ${d.cta ? `<button type="button" class="kmp-btn kmp-btn-primary" style="margin-top:8px">${esc(d.cta.label || 'Explore')}</button>` : ''}
              </div>
              ${d.image?.url ? `<img class="kmp-banner-img" src="${esc(d.image.url)}" alt="${esc(d.image.content_description || '')}" />` : ''}
            </div>
          </div>`;

      case 'metric_card': {
        const pct = Math.round(((d.progress?.value ?? 0) * 100));
        return `
          <div class="kmp-metric-card">
            <div class="card-row">
              <span class="card-sub">${esc(textOf(d.label))}</span>
              ${d.delta ? `<span class="kmp-tag">${esc(d.delta.label || textOf(d.delta))}</span>` : ''}
            </div>
            <div class="kmp-metric-val">${esc(textOf(d.value))}</div>
            ${d.progress ? `<div class="kmp-progress-track"><div class="kmp-progress-fill" style="width:${pct}%"></div></div>` : ''}
          </div>`;
      }

      case 'product_card': {
        const pid = d.product_id || env.snippet_id;
        const qty = state.cartCounts[pid] ?? (d.stepper?.value || 0);
        const unit = d.unit_price_cents || 0;
        const stepperHtml = qty > 0
          ? `<div class="kmp-stepper has-qty">
               <button type="button" class="kmp-stepper-btn" data-cart-dec="${esc(pid)}" data-unit="${unit}">−</button>
               <span class="kmp-stepper-qty">${qty}</span>
               <button type="button" class="kmp-stepper-btn" data-cart-inc="${esc(pid)}" data-unit="${unit}">+</button>
             </div>`
          : `<div class="kmp-stepper">
               <button type="button" class="kmp-stepper-add" data-cart-inc="${esc(pid)}" data-unit="${unit}">ADD</button>
             </div>`;
        return `
          <div class="kmp-product-card">
            <div class="kmp-product-thumb">
              ${d.badge ? `<span class="kmp-product-badge">${esc(d.badge.label || textOf(d.badge))}</span>` : ''}
              ${d.image?.url ? `<img src="${esc(d.image.url)}" alt="" />` : `<span>🥬</span>`}
            </div>
            <div class="kmp-product-title">${esc(textOf(d.title))}</div>
            ${d.subtitle ? `<div class="kmp-product-sub">${esc(textOf(d.subtitle))}</div>` : ''}
            <div class="kmp-product-footer">
              <div>
                <span class="kmp-price">${esc(textOf(d.price))}</span>
                ${d.original_price ? `<span class="kmp-orig-price">${esc(textOf(d.original_price))}</span>` : ''}
              </div>
              ${stepperHtml}
            </div>
          </div>`;
      }

      case 'grid': {
        const cols = state.device === 'watch' ? 1 : (state.device === 'desktop' ? (d.columns_by_width?.expanded || d.columns || 3) : (d.columns_by_width?.compact || d.columns || 2));
        return `
          <div>
            ${d.title ? `<div class="kmp-section-title">${esc(textOf(d.title))}</div>` : ''}
            <div class="kmp-grid" style="grid-template-columns:repeat(${cols},minmax(0,1fr))">
              ${(d.items || []).map(renderSnippet).join('')}
            </div>
          </div>`;
      }

      case 'carousel':
        return `
          <div>
            ${d.title ? `<div class="kmp-section-title">${esc(textOf(d.title))}</div>` : ''}
            <div class="kmp-carousel">
              ${(d.items || []).map((it) => `<div class="kmp-carousel-item">${renderSnippet(it)}</div>`).join('')}
            </div>
          </div>`;

      case 'tabs': {
        const tabs = d.tabs || [];
        const activeIdx = Math.min(d.selected_index || 0, Math.max(0, tabs.length - 1));
        const activeTab = tabs[activeIdx];
        return `
          <div>
            <div class="kmp-tabs-bar">
              ${tabs.map((t, i) => `<button type="button" class="kmp-tab-pill ${i === activeIdx ? 'active' : ''}">${esc(t.title || t.id)}</button>`).join('')}
            </div>
            <div class="kmp-tabs-body">${activeTab ? (activeTab.items || []).map(renderSnippet).join('') : ''}</div>
          </div>`;
      }

      case 'accordion':
        return `
          <div class="kmp-accordion">
            ${(d.items || []).map((it, i) => `
              <div>
                <div class="kmp-accordion-head"><span>${esc(textOf(it.title))}</span><span>${i === 0 ? '▾' : '▸'}</span></div>
                ${i === 0 ? `<div class="kmp-accordion-body">${esc(textOf(it.body))}${(it.content || []).map(renderSnippet).join('')}</div>` : ''}
              </div>`).join('')}
          </div>`;

      case 'card_v1':
        return `
          <div class="kmp-card">
            <div class="card-row">
              <div class="kmp-h-title">${esc(textOf(d.title))}</div>
              ${d.badge ? `<span class="kmp-tag">${esc(d.badge.label || textOf(d.badge))}</span>` : ''}
            </div>
            ${d.subtitle ? `<div class="kmp-h-sub">${esc(textOf(d.subtitle))}</div>` : ''}
            ${d.description ? `<div class="kmp-banner-body">${esc(textOf(d.description))}</div>` : ''}
            ${(d.actions || []).length ? `<div class="kmp-card-actions">${d.actions.map((a) => `<button type="button" class="kmp-btn kmp-btn-primary">${esc(a.label)}</button>`).join('')}</div>` : ''}
          </div>`;

      case 'action_row':
        return `
          <div class="kmp-action-row">
            <div>
              <div class="kmp-h-title">${esc(textOf(d.title))}</div>
              ${d.subtitle ? `<div class="kmp-h-sub">${esc(textOf(d.subtitle))}</div>` : ''}
            </div>
            <span>›</span>
          </div>`;

      case 'form_v1':
        return `
          <div class="kmp-form">
            ${d.title ? `<div class="kmp-h-title">${esc(textOf(d.title))}</div>` : ''}
            ${(d.fields || []).map((f) => `<input type="text" placeholder="${esc(f.placeholder || f.label || f.field_id)}" />`).join('')}
            ${d.submit_button ? `<button type="button" class="kmp-btn kmp-btn-primary">${esc(d.submit_button.label || 'Submit')}</button>` : ''}
          </div>`;

      case 'empty_state':
        return `
          <div class="kmp-empty-state">
            <div class="kmp-h-title">${esc(textOf(d.title))}</div>
            ${d.message ? `<div class="kmp-h-sub">${esc(textOf(d.message))}</div>` : ''}
          </div>`;

      case 'divider':
        return `<hr class="kmp-divider" />`;

      case 'spacer':
        return `<div style="height:${Number(d.height_dp || 12)}px"></div>`;

      default:
        return `<div class="kmp-card"><span class="mono">${esc(env.snippet_type)}</span></div>`;
    }
  }

  function renderDevicePreview(page) {
    const screen = $('device-screen');
    const frame = $('device-frame');
    if (!screen || !frame || !page) return;

    frame.className = `device-frame device-${state.device}`;
    screen.className = `device-screen ${$('ctx-dark-mode')?.checked ? 'kmp-dark' : ''}`;

    const toolbar = page.toolbar ? `
      <div class="kmp-toolbar">
        <div class="kmp-toolbar-left">
          <div class="kmp-toolbar-title">${esc(textOf(page.toolbar.title) || page.title)}</div>
          ${page.toolbar.subtitle ? `<div class="kmp-toolbar-sub">${esc(textOf(page.toolbar.subtitle))}</div>` : ''}
        </div>
        <div class="kmp-toolbar-actions"><span class="kmp-icon-pill">v1</span></div>
      </div>` : '';

    const snippets = page.snippets || [];
    const splitEnabled = state.device === 'desktop' && page.adaptive_layout?.split_pane_enabled;
    const secondarySet = new Set(page.adaptive_layout?.secondary_snippet_ids || []);
    let bodyHtml = '';
    if (splitEnabled && secondarySet.size > 0) {
      const primary = snippets.filter((s) => !secondarySet.has(s.snippet_id));
      const secondary = snippets.filter((s) => secondarySet.has(s.snippet_id));
      bodyHtml = `
        <div class="kmp-split-pane">
          <div class="kmp-pane-col">${primary.map(renderSnippet).join('')}</div>
          <div class="kmp-pane-col kmp-pane-secondary">${secondary.map(renderSnippet).join('')}</div>
        </div>`;
    } else {
      bodyHtml = snippets.map(renderSnippet).join('');
    }

    const totalItems = Object.values(state.cartCounts).reduce((a, b) => a + b, 0);
    const showCart = Boolean(page.floating_cart);
    const cartHtml = showCart ? `
      <div class="kmp-floating-cart">
        <div class="kmp-cart-info">
          <span class="kmp-cart-title">${totalItems > 0 ? `${totalItems} item${totalItems > 1 ? 's' : ''} in cart` : esc(page.floating_cart.label || 'Express Cart')}</span>
          ${page.floating_cart.eta_badge ? `<span class="kmp-cart-eta">${esc(page.floating_cart.eta_badge.label || textOf(page.floating_cart.eta_badge))}</span>` : ''}
        </div>
        <button type="button" class="kmp-cart-cta" data-express-checkout="1">${esc(page.floating_cart.cta?.label || 'Checkout')}</button>
      </div>` : '';

    const bottomBar = page.bottom_bar?.primary_button ? `
      <div class="kmp-bottom-bar">
        <button type="button" class="kmp-btn kmp-btn-primary">${esc(page.bottom_bar.primary_button.label)}</button>
      </div>` : '';

    screen.innerHTML = `
      ${toolbar}
      <div class="kmp-scroll-body ${showCart ? 'has-floating-cart' : ''}">${bodyHtml || '<div class="kmp-empty-state">No snippets matched current rules.</div>'}</div>
      ${bottomBar}
      ${cartHtml}`;
  }

  // ── Right Inspector — Trace & JSON View ─────────────────────────────────────

  function renderTrace(trace, page) {
    if (!trace) return;
    if ($('trace-latency')) $('trace-latency').textContent = `${trace.duration_ms ?? 0} ms`;
    if ($('trace-snapshot-version')) $('trace-snapshot-version').textContent = trace.preview_draft ? `Draft v${trace.snapshot_version}` : `v${trace.snapshot_version}`;
    if ($('trace-fallback-badge')) {
      $('trace-fallback-badge').textContent = trace.used_fallback ? 'FALLBACK' : 'LIVE';
      $('trace-fallback-badge').className = `status-badge ${trace.used_fallback ? 'badge-amber' : 'badge-emerald'}`;
    }
    if ($('trace-snippet-count')) $('trace-snippet-count').textContent = String((page?.snippets || []).length);
    if ($('trace-resolved-at')) $('trace-resolved-at').textContent = trace.resolved_at ? new Date(trace.resolved_at).toLocaleTimeString() : '—';
    const fbReason = $('trace-fallback-reason');
    if (fbReason) {
      fbReason.hidden = !trace.fallback_reason;
      fbReason.textContent = trace.fallback_reason || '';
    }
    const chipList = (arr, cls, emptyText) => (arr && arr.length)
      ? arr.map((x) => `<span class="mini-chip ${cls}">${esc(x)}</span>`).join('')
      : `<span class="card-sub">${emptyText}</span>`;
    if ($('trace-matched-rules')) $('trace-matched-rules').innerHTML = chipList(trace.matched_rule_ids, 'badge-emerald', 'None matched');
    if ($('trace-applied-slotting')) $('trace-applied-slotting').innerHTML = chipList(trace.applied_slotting_ids, 'badge-amber', 'None injected');
    if ($('trace-skipped-slots')) $('trace-skipped-slots').innerHTML = chipList(trace.skipped_slot_ids, 'badge-neutral', 'None skipped');
  }

  function highlightJson(obj) {
    const raw = esc(JSON.stringify(obj, null, 2));
    return raw.replace(/(&quot;(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\&]|&(?!quot;))*&quot;(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (m) => {
      let cls = 'json-number';
      if (m.startsWith('&quot;')) cls = m.endsWith(':') ? 'json-key' : 'json-string';
      else if (/true|false/.test(m)) cls = 'json-boolean';
      else if (/null/.test(m)) cls = 'json-null';
      return `<span class="${cls}">${m}</span>`;
    });
  }

  function renderJsonView(res) {
    if (!res) return;
    const ctx = currentContext();
    const qs = new URLSearchParams({
      tenant: state.tenant,
      platform: ctx.platform,
      app_version: ctx.app_version,
      city: ctx.city,
      segment: ctx.segment,
      user_id: ctx.user_id,
      window_class: ctx.window_class,
    });
    if (ctx.at) qs.set('at', ctx.at);
    const origin = window.location.origin;
    if ($('kmp-curl-cmd')) {
      $('kmp-curl-cmd').textContent = `curl -sS "${origin}/api/v1/screens/${encodeURIComponent(state.activeSlug)}?${qs.toString()}" | jq .`;
    }
    const bytes = new TextEncoder().encode(JSON.stringify(res.page)).length;
    if ($('json-meta-stats')) {
      $('json-meta-stats').textContent = `${(res.page?.snippets || []).length} snippets • ${(bytes / 1024).toFixed(1)} KB`;
    }
    if ($('json-viewer')) $('json-viewer').innerHTML = highlightJson(res.page);
  }

  // ── Layout Authoring Actions ────────────────────────────────────────────────

  async function selectLayout(slug) {
    state.activeSlug = slug;
    const found = state.layouts.find((l) => l.slug === slug);
    state.activeLayout = found ? structuredClone(found) : null;
    setDirty(false);
    renderAll();
    await refreshPreview();
  }

  async function saveDraft() {
    if (!state.activeLayout) return false;
    try {
      const saved = await api('PUT', `/api/v1/studio/layouts/${encodeURIComponent(state.activeLayout.slug)}`, state.activeLayout);
      state.activeLayout = structuredClone(saved);
      const idx = state.layouts.findIndex((l) => l.slug === saved.slug);
      if (idx >= 0) state.layouts[idx] = saved;
      else state.layouts.push(saved);
      setDirty(false);
      renderAll();
      await refreshPreview();
      toast(`Saved draft v${saved.draft_version} for ${saved.slug}`, 'success');
      return true;
    } catch (err) {
      toast(`Save failed: ${err.message}`, 'error');
      return false;
    }
  }

  async function validateDraft() {
    if (!state.activeLayout) return;
    if (state.dirty) await saveDraft();
    try {
      const res = await api('POST', `/api/v1/studio/layouts/${encodeURIComponent(state.activeLayout.slug)}/validate`);
      const banner = $('validation-banner');
      const body = $('validation-banner-body');
      if (banner && body) {
        banner.hidden = false;
        if (res.valid) {
          body.innerHTML = `<strong style="color:var(--emerald)">✓ Valid!</strong> Ready to publish v1 KMP PageResponse.`;
          toast('Layout draft passed validation', 'success');
        } else {
          body.innerHTML = `<strong style="color:var(--rose)">✕ ${(res.issues || []).length} issue(s):</strong> ` +
            (res.issues || []).map((i) => `${esc(i.path)}: ${esc(i.message)}`).join(' • ');
          toast(`Validation found ${(res.issues || []).length} issue(s)`, 'warning');
        }
      }
    } catch (err) {
      toast(`Validation error: ${err.message}`, 'error');
    }
  }

  async function openHistoryModal() {
    if (!state.activeSlug) return;
    if ($('history-layout-slug')) $('history-layout-slug').textContent = state.activeSlug;
    openModal('modal-history');
    try {
      const data = await api('GET', `/api/v1/studio/layouts/${encodeURIComponent(state.activeSlug)}/versions`);
      const snaps = data?.snapshots || [];
      const host = $('history-list');
      if (!host) return;
      if (snaps.length === 0) {
        host.innerHTML = `<div class="empty-state">No published versions yet.</div>`;
        return;
      }
      host.innerHTML = snaps.map((s) => `
        <div class="history-card">
          <div class="card-row">
            <span class="card-title">Version v${s.version} — ${esc(s.title)}</span>
            <button type="button" class="chip-btn" data-rollback-ver="${s.version}">Rollback to v${s.version}</button>
          </div>
          <div class="card-sub">${esc(s.comment || 'Published snapshot')} • by ${esc(s.published_by)} • ${new Date(s.published_at).toLocaleString()} • ${(s.slots || []).length} slots</div>
        </div>`).join('');
    } catch (err) {
      toast(`Failed to load history: ${err.message}`, 'error');
    }
  }

  function renderBrandConfig() {
    const spHost = $('splash-list');
    if (spHost) {
      spHost.innerHTML = state.splash.map((s) => `
        <div class="resource-card">
          <div class="card-row">
            <span class="card-title">${esc(s.key)}</span>
            <span class="status-badge ${s.is_default ? 'badge-emerald' : 'badge-amber'}">${s.is_default ? 'Default' : 'Targeted'}</span>
          </div>
          <div class="card-sub mono">${esc(s.background_color_hex)} • ${s.duration_ms}ms</div>
        </div>`).join('') || `<div class="empty-state">No splash configs.</div>`;
    }
    const tkHost = $('token-list');
    if (tkHost) {
      tkHost.innerHTML = state.tokens.map((t) => `
        <div class="resource-card">
          <div class="card-row">
            <span class="card-title">${esc(t.key)}</span>
            <span class="status-badge badge-violet">${esc(t.theme_preset)}</span>
          </div>
          <div class="card-sub mono">${esc(JSON.stringify(t.tokens || {}))}</div>
        </div>`).join('') || `<div class="empty-state">No token overrides.</div>`;
    }
  }

  function addConditionRow() {
    const host = $('nr-conditions-list');
    if (!host) return;
    const row = document.createElement('div');
    row.className = 'condition-row';
    row.innerHTML = `
      <select class="cond-field">
        <option value="segment">segment</option><option value="city">city</option>
        <option value="platform">platform</option><option value="app_version">app_version</option>
        <option value="user_id">user_id</option><option value="window_class">window_class</option>
        <option value="dark_mode">dark_mode</option><option value="zone">zone</option>
      </select>
      <select class="cond-op">
        <option value="eq">eq</option><option value="neq">neq</option>
        <option value="in">in</option><option value="semver_gte">semver_gte</option>
        <option value="bucket_pct">bucket_pct</option><option value="contains">contains</option>
        <option value="regex">regex</option>
      </select>
      <input type="text" class="cond-val" placeholder="vip" />
      <button type="button" class="btn-icon-xs danger" onclick="this.parentElement.remove()">✕</button>`;
    host.appendChild(row);
  }

  // ── Event Wiring ────────────────────────────────────────────────────────────

  function initEvents() {
    // Theme toggle
    $('btn-theme-toggle')?.addEventListener('click', () => {
      const root = document.documentElement;
      const next = root.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
      root.setAttribute('data-theme', next);
      if ($('icon-sun')) $('icon-sun').hidden = next === 'dark';
      if ($('icon-moon')) $('icon-moon').hidden = next !== 'dark';
    });

    // Modal close buttons
    document.querySelectorAll('[data-close-modal]').forEach((btn) => {
      btn.addEventListener('click', () => closeModal(btn.dataset.closeModal));
    });

    // Left rail tabs
    document.querySelectorAll('[data-left-tab]').forEach((btn) => {
      btn.addEventListener('click', () => {
        state.leftTab = btn.dataset.leftTab;
        document.querySelectorAll('[data-left-tab]').forEach((b) => b.classList.toggle('active', b === btn));
        ['pages', 'widgets', 'rules', 'slotting', 'audit'].forEach((t) => {
          const p = $(`panel-${t}`);
          if (p) p.hidden = t !== state.leftTab;
        });
      });
    });

    // Center stage tabs
    document.querySelectorAll('[data-center-tab]').forEach((btn) => {
      btn.addEventListener('click', () => {
        state.centerTab = btn.dataset.centerTab;
        document.querySelectorAll('[data-center-tab]').forEach((b) => b.classList.toggle('active', b === btn));
        ['composer', 'preview', 'json'].forEach((t) => {
          const v = $(`view-${t}`);
          if (v) v.hidden = t !== state.centerTab;
        });
        if (state.centerTab !== 'composer') refreshPreview();
      });
    });

    // Device switcher
    document.querySelectorAll('[data-device]').forEach((btn) => {
      btn.addEventListener('click', () => {
        state.device = btn.dataset.device;
        document.querySelectorAll('[data-device]').forEach((b) => b.classList.toggle('active', b === btn));
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
      });
    });

    // Draft vs Published preview toggle (synced between preview bar and inspector)
    const setPreviewMode = (isDraft) => {
      state.previewDraft = isDraft;
      ['btn-preview-draft', 'ctx-mode-draft'].forEach((id) => $(id)?.classList.toggle('active', isDraft));
      ['btn-preview-published', 'ctx-mode-published'].forEach((id) => $(id)?.classList.toggle('active', !isDraft));
      setDirty(state.dirty);
      refreshPreview();
    };
    $('btn-preview-draft')?.addEventListener('click', () => setPreviewMode(true));
    $('ctx-mode-draft')?.addEventListener('click', () => setPreviewMode(true));
    $('btn-preview-published')?.addEventListener('click', () => setPreviewMode(false));
    $('ctx-mode-published')?.addEventListener('click', () => setPreviewMode(false));

    // Tenant change & creation
    $('tenant-select')?.addEventListener('change', (e) => {
      state.tenant = e.target.value;
      loadAll();
    });
    $('btn-new-tenant')?.addEventListener('click', () => openModal('modal-tenant'));
    $('tenant-form')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      try {
        const saved = await api('POST', '/api/v1/studio/tenants', {
          slug: $('tf-slug').value.trim(),
          name: $('tf-name').value.trim(),
          description: $('tf-description').value.trim(),
          active: true,
        });
        state.tenant = saved.slug;
        closeModal('modal-tenant');
        toast(`Created tenant ${saved.slug}`, 'success');
        await loadAll();
      } catch (err) { toast(err.message, 'error'); }
    });

    // Header actions
    $('btn-validate')?.addEventListener('click', validateDraft);
    $('btn-save-draft')?.addEventListener('click', saveDraft);
    $('btn-preview-save-refresh')?.addEventListener('click', saveDraft);
    $('btn-refresh-preview')?.addEventListener('click', refreshPreview);
    $('btn-close-validation')?.addEventListener('click', () => { $('validation-banner').hidden = true; });
    $('btn-history')?.addEventListener('click', openHistoryModal);
    $('btn-brand-config')?.addEventListener('click', () => openModal('modal-brand-config'));
    $('btn-set-fallback')?.addEventListener('click', async () => {
      if (!state.activeSlug) return;
      try {
        const fb = await api('POST', `/api/v1/studio/layouts/${encodeURIComponent(state.activeSlug)}/fallback`, { actor: 'studio-admin' });
        toast(`Promoted v${fb.snapshot_version} as safe fallback`, 'success');
      } catch (err) { toast(err.message, 'error'); }
    });

    // Publish modal
    $('btn-publish')?.addEventListener('click', () => {
      if ($('publish-diff-section')) $('publish-diff-section').hidden = true;
      openModal('modal-publish');
    });
    $('publish-form')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      if (!state.activeSlug) return;
      if (state.dirty) {
        const ok = await saveDraft();
        if (!ok) return;
      }
      try {
        const res = await api('POST', `/api/v1/studio/layouts/${encodeURIComponent(state.activeSlug)}/publish`, {
          actor: $('pf-actor')?.value?.trim() || 'studio-admin',
          comment: $('pf-comment')?.value?.trim() || '',
        });
        toast(`Published ${state.activeSlug} v${res.snapshot.version}!`, 'success');
        if ($('publish-diff-section') && $('publish-diff-list') && (res.diff || []).length) {
          $('publish-diff-section').hidden = false;
          $('publish-diff-list').innerHTML = res.diff.slice(0, 12)
            .map((d) => `<div class="diff-item diff-${esc(d.op)}"><span class="mono">${esc(d.op)}</span> ${esc(d.path)}</div>`).join('');
        } else {
          closeModal('modal-publish');
        }
        await loadAll();
      } catch (err) { toast(`Publish failed: ${err.message}`, 'error'); }
    });

    // New Page inline form
    $('btn-toggle-new-page')?.addEventListener('click', () => { $('new-page-form').hidden = !$('new-page-form').hidden; });
    $('btn-cancel-new-page')?.addEventListener('click', () => { $('new-page-form').hidden = true; });
    $('new-page-form')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      const slug = $('np-slug').value.trim().toLowerCase();
      const title = $('np-title').value.trim() || slug;
      const mode = $('np-layout-mode').value;
      const layout = {
        slug,
        title,
        description: $('np-description').value.trim(),
        page_config: {
          toolbar: { title: { text: title, style: 'title_large' } },
          adaptive_layout: { layout_mode: mode, split_pane_enabled: mode === 'split_pane' },
          floating_cart: $('np-floating-cart').checked
            ? { label: 'Express Cart', eta_badge: { label: '8 min' }, cta: { id: 'cart_cta', label: 'Checkout' } }
            : undefined,
        },
        slots: state.widgets[0]
          ? [{ slot_id: 'slot_0', widget_id: state.widgets[0].id, position: 0, enabled: true }]
          : [],
      };
      try {
        const saved = await api('PUT', `/api/v1/studio/layouts/${encodeURIComponent(slug)}`, layout);
        $('new-page-form').hidden = true;
        $('new-page-form').reset();
        state.activeSlug = saved.slug;
        toast(`Created dynamic page "${saved.slug}"`, 'success');
        await loadAll();
      } catch (err) { toast(err.message, 'error'); }
    });

    // Page Config modal
    $('btn-page-config')?.addEventListener('click', () => {
      const l = state.activeLayout;
      if (!l) return;
      $('pcf-title').value = l.title || '';
      $('pcf-description').value = l.description || '';
      $('pcf-toolbar-title').value = l.page_config?.toolbar?.title?.text || '';
      $('pcf-toolbar-subtitle').value = l.page_config?.toolbar?.subtitle?.text || '';
      $('pcf-layout-mode').value = l.page_config?.adaptive_layout?.layout_mode || 'single_column';
      $('pcf-max-width').value = l.page_config?.adaptive_layout?.max_width_dp || '';
      $('pcf-split-pane').checked = Boolean(l.page_config?.adaptive_layout?.split_pane_enabled);
      $('pcf-secondary-ids').value = (l.page_config?.adaptive_layout?.secondary_snippet_ids || []).join(', ');
      $('pcf-floating-cart-enabled').checked = Boolean(l.page_config?.floating_cart);
      $('pcf-cart-label').value = l.page_config?.floating_cart?.label || 'Express Cart';
      $('pcf-cart-eta').value = l.page_config?.floating_cart?.eta_badge?.label || '8 min';
      $('pcf-cart-cta').value = l.page_config?.floating_cart?.cta?.label || 'Checkout';
      openModal('modal-page-config');
    });
    $('page-config-form')?.addEventListener('submit', (e) => {
      e.preventDefault();
      const l = state.activeLayout;
      if (!l) return;
      l.title = $('pcf-title').value.trim();
      l.description = $('pcf-description').value.trim();
      l.page_config = l.page_config || {};
      const tbTitle = $('pcf-toolbar-title').value.trim();
      const tbSub = $('pcf-toolbar-subtitle').value.trim();
      l.page_config.toolbar = tbTitle ? {
        title: { text: tbTitle, style: 'title_large' },
        subtitle: tbSub ? { text: tbSub, style: 'body_small' } : undefined,
      } : undefined;
      const maxW = Number($('pcf-max-width').value);
      l.page_config.adaptive_layout = {
        layout_mode: $('pcf-layout-mode').value,
        max_width_dp: maxW > 0 ? maxW : undefined,
        split_pane_enabled: $('pcf-split-pane').checked,
        secondary_snippet_ids: $('pcf-secondary-ids').value.split(',').map((s) => s.trim()).filter(Boolean),
      };
      l.page_config.floating_cart = $('pcf-floating-cart-enabled').checked ? {
        label: $('pcf-cart-label').value.trim() || 'Express Cart',
        eta_badge: { label: $('pcf-cart-eta').value.trim() || '8 min' },
        cta: { id: 'cart_checkout', label: $('pcf-cart-cta').value.trim() || 'Checkout' },
      } : undefined;
      closeModal('modal-page-config');
      setDirty(true);
      renderHeader();
    });

    // Widgets filter & modal
    $('widget-search')?.addEventListener('input', renderWidgetsList);
    $('widget-type-filter')?.addEventListener('change', renderWidgetsList);
    $('btn-composer-add-widget')?.addEventListener('click', () => {
      document.querySelector('[data-left-tab="widgets"]')?.click();
    });
    $('btn-open-new-widget')?.addEventListener('click', () => {
      $('modal-widget-title').textContent = 'Create Reusable Widget';
      $('wf-id').value = '';
      $('wf-name').value = '';
      $('wf-slug').value = '';
      $('wf-tags').value = '';
      $('wf-active').checked = true;
      $('wf-props').value = '{\n  "title": { "text": "New Dynamic Banner • {{city}}" }\n}';
      $('wf-motion').value = '';
      $('wf-json-error').hidden = true;
      openModal('modal-widget');
    });
    $('wf-template')?.addEventListener('change', (e) => {
      const tpl = state.templates.find((t) => t.key === e.target.value);
      if (!tpl) return;
      $('wf-snippet-type').value = tpl.snippet_type;
      $('wf-props').value = JSON.stringify(tpl.default_props || {}, null, 2);
    });
    $('btn-format-widget-json')?.addEventListener('click', () => {
      try {
        $('wf-props').value = JSON.stringify(JSON.parse($('wf-props').value || '{}'), null, 2);
        $('wf-json-error').hidden = true;
      } catch (err) {
        $('wf-json-error').textContent = err.message;
        $('wf-json-error').hidden = false;
      }
    });
    $('widget-form')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      try {
        const props = JSON.parse($('wf-props').value || '{}');
        const motionRaw = $('wf-motion').value.trim();
        const motion = motionRaw ? JSON.parse(motionRaw) : undefined;
        await api('POST', '/api/v1/studio/widgets', {
          id: $('wf-id').value || undefined,
          name: $('wf-name').value.trim(),
          slug: $('wf-slug').value.trim(),
          template_key: $('wf-template').value || undefined,
          snippet_type: $('wf-snippet-type').value,
          props,
          motion,
          tags: $('wf-tags').value.split(',').map((s) => s.trim()).filter(Boolean),
          active: $('wf-active').checked,
        });
        closeModal('modal-widget');
        toast('Widget saved', 'success');
        await loadAll();
      } catch (err) {
        $('wf-json-error').textContent = err.message;
        $('wf-json-error').hidden = false;
      }
    });

    // Slot Edit modal
    $('btn-format-slot-json')?.addEventListener('click', () => {
      try {
        $('sf-overrides').value = JSON.stringify(JSON.parse($('sf-overrides').value || '{}'), null, 2);
        $('sf-json-error').hidden = true;
      } catch (err) {
        $('sf-json-error').textContent = err.message;
        $('sf-json-error').hidden = false;
      }
    });
    $('slot-form')?.addEventListener('submit', (e) => {
      e.preventDefault();
      const idx = Number($('sf-index').value);
      const slot = state.activeLayout?.slots?.[idx];
      if (!slot) return;
      try {
        const ovRaw = $('sf-overrides').value.trim();
        slot.overrides = ovRaw ? JSON.parse(ovRaw) : undefined;
        slot.rule_id = $('sf-rule-id').value || '';
        slot.pinned = $('sf-pinned').checked;
        slot.enabled = $('sf-enabled').checked;
        const sh = $('sf-start-hour').value;
        const eh = $('sf-end-hour').value;
        slot.schedule = (sh !== '' && eh !== '') ? { start_hour_utc: Number(sh), end_hour_utc: Number(eh) } : undefined;
        closeModal('modal-slot');
        setDirty(true);
        renderComposer();
      } catch (err) {
        $('sf-json-error').textContent = err.message;
        $('sf-json-error').hidden = false;
      }
    });

    // Rules form
    $('btn-toggle-new-rule')?.addEventListener('click', () => {
      $('new-rule-form').hidden = !$('new-rule-form').hidden;
      if (!$('new-rule-form').hidden && $('nr-conditions-list').children.length === 0) addConditionRow();
    });
    $('btn-cancel-new-rule')?.addEventListener('click', () => { $('new-rule-form').hidden = true; });
    $('btn-add-condition')?.addEventListener('click', addConditionRow);
    $('new-rule-form')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      const conditions = [...$('nr-conditions-list').querySelectorAll('.condition-row')].map((row) => {
        const op = row.querySelector('.cond-op').value;
        const raw = row.querySelector('.cond-val').value.trim();
        let val = raw;
        if (op === 'in' || op === 'not_in') val = raw.split(',').map((s) => s.trim()).filter(Boolean);
        else if (op === 'bucket_pct') val = Number(raw);
        return { field: row.querySelector('.cond-field').value, operator: op, value: val };
      });
      const sh = $('nr-start-hour').value;
      const eh = $('nr-end-hour').value;
      try {
        await api('POST', '/api/v1/studio/rules', {
          key: $('nr-key').value.trim(),
          name: $('nr-name').value.trim(),
          description: $('nr-description').value.trim(),
          priority: Number($('nr-priority').value || 50),
          match_mode: $('nr-match-mode').value,
          enabled: $('nr-enabled').checked,
          conditions,
          schedule: (sh !== '' && eh !== '') ? { start_hour_utc: Number(sh), end_hour_utc: Number(eh) } : undefined,
        });
        $('new-rule-form').hidden = true;
        $('new-rule-form').reset();
        $('nr-conditions-list').innerHTML = '';
        toast('Targeting rule created', 'success');
        await loadAll();
      } catch (err) { toast(err.message, 'error'); }
    });

    // Slotting form
    $('btn-toggle-new-slotting')?.addEventListener('click', () => {
      $('new-slotting-form').hidden = !$('new-slotting-form').hidden;
      if ($('ns-layout-slug') && state.activeSlug) $('ns-layout-slug').value = state.activeSlug;
    });
    $('btn-cancel-new-slotting')?.addEventListener('click', () => { $('new-slotting-form').hidden = true; });
    $('new-slotting-form')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      const sh = $('ns-start-hour').value;
      const eh = $('ns-end-hour').value;
      try {
        await api('POST', '/api/v1/studio/slotting', {
          name: $('ns-name').value.trim(),
          layout_slug: $('ns-layout-slug').value.trim() || state.activeSlug,
          priority: Number($('ns-priority').value || 90),
          action: $('ns-action').value,
          target_position: Number($('ns-position').value || 1),
          widget_id: $('ns-widget-id').value,
          rule_id: $('ns-rule-id').value || '',
          enabled: $('ns-enabled').checked,
          schedule: (sh !== '' && eh !== '') ? { start_hour_utc: Number(sh), end_hour_utc: Number(eh) } : undefined,
        });
        $('new-slotting-form').hidden = true;
        $('new-slotting-form').reset();
        toast('Dynamic slotting rule saved', 'success');
        await loadAll();
      } catch (err) { toast(err.message, 'error'); }
    });

    // Brand Config tabs & forms
    $('brand-tab-splash')?.addEventListener('click', () => {
      $('brand-tab-splash').classList.add('active');
      $('brand-tab-tokens').classList.remove('active');
      $('brand-panel-splash').hidden = false;
      $('brand-panel-tokens').hidden = true;
    });
    $('brand-tab-tokens')?.addEventListener('click', () => {
      $('brand-tab-tokens').classList.add('active');
      $('brand-tab-splash').classList.remove('active');
      $('brand-panel-tokens').hidden = false;
      $('brand-panel-splash').hidden = true;
    });
    $('splash-form')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      try {
        await api('POST', '/api/v1/studio/splash', {
          key: $('sp-key').value.trim(),
          background_color_hex: $('sp-bg-hex').value.trim() || '#0C831F',
          duration_ms: Number($('sp-duration').value || 1500),
          image_url: $('sp-image-url').value.trim(),
          rule_id: $('sp-rule-id').value || '',
          is_default: $('sp-default').checked,
          enabled: $('sp-enabled').checked,
        });
        toast('Splash config saved', 'success');
        await loadAll();
      } catch (err) { toast(err.message, 'error'); }
    });
    $('token-form')?.addEventListener('submit', async (e) => {
      e.preventDefault();
      try {
        const tokens = JSON.parse($('tk-tokens-json').value || '{}');
        $('tk-json-error').hidden = true;
        await api('POST', '/api/v1/studio/tokens', {
          key: $('tk-key').value.trim(),
          theme_preset: $('tk-preset').value.trim() || 'abeta-fresh',
          rule_id: $('tk-rule-id').value || '',
          tokens,
          enabled: $('tk-enabled').checked,
        });
        toast('Token override saved', 'success');
        await loadAll();
      } catch (err) {
        $('tk-json-error').textContent = err.message;
        $('tk-json-error').hidden = false;
      }
    });

    // Audit refresh & JSON copy buttons
    $('btn-refresh-audit')?.addEventListener('click', loadAll);
    $('btn-copy-curl')?.addEventListener('click', () => {
      navigator.clipboard?.writeText($('kmp-curl-cmd')?.textContent || '');
      toast('Copied curl command', 'success');
    });
    $('btn-copy-page-json')?.addEventListener('click', () => {
      if (state.previewResult?.page) {
        navigator.clipboard?.writeText(JSON.stringify(state.previewResult.page, null, 2));
        toast('Copied PageResponse JSON', 'success');
      }
    });
    $('btn-copy-full-json')?.addEventListener('click', () => {
      if (state.previewResult) {
        navigator.clipboard?.writeText(JSON.stringify(state.previewResult, null, 2));
        toast('Copied Page + Trace JSON', 'success');
      }
    });

    // Context Simulator live bindings
    ['ctx-platform', 'ctx-app-version', 'ctx-city', 'ctx-zone', 'ctx-segment', 'ctx-window-class', 'ctx-user-id', 'ctx-dark-mode', 'ctx-hour-enabled'].forEach((id) => {
      $(id)?.addEventListener('change', schedulePreview);
      $(id)?.addEventListener('input', schedulePreview);
    });
    $('ctx-utc-hour')?.addEventListener('input', (e) => {
      if ($('ctx-utc-hour-label')) $('ctx-utc-hour-label').textContent = `${String(e.target.value).padStart(2, '0')}:30 UTC`;
      if ($('ctx-hour-enabled')) $('ctx-hour-enabled').checked = true;
      schedulePreview();
    });
    document.querySelectorAll('[data-hour-preset]').forEach((btn) => {
      btn.addEventListener('click', () => {
        const val = btn.dataset.hourPreset;
        if (val === 'live') {
          $('ctx-hour-enabled').checked = false;
          if ($('ctx-utc-hour-label')) $('ctx-utc-hour-label').textContent = 'Live UTC';
        } else {
          $('ctx-hour-enabled').checked = true;
          $('ctx-utc-hour').value = val;
          if ($('ctx-utc-hour-label')) $('ctx-utc-hour-label').textContent = `${String(val).padStart(2, '0')}:30 UTC`;
        }
        schedulePreview();
      });
    });
    $('btn-reset-context')?.addEventListener('click', () => {
      $('ctx-platform').value = 'ios';
      $('ctx-app-version').value = '1.2.0';
      $('ctx-city').value = 'Bengaluru';
      $('ctx-zone').value = 'koramangala-01';
      $('ctx-segment').value = 'standard';
      $('ctx-window-class').value = 'compact';
      $('ctx-user-id').value = 'user-42';
      $('ctx-dark-mode').checked = false;
      $('ctx-hour-enabled').checked = false;
      refreshPreview();
    });

    // Delegated clicks across lists, slots, preview steppers, and history rollback
    document.addEventListener('click', async (e) => {
      const target = e.target.closest('button, [data-select-layout]');
      if (!target) return;

      if (target.dataset.selectLayout) {
        selectLayout(target.dataset.selectLayout);
        return;
      }
      if (target.dataset.addWidget && state.activeLayout) {
        const wid = target.dataset.addWidget;
        const slots = state.activeLayout.slots || (state.activeLayout.slots = []);
        slots.push({ slot_id: `slot_${Date.now().toString(36)}`, widget_id: wid, position: slots.length, enabled: true });
        setDirty(true);
        renderComposer();
        toast('Widget added to draft layout', 'info');
        return;
      }
      if (target.dataset.editWidget) {
        const w = widgetById(target.dataset.editWidget);
        if (!w) return;
        $('modal-widget-title').textContent = `Edit Widget: ${w.name}`;
        $('wf-id').value = w.id;
        $('wf-template').value = w.template_key || '';
        $('wf-snippet-type').value = w.snippet_type;
        $('wf-name').value = w.name;
        $('wf-slug').value = w.slug;
        $('wf-tags').value = (w.tags || []).join(', ');
        $('wf-active').checked = w.active;
        $('wf-props').value = JSON.stringify(w.props || {}, null, 2);
        $('wf-motion').value = w.motion ? JSON.stringify(w.motion, null, 2) : '';
        $('wf-json-error').hidden = true;
        openModal('modal-widget');
        return;
      }
      if (target.dataset.delWidget) {
        try {
          await api('DELETE', `/api/v1/studio/widgets/${encodeURIComponent(target.dataset.delWidget)}`);
          toast('Widget deleted', 'info');
          await loadAll();
        } catch (err) { toast(err.message, 'error'); }
        return;
      }
      if (target.dataset.delRule) {
        try {
          await api('DELETE', `/api/v1/studio/rules/${encodeURIComponent(target.dataset.delRule)}`);
          toast('Rule deleted', 'info');
          await loadAll();
        } catch (err) { toast(err.message, 'error'); }
        return;
      }
      if (target.dataset.delSlotting) {
        try {
          await api('DELETE', `/api/v1/studio/slotting/${encodeURIComponent(target.dataset.delSlotting)}`);
          toast('Slotting rule deleted', 'info');
          await loadAll();
        } catch (err) { toast(err.message, 'error'); }
        return;
      }
      if (target.dataset.slotUp !== undefined && state.activeLayout) {
        const i = Number(target.dataset.slotUp);
        if (i > 0) {
          const a = state.activeLayout.slots;
          [a[i - 1], a[i]] = [a[i], a[i - 1]];
          a.forEach((s, idx) => { s.position = idx; });
          setDirty(true);
          renderComposer();
        }
        return;
      }
      if (target.dataset.slotDown !== undefined && state.activeLayout) {
        const i = Number(target.dataset.slotDown);
        const a = state.activeLayout.slots;
        if (i < a.length - 1) {
          [a[i], a[i + 1]] = [a[i + 1], a[i]];
          a.forEach((s, idx) => { s.position = idx; });
          setDirty(true);
          renderComposer();
        }
        return;
      }
      if (target.dataset.slotPin !== undefined && state.activeLayout) {
        const slot = state.activeLayout.slots[Number(target.dataset.slotPin)];
        if (slot) { slot.pinned = !slot.pinned; setDirty(true); renderComposer(); }
        return;
      }
      if (target.dataset.slotToggle !== undefined && state.activeLayout) {
        const slot = state.activeLayout.slots[Number(target.dataset.slotToggle)];
        if (slot) { slot.enabled = !slot.enabled; setDirty(true); renderComposer(); }
        return;
      }
      if (target.dataset.slotRemove !== undefined && state.activeLayout) {
        state.activeLayout.slots.splice(Number(target.dataset.slotRemove), 1);
        state.activeLayout.slots.forEach((s, idx) => { s.position = idx; });
        setDirty(true);
        renderComposer();
        return;
      }
      if (target.dataset.slotEdit !== undefined && state.activeLayout) {
        const idx = Number(target.dataset.slotEdit);
        const slot = state.activeLayout.slots[idx];
        if (!slot) return;
        const w = widgetById(slot.widget_id);
        $('sf-index').value = String(idx);
        $('sf-slot-id').value = slot.slot_id;
        $('sf-widget-label').textContent = w ? `${w.name} (${w.snippet_type})` : slot.widget_id;
        $('sf-rule-id').value = slot.rule_id || '';
        $('sf-start-hour').value = slot.schedule?.start_hour_utc ?? '';
        $('sf-end-hour').value = slot.schedule?.end_hour_utc ?? '';
        $('sf-pinned').checked = Boolean(slot.pinned);
        $('sf-enabled').checked = Boolean(slot.enabled);
        $('sf-overrides').value = slot.overrides ? JSON.stringify(slot.overrides, null, 2) : '';
        $('sf-json-error').hidden = true;
        openModal('modal-slot');
        return;
      }
      if (target.dataset.rollbackVer) {
        try {
          const ver = Number(target.dataset.rollbackVer);
          await api('POST', `/api/v1/studio/layouts/${encodeURIComponent(state.activeSlug)}/rollback`, { version: ver, actor: 'studio-admin' });
          closeModal('modal-history');
          toast(`Rolled back ${state.activeSlug} to v${ver}`, 'success');
          await loadAll();
        } catch (err) { toast(err.message, 'error'); }
        return;
      }
      if (target.dataset.cartInc) {
        const pid = target.dataset.cartInc;
        state.cartCounts[pid] = (state.cartCounts[pid] || 0) + 1;
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.cartDec) {
        const pid = target.dataset.cartDec;
        state.cartCounts[pid] = Math.max(0, (state.cartCounts[pid] || 0) - 1);
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.expressCheckout) {
        try {
          const res = await api('POST', '/api/v1/orders/express', { cart: state.cartCounts });
          toast(res?.follow_up_actions?.[0]?.payload?.title || 'Express order confirmed!', 'success');
          state.cartCounts = {};
          if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        } catch (err) { toast(err.message, 'error'); }
      }
    });
  }

  document.addEventListener('error', (e) => {
    if (e.target && e.target.tagName === 'IMG') {
      e.target.style.display = 'none';
    }
  }, true);

  document.addEventListener('DOMContentLoaded', () => {
    initEvents();
    loadAll();
  });
})();
