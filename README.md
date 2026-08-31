# scribe

Multi-creator blog platform where creators share one domain — like a YouTube for blog posts.
Write in markdown, publish with a click.

Built with Go/Gin backend, React frontend, and fully containerized for self-hosted deployment.

## Tech Stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.27, Gin, GORM, JWT |
| **Frontend** | React 18, Vite, TypeScript, Tailwind CSS |
| **Database** | PostgreSQL 16 |
| **Cache** | Redis 7 |
| **Storage** | MinIO (object storage for images) |
| **Deployment** | Docker Compose, self-hosted registry |

---

## Quick Start (Local Testing)

```bash
# 1. Clone
git clone <repo-url> scribe
cd scribe

# 2. Copy env file
cp .env.example .env

# 3. Start all services
docker compose up -d
```

Open **http://localhost** in your browser.

### Test Creator Account

| Field | Value |
|---|---|
| Login page | http://localhost/login |
| Email | `creator@test.com` |
| Password | `Test123!` |

---

## Deployment Guide

### Prerequisites

Before deploying, you need:

- **Docker** and **Docker Compose** installed on your server
- **A domain name** (optional, but recommended for production)
- **A reverse proxy** (like Nginx, Caddy, or Traefik) if you want HTTPS

---

### Option A: Deploy with Docker Compose (Simplest)

Best for: single-server deployment, homelab, or VPS.

**Step 1 — Clone on your server**

```bash
git clone <repo-url> scribe
cd scribe
```

**Step 2 — Configure environment**

```bash
cp .env.example .env
nano .env
```

At minimum, change these values:

```env
DB_PASSWORD=your-strong-password
JWT_SECRET=your-random-secret-key
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your-minio-password
```

Generate a strong JWT secret:

```bash
openssl rand -base64 32
```

**Step 3 — Set up a reverse proxy for HTTPS**

Example with Caddy (create `Caddyfile`):

```caddyfile
blog.yourdomain.com {
    reverse_proxy frontend:80
}
```

Then run Caddy:

```bash
docker run -d \
  --name caddy \
  --network scribe_blog-network \
  -p 80:80 -p 443:443 \
  -v $PWD/Caddyfile:/etc/caddy/Caddyfile \
  -v caddy_data:/data \
  caddy:latest
```

> Or use Nginx, Traefik, or whatever you prefer. The frontend serves on port 80.

**Step 4 — Start the stack**

```bash
docker compose up -d
```

**Step 5 — Access MinIO console (optional)**

Open http://your-server-ip:9001 to manage uploaded images.
Login with the credentials from your `.env` file.

---

### Option B: Deploy with Coolify

Best for: easier management, web UI, automatic HTTPS.

Coolify is an open-source PaaS that runs on your own server.

**Step 1 — Clone the repo in Coolify**

- In Coolify, create a new project
- Add a new resource → "Docker Compose"
- Point it to your Git repository
- Coolify will detect the `docker-compose.yml`

**Step 2 — Set environment variables**

In Coolify's environment variables section, add:

```env
DB_PASSWORD=your-strong-password
JWT_SECRET=your-random-secret-key
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your-minio-password
```

**Step 3 — Set up a GitHub Actions runner (for CI/CD)**

If you want automatic deployments on `git push`, you need a runner.

**Using the official Docker runner image:**

In Coolify, deploy a new service with:

- **Image:** `myoung34/github-runner:latest`
- **Environment variables:**

```env
RUNNER_NAME=scribe-runner
RUNNER_REPOSITORY_URL=https://github.com/your-org/scribe
RUNNER_TOKEN=your_github_pat
RUNNER_LABELS=self-hosted,docker
RUNNER_WORK_DIRECTORY=/tmp/runner
```

- **Volume:** `/var/run/docker.sock:/var/run/docker.sock` (so the runner can build Docker images)
- **Restart policy:** `no` (start it manually only when you need to deploy)

> **Getting a GitHub PAT:**
> 1. Go to GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens
> 2. Scope: `Repository` → select your repo → `Read and write` for `Administration` (so it can register itself)
> 3. Copy the token and use it as `RUNNER_TOKEN`

**Step 4 — Set up GitHub Actions secrets**

In your GitHub repo → Settings → Secrets and variables → Actions, add:

| Secret | Value |
|---|---|
| `REGISTRY_URL` | Your registry host (e.g., `registry.example.com`) |
| `REGISTRY_USERNAME` | Registry login username |
| `REGISTRY_PASSWORD` | Registry login password |

**Step 5 — Deploy**

Push to `main` — the workflow will:
1. Build the Docker images
2. Push them to your registry
3. (You manually update the server by pulling the new images)

Or trigger manually: GitHub → Actions → Deploy → Run workflow.

---

### Option C: Manual Deployment with CI/CD

Best for: full control, learning, custom setups.

This is the same as Option A, but you add CI/CD on top.

**Architecture:**

```
Git push → GitHub Actions → self-hosted runner
                                ↓
                        builds Docker images
                                ↓
                        pushes to registry
                                ↓
                        pull on server → docker compose up -d
```

