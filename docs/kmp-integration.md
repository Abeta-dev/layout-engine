# `abeta-dev/kmp` Client Integration & Wire Protocol Guide

**Abeta Dynamic Layout Engine (`v0.3.0`)** consumes [`github.com/abeta-dev/kmp/bff@v0.5.0`](https://github.com/Abeta-dev/kmp) (`pkg/sdui/{actions,atoms,snippets,theming}`) and [`github.com/abeta-dev/go-libs@v0.4.0`](https://github.com/Abeta-dev/go-libs) (`ginmw.Envelope`) directly. Because both repositories share the exact same Go `v1` wire contracts, any **`abeta-dev/kmp`** client can switch from the static sample BFF to the Dynamic Layout Engine by changing a single base URL.

---

## 1. Drop-In Replacement for `abeta-dev/kmp` Apps

### Start the Layout Engine

```bash
go run ./cmd/layout-engine
# Listening on http://localhost:8090 (Visual Studio at /studio, Docs at /docs)
```

### Point `kmp/sample-app` at the Layout Engine

From your `abeta-dev/kmp` checkout:

```bash
(cd kmp && ./gradlew :sample-app:run --args="--bff-url=http://localhost:8090")
```

### Point `kmp/storybook` (`LiveBffStory`) at the Layout Engine

In `kmp/storybook` (Desktop JVM, macOS Cocoa, iOS `AbetaStorybookKit`, or WasmJS), open the **Live BFF** story and set the base URL to `http://localhost:8090`. Select screens across all **7 vertical workspaces** (`store-home`, `fresh-home`, `bites-home`, `pulse-home`, `mart-home`, `spark-deck`, `trade-home`, `pay-home`, `pay-transfer`, `pay-cards`, `pay-receipt`, `watch-status`) to render live dynamic layouts edited in `/studio`.

### Point `kmp/watch-companion` (`AbetaWatchKit` / Wear OS) at the Layout Engine

`RemoteWatchPageSource` fetches `GET /api/v1/screens/watch-status` (or any published screen slug) and maps the returned `PageResponse` via `WatchScreenMapper` into compact wearable glance cards for watchOS and Wear OS.

---

## 2. 1:1 `v1` SDUI Wire Contract Parity

| Contract Surface | Source Package (`github.com/abeta-dev/kmp/bff@v0.5.0`) | Supported Elements |
| :--- | :--- | :--- |
| **21 Snippet Types** | `pkg/sdui/snippets` | `header_v1`, `banner_v1`, `metric_card`, `card_v1`, `action_row`, `form_v1`, `empty_state`, `divider`, `spacer`, `carousel`, `grid`, `tabs`, `accordion`, `product_card`, `search_bar`, `chip_group`, `swipe_deck`, `nudge_bar`, `marquee_strip`, `order_tracker`, `composed_block` (29 `ComposedNodeKind`s, `maxDepth <= 6`) |
| **28 Wire Atoms / 30 UI Atoms** | `pkg/sdui/atoms` | Text, Icon, Tag, Button, Image, TextField, Divider, Spacer, Progress, Switch, Checkbox, Slider, Avatar, Skeleton, Stepper (`StepperAtom`), Countdown (`CountdownAtom`), plus 12 extended atoms (`RadioGroup`, `Select`, `OtpInput`, `Chip`, `SegmentedControl`, `Rating`, `Price`, `Badge`, `AlertBanner`, `StatTile`, `KeyValueRow`, `QrCode`) + `MotionSpec` (7 ambient `MotionLoopEffect`s), `LayoutConfig`, `TrackingMeta` |
| **18 Action Types** | `pkg/sdui/actions` | `navigate`, `back_press`, `open_dialog`, `dismiss_overlay`, `show_toast`, `call_api`, `save_state`, `delete_state`, `track_event`, `refresh_page`, `copy_to_clipboard`, `update_snippets`, `batch`, `trigger_feedback`, `open_url`, `share`, `request_permission`, `biometric_auth` |
| **Page Chrome & Adaptive Specs** | `pkg/sdui/snippets` | `ToolbarSpec`, `BottomBarSpec`, `FloatingCartSpec` (incl. free-delivery threshold nudge fields), `AdaptivePageSpec` (`single_column`, `centered`, `split_pane`), `ResponsiveColumnsSpec`, `ResponsiveIntSpec`, `PaginationSpec`, `PullToRefresh`, `PageActions` |
| **Design Token Catalog** | `pkg/sdui/theming` | Base token catalog merged dynamically with rule-matched `TokenOverride` entries |
| **Response Envelope** | `github.com/abeta-dev/go-libs/ginmw` | Standard `ginmw.Envelope` (`success`, `data`, `error`, `meta.request_id`) across all `/api/v1/*` routes |

---

## 3. Runtime Endpoint Mapping for KMP Clients

| KMP Client Call | Layout Engine Endpoint | Behavior |
| :--- | :--- | :--- |
| Screen discovery (`SampleContainer`, `LiveBffStory`) | `GET /api/v1/screens` | Returns `{ "screens": [...], "tenant": "abeta-fresh" }` listing all published dynamic slugs. |
| Screen fetch (`LayoutController.load(screenId)`) | `GET /api/v1/screens/:screenId` | Resolves the active `PublishedSnapshot` for `:screenId`, evaluates rules & slotting, applies fallback if needed, sets `X-Layout-Version` (and `X-Layout-Fallback: true` on failover), and returns `PageResponse` inside `ginmw.Envelope`. |
| Debug / telemetry screen fetch | `GET /api/v1/pages/:slug?include_trace=true` | Returns `{ "page": PageResponse, "trace": ResolutionTrace }` (or raw `PageResponse` without `?include_trace=true`). |
| Launch splash config | `GET /api/v1/config/splash` | Returns the highest-priority enabled `SplashConfig` matching the caller's rules and UTC schedule (falling back to the default splash). |
| Runtime theme & token catalog | `GET /api/v1/catalog/tokens` | Returns the base `v1` design token catalog merged with any matching `TokenOverride`s. |
| Remote action execution (`ApiCallExecutor`) | `POST /api/v1/actions/execute` | Executes `ActionTrigger` payloads (all 18 `v1` action types) and returns follow-up client actions. |
| Quick-commerce express checkout | `POST /api/v1/orders/express` | Places an express order and emits celebration (`trigger_feedback`) + confirmation (`show_toast`) follow-ups consumed by `sdui-layout`'s `UiCelebrationOverlay`. |

---

## 4. Sending Targeting Headers from KMP `NetworkConfig`

To unlock per-device, per-city, semver, and percentage-rollout targeting automatically from your KMP app, attach standard context headers in `kmp/network` (`NetworkConfig.defaultHeaders`) or per-request headers in your `RemotePageSource`:

```kotlin
val networkConfig = NetworkConfig(
    baseUrl = "http://localhost:8090", // or https://layout.your-domain.example
    defaultHeaders = mapOf(
        "X-Tenant-ID" to "abeta-fresh",
        "X-Platform" to platformInfo.kind.name.lowercase(), // android, ios, macos, watchos, wearos, web, desktop
        "X-App-Version" to appVersion,                      // e.g. "2.1.0"
        "X-City" to userLocation.city,                      // e.g. "Bengaluru"
        "X-Zone" to userLocation.zone,                      // e.g. "koramangala-01"
        "X-User-Segment" to session.segment,                // e.g. "vip"
        "X-User-ID" to session.userId,                      // stable hash key for bucket_pct rollouts
        "X-Window-Class" to adaptiveMetrics.widthClass.name.lowercase(), // compact, medium, expanded
        "X-Dark-Mode" to isDarkTheme.toString(),
    ),
)
```

### Verifying Context Targeting via `curl`

```bash
# Compact mobile client in Bengaluru (VIP segment)
curl -s "http://localhost:8090/api/v1/screens/store-home" \
  -H "X-Tenant-ID: abeta-fresh" \
  -H "X-Platform: ios" \
  -H "X-App-Version: 2.1.0" \
  -H "X-City: Bengaluru" \
  -H "X-User-Segment: vip" \
  -H "X-User-ID: usr_1001" \
  -H "X-Window-Class: compact" \
  -i | head -n 20
```

Look for `X-Layout-Version`, `ETag`, and `Content-Encoding: br` (when passing `--compressed` or `Accept-Encoding: br, gzip`).

---

## 5. Transport, Caching & Offline Synergy with KMP

The Layout Engine's transport optimizations pair directly with `abeta-dev/kmp`'s networking and local database layers:

- **HTTP/2 & Brotli (`kmp/network`)**: `KtorHttpTransport` negotiates HTTP/2 ALPN and `Accept-Encoding: br, gzip`; the Layout Engine responds with Brotli-compressed payloads (`COMPRESSION_ENABLED=true`) to minimize mobile cellular payload size.
- **ETag Conditional GETs (`ETAG_ENABLED=true`)**: Unchanged rendered pages return `304 Not Modified`, cutting bandwidth on pull-to-refresh and app-resume checks.
- **Offline Cache & Room 3 (`kmp/database` + `kmp/sdui-layout`)**: Pair `CachedScreenRepository` (backed by Room 3 `AbetaDatabase`'s `cached_screens` table) with the Layout Engine's server-side `FallbackLayout` chain so users get **two layers of resilience**:
  1. **Server-side failover**: Layout Engine serves pre-validated `FallbackLayout` (`X-Layout-Fallback: true`) if a snapshot fails to compose.
  2. **Client-side offline failover**: KMP serves the last cached `PageResponse` from SQLite/OPFS whenever device connectivity drops (`ConnectivityMonitor` -> `OfflineStatusBanner`) and replays queued mutations via `OfflineActionCoordinator` upon reconnect.

---

## Related Guides

- [Clean Architecture & Persistence](architecture.md)
- [Visual Layout Studio, Rules & Slotting](studio-and-rules.md)
- [Repository Overview](../README.md) · [AI Agent Contract](../AGENTS.md)
