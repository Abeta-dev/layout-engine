(() => {
  'use strict';

  const CANONICAL_SNIPPET_TYPES = [
    'accordion',
    'action_row',
    'banner_v1',
    'card_v1',
    'carousel',
    'chip_group',
    'composed_block',
    'divider',
    'empty_state',
    'form_v1',
    'grid',
    'header_v1',
    'marquee_strip',
    'metric_card',
    'nudge_bar',
    'order_tracker',
    'product_card',
    'search_bar',
    'spacer',
    'swipe_deck',
    'tabs',
  ];

  const CANONICAL_COMPOSED_NODE_KINDS = [
    'column',
    'row',
    'box',
    'card',
    'flow_row',
    'scroll_row',
    'grid',
    'text',
    'button',
    'tag',
    'icon',
    'image',
    'avatar',
    'progress',
    'divider',
    'spacer',
    'stepper',
    'countdown',
    'rolling_counter',
    'radio_group',
    'select',
    'otp_input',
    'chip',
    'segmented_control',
    'rating',
    'price',
    'badge',
    'alert_banner',
    'stat_tile',
    'key_value_row',
    'qr_code',
    'text_field',
    'switch',
    'checkbox',
    'slider',
    'media_pager',
  ];

  const CANONICAL_BOX_ALIGNMENTS = new Set([
    'top_start',
    'top_center',
    'top_end',
    'center_start',
    'center',
    'center_end',
    'bottom_start',
    'bottom_center',
    'bottom_end',
  ]);

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
    deckIndex: {},
    searchQueries: {},
    chipSelections: {},
    dismissedNudges: {},
    composedSelections: {},
    dismissedAlerts: {},
    dragIndex: null,
    previewTimer: null,
  };

  function resetPreviewInteractions() {
    state.cartCounts = {};
    state.deckIndex = {};
    state.searchQueries = {};
    state.chipSelections = {};
    state.dismissedNudges = {};
    state.composedSelections = {};
    state.dismissedAlerts = {};
  }

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
      case 'header_v1': case 'banner_v1': case 'marquee_strip': return 'badge-emerald';
      case 'product_card': case 'metric_card': case 'nudge_bar': return 'badge-amber';
      case 'grid': case 'carousel': case 'tabs': case 'accordion': case 'order_tracker': return 'badge-sky';
      case 'form_v1': case 'action_row': case 'chip_group': case 'composed_block': return 'badge-violet';
      case 'card_v1': case 'search_bar': return 'badge-teal';
      case 'empty_state': case 'swipe_deck': return 'badge-rose';
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
      if (state.tenants.length > 0 && !state.tenants.some((t) => t.slug === state.tenant)) {
        state.tenant = state.tenants[0].slug;
      }
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

      if (state.layouts.length === 0) {
        state.activeSlug = '';
        state.activeLayout = null;
        state.previewResult = null;
      } else {
        if (!state.layouts.some((l) => l.slug === state.activeSlug)) {
          state.activeSlug = state.layouts[0].slug;
        }
        const active = state.layouts.find((l) => l.slug === state.activeSlug);
        state.activeLayout = active ? structuredClone(active) : null;
      }
      if ($('tk-preset')) $('tk-preset').value = state.tenant;
      if ($('ns-layout-slug')) $('ns-layout-slug').value = state.activeSlug || '';
      setDirty(false);
      populateSelects();
      renderAll();
      if (state.activeLayout) {
        await refreshPreview();
      } else {
        renderEmptyPreview();
      }
    } catch (err) {
      toast(`Failed to load studio workspace: ${err.message}`, 'error');
    }
  }

  function renderTenants() {
    const sel = $('tenant-select');
    if (!sel) return;
    if (state.tenants.length === 0) {
      sel.innerHTML = `<option value="${esc(state.tenant)}" selected>${esc(state.tenant)}</option>`;
      sel.value = state.tenant;
      return;
    }
    sel.innerHTML = state.tenants
      .map((t) => `<option value="${esc(t.slug)}" ${t.slug === state.tenant ? 'selected' : ''}>${esc(t.name)} (${esc(t.slug)})</option>`)
      .join('');
    sel.value = state.tenant;
  }

  function populateSelects() {
    const typeFilter = $('widget-type-filter');
    const types = [...new Set([
      ...CANONICAL_SNIPPET_TYPES,
      ...state.templates.map((t) => t.snippet_type),
      ...state.widgets.map((w) => w.snippet_type),
    ].filter(Boolean))].sort();
    if (typeFilter) {
      const cur = typeFilter.value;
      typeFilter.innerHTML = `<option value="">All ${types.length} snippet types</option>` +
        types.map((t) => `<option value="${esc(t)}" ${t === cur ? 'selected' : ''}>${esc(t)}</option>`).join('');
    }
    const wfTpl = $('wf-template');
    if (wfTpl) {
      wfTpl.innerHTML = `<option value="">Custom (manual props)</option>` +
        state.templates.map((t) => `<option value="${esc(t.key)}">${esc(t.name)} (${esc(t.snippet_type)})</option>`).join('');
    }
    const wfType = $('wf-snippet-type');
    if (wfType) {
      const curType = wfType.value;
      wfType.innerHTML = types.map((t) => `<option value="${esc(t)}" ${t === curType ? 'selected' : ''}>${esc(t)}</option>`).join('');
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
    if (meta) {
      if (l) {
        const mode = l.page_config?.adaptive_layout?.layout_mode || 'single_column';
        const cart = l.page_config?.floating_cart ? ' • Floating Cart' : '';
        meta.textContent = `${l.title} • ${mode}${cart}`;
      } else {
        meta.textContent = '';
      }
    }
    if ($('slotting-layout-label')) $('slotting-layout-label').textContent = l ? l.slug : 'all';
    if ($('ns-layout-slug') && l) $('ns-layout-slug').value = l.slug;
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
    if (!host) return;
    if (!l) {
      if ($('composer-slot-count')) $('composer-slot-count').textContent = '0 slots';
      host.innerHTML = `<div class="empty-state">No dynamic layouts in ${esc(state.tenant)} yet. Create a page in the left rail to start composing.</div>`;
      return;
    }
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
      const titlePreview = textOf(w?.props?.title)
        || textOf(w?.props?.headline)
        || textOf(w?.props?.label)
        || w?.props?.placeholder
        || (w?.snippet_type === 'marquee_strip' && Array.isArray(w?.props?.items) ? `${w.props.items.length} scrolling offer${w.props.items.length === 1 ? '' : 's'}` : '')
        || (w?.snippet_type === 'composed_block' ? (w?.props?.recipe_name ? `Recipe: ${w.props.recipe_name}` : `Composed (${w?.props?.root?.kind || 'tree'})`) : '')
        || (Array.isArray(w?.props?.cards) ? `${w.props.cards.length} discovery card${w.props.cards.length === 1 ? '' : 's'}` : '')
        || (Array.isArray(w?.props?.chips) ? `${w.props.chips.length} filter chip${w.props.chips.length === 1 ? '' : 's'}` : '')
        || w?.slug
        || '';
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

  // ── Center Canvas — Live Device Preview (All 21 KMP v1 Snippets) ────────────

  function renderEmptyPreview() {
    const screen = $('device-screen');
    const frame = $('device-frame');
    if (frame) frame.className = `device-frame device-${state.device}`;
    if (screen) {
      screen.className = `device-screen ${$('ctx-dark-mode')?.checked ? 'kmp-dark' : ''}`;
      screen.innerHTML = `<div class="kmp-scroll-body"><div class="kmp-empty-state"><div class="kmp-h-title">No layout selected</div><div class="kmp-h-sub">Select or create a dynamic layout in ${esc(state.tenant)}.</div></div></div>`;
    }
    if ($('json-viewer')) $('json-viewer').textContent = '{}';
    if ($('json-meta-stats')) $('json-meta-stats').textContent = '0 snippets • 0.0 KB';
  }

  async function refreshPreview() {
    if (!state.activeSlug) {
      renderEmptyPreview();
      return;
    }
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

  function previewIconGlyph(icon) {
    const name = (typeof icon === 'string' ? icon : icon?.name || '').toLowerCase();
    switch (name) {
      case 'search': return '🔍';
      case 'mic': return '🎙';
      case 'tune': case 'filter_list': case 'sliders': return '⚙';
      case 'bolt': case 'flash_on': return '⚡';
      case 'local_fire_department': case 'flame': case 'whatshot': return '🔥';
      case 'eco': case 'leaf': return '🌿';
      case 'star': case 'grade': case 'sparkles': return '★';
      case 'location_on': case 'place': case 'pin': return '📍';
      case 'schedule': case 'timer': return '⏱';
      case 'verified': case 'check_circle': case 'check': return '✓';
      case 'radio_button_unchecked': return '○';
      case 'local_shipping': return '🛵';
      case 'home': return '🏠';
      case 'restaurant': case 'fastfood': return '🍽';
      case 'shopping_bag': case 'storefront': case 'inventory_2': return '🛍';
      case 'music_note': case 'confirmation_number': return '♪';
      case 'trending_up': return '↗';
      case 'wb_sunny': return '☀';
      case 'notifications': return '🔔';
      case 'settings': return '⚙';
      case 'info': return 'ℹ';
      case 'person': return '👤';
      case 'arrow_forward': return '→';
      default: return name ? '•' : '';
    }
  }

  function motionLoopClass(motion) {
    const loop = motion?.loop_effect;
    switch (loop) {
      case 'pulse':
      case 'shimmer':
      case 'wiggle':
      case 'float':
      case 'heartbeat':
      case 'glow_ring':
      case 'marquee':
        return `kmp-loop-${loop}`;
      default:
        return '';
    }
  }

  function formatCountdown(cd) {
    if (!cd) return '';
    const secs = Math.max(0, Number(cd.remaining_seconds || 0));
    if (secs === 0 && cd.expired_text) return cd.expired_text;
    const hrs = Math.floor(secs / 3600);
    const mins = Math.floor((secs % 3600) / 60);
    const rem = secs % 60;
    const pad = (n) => String(n).padStart(2, '0');
    if (cd.show_hours || hrs > 0) return `${pad(hrs)}:${pad(mins)}:${pad(rem)}`;
    return `${pad(mins)}:${pad(rem)}`;
  }

  function renderCountdownPill(cd) {
    if (!cd) return '';
    const secs = Math.max(0, Number(cd.remaining_seconds || 0));
    const urgent = secs <= Number(cd.urgent_threshold_seconds || 60);
    const loopCls = motionLoopClass(cd.motion);
    return `
      <span class="kmp-countdown-pill ${urgent ? 'urgent' : ''} ${loopCls}">
        ${cd.label ? `<span class="kmp-countdown-label">${esc(cd.label)}</span>` : ''}
        <span class="mono">${esc(formatCountdown(cd))}</span>
      </span>`;
  }

  function mapArrangementStyle(arr) {
    switch (arr) {
      case 'space_between': return 'justify-content:space-between';
      case 'space_around': return 'justify-content:space-around';
      case 'space_evenly': return 'justify-content:space-evenly';
      case 'center': return 'justify-content:center';
      case 'end': return 'justify-content:flex-end';
      default: return '';
    }
  }

  function normalizeVariant(v) {
    switch ((v || '').toLowerCase()) {
      case 'brand':
      case 'success':
      case 'warning':
      case 'error':
      case 'info':
        return v.toLowerCase();
      default:
        return 'neutral';
    }
  }

  function mapValueColorClass(colorRef) {
    const token = (typeof colorRef === 'string' ? colorRef : colorRef?.token || '').toLowerCase();
    if (token.includes('success')) return 'kmp-color-success';
    if (token.includes('error')) return 'kmp-color-error';
    if (token.includes('warning')) return 'kmp-color-warning';
    if (token.includes('brand') || token.includes('accent')) return 'kmp-color-brand';
    return '';
  }

  function renderQrSvg(payload, sizeDp) {
    const dim = Math.max(72, Math.min(148, Number(sizeDp || 120)));
    const cells = 15;
    let hash = 2166136261;
    const str = String(payload || 'abeta://pay');
    for (let i = 0; i < str.length; i++) {
      hash ^= str.charCodeAt(i);
      hash = Math.imul(hash, 16777619) >>> 0;
    }
    const isFinder = (r, c) =>
      (r < 4 && c < 4) || (r < 4 && c >= cells - 4) || (r >= cells - 4 && c < 4);
    const rects = [];
    for (let r = 0; r < cells; r++) {
      for (let c = 0; c < cells; c++) {
        let on = false;
        if (isFinder(r, c)) {
          const lr = r < 4 ? r : r - (cells - 4);
          const lc = c < 4 ? c : c - (cells - 4);
          on = lr === 0 || lr === 3 || lc === 0 || lc === 3 || (lr === 1 && lc === 1) || (lr === 2 && lc === 2) || (lr === 1 && lc === 2) || (lr === 2 && lc === 1);
        } else {
          const bit = ((hash >>> ((r * 7 + c * 13) % 24)) ^ (r * 31 + c * 17)) & 1;
          on = bit === 1;
        }
        if (on) rects.push(`<rect x="${c + 1}" y="${r + 1}" width="1" height="1" rx="0.15"/>`);
      }
    }
    return `<svg class="kmp-qr-svg" width="${dim}" height="${dim}" viewBox="0 0 ${cells + 2} ${cells + 2}" role="img" aria-label="QR Code">${rects.join('')}</svg>`;
  }

  function buildComposedNodeTemplate(kind) {
    switch (kind) {
      case 'column':
      case 'row':
      case 'box':
      case 'card':
      case 'flow_row':
      case 'scroll_row':
        return { kind, spacing_dp: 8, children: [] };
      case 'grid':
        return { kind, columns: 2, spacing_dp: 8, columns_by_width: { compact: 2, medium: 3, expanded: 4 }, children: [] };
      case 'text':
        return { kind, text: { text: 'Sample Heading' } };
      case 'button':
        return { kind, button: { id: 'btn_new', label: 'Continue', variant: 'primary' } };
      case 'tag':
        return { kind, tag: { label: 'NEW' } };
      case 'icon':
        return { kind, icon: { name: 'verified', size_dp: 20 } };
      case 'image':
        return { kind, image: { url: 'https://cdn.abeta.dev/fresh/hero.png', content_description: 'Banner' } };
      case 'avatar':
        return { kind, avatar: { initials: 'AB', size_dp: 36 } };
      case 'progress':
        return { kind, progress: { kind: 'linear', progress: 0.72, label: '72% complete' } };
      case 'divider':
        return { kind, divider: { thickness_dp: 1 } };
      case 'spacer':
        return { kind, spacer: { height_dp: 8 } };
      case 'stepper':
        return { kind, stepper: { field_id: 'qty_1', label: 'ADD', quantity: 1 } };
      case 'countdown':
        return { kind, countdown: { countdown_id: 'cd_1', remaining_seconds: 300, label: 'Ends in' } };
      case 'rolling_counter':
        return { kind, counter_value: '$2,480.50' };
      case 'text_field':
        return { kind, text_field: { field_id: 'delivery_note', label: 'Delivery Instructions', placeholder: 'Gate code or drop-off note...' } };
      case 'switch':
        return { kind, switch: { field_id: 'leave_at_door', label: 'Leave at Door', subtitle: 'Contactless express drop-off', checked: true } };
      case 'checkbox':
        return { kind, checkbox: { field_id: 'no_cutlery', label: 'Skip disposable cutlery', checked: true } };
      case 'slider':
        return { kind, slider: { field_id: 'rider_tip', label: 'Rider Tip', value: 2, min_value: 0, max_value: 10, steps: 10, unit_label: 'USD' } };
      case 'media_pager':
        return {
          kind,
          media_pager: {
            id: 'media_pager_1',
            height_dp: 140,
            show_indicators: true,
            items: [
              { id: 'slide_1', url: 'https://cdn.abeta.dev/fresh/hero_1.png', content_description: 'Express Hero 1' },
              { id: 'slide_2', url: 'https://cdn.abeta.dev/fresh/hero_2.png', content_description: 'Express Hero 2' },
            ],
          },
        };
      case 'radio_group':
        return { kind, radio_group: { field_id: 'funding_source', label: 'Choose Option', selected_id: 'opt_1', options: [{ id: 'opt_1', label: 'Option One', badge: 'BEST' }, { id: 'opt_2', label: 'Option Two' }] } };
      case 'select':
        return { kind, select: { field_id: 'currency', label: 'Currency', placeholder: 'Select currency', selected_id: 'usd', options: [{ id: 'usd', label: 'USD ($)' }, { id: 'eur', label: 'EUR (€)' }] } };
      case 'otp_input':
        return { kind, otp_input: { field_id: 'pin_code', label: '6-Digit PIN', length: 6, masked: true } };
      case 'chip':
        return { kind, chip: { id: 'chip_1', label: 'Instant 0% Fee', selected: true, leading_icon: { name: 'bolt' } } };
      case 'segmented_control':
        return { kind, segmented_control: { field_id: 'mode', selected_id: 'instant', options: [{ id: 'instant', label: 'Instant' }, { id: 'scheduled', label: 'Scheduled' }] } };
      case 'rating':
        return { kind, rating: { id: 'rt_1', value: 4.9, max_stars: 5, review_count: 1240, label: 'Trust Score' } };
      case 'price':
        return { kind, price: { id: 'pr_1', amount_text: '$125.00', original_amount_text: '$127.50', discount_label: 'ZERO FEE' } };
      case 'badge':
        return { kind, badge: { id: 'bdg_1', count: 3, max_count: 99, variant: 'brand' } };
      case 'alert_banner':
        return { kind, alert_banner: { id: 'al_1', title: 'Instant Clearing Active', message: 'Zero-fee transfers enabled.', variant: 'success', icon: { name: 'verified' } } };
      case 'stat_tile':
        return { kind, stat_tile: { id: 'st_1', label: 'Available Balance', value: '$2,480.50', delta_label: '+5.0% APY', delta_positive: true, icon: { name: 'star' } } };
      case 'key_value_row':
        return { kind, key_value_row: { id: 'kv_1', label: 'Network Fee', value: '$0.00', emphasized: true, show_divider: true } };
      case 'qr_code':
        return { kind, qr_code: { id: 'qr_1', payload: 'abeta://pay/@alex.rivera', label: '@alex.rivera', caption: 'Scan to pay instantly', size_dp: 140 } };
      default:
        return { kind };
    }
  }

  function resolveComposedGridColumns(node) {
    const bw = node?.columns_by_width || {};
    let cols = Number(node?.columns || 2);
    if (state.device === 'desktop' && Number(bw.expanded) > 0) cols = Number(bw.expanded);
    else if (state.device === 'tablet' && Number(bw.medium) > 0) cols = Number(bw.medium);
    else if (state.device === 'phone' && Number(bw.compact) > 0) cols = Number(bw.compact);
    return Math.max(1, Math.min(6, cols));
  }

  function renderComposedNode(node, depth = 0, parentKind = '') {
    if (!node || depth > 6) return '';
    const kind = node.kind || '';
    const loopCls = motionLoopClass(node.motion);
    const gap = node.spacing_dp != null ? `gap:${Number(node.spacing_dp)}px;` : '';
    const weight = node.weight != null && Number(node.weight) > 0 ? `flex:${Number(node.weight)} 1 0%;min-width:0;` : '';
    const arr = mapArrangementStyle(node.main_axis_arrangement);
    const alignItems = node.vertical_alignment === 'center' || node.alignment === 'center'
      ? 'align-items:center;'
      : (node.vertical_alignment === 'bottom' || node.alignment === 'end' ? 'align-items:flex-end;' : '');
    const baseStyle = `${gap}${weight}${arr ? `${arr};` : ''}${alignItems}`;
    const childrenArr = Array.isArray(node.children) ? node.children : [];
    const childrenHtml = childrenArr.map((ch) => renderComposedNode(ch, depth + 1, kind)).join('');

    const wrapBoxOverlay = (html) => {
      if (!html) return '';
      const rawAlign = typeof node.box_alignment === 'string' ? node.box_alignment.trim() : '';
      if (parentKind === 'box' && CANONICAL_BOX_ALIGNMENTS.has(rawAlign)) {
        return `<div class="kmp-box-child kmp-box-align-${rawAlign}">${html}</div>`;
      }
      return html;
    };

    switch (kind) {
      case 'card':
        return wrapBoxOverlay(`<div class="kmp-composed-card ${loopCls}" style="${baseStyle}">${childrenHtml}</div>`);
      case 'column':
        return wrapBoxOverlay(`<div class="kmp-composed-column ${loopCls}" style="${baseStyle}">${childrenHtml}</div>`);
      case 'row':
        return wrapBoxOverlay(`<div class="kmp-composed-row ${loopCls}" style="${baseStyle}">${childrenHtml}</div>`);
      case 'box': {
        const hasOverlay = childrenArr.some((ch) => typeof ch?.box_alignment === 'string' && CANONICAL_BOX_ALIGNMENTS.has(ch.box_alignment.trim()));
        return wrapBoxOverlay(`<div class="kmp-composed-box ${hasOverlay ? 'has-overlay' : ''} ${loopCls}" style="${baseStyle}">${childrenHtml}</div>`);
      }
      case 'flow_row':
        return wrapBoxOverlay(`<div class="kmp-composed-flow-row ${loopCls}" style="${baseStyle}">${childrenHtml}</div>`);
      case 'scroll_row':
        return wrapBoxOverlay(`<div class="kmp-composed-scroll-row ${loopCls}" style="${baseStyle}">${childrenHtml}</div>`);
      case 'grid': {
        const cols = resolveComposedGridColumns(node);
        return wrapBoxOverlay(`<div class="kmp-composed-grid ${loopCls}" style="grid-template-columns:repeat(${cols},minmax(0,1fr));${baseStyle}">${childrenHtml}</div>`);
      }
      case 'text': {
        const token = node.text?.typography?.token || '';
        const isTitle = token.includes('headline') || token.includes('title');
        const isSmall = token.includes('small') || token.includes('label');
        const cls = isTitle ? 'kmp-h-title' : (isSmall ? 'kmp-h-sub' : 'kmp-banner-body');
        return wrapBoxOverlay(`<div class="${cls} ${loopCls}" style="${weight}">${esc(textOf(node.text))}</div>`);
      }
      case 'button':
        if (!node.button) return '';
        return wrapBoxOverlay(`<button type="button" class="kmp-btn ${node.button.variant === 'secondary' || node.button.variant === 'outline' ? 'kmp-btn-secondary' : 'kmp-btn-primary'} ${motionLoopClass(node.button.motion) || loopCls}">${esc(node.button.label || 'Action')}</button>`);
      case 'tag':
        if (!node.tag) return '';
        return wrapBoxOverlay(`<span class="kmp-tag ${motionLoopClass(node.tag.motion) || loopCls}">${esc(node.tag.label || textOf(node.tag))}</span>`);
      case 'icon': {
        const glyph = previewIconGlyph(node.icon);
        const iconName = (node.icon?.name || '').toLowerCase();
        const isDone = iconName === 'check_circle' || iconName === 'verified';
        return wrapBoxOverlay(`<span class="kmp-composed-icon ${isDone ? 'done' : ''} ${loopCls}" title="${esc(node.icon?.name || '')}">${esc(glyph || '•')}</span>`);
      }
      case 'image': {
        if (!node.image?.url) return '';
        const hStyle = node.image.height_dp ? `height:${Math.min(180, Number(node.image.height_dp))}px;` : '';
        const boxImgCls = parentKind === 'box' ? 'kmp-composed-img-box' : '';
        return wrapBoxOverlay(`<img class="kmp-composed-img ${boxImgCls} ${loopCls}" style="${hStyle}" src="${esc(node.image.url)}" alt="${esc(node.image.content_description || '')}" />`);
      }
      case 'avatar':
        return wrapBoxOverlay(`<span class="kmp-avatar ${loopCls}">${esc(node.avatar?.initials || 'AB')}</span>`);
      case 'progress': {
        const pct = Math.min(100, Math.max(0, Math.round((Number(node.progress?.progress ?? node.progress?.value ?? 0)) * 100)));
        return wrapBoxOverlay(`
          <div class="kmp-composed-progress ${loopCls}" style="${weight}">
            <div class="kmp-progress-track"><div class="kmp-progress-fill" style="width:${pct}%"></div></div>
            ${node.progress?.label ? `<div class="kmp-h-sub">${esc(node.progress.label)}</div>` : ''}
          </div>`);
      }
      case 'divider':
        return wrapBoxOverlay(`<hr class="kmp-divider" />`);
      case 'spacer':
        return wrapBoxOverlay(`<div style="height:${Number(node.spacer?.height_dp || 8)}px;width:${Number(node.spacer?.width_dp || 0)}px"></div>`);
      case 'stepper': {
        const qty = Number(node.stepper?.quantity || 0);
        return wrapBoxOverlay(`<span class="kmp-tag ${loopCls}">${qty > 0 ? `Qty ${qty}` : esc(node.stepper?.label || 'ADD')}</span>`);
      }
      case 'countdown':
        return wrapBoxOverlay(renderCountdownPill(node.countdown));
      case 'rolling_counter':
        return wrapBoxOverlay(`<span class="kmp-rolling-counter mono ${loopCls}">${esc(node.counter_value || '0')}</span>`);

      case 'text_field': {
        const tf = node.text_field || {};
        return wrapBoxOverlay(`
          <div class="kmp-composed-textfield ${loopCls}" style="${weight}">
            ${tf.label ? `<label class="kmp-atom-label">${esc(tf.label)}${tf.required ? ' *' : ''}</label>` : ''}
            <input type="text" class="kmp-input-box" placeholder="${esc(tf.placeholder || tf.label || tf.field_id || '')}" value="${esc(tf.initial_value || '')}" />
            ${tf.helper_text ? `<div class="kmp-h-sub">${esc(tf.helper_text)}</div>` : ''}
          </div>`);
      }

      case 'switch': {
        const sw = node.switch || {};
        return wrapBoxOverlay(`
          <label class="kmp-composed-switch kmp-form-switch ${loopCls}" style="${weight}">
            <span>
              <div class="kmp-kv-label">${esc(sw.label || sw.field_id || 'Toggle')}</div>
              ${sw.subtitle ? `<div class="kmp-h-sub">${esc(sw.subtitle)}</div>` : ''}
            </span>
            <input type="checkbox" ${sw.checked ? 'checked' : ''} />
          </label>`);
      }

      case 'checkbox': {
        const cb = node.checkbox || {};
        return wrapBoxOverlay(`
          <label class="kmp-composed-checkbox kmp-form-switch ${loopCls}" style="${weight}">
            <span>
              <div class="kmp-kv-label">${esc(cb.label || cb.field_id || 'Check')}</div>
              ${cb.subtitle ? `<div class="kmp-h-sub">${esc(cb.subtitle)}</div>` : ''}
            </span>
            <input type="checkbox" ${cb.checked ? 'checked' : ''} />
          </label>`);
      }

      case 'slider': {
        const sl = node.slider || {};
        const minV = Number(sl.min_value ?? 0);
        const maxV = Number(sl.max_value ?? 100);
        const curV = Number(sl.value ?? minV);
        return wrapBoxOverlay(`
          <div class="kmp-composed-slider kmp-form-slider ${loopCls}" style="${weight}">
            <div class="card-row">
              <span class="kmp-kv-label">${esc(sl.label || sl.field_id || 'Slider')}</span>
              ${sl.show_value === false ? '' : `<span class="mono kmp-h-sub">${esc(curV)}${sl.unit_label ? ` ${esc(sl.unit_label)}` : ''}</span>`}
            </div>
            <input type="range" min="${minV}" max="${maxV}" value="${curV}" />
          </div>`);
      }

      case 'media_pager': {
        const mp = node.media_pager || {};
        const items = Array.isArray(mp.items) ? mp.items : [];
        const height = Math.max(84, Math.min(220, Number(mp.height_dp || 140)));
        const showDots = mp.show_indicators !== false && items.length > 1;
        return wrapBoxOverlay(`
          <div class="kmp-media-pager ${loopCls}" style="${weight}">
            <div class="kmp-media-pager-track">
              ${items.length ? items.map((img, idx) => `
                <div class="kmp-media-pager-slide" style="height:${height}px">
                  ${img?.url ? `<img class="kmp-media-pager-img" src="${esc(img.url)}" alt="${esc(img.content_description || `Slide ${idx + 1}`)}" />` : ''}
                  <span class="kmp-media-pager-caption">${esc(img?.content_description || `Slide ${idx + 1}`)}</span>
                </div>`).join('') : `<div class="kmp-media-pager-slide" style="height:${height}px"><span class="kmp-media-pager-caption">Media Pager</span></div>`}
            </div>
            ${showDots ? `<div class="kmp-media-pager-dots">${items.map((_, i) => `<span class="kmp-media-pager-dot ${i === 0 ? 'active' : ''}"></span>`).join('')}</div>` : ''}
          </div>`);
      }

      case 'radio_group': {
        const rg = node.radio_group || {};
        const key = `rg:${rg.field_id || node.node_id || 'rg'}`;
        const selId = key in state.composedSelections ? state.composedSelections[key] : rg.selected_id;
        const opts = Array.isArray(rg.options) ? rg.options : [];
        return wrapBoxOverlay(`
          <div class="kmp-radio-group ${loopCls}" style="${weight}">
            ${rg.label ? `<div class="kmp-atom-label">${esc(rg.label)}</div>` : ''}
            <div class="kmp-radio-list">
              ${opts.map((o) => {
                const active = o.id === selId;
                return `
                  <button type="button" class="kmp-radio-opt ${active ? 'selected' : ''}" data-composed-key="${esc(key)}" data-composed-val="${esc(o.id)}">
                    <span class="kmp-radio-dot ${active ? 'selected' : ''}" aria-hidden="true"></span>
                    <span class="kmp-radio-copy">
                      <span class="kmp-radio-title">${esc(o.label || o.id)}</span>
                      ${o.subtitle ? `<span class="kmp-h-sub">${esc(o.subtitle)}</span>` : ''}
                    </span>
                    ${o.badge ? `<span class="kmp-tag">${esc(o.badge)}</span>` : ''}
                  </button>`;
              }).join('')}
            </div>
          </div>`);
      }

      case 'select': {
        const sel = node.select || {};
        const key = `sel:${sel.field_id || node.node_id || 'sel'}`;
        const curVal = key in state.composedSelections ? state.composedSelections[key] : (sel.selected_id || '');
        const opts = Array.isArray(sel.options) ? sel.options : [];
        return wrapBoxOverlay(`
          <div class="kmp-select ${loopCls}" style="${weight}">
            ${sel.label ? `<div class="kmp-atom-label">${esc(sel.label)}</div>` : ''}
            <select class="kmp-select-box" data-composed-select="${esc(key)}" aria-label="${esc(sel.label || sel.placeholder || 'Select')}">
              ${sel.placeholder && !curVal ? `<option value="">${esc(sel.placeholder)}</option>` : ''}
              ${opts.map((o) => `<option value="${esc(o.id)}" ${o.id === curVal ? 'selected' : ''}>${esc(o.label || o.id)}${o.badge ? ` (${esc(o.badge)})` : ''}</option>`).join('')}
            </select>
            ${sel.error_text ? `<div class="kmp-atom-error">${esc(sel.error_text)}</div>` : (sel.helper_text ? `<div class="kmp-h-sub">${esc(sel.helper_text)}</div>` : '')}
          </div>`);
      }

      case 'otp_input': {
        const otp = node.otp_input || {};
        const len = Math.max(4, Math.min(8, Number(otp.length || 6)));
        const val = String(otp.initial_value || '');
        const boxes = [];
        for (let i = 0; i < len; i++) {
          const ch = val[i] || '';
          const display = ch ? (otp.masked ? '•' : ch) : (otp.masked && i < 2 ? '•' : '');
          boxes.push(`<span class="kmp-otp-cell ${display ? 'filled' : ''}">${esc(display)}</span>`);
        }
        return wrapBoxOverlay(`
          <div class="kmp-otp-input ${loopCls}" style="${weight}">
            ${otp.label ? `<div class="kmp-atom-label">${esc(otp.label)}</div>` : ''}
            <div class="kmp-otp-cells" aria-label="${esc(otp.label || 'OTP')}">${boxes.join('')}</div>
            ${otp.error_text ? `<div class="kmp-atom-error">${esc(otp.error_text)}</div>` : (otp.helper_text ? `<div class="kmp-h-sub">${esc(otp.helper_text)}</div>` : '')}
          </div>`);
      }

      case 'chip': {
        const ch = node.chip || {};
        const key = `chip:${ch.id || node.node_id || ch.label || 'c'}`;
        const active = key in state.composedSelections ? state.composedSelections[key] === '1' : Boolean(ch.selected);
        const glyph = ch.leading_icon ? previewIconGlyph(ch.leading_icon) : '';
        return wrapBoxOverlay(`
          <button type="button" class="kmp-chip ${active ? 'selected' : ''} ${loopCls}" data-composed-toggle="${esc(key)}" data-composed-cur="${active ? '1' : '0'}">
            ${glyph ? `<span aria-hidden="true">${esc(glyph)}</span>` : ''}
            <span>${esc(ch.label || ch.id || 'Chip')}</span>
            ${ch.removable ? '<span class="kmp-chip-rm" aria-hidden="true">×</span>' : ''}
          </button>`);
      }

      case 'segmented_control': {
        const sc = node.segmented_control || {};
        const key = `seg:${sc.field_id || node.node_id || 'seg'}`;
        const selId = key in state.composedSelections ? state.composedSelections[key] : sc.selected_id;
        const opts = Array.isArray(sc.options) ? sc.options : [];
        return wrapBoxOverlay(`
          <div class="kmp-segmented ${loopCls}" style="${weight}" role="group">
            ${opts.map((o) => {
              const active = o.id === selId;
              return `<button type="button" class="kmp-seg-pill ${active ? 'selected' : ''}" data-composed-key="${esc(key)}" data-composed-val="${esc(o.id)}">${esc(o.label || o.id)}</button>`;
            }).join('')}
          </div>`);
      }

      case 'rating': {
        const rt = node.rating || {};
        const maxStars = Math.max(1, Math.min(10, Number(rt.max_stars || 5)));
        const val = Number(rt.value || 0);
        const full = Math.round(val);
        let stars = '';
        for (let i = 1; i <= maxStars; i++) {
          stars += `<span class="kmp-star ${i <= full ? 'filled' : ''}">★</span>`;
        }
        return wrapBoxOverlay(`
          <div class="kmp-rating ${loopCls}">
            <span class="kmp-rating-stars" aria-label="${val} of ${maxStars} stars">${stars}</span>
            <span class="kmp-rating-val mono">${val.toFixed(1)}</span>
            ${rt.review_count != null ? `<span class="kmp-h-sub">(${Number(rt.review_count).toLocaleString()})</span>` : ''}
            ${rt.label ? `<span class="kmp-h-sub">${esc(rt.label)}</span>` : ''}
          </div>`);
      }

      case 'price': {
        const pr = node.price || {};
        return wrapBoxOverlay(`
          <div class="kmp-price-atom ${loopCls}">
            <span class="kmp-price-amt ${pr.emphasized === false ? '' : 'emphasized'}">${esc(pr.amount_text || '$0.00')}</span>
            ${pr.unit_suffix ? `<span class="kmp-h-sub">${esc(pr.unit_suffix)}</span>` : ''}
            ${pr.original_amount_text ? `<span class="kmp-orig-price">${esc(pr.original_amount_text)}</span>` : ''}
            ${pr.discount_label ? `<span class="kmp-price-disc">${esc(pr.discount_label)}</span>` : ''}
          </div>`);
      }

      case 'badge': {
        const b = node.badge || {};
        const variant = normalizeVariant(b.variant);
        const bLoop = motionLoopClass(b.motion) || loopCls;
        if (b.dot_only) {
          return wrapBoxOverlay(`<span class="kmp-badge-dot kmp-var-${variant} ${bLoop}" aria-label="${esc(b.label || variant)}"></span>`);
        }
        let text = b.label || '';
        if (b.count != null) {
          const maxC = Number(b.max_count || 99);
          text = Number(b.count) > maxC ? `${maxC}+` : String(b.count);
        }
        return wrapBoxOverlay(`<span class="kmp-badge-atom kmp-var-${variant} ${bLoop}">${esc(text || '•')}</span>`);
      }

      case 'alert_banner': {
        const al = node.alert_banner || {};
        const alertKey = al.id || node.node_id || al.title || 'alert';
        if (state.dismissedAlerts[alertKey]) return '';
        const variant = normalizeVariant(al.variant);
        const glyph = al.icon ? previewIconGlyph(al.icon) : '';
        return wrapBoxOverlay(`
          <div class="kmp-alert-banner kmp-var-${variant} ${loopCls}" style="${weight}">
            ${glyph ? `<span class="kmp-alert-icon" aria-hidden="true">${esc(glyph)}</span>` : ''}
            <div class="kmp-alert-copy">
              <div class="kmp-alert-title">${esc(al.title || '')}</div>
              ${al.message ? `<div class="kmp-alert-msg">${esc(al.message)}</div>` : ''}
            </div>
            ${al.action_label ? `<button type="button" class="kmp-btn kmp-btn-secondary">${esc(al.action_label)}</button>` : ''}
            ${al.dismissible ? `<button type="button" class="kmp-nudge-dismiss" data-dismiss-alert="${esc(alertKey)}" aria-label="Dismiss">&times;</button>` : ''}
          </div>`);
      }

      case 'stat_tile': {
        const st = node.stat_tile || {};
        const glyph = st.icon ? previewIconGlyph(st.icon) : '';
        const deltaCls = st.delta_positive === false ? 'neg' : 'pos';
        return wrapBoxOverlay(`
          <div class="kmp-stat-tile ${loopCls}" style="${weight}">
            <div class="kmp-stat-head">
              <span class="kmp-h-sub">${esc(st.label || '')}</span>
              ${glyph ? `<span class="kmp-stat-icon" aria-hidden="true">${esc(glyph)}</span>` : ''}
            </div>
            <div class="kmp-stat-val">${esc(st.value || '')}</div>
            ${st.delta_label ? `<span class="kmp-stat-delta ${deltaCls}">${esc(st.delta_label)}</span>` : ''}
            ${st.caption ? `<div class="kmp-h-sub">${esc(st.caption)}</div>` : ''}
          </div>`);
      }

      case 'key_value_row': {
        const kv = node.key_value_row || {};
        const valColorCls = mapValueColorClass(kv.value_color);
        return wrapBoxOverlay(`
          <div class="kmp-kv-row ${kv.show_divider ? 'has-divider' : ''} ${loopCls}" style="${weight}">
            <div class="kmp-kv-left">
              <span class="kmp-kv-label">${esc(kv.label || '')}</span>
              ${kv.sublabel ? `<span class="kmp-h-sub">${esc(kv.sublabel)}</span>` : ''}
            </div>
            <span class="kmp-kv-val ${kv.emphasized ? 'emphasized' : ''} ${valColorCls}">${esc(kv.value || '')}</span>
          </div>`);
      }

      case 'qr_code': {
        const qr = node.qr_code || {};
        return wrapBoxOverlay(`
          <div class="kmp-qr-code ${loopCls}">
            <div class="kmp-qr-frame">${renderQrSvg(qr.payload, qr.size_dp)}</div>
            ${qr.label ? `<div class="kmp-qr-label mono">${esc(qr.label)}</div>` : ''}
            ${qr.caption ? `<div class="kmp-h-sub">${esc(qr.caption)}</div>` : ''}
          </div>`);
      }

      default:
        return wrapBoxOverlay(`<span class="mini-chip">${esc(kind || 'node')}</span>`);
    }
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
        const initQty = Number(d.stepper?.quantity ?? d.stepper?.value ?? 0);
        const qty = pid in state.cartCounts ? state.cartCounts[pid] : initQty;
        const unit = d.unit_price_cents || 0;
        const stepperHtml = qty > 0
          ? `<div class="kmp-stepper has-qty">
               <button type="button" class="kmp-stepper-btn" data-cart-dec="${esc(pid)}" data-init-qty="${initQty}" data-unit="${unit}">−</button>
               <span class="kmp-stepper-qty">${qty}</span>
               <button type="button" class="kmp-stepper-btn" data-cart-inc="${esc(pid)}" data-init-qty="${initQty}" data-unit="${unit}">+</button>
             </div>`
          : `<div class="kmp-stepper">
               <button type="button" class="kmp-stepper-add" data-cart-inc="${esc(pid)}" data-init-qty="${initQty}" data-unit="${unit}">ADD</button>
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
        const tabSnippets = activeTab ? (activeTab.snippets || activeTab.items || []) : [];
        return `
          <div>
            <div class="kmp-tabs-bar">
              ${tabs.map((t, i) => `<button type="button" class="kmp-tab-pill ${i === activeIdx ? 'active' : ''}">${esc(t.title || t.id)}</button>`).join('')}
            </div>
            <div class="kmp-tabs-body">${tabSnippets.map(renderSnippet).join('')}</div>
          </div>`;
      }

      case 'accordion': {
        if (Array.isArray(d.snippets) || d.title) {
          const expanded = d.initially_expanded !== false;
          return `
            <div class="kmp-accordion">
              <div class="kmp-accordion-head">
                <div>
                  <div>${esc(textOf(d.title))}</div>
                  ${d.subtitle ? `<div class="kmp-h-sub">${esc(textOf(d.subtitle))}</div>` : ''}
                </div>
                <span>${expanded ? '▾' : '▸'}</span>
              </div>
              ${expanded ? `<div class="kmp-accordion-body">${(d.snippets || []).map(renderSnippet).join('')}</div>` : ''}
            </div>`;
        }
        return `
          <div class="kmp-accordion">
            ${(d.items || []).map((it, i) => `
              <div>
                <div class="kmp-accordion-head"><span>${esc(textOf(it.title))}</span><span>${i === 0 ? '▾' : '▸'}</span></div>
                ${i === 0 ? `<div class="kmp-accordion-body">${esc(textOf(it.body))}${(it.content || []).map(renderSnippet).join('')}</div>` : ''}
              </div>`).join('')}
          </div>`;
      }

      case 'swipe_deck': {
        const deckId = d.deck_id || env.snippet_id || 'deck';
        const cards = Array.isArray(d.cards) ? d.cards : [];
        const idx = Math.max(0, state.deckIndex[deckId] || 0);
        if (cards.length === 0 || idx >= cards.length) {
          return `
            <div class="kmp-swipe-deck">
              <div class="kmp-swipe-empty">
                <div class="kmp-swipe-empty-icon" aria-hidden="true">✦</div>
                <div class="kmp-h-title">${esc(textOf(d.empty_title) || "You're all caught up")}</div>
                <div class="kmp-h-sub">${esc(textOf(d.empty_message) || 'No more discovery cards in this deck.')}</div>
                ${cards.length > 0 ? `<button type="button" class="kmp-btn kmp-btn-secondary" data-swipe-reset="${esc(deckId)}">Replay Deck (${cards.length})</button>` : ''}
              </div>
            </div>`;
        }
        const card = cards[idx];
        const remaining = cards.length - idx;
        const cardTitle = textOf(card.title) || card.card_id || 'Profile';
        const badgeLabel = card.badge ? (card.badge.label || textOf(card.badge)) : '';
        const distance = textOf(card.distance_text);
        const tags = Array.isArray(card.tags) ? card.tags : [];
        return `
          <div class="kmp-swipe-deck">
            <div class="kmp-swipe-stack">
              ${remaining > 2 ? '<div class="kmp-swipe-underlay kmp-swipe-underlay-2" aria-hidden="true"></div>' : ''}
              ${remaining > 1 ? '<div class="kmp-swipe-underlay kmp-swipe-underlay-1" aria-hidden="true"></div>' : ''}
              <article class="kmp-swipe-card kmp-swipe-grad-${idx % 4}">
                <div class="kmp-swipe-media">
                  ${card.image?.url ? `<img src="${esc(card.image.url)}" alt="${esc(card.image.content_description || cardTitle)}" />` : ''}
                  <div class="kmp-swipe-scrim" aria-hidden="true"></div>
                  <div class="kmp-swipe-media-top">
                    ${badgeLabel ? `<span class="kmp-swipe-badge">${esc(badgeLabel)}</span>` : '<span></span>'}
                    <div class="kmp-swipe-media-pills">
                      ${distance ? `<span class="kmp-swipe-distance">📍 ${esc(distance)}</span>` : ''}
                      <span class="kmp-swipe-counter">${idx + 1}/${cards.length}</span>
                    </div>
                  </div>
                </div>
                <div class="kmp-swipe-body">
                  <div class="kmp-swipe-title">${esc(cardTitle)}</div>
                  ${card.subtitle ? `<div class="kmp-swipe-sub">${esc(textOf(card.subtitle))}</div>` : ''}
                  ${card.bio ? `<div class="kmp-swipe-bio">${esc(textOf(card.bio))}</div>` : ''}
                  ${tags.length ? `<div class="kmp-swipe-tags">${tags.slice(0, 6).map((t) => `<span class="kmp-swipe-tag">${esc(t?.label || textOf(t))}</span>`).join('')}</div>` : ''}
                </div>
              </article>
            </div>
            <div class="kmp-swipe-actions">
              ${idx > 0 ? `<button type="button" class="kmp-swipe-btn kmp-swipe-btn-undo" data-swipe-undo="${esc(deckId)}" title="Undo" aria-label="Undo">↺</button>` : ''}
              <button type="button" class="kmp-swipe-btn kmp-swipe-btn-nope" data-swipe-deck="${esc(deckId)}" data-swipe-dir="nope" data-swipe-title="${esc(cardTitle)}" title="${esc(d.nope_label || 'Nope')}" aria-label="${esc(d.nope_label || 'Nope')}">✕</button>
              <button type="button" class="kmp-swipe-btn kmp-swipe-btn-super" data-swipe-deck="${esc(deckId)}" data-swipe-dir="super_like" data-swipe-title="${esc(cardTitle)}" title="${esc(d.super_like_label || 'Super Like')}" aria-label="${esc(d.super_like_label || 'Super Like')}">★</button>
              <button type="button" class="kmp-swipe-btn kmp-swipe-btn-like" data-swipe-deck="${esc(deckId)}" data-swipe-dir="like" data-swipe-title="${esc(cardTitle)}" title="${esc(d.like_label || 'Like')}" aria-label="${esc(d.like_label || 'Like')}">♥</button>
            </div>
          </div>`;
      }

      case 'search_bar': {
        const searchId = d.search_id || env.snippet_id || 'search';
        const queryVal = searchId in state.searchQueries ? state.searchQueries[searchId] : (d.query || '');
        const placeholder = d.placeholder || 'Search...';
        const clearable = d.clearable !== false;
        const trailingGlyph = d.trailing_icon ? previewIconGlyph(d.trailing_icon) : '';
        return `
          <div class="kmp-search-bar" data-search-wrap="${esc(searchId)}">
            <span class="kmp-search-icon" aria-hidden="true">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.3" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="11" cy="11" r="7.5"></circle>
                <line x1="20" y1="20" x2="16.4" y2="16.4"></line>
              </svg>
            </span>
            <input type="text" class="kmp-search-input" data-search-input="${esc(searchId)}" value="${esc(queryVal)}" placeholder="${esc(placeholder)}" aria-label="${esc(placeholder)}" />
            ${clearable ? `<button type="button" class="kmp-search-clear" data-search-clear="${esc(searchId)}" aria-label="Clear search" ${queryVal ? '' : 'hidden'}>✕</button>` : ''}
            ${trailingGlyph && !d.submit_label ? `<span class="kmp-search-trailing" aria-hidden="true">${esc(trailingGlyph)}</span>` : ''}
            ${d.submit_label ? `<button type="button" class="kmp-search-submit" data-search-submit="${esc(searchId)}">${esc(d.submit_label)}</button>` : ''}
          </div>`;
      }

      case 'chip_group': {
        const groupId = d.group_id || env.snippet_id || 'chips';
        const chips = Array.isArray(d.chips) ? d.chips : [];
        const selectedSet = state.chipSelections[groupId]
          || new Set(chips.filter((c) => c.selected).map((c) => c.id));
        return `
          <div class="kmp-chip-group">
            ${d.title ? `<div class="kmp-section-title">${esc(textOf(d.title))}</div>` : ''}
            <div class="kmp-chip-row" role="group" aria-label="${esc(textOf(d.title) || groupId)}">
              ${chips.map((c) => {
                const isSel = selectedSet.has(c.id);
                const glyph = c.icon ? previewIconGlyph(c.icon) : '';
                return `
                  <button type="button" class="kmp-chip-pill ${isSel ? 'selected' : ''}" data-chip-group="${esc(groupId)}" data-chip-id="${esc(c.id)}" data-chip-multi="${d.multi_select ? '1' : '0'}" aria-pressed="${isSel ? 'true' : 'false'}">
                    ${glyph ? `<span class="kmp-chip-icon" aria-hidden="true">${esc(glyph)}</span>` : ''}
                    <span class="kmp-chip-label">${esc(c.label || c.id)}</span>
                    ${c.badge ? `<span class="kmp-chip-badge">${esc(c.badge)}</span>` : ''}
                  </button>`;
              }).join('')}
            </div>
          </div>`;
      }

      case 'nudge_bar': {
        const nudgeId = d.nudge_id || env.snippet_id || 'nudge';
        if (state.dismissedNudges[nudgeId]) return '';
        const iconGlyph = d.icon ? previewIconGlyph(d.icon) : '';
        const badgeLoop = motionLoopClass(d.badge?.motion);
        const hasProg = typeof d.progress === 'number';
        const pct = hasProg ? Math.min(100, Math.max(0, Math.round(d.progress * 100))) : 0;
        const envLoop = motionLoopClass(env.motion);
        return `
          <div class="kmp-nudge-bar ${envLoop}">
            <div class="kmp-nudge-top">
              <div class="kmp-nudge-lead">
                ${iconGlyph ? `<span class="kmp-nudge-icon" aria-hidden="true">${esc(iconGlyph)}</span>` : ''}
                <div class="kmp-nudge-copy">
                  <div class="kmp-nudge-pills">
                    ${d.badge ? `<span class="kmp-nudge-badge ${badgeLoop}">${esc(d.badge.label || textOf(d.badge))}</span>` : ''}
                    ${d.countdown ? renderCountdownPill(d.countdown) : ''}
                  </div>
                  <div class="kmp-nudge-title">${esc(textOf(d.title))}</div>
                  ${d.subtitle ? `<div class="kmp-nudge-sub">${esc(textOf(d.subtitle))}</div>` : ''}
                </div>
              </div>
              <div class="kmp-nudge-actions">
                ${d.cta ? `<button type="button" class="kmp-btn kmp-btn-primary kmp-nudge-cta">${esc(d.cta.label || 'View')}</button>` : ''}
                ${d.dismissible !== false ? `<button type="button" class="kmp-nudge-dismiss" data-dismiss-nudge="${esc(nudgeId)}" aria-label="Dismiss nudge">&times;</button>` : ''}
              </div>
            </div>
            ${hasProg ? `<div class="kmp-nudge-progress"><div class="kmp-nudge-progress-fill" style="width:${pct}%"></div></div>` : ''}
          </div>`;
      }

      case 'marquee_strip': {
        const items = Array.isArray(d.items) ? d.items : [];
        const sep = d.separator || '•';
        const vel = Math.max(16, Number(d.velocity_dp_per_sec || 48));
        const durationSec = Math.max(8, Math.min(32, Math.round((Math.max(items.length, 2) * 260) / vel)));
        const renderItem = (it) => {
          const glyph = it.icon ? previewIconGlyph(it.icon) : '';
          return `
            <span class="kmp-marquee-item">
              ${glyph ? `<span class="kmp-marquee-icon" aria-hidden="true">${esc(glyph)}</span>` : ''}
              ${it.badge ? `<span class="kmp-marquee-badge">${esc(it.badge.label || textOf(it.badge))}</span>` : ''}
              <span>${esc(it.text || '')}</span>
              <span class="kmp-marquee-sep" aria-hidden="true">${esc(sep)}</span>
            </span>`;
        };
        const chunk = items.length ? items.map(renderItem).join('') : '<span class="kmp-marquee-item">Live Offer Ticker</span>';
        return `
          <div class="kmp-marquee-strip" aria-label="Offer ticker">
            <div class="kmp-marquee-track" style="animation-duration:${durationSec}s">
              <div class="kmp-marquee-group">${chunk}</div>
              <div class="kmp-marquee-group" aria-hidden="true">${chunk}</div>
            </div>
          </div>`;
      }

      case 'order_tracker': {
        const steps = Array.isArray(d.steps) ? d.steps : [];
        const riderInitials = (d.rider_name || 'R').split(/\s+/).map((p) => p[0] || '').join('').slice(0, 2).toUpperCase();
        return `
          <div class="kmp-order-tracker ${motionLoopClass(env.motion)}">
            <div class="kmp-tracker-head">
              <div>
                <div class="kmp-tracker-meta">
                  ${d.order_code ? `<span class="kmp-tracker-code mono">${esc(d.order_code)}</span>` : ''}
                  ${d.eta_badge ? `<span class="kmp-tag ${motionLoopClass(d.eta_badge.motion)}">${esc(d.eta_badge.label || textOf(d.eta_badge))}</span>` : ''}
                </div>
                <div class="kmp-h-title">${esc(textOf(d.headline))}</div>
              </div>
              ${d.eta_countdown ? renderCountdownPill(d.eta_countdown) : ''}
            </div>
            <div class="kmp-tracker-steps">
              ${steps.map((st, idx) => {
                const statusCls = st.completed ? 'completed' : (st.active ? 'active' : 'pending');
                const glyph = st.completed ? '✓' : (st.icon ? previewIconGlyph(st.icon) : String(idx + 1));
                return `
                  <div class="kmp-tracker-step ${statusCls}">
                    <div class="kmp-step-rail">
                      <span class="kmp-step-dot ${statusCls}">${esc(glyph)}</span>
                      ${idx < steps.length - 1 ? `<span class="kmp-step-line ${st.completed ? 'completed' : ''}"></span>` : ''}
                    </div>
                    <div class="kmp-step-body">
                      <div class="kmp-step-label">${esc(st.label || st.id)}</div>
                      ${st.caption ? `<div class="kmp-step-caption">${esc(st.caption)}</div>` : ''}
                    </div>
                  </div>`;
              }).join('')}
            </div>
            ${d.rider_name ? `
              <div class="kmp-tracker-rider">
                <span class="kmp-avatar">${esc(riderInitials)}</span>
                <div class="kmp-tracker-rider-copy">
                  <div class="kmp-tracker-rider-name">${esc(d.rider_name)}</div>
                  ${d.rider_caption ? `<div class="kmp-h-sub">${esc(d.rider_caption)}</div>` : ''}
                </div>
              </div>` : ''}
            ${(d.primary_action || d.secondary_action) ? `
              <div class="kmp-tracker-actions">
                ${d.primary_action ? `<button type="button" class="kmp-btn kmp-btn-primary">${esc(d.primary_action.label || 'Track')}</button>` : ''}
                ${d.secondary_action ? `<button type="button" class="kmp-btn kmp-btn-secondary">${esc(d.secondary_action.label || 'Share')}</button>` : ''}
              </div>` : ''}
          </div>`;
      }

      case 'composed_block': {
        const envLoop = motionLoopClass(env.motion);
        return `
          <div class="kmp-composed-block ${envLoop}">
            ${d.recipe_name ? `<div class="kmp-composed-recipe mono">${esc(d.recipe_name)}</div>` : ''}
            ${renderComposedNode(d.root, 0)}
          </div>`;
      }

      case 'card_v1':
        return `
          <div class="kmp-card">
            <div class="card-row">
              <div class="kmp-h-title">${esc(textOf(d.title))}</div>
              ${d.badge ? `<span class="kmp-tag">${esc(d.badge.label || textOf(d.badge))}</span>` : ''}
            </div>
            ${d.subtitle ? `<div class="kmp-h-sub">${esc(textOf(d.subtitle))}</div>` : ''}
            ${d.description ? `<div class="kmp-banner-body">${esc(textOf(d.description))}</div>` : ''}
            ${(d.primary_button || d.secondary_button || (d.actions || []).length) ? `
              <div class="kmp-card-actions">
                ${d.primary_button ? `<button type="button" class="kmp-btn kmp-btn-primary">${esc(d.primary_button.label || 'Open')}</button>` : ''}
                ${d.secondary_button ? `<button type="button" class="kmp-btn kmp-btn-secondary">${esc(d.secondary_button.label || 'More')}</button>` : ''}
                ${(d.actions || []).map((a) => `<button type="button" class="kmp-btn kmp-btn-primary">${esc(a.label)}</button>`).join('')}
              </div>` : ''}
          </div>`;

      case 'action_row':
        return `
          <div class="kmp-action-row">
            <div>
              <div class="kmp-h-title">${esc(textOf(d.title))}</div>
              ${d.subtitle ? `<div class="kmp-h-sub">${esc(textOf(d.subtitle))}</div>` : ''}
            </div>
            <div class="card-row">
              ${d.trailing_tag ? `<span class="kmp-tag">${esc(d.trailing_tag.label || textOf(d.trailing_tag))}</span>` : ''}
              <span>›</span>
            </div>
          </div>`;

      case 'form_v1':
        return `
          <div class="kmp-form">
            ${d.title ? `<div class="kmp-h-title">${esc(textOf(d.title))}</div>` : ''}
            ${d.description ? `<div class="kmp-h-sub">${esc(textOf(d.description))}</div>` : ''}
            ${(d.fields || []).map((f) => `<input type="text" placeholder="${esc(f.placeholder || f.label || f.field_id)}" value="${esc(f.initial_value || '')}" />`).join('')}
            ${(d.switches || []).map((sw) => `
              <label class="kmp-form-switch">
                <span>
                  <div class="kmp-kv-label">${esc(sw.label || sw.field_id)}</div>
                  ${sw.subtitle ? `<div class="kmp-h-sub">${esc(sw.subtitle)}</div>` : ''}
                </span>
                <input type="checkbox" ${sw.checked ? 'checked' : ''} />
              </label>`).join('')}
            ${(d.checkboxes || []).map((cb) => `
              <label class="kmp-form-switch">
                <span class="kmp-kv-label">${esc(cb.label || cb.field_id)}</span>
                <input type="checkbox" ${cb.checked ? 'checked' : ''} />
              </label>`).join('')}
            ${(d.sliders || []).map((sl) => `
              <div class="kmp-form-slider">
                <div class="card-row">
                  <span class="kmp-kv-label">${esc(sl.label || sl.field_id)}</span>
                  <span class="mono kmp-h-sub">${esc(sl.value ?? sl.min_value ?? 0)} ${esc(sl.unit_label || '')}</span>
                </div>
                <input type="range" min="${Number(sl.min_value ?? 0)}" max="${Number(sl.max_value ?? 100)}" value="${Number(sl.value ?? 0)}" />
              </div>`).join('')}
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

  function collectCartSummary(snippets) {
    const byProduct = new Map();
    const walk = (list) => {
      if (!Array.isArray(list)) return;
      for (const env of list) {
        if (!env || typeof env !== 'object') continue;
        const d = env.data || {};
        if (env.snippet_type === 'product_card') {
          const pid = d.product_id || env.snippet_id;
          if (pid && !byProduct.has(pid)) {
            const initQty = Number(d.stepper?.quantity ?? d.stepper?.value ?? 0);
            const qty = pid in state.cartCounts ? state.cartCounts[pid] : initQty;
            byProduct.set(pid, { qty, unit: Number(d.unit_price_cents || 0) });
          }
        }
        if (Array.isArray(d.items)) walk(d.items);
        if (Array.isArray(d.snippets)) walk(d.snippets);
        if (Array.isArray(d.tabs)) {
          for (const t of d.tabs) {
            if (Array.isArray(t.snippets)) walk(t.snippets);
            if (Array.isArray(t.items)) walk(t.items);
          }
        }
      }
    };
    walk(snippets);
    let totalItems = 0;
    let totalCents = 0;
    byProduct.forEach(({ qty, unit }) => {
      if (qty > 0) {
        totalItems += qty;
        totalCents += qty * unit;
      }
    });
    return { totalItems, totalCents };
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

    const renderEnvelopeItem = (env) => {
      const inner = renderSnippet(env);
      if (env && env.sticky) {
        return `<div class="kmp-sticky-wrap" data-sticky="1">${inner}</div>`;
      }
      return inner;
    };

    const snippets = page.snippets || [];
    const splitEnabled = state.device === 'desktop' && page.adaptive_layout?.split_pane_enabled;
    const secondarySet = new Set(page.adaptive_layout?.secondary_snippet_ids || []);
    let bodyHtml = '';
    if (splitEnabled && secondarySet.size > 0) {
      const primary = snippets.filter((s) => !secondarySet.has(s.snippet_id));
      const secondary = snippets.filter((s) => secondarySet.has(s.snippet_id));
      bodyHtml = `
        <div class="kmp-split-pane">
          <div class="kmp-pane-col">${primary.map(renderEnvelopeItem).join('')}</div>
          <div class="kmp-pane-col kmp-pane-secondary">${secondary.map(renderEnvelopeItem).join('')}</div>
        </div>`;
    } else {
      bodyHtml = snippets.map(renderEnvelopeItem).join('');
    }

    const { totalItems, totalCents } = collectCartSummary(snippets);
    const showCart = Boolean(page.floating_cart);
    let cartNudgeHtml = '';
    if (showCart && Number(page.floating_cart.free_delivery_threshold_cents || 0) > 0) {
      const threshold = Number(page.floating_cart.free_delivery_threshold_cents);
      const remaining = Math.max(0, threshold - totalCents);
      const pct = Math.min(100, Math.round((totalCents / threshold) * 100));
      const unlocked = remaining === 0;
      const prefix = page.floating_cart.nudge_prefix || 'Add';
      const unlockedLabel = page.floating_cart.unlocked_label || 'Free delivery unlocked!';
      const msg = unlocked
        ? `🎉 ${unlockedLabel}`
        : `${prefix} $${(remaining / 100).toFixed(2)} more for free delivery`;
      cartNudgeHtml = `
        <div class="kmp-cart-nudge ${unlocked ? 'unlocked' : ''}">
          <div class="kmp-cart-nudge-row">
            <span>${esc(msg)}</span>
            <span class="mono">${pct}%</span>
          </div>
          <div class="kmp-cart-nudge-track">
            <div class="kmp-cart-nudge-fill" style="width:${pct}%"></div>
          </div>
        </div>`;
    }
    const cartLabel = totalItems > 0
      ? `${totalItems} item${totalItems > 1 ? 's' : ''}${totalCents > 0 ? ` · $${(totalCents / 100).toFixed(2)}` : ''}`
      : (page.floating_cart?.label || 'Express Cart');
    const cartHtml = showCart ? `
      <div class="kmp-floating-cart ${cartNudgeHtml ? 'has-nudge' : ''}">
        ${cartNudgeHtml}
        <div class="kmp-floating-cart-row">
          <div class="kmp-cart-info">
            <span class="kmp-cart-title">${esc(cartLabel)}</span>
            ${page.floating_cart.eta_badge ? `<span class="kmp-cart-eta">${esc(page.floating_cart.eta_badge.label || textOf(page.floating_cart.eta_badge))}</span>` : ''}
          </div>
          <button type="button" class="kmp-cart-cta" data-express-checkout="1">${esc(page.floating_cart.cta?.label || 'Checkout')}</button>
        </div>
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
    resetPreviewInteractions();
    if ($('validation-banner')) $('validation-banner').hidden = true;
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
      state.activeSlug = '';
      state.activeLayout = null;
      state.previewResult = null;
      resetPreviewInteractions();
      if ($('validation-banner')) $('validation-banner').hidden = true;
      if ($('widget-search')) $('widget-search').value = '';
      if ($('widget-type-filter')) $('widget-type-filter').value = '';
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
        state.activeSlug = '';
        state.activeLayout = null;
        state.previewResult = null;
        resetPreviewInteractions();
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
      if ($('pcf-cart-threshold')) $('pcf-cart-threshold').value = l.page_config?.floating_cart?.free_delivery_threshold_cents || '';
      if ($('pcf-cart-nudge-prefix')) $('pcf-cart-nudge-prefix').value = l.page_config?.floating_cart?.nudge_prefix || '';
      if ($('pcf-cart-unlocked')) $('pcf-cart-unlocked').value = l.page_config?.floating_cart?.unlocked_label || '';
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
      const thresholdCents = Number($('pcf-cart-threshold')?.value || 0);
      const nudgePrefix = $('pcf-cart-nudge-prefix')?.value?.trim() || '';
      const unlockedLabel = $('pcf-cart-unlocked')?.value?.trim() || '';
      l.page_config.floating_cart = $('pcf-floating-cart-enabled').checked ? {
        label: $('pcf-cart-label').value.trim() || 'Express Cart',
        eta_badge: { label: $('pcf-cart-eta').value.trim() || '8 min' },
        cta: { id: 'cart_checkout', label: $('pcf-cart-cta').value.trim() || 'Checkout' },
        free_delivery_threshold_cents: thresholdCents > 0 ? thresholdCents : undefined,
        nudge_prefix: nudgePrefix || undefined,
        unlocked_label: unlockedLabel || undefined,
      } : undefined;
      closeModal('modal-page-config');
      setDirty(true);
      renderHeader();
    });

    // Widgets filter & modal
    const syncComposedMolderBar = () => {
      const bar = $('wf-composed-molder-bar');
      if (bar) bar.hidden = $('wf-snippet-type')?.value !== 'composed_block';
    };
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
      syncComposedMolderBar();
      openModal('modal-widget');
    });
    $('wf-snippet-type')?.addEventListener('change', syncComposedMolderBar);
    $('wf-template')?.addEventListener('change', (e) => {
      const tpl = state.templates.find((t) => t.key === e.target.value);
      if (!tpl) return;
      $('wf-snippet-type').value = tpl.snippet_type;
      $('wf-props').value = JSON.stringify(tpl.default_props || {}, null, 2);
      syncComposedMolderBar();
    });
    $('btn-insert-composed-node')?.addEventListener('click', () => {
      const kind = $('wf-composed-node-kind')?.value;
      if (!kind || !CANONICAL_COMPOSED_NODE_KINDS.includes(kind)) return;
      try {
        const cur = JSON.parse($('wf-props').value || '{}');
        if (!cur.root || typeof cur.root !== 'object') {
          cur.block_id = cur.block_id || 'molded_block';
          cur.root = { kind: 'card', children: [{ kind: 'column', spacing_dp: 8, children: [] }] };
        }
        const target = Array.isArray(cur.root.children) && cur.root.children[0]?.children
          ? cur.root.children[0]
          : cur.root;
        if (!Array.isArray(target.children)) target.children = [];
        target.children.push(buildComposedNodeTemplate(kind));
        $('wf-props').value = JSON.stringify(cur, null, 2);
        $('wf-json-error').hidden = true;
        toast(`Inserted ${kind} node into composed_block`, 'info');
      } catch (err) {
        $('wf-json-error').textContent = `Cannot insert node — fix Props JSON first: ${err.message}`;
        $('wf-json-error').hidden = false;
      }
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
        syncComposedMolderBar();
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
        const base = pid in state.cartCounts ? state.cartCounts[pid] : Number(target.dataset.initQty || 0);
        state.cartCounts[pid] = base + 1;
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.cartDec) {
        const pid = target.dataset.cartDec;
        const base = pid in state.cartCounts ? state.cartCounts[pid] : Number(target.dataset.initQty || 0);
        state.cartCounts[pid] = Math.max(0, base - 1);
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.dismissNudge) {
        state.dismissedNudges[target.dataset.dismissNudge] = true;
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.dismissAlert) {
        state.dismissedAlerts[target.dataset.dismissAlert] = true;
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.composedKey) {
        state.composedSelections[target.dataset.composedKey] = target.dataset.composedVal;
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.composedToggle) {
        const key = target.dataset.composedToggle;
        state.composedSelections[key] = target.dataset.composedCur === '1' ? '0' : '1';
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.swipeDeck) {
        const deckId = target.dataset.swipeDeck;
        const dir = target.dataset.swipeDir || 'like';
        const title = target.dataset.swipeTitle || 'card';
        state.deckIndex[deckId] = (state.deckIndex[deckId] || 0) + 1;
        const labels = { like: 'Liked', nope: 'Passed', super_like: 'Super Liked' };
        toast(`${labels[dir] || 'Swiped'} ${title}`, dir === 'nope' ? 'info' : 'success');
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.swipeUndo) {
        const deckId = target.dataset.swipeUndo;
        state.deckIndex[deckId] = Math.max(0, (state.deckIndex[deckId] || 0) - 1);
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.swipeReset) {
        const deckId = target.dataset.swipeReset;
        state.deckIndex[deckId] = 0;
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.searchClear) {
        const searchId = target.dataset.searchClear;
        state.searchQueries[searchId] = '';
        if (state.previewResult?.page) renderDevicePreview(state.previewResult.page);
        return;
      }
      if (target.dataset.searchSubmit) {
        const searchId = target.dataset.searchSubmit;
        const q = (state.searchQueries[searchId] || '').trim();
        toast(q ? `Preview search: "${q}"` : 'Preview search submitted', 'info');
        return;
      }
      if (target.dataset.chipGroup) {
        const groupId = target.dataset.chipGroup;
        const chipId = target.dataset.chipId;
        const isMulti = target.dataset.chipMulti === '1';
        let set = state.chipSelections[groupId];
        if (!set) {
          set = new Set();
          document.querySelectorAll(`[data-chip-group="${CSS.escape(groupId)}"].selected`).forEach((btn) => {
            if (btn.dataset.chipId) set.add(btn.dataset.chipId);
          });
          state.chipSelections[groupId] = set;
        }
        if (isMulti) {
          if (set.has(chipId)) set.delete(chipId);
          else set.add(chipId);
        } else {
          set.clear();
          set.add(chipId);
        }
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

    // Live interactive preview input bindings for search_bar snippets
    document.addEventListener('input', (e) => {
      const inp = e.target.closest?.('[data-search-input]');
      if (!inp) return;
      const searchId = inp.dataset.searchInput;
      state.searchQueries[searchId] = inp.value;
      const wrap = inp.closest('[data-search-wrap]');
      const clearBtn = wrap?.querySelector('[data-search-clear]');
      if (clearBtn) clearBtn.hidden = !inp.value;
    });
    document.addEventListener('change', (e) => {
      const sel = e.target.closest?.('[data-composed-select]');
      if (!sel) return;
      state.composedSelections[sel.dataset.composedSelect] = sel.value;
    });
    document.addEventListener('keydown', (e) => {
      const inp = e.target.closest?.('[data-search-input]');
      if (!inp || e.key !== 'Enter') return;
      e.preventDefault();
      const q = inp.value.trim();
      toast(q ? `Preview search: "${q}"` : 'Preview search submitted', 'info');
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