**Step 1 — Set up a Docker registry**

You need a place to store your built images. Options:

- **Self-hosted registry** (like this project uses):
  ```bash
  docker run -d \
    --name registry \
    -p 5000:5000 \
    -v registry_data:/var/lib/registry \
    registry:2
  ```

- **Docker Hub** — simpler but public (unless you pay)

- **GitHub Container Registry (ghcr.io)** — free for public repos

**Step 2 — Set up a self-hosted runner**

On your server, run the runner container:

```bash
docker run -d \
  --name github-runner \
  -e RUNNER_NAME=my-runner \
  -e RUNNER_REPOSITORY_URL=https://github.com/your-org/scribe \
  -e RUNNER_TOKEN=your_github_pat \
  -e RUNNER_LABELS=self-hosted,docker \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --restart no \
  myoung34/github-runner:latest
```

When you want to deploy, start the container:

```bash
docker start github-runner
```

When done, stop it:

```bash
docker stop github-runner
```

**Step 3 — Configure GitHub Actions secrets**

Same as Option B Step 4.

**Step 4 — Deploy**

Push to `main` or trigger the workflow manually.

---

## Environment Variables Reference

| Variable | Default | Description |
|---|---|---|
| `DB_PASSWORD` | `blogpassword` | PostgreSQL password |
| `JWT_SECRET` | `dev-secret-change-in-production` | Secret key for JWT tokens (change in production!) |
| `MINIO_ROOT_USER` | `minioadmin` | MinIO admin username |
| `MINIO_ROOT_PASSWORD` | `minioadmin` | MinIO admin password |
| `REGISTRY_URL` | — | Your Docker registry host (for CI/CD) |
| `REGISTRY_USERNAME` | — | Registry login username |
| `REGISTRY_PASSWORD` | — | Registry login password |

---

## Production Checklist

Before going live, make sure to:

- [ ] **Change `JWT_SECRET`** — use a long random string (`openssl rand -base64 32`)
- [ ] **Change `DB_PASSWORD`** — use a strong password
- [ ] **Change `MINIO_ROOT_PASSWORD`** — don't leave it as default
- [ ] **Set up HTTPS** — use Caddy, Nginx + Let's Encrypt, or Traefik
- [ ] **Limit MinIO console access** — port 9001 should not be public
- [ ] **Set up backups** — the PostgreSQL volume contains all your data
- [ ] **Remove the test creator** — or change their password after first login

---

## CI/CD

The project uses `ci.sh` — a single script that works with any CI platform:

```bash
./ci.sh test              # Go vet/build + frontend build
./ci.sh docker-build      # Build Docker images
./ci.sh docker-push       # Push images to registry
./ci.sh docker-build-push # Build + push in one step
```

The GitHub Actions workflows in `.github/workflows/` call this script.
To switch to GitLab CI, Woodpecker, or Drone, just create a new workflow file that calls `./ci.sh` — no changes to the build logic needed.

---

## Project Structure

```
scribe/
├── backend/              # Go/Gin API server
│   ├── cmd/server/       # Entry point, routes, seeding
│   └── internal/
│       ├── config/       # Environment config
│       ├── database/     # PostgreSQL connection
│       ├── handlers/     # Auth, Posts, Categories, Upload
│       ├── middleware/    # JWT auth, CORS
│       ├── models/       # Creator, Post, Category
│       └── storage/      # MinIO, Redis clients
├── frontend/             # React/Vite SPA
│   └── src/
│       ├── api/          # API client (fetch-based)
│       ├── components/   # Navbar, PostCard, CategoryFilter, etc.
│       ├── pages/        # Landing, Directory, PostView, Login, Dashboard
│       └── context/      # Auth context
├── .github/workflows/    # CI/CD pipelines
│   ├── test.yml          # Runs on PR/push: go vet + build, npm build
│   └── deploy.yml        # Builds & pushes Docker images to registry
├── ci.sh                 # Platform-agnostic CI/CD script
├── docker-compose.yml    # 5 services: postgres, redis, minio, backend, frontend
└── .env.example          # Environment template
```

## API Endpoints

### Public
- `GET /api/v1/posts?page=1&limit=12&category=slug` — Paginated published posts
- `GET /api/v1/posts/:slug` — Single post
- `GET /api/v1/categories` — Categories with post counts

### Auth
- `POST /api/v1/auth/login` — Login (email + password)
- `GET /api/v1/auth/me` — Current creator profile

### Creator (requires JWT)
- `GET /api/v1/creator/posts` — Creator's own posts (including drafts)
- `POST /api/v1/creator/posts` — Create post
- `PUT /api/v1/creator/posts/:id` — Update post
- `DELETE /api/v1/creator/posts/:id` — Delete post
- `POST /api/v1/creator/upload` — Upload image (to MinIO)

## Development

```bash
# Backend
cd backend
go mod tidy
go run ./cmd/server

# Frontend (separate terminal)
cd frontend
npm install
npm run dev        # Proxies /api to localhost:8080
```

The frontend dev server automatically proxies `/api` requests to the backend at `localhost:8080`.

## License

MIT