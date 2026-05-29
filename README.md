<div><h1 align="center">guardian</h1><p align="center">A self-hosted image gallery with referrer analytics and access control. Built as a single Go binary with an embedded Vue 3 SPA. Deploy with one Docker container.</p></div>

### Features

- **Image Gallery** — Create albums, bulk import images from URLs, drag-and-drop reordering.
- **Asset Proxy** — Images are proxied through the server with automatic view tracking.
- **Referrer Analytics** — See which domains are viewing your images, with per-asset breakdowns.
- **Domain Firewall** — Ban/unban/forget referrer domains; set a redirect image for banned domains.
- **Secure Auth** — bcrypt-hashed passwords with signed HMAC session cookies (no server-side store).
- **Theme Support** — 12 selectable color schemes (Gruvbox, Nord, Catppuccin, Tokyo Night, etc.).
- **Setup Wizard** — First-boot guided setup: create admin password, pick theme (hot-swappable).

### Tech stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.26, chi, modernc SQLite |
| Frontend | Vue 3, Vite, Tailwind v4, shadcn-vue (reka-ui) |

### Quick start

```bash
docker compose up --build -d
```

Open `http://localhost:8082` and complete the setup wizard.

### Development

#### Frontend (Vite dev server)

```bash
cd dashboard
bun install
bun run dev
```

#### Backend (Go)

```bash
# Build the SPA first
cd dashboard && bun run build && cd ..

# Run the Go server (serves the built SPA)
go run .
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8082` | HTTP listen port |
| `DB_PATH` | `/app/data/assets.db` | SQLite database path |
| `DATA_PATH` | `/app/data` | Directory for `config.json` and other data |

### API endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/config` | Get app config |
| `PATCH` | `/api/config` | Update app config |
| `POST` | `/api/setup` | First-time setup |
| `POST` | `/api/auth/login` | Authenticate |
| `POST` | `/api/auth/logout` | Clear session |
| `GET` | `/api/auth/me` | Check admin status |
| `GET` | `/assets/{id}.{ext}` | Proxy an image by ID (with extension) |
| `GET` | `/assets/{id}` | Proxy an image by ID (legacy, no extension) |
| `GET` | `/art/{id}` | Proxy an image by ID (legacy route) |
| `GET` | `/api/assets/{id}/stats` | View stats for an asset |
| `GET` | `/api/albums` | List all albums |
| `POST` | `/api/albums` | Bulk import images |
| `DELETE` | `/api/albums/{album}` | Delete an album |
| `GET` | `/api/albums/{album}/images` | List images in an album |
| `PUT` | `/api/albums/{album}/reorder` | Reorder images |
| `PUT` | `/api/albums/reorder` | Reorder albums |
| `DELETE` | `/api/albums/{album}/images/{id}` | Delete an image |
| `PUT` | `/api/albums/{album}/images/{id}` | Update image source URL |
| `GET` | `/api/referrers` | List referrer domains with stats |
| `POST` | `/api/referrers/ban` | Ban a domain |
| `POST` | `/api/referrers/unban` | Unban a domain |
| `POST` | `/api/referrers/forget` | Forget a domain |
| `POST` | `/api/referrers/set-redirect` | Set redirect image URL |
| `GET` | `/api/tools` | Utility tools |
