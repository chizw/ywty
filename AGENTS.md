# AGENTS.md

Guidance for AI agents working in this repo. Docs (`README.md`, `CONTRIBUTING.md`) are zh-CN; commit messages follow Conventional Commits (`feat:` / `fix:` / ...).

## Layout

- `server-rust/` — Rust + axum + sqlx backend (Cargo workspace)
- `web-astro/` — Astro 5 SSR frontend (Vue/React Islands); replaced the old Nuxt 3 app (`web-nuxt/`, deleted)
- Root `package.json` only holds Playwright for scratch scripts; it is not part of the app.
- `deploy/` — deploy scripts, nginx/supervisor templates, prod docker-compose.

## Backend (server-rust/)

- Workspace members are **`crates/*` only** (the stale top-level `server-rust/src/` duplicate was removed). All live code is in:
  - `crates/api` — thin binary shell (`main.rs` just calls `core_lib::app::run`)
  - `crates/core_lib` — everything else: `handlers/`, `services/`, `models/`, `dto/`, `middleware/`, `auth/`
- Keep the `api` crate thin. Large binary crates trigger STATUS_ACCESS_VIOLATION during codegen on Rust 1.97/LLVM 22; putting logic in the lib crate avoids it.
- Migrations live in `crates/core_lib/src/db/migrations/` (SQLite, default) and `crates/core_lib/src/db/migrations-mysql/` (MariaDB/MySQL ≥10.6), embedded via `sqlx::migrate!` in `db::MIGRATOR` and run automatically at startup. Never edit an applied migration — add a new pair to **both** dirs. Dialect is a compile-time feature: default = SQLite; `cargo build --release --bin api --features mysql` builds the MySQL variant. The Docker image ships **both** binaries (`api-sqlite` / `api-mysql`) and the entrypoint picks one by `DB_DRIVER` at runtime — single image, no flavor tags. Timestamps must be written via `db::now_str()` (`YYYY-MM-DD HH:MM:SS`, fits both TEXT and DATETIME columns); use `db::last_id(&result)` for insert IDs and `db::sql_insert_ignore()` for conflict-skipping inserts. PostgreSQL is NOT supported.
- Config: loads `config.yaml` from CWD (so run cargo from `server-rust/`), then short env vars override (see Feature status notes). `dotenvy` reads `server-rust/.env`.
- Build/test do not require any env (config.yaml ships a default JWT secret); CI sets `JWT_SECRET` anyway.
- **Windows toolchain gotcha**: PATH may resolve `cargo`/`rustc` to a standalone install (`C:\Program Files\Rust stable LLVM 1.97`) whose msvc std is broken. Prepend rustup's bin: `$env:PATH = "$env:USERPROFILE\.cargo\bin;$env:PATH"` before cargo commands. `.cargo/config.toml` pins target `x86_64-pc-windows-msvc`.
- Verify commands (run from `server-rust/`):
  ```
  cargo fmt --all
  cargo clippy --workspace --all-targets -- -D warnings   # RUSTFLAGS -D warnings in CI too
  cargo test --workspace
  ```
- On first startup with an empty users table, a default admin is seeded: `admin` / `admin123456`.

## Frontend (web-astro/)

- Scripts: `dev`, `build`, `preview`, `astro`. There is **no lint or test script**. Typecheck: `npx astro check`.
- Astro 5 SSR (node adapter, output `server`). Public pages are **Vue 3 Islands** (`src/components/vue/`); user center + admin are **React 19 Islands** (`src/components/react/`, shadcn/ui). State: Pinia (Vue) / Zustand (React).
- API calls go to relative `/api/**`. In dev, vite proxy forwards `/api`, `/uploads`, `/i/`, `/s/`, `/healthz` to the Rust backend (`API_INTERNAL`, default `http://127.0.0.1:3000`). In production the same prefixes must be reverse-proxied by `src/middleware.ts` (vite proxy does NOT work in build output — keep both in sync when adding prefixes).
- Port gotcha: both the Rust API and the Astro server can default toward port 3000. For side-by-side dev set `PORT` for the backend and point `API_INTERNAL` at it before `npm run dev`; Astro dev listens on 4321.
- Route guards live in `src/middleware.ts`: unauthenticated `/dashboard/**` and `/admin/**` redirect to `/auth/login`; non-admins are kicked out of `/admin`. Keep new admin/user-center pages under those paths.
- Auth cookie is `ywty.auth`: `encodeURIComponent(JSON.stringify({access_token, refresh_token, user}))` — parsed in `src/lib/auth.ts`.

## CI (.github/workflows/ci.yml)

- Backend job: fmt → clippy `-D warnings` → build → test, all with `JWT_SECRET` set.
- Frontend job: `npx astro check` then `npm run build`, both required (run in `web-astro/`).
- Docker job builds the root merged image (API + Web in one container), no push.

## Feature status notes

- SMS / phone verification chains were removed (email-only verify codes); users.phone column remains as an optional login identifier.
- License activation and image-scan drivers were removed; image processing is local thumbnails/watermark only.
- Storage drivers: local/s3/oss/cos/qiniu implemented and wired into upload; direct-upload signing in `services/storage.rs`.
- Payments: `services/payment.rs` defines the `PaymentDriver` trait + Mock driver; order notify requires HMAC signature (`X-Signature`) + amount check + unpaid→paid state machine. Secret reuses the JWT secret (`handlers/order.rs` `payment_driver()` is the single place to change).
- Global settings live in the `settings` table (migration 0009), admin UI at `/admin/settings`; mail SMTP config falls back to `config.yaml`. Keys whitelisted in `services/settings.rs`.
- Quotas: `groups.max_storage` + per-user `users.quota_override` (bytes); resolution in `services/capacity.rs::effective_limit_bytes`, enforced at upload.
- Branding (site name/description/footer/ICP) comes from settings via public `GET /api/v1/site/info`; frontend fetches it in `web-astro/src/lib/site.ts` — do not hardcode "云雾图驿" in pages.
- Short env vars (`PORT`, `HOST`, `APP_URL`, `JWT_SECRET`, `DB_*`, `REDIS_*`, `STORAGE_*`, `RATELIMIT_ENABLE`) override config.yaml in `config.rs::apply_env_overrides` — this is the ONLY env mechanism (the old `YWTY_*` prefix was removed). Docker entrypoint auto-generates + persists the JWT secret to `/app/data/.jwt_secret`; Astro's internal backend address is `API_INTERNAL`. Compose deploy needs zero env config.

## Migration gotcha

sqlx validates checksums of applied migrations; editing any applied `NNNN_*.sql` makes every startup log `migration N was previously applied but has been modified` and silently SKIP all newer migrations (they are only logged as a warning, non-fatal!). If settings/quota columns go missing at runtime, this is why. For dev, move the stale db aside and let a fresh one be created.
