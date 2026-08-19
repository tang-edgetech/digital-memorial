# Digital Memorial — Cemetery / Columbarium Management System

## Objective

A backoffice system for managing tombstones, columbarium niches, ancestor plaques, and inscriptions — similar in scope to Nirvana Memorial Park's cemetery management systems. Staff (Super Admin / Admin / Agent) use it to record, search, and maintain records tied to physical plots, niches, and plaques, including the inscription text engraved on each.

## Platforms

Two codebases make up the whole system for now — no public-facing site is in scope:

1. **Backoffice** — Next.js + Tailwind CSS + Ant Design
2. **API** — Go + MySQL

## Deployment Targets

- **Local development:** XAMPP, running on this machine. XAMPP supplies MySQL (and the project happens to live under XAMPP's `htdocs`), but Apache does not serve the Next.js/Go apps the way it would serve PHP — the frontend (`next dev`) and API (`go run`) run as their own local processes on their own ports (e.g. `:3000` and `:8080`) and talk to XAMPP's MySQL directly. No Apache reverse-proxy is required for local dev, though one can be added later for convenience.
- **Production:** cPanel or a self-hosted Linux VPS.
  - On a **VPS with root/SSH access** (self-hosted, or "VPS with cPanel/WHM" rather than shared hosting): Nginx or Apache reverse-proxies to the Next.js process (`next start`, kept alive via PM2 or a systemd unit) and the compiled Go binary (its own systemd unit); MySQL runs natively on the box; TLS via Let's Encrypt/certbot. This is the straightforward path and is recommended.
  - On **shared cPanel hosting** (no root/SSH): cPanel's "Setup Node.js App" (Passenger) can host the Next.js app directly. The Go API is the open risk here — typical shared cPanel plans don't support running an arbitrary persistent background binary/service (no systemd, no root). ⚠️ Needs confirming against the actual hosting plan before Phase 1 production deploy: if it's truly shared hosting, the Go API may need to move behind a supported mechanism (e.g. a long-running process manager cPanel explicitly allows) or the hosting plan needs to be a VPS-backed one instead.

## Development Roadmap

Built phase by phase; each phase should be usable/demoable on its own before moving to the next. Enhancements can be slotted in mid-phase as they come up — this list is expected to evolve.

### Phase 1 — Foundation & Setup
- Project scaffolding: Next.js + Tailwind + AntD shell; Go API skeleton; MySQL connection.
- `/setup` first-run wizard: DB connection + schema init, create Super Admin, set site title/logo.
- Site settings table + settings page.
- Auth: `/login` (route configurable later), session/token handling.
- App shell: collapsible sidebar (hamburger), per-user Light/Dark theme (not on login page).
- Shared interaction patterns to reuse everywhere from here on: AJAX-driven actions, Toast notifications, confirmation-before-execute prompts, icon buttons with hover tooltips.
- Audit log infrastructure (table + logging hook/middleware) — built early so every later phase logs through it from day one.

### Phase 2 — User & Access Management *(done)*
- User Listing page (Super Admin and Admin each get one; filter, sort, bulk actions, quick actions).
- Create/Edit user as full pages: `/users/create`, `/users/:id/edit`.
- Roles: Super Admin (hidden from Admin/Agent views), Admin (one flagged as Owner — only the Owner can remove other Admins), Agent.
- Enable/disable account toggle (available to roles with sufficient permission).
- Roles × permissions matrix (granular, per-module/per-action), editable by Super Admin and Owner.

### Phase 3 — Master Data
- Provinces: seed with Guang Dong, Guang Xi, Hokkien, Hakka, Hainan; Admin/Agent can add more.
- Other master/reference data as identified.

### Phase 4 — Core Cemetery Modules *(not yet defined)*
- Tombstones, Columbarium, Ancestor Plaques, Inscriptions.
- Blocked on: data model and relationships between these entities (and any customer/family/deceased-person record) — to be scoped when we get here.

### Later / Unscheduled
- Anything raised mid-development that isn't urgent enough to reprioritize into the current phase goes here until picked up.

## Cross-Cutting Requirements

These apply across every phase, not just one module:

- **Interaction model:** all functional actions run via AJAX (no full page reloads for actions); results/errors surface as **Toast** notifications; every state-changing action (create/update/delete/enable-disable/etc.) requires a **confirmation prompt** before executing.
- **Security:** standard-to-strong measures — CSRF protection, input validation/sanitization, rate limiting, secure session/token handling, audit logging.
- **Listing pages:** every list view needs filtering, column sorting, bulk actions, and per-row quick-action buttons.
- **Buttons:** icon-based, with hover tooltips showing the action label.
- **Editing:** full-page forms for create/edit, not modals.
- **Theme:** Light/Dark, per-user, applied everywhere except the login page.

## Site Settings

Stored in a dedicated settings table (key/value or structured), editable by Super Admin/Admin. Suggested items beyond title/logo:

- Site title, logo, favicon
- Company/organization name, address, contact phone, email
- Login route (customizable)
- Default timezone, date format
- Default language (if multi-language is ever added)
- Default theme for new users (light/dark)
- Session timeout duration
- Password policy (min length, complexity, expiry)
- Failed-login lockout threshold / lockout duration
- File upload limits (max size, allowed types)
- Pagination default (rows per page on listing screens)
- SMTP settings (for system emails/notifications)
- Maintenance mode toggle
- Audit log retention period

## Audit Logs

Every significant action (create/update/delete, permission changes, login/logout, enable/disable) is logged for the dev team to trace and debug — actor, action, target record, before/after values (where practical), timestamp, IP/user agent.

## Open Questions

1. Data model for the Phase 4 entities (Tombstone, Columbarium, Ancestor Plaque, Inscription) — deferred until Phase 4; not blocking Phase 1–3 work.
2. Does an Agent need to be scoped to specific plots/sections/branches, or do they see everything they have permission for?
3. Any payment/billing module (purchase of plots/niches, maintenance fees), or is this purely records management?
4. Deployment target (on-prem server, cloud VM, etc.) and whether this replaces or coexists with an existing system/data.
5. Exact permission granularity expected — module-level (view/create/edit/delete) is the common baseline; confirm if finer control (e.g. field-level, per-record) is needed.

## Running Locally

1. **API** (`api/`): `go run ./cmd/server` (from the `api/` directory). Boots even with no `.env` — it just runs in "unconfigured" mode until `/setup` is completed. Listens on `:8080` by default.
2. **Web** (`web/`): `npm run dev` (from the `web/` directory). Listens on `:3000` by default; copy `.env.local.example` to `.env.local` if the API isn't on the default URL.
3. Visit `http://localhost:3000` — you'll be redirected to `/setup` automatically until the wizard completes (DB connection → Super Admin → site settings), then to `/login`.
4. XAMPP's MySQL (`mysqld`, port `3306`) just needs to be running; the setup wizard creates the database itself (`CREATE DATABASE IF NOT EXISTS`) if it doesn't already exist.

## Phase 1 Implementation Notes

Phase 1 (Foundation & Setup) is implemented and has been verified end-to-end (setup wizard, login, sliding session/idle-timeout, CSRF, audit logging). A few simplifications were made versus the original plan, to be revisited later:

- **Logo/favicon** are plain text path/URL fields in the setup wizard and settings page, not a file upload — actual upload handling is deferred to a follow-up enhancement rather than Phase 1.
- **Login route customization** is stored and editable in Site Settings, but changing it doesn't yet move the actual `/login` page (no catch-all redirect-alias route yet) — `/login` remains the only working path. Full alias-redirect support is deferred.
- **No `docker-compose.yml`** for local MySQL — local dev uses XAMPP's MySQL directly per the Deployment Targets above, so a containerized DB would just be redundant.
- Frontend/backend run on separate ports (`:3000` / `:8080`) sharing cookies via the browser's port-agnostic cookie scoping (both are `localhost`). In production, reverse-proxying both under one domain (e.g. `/api/*` → Go, everything else → Next.js) is simpler and avoids relying on that behavior — see Deployment Targets.

## Phase 2 Implementation Notes

Phase 2 (User & Access Management) is implemented; the backend rules (owner assignment/transfer, admin-on-admin protection, self-protection, super_admin row hiding, live permission-matrix updates, bulk actions, audit logging) were all verified end-to-end via direct API testing. The frontend builds/lints/typechecks cleanly and the pages/routes are wired up, but the UI itself (forms, table filtering/sorting, confirm dialogs) has **not** been manually clicked through in a browser — worth a pass before considering this fully done from a UX standpoint.

- Editing your own account through `/users/:id/edit` is blocked entirely (not just role changes) — there's no "my profile" self-service page yet; that'd be a separate future feature if needed.
- The `permissions` module (managing the matrix itself) is intentionally hardcoded to Super Admin/Owner and never appears as a row in the matrix — avoids a self-referential escalation path.
- `GET /api/settings` stays open to any authenticated role (Agents' idle-timeout logic depends on it); only `PUT /api/settings` is permission-gated.

## Status

**Phase 1 (Foundation & Setup): done.** Next.js + Go scaffolding, `/setup` wizard, site settings (with the ~27 suggested fields seeded), JWT auth with sliding 2-hour idle-timeout (server-enforced + client-side idle modal), CSRF protection, per-IP rate limiting on login/setup, audit logging, dashboard shell (collapsible sidebar, per-user light/dark theme), and the shared Toast/Confirm/IconButtonWithTooltip patterns are all in place and build/lint/typecheck cleanly.

**Phase 2 (User & Access Management): done** (backend verified end-to-end; frontend UI not yet manually browser-tested — see notes above). User listing with filter/sort/bulk-actions/quick-actions, full-page create/edit, Owner/Admin/Agent role rules, and the roles × permissions matrix are all in place.

Phases 3–4 remain to be detailed as we approach them.
