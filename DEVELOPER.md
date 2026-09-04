# Developer Guide — Brambler

## Architecture

```
┌──────────┐     ┌──────────┐     ┌──────────────┐
│  Browser │────▶│  Nginx   │────▶│  Go Backend   │
│ (React)  │     │ (static) │     │  (Gin/GORM)   │
└──────────┘     └──────────┘     └───────┬───────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
                    ▼                     ▼                     ▼
              ┌──────────┐         ┌──────────┐         ┌──────────┐
              │PostgreSQL│         │  Redis   │         │  MinIO   │
              │ (data)   │         │ (cache)  │         │(images)  │
              └──────────┘         └──────────┘         └──────────┘
```

---

## Authentication System

### Current Auth Methods

| Method | Status | Notes |
|--------|--------|-------|
| Email + Password | ✅ Active | Test account only: `creator@test.com` / `Test123!` |
| Google SSO | ✅ Active | Primary signup/signin for real users |
| Apple SSO | 🔜 Planned | See "Adding a New Provider" below |
| GitHub SSO | 🔜 Planned | See "Adding a New Provider" below |
| Email + Password (SMTP) | 🔜 Planned | Requires SMTP server for verification |

### How Google SSO Works

```
Login Page                      Backend                     Google
    │                              │                          │
    │  Click "Sign in with Google" │                          │
    │ ──────────────────────────▶  │                          │
    │                              │  Redirect to Google      │
    │                              │ ────────────────────────▶│
    │                              │                          │
    │                              │  User approves,          │
    │                              │  Google sends code back  │
    │                              │ ◀────────────────────────│
    │                              │                          │
    │                              │  Exchange code for token │
    │                              │ ────────────────────────▶│
    │                              │  Get user info           │
    │                              │ ◀────────────────────────│
    │                              │                          │
    │                              │  Find or create Creator  │
    │                              │  Generate JWT            │
    │                              │                          │
    │  Redirect with token         │                          │
    │ ◀────────────────────────────│                          │
    │                              │                          │
    │  Store token, redirect to    │                          │
    │  dashboard                   │                          │
```

### Auth Flow (Code)

1. **Frontend** — Login page has a "Sign in with Google" button
2. **Frontend** — Click redirects to `GET /api/auth/google/login`
3. **Backend** — `GoogleLogin` handler:
   - Generates a random state string (CSRF protection)
   - Stores state in a cookie (HttpOnly, 10 min expiry)
   - Redirects to `https://accounts.google.com/o/oauth2/v2/auth`
4. **Google** — User approves, Google redirects to callback URL
5. **Backend** — `GoogleCallback` handler:
   - Verifies state cookie matches query param
   - Exchanges auth code for access token (`POST /token`)
   - Fetches user info (`GET /oauth2/v2/userinfo`)
   - Finds Creator by email, or creates a new one
   - Generates JWT (same as email/password login)
   - Redirects to `{FRONTEND_URL}/auth/callback?token=xxx`
6. **Frontend** — `AuthCallback` page reads token from URL, saves to localStorage, redirects to `/dashboard`

### JWT Token

- Signed with HS256 using `JWT_SECRET`
- Contains: `creator_id`, `email`, `exp` (24h), `iat`
- Sent as `Authorization: Bearer <token>` header

---

## Adding a New Auth Provider

To add a new SSO provider (Apple, GitHub, etc.):

### 1. Backend — Config

Edit `backend/internal/config/config.go`:

```go
type Config struct {
    // ... existing fields
    AppleClientID     string
    AppleClientSecret string
    // or
    GitHubClientID     string
    GitHubClientSecret string
}
```

### 2. Backend — Model

Edit `backend/internal/models/creator.go`:

```go
type Creator struct {
    // ... existing fields
    AppleID  string `gorm:"uniqueIndex;size:255;default:''"`
    // or
    GitHubID string `gorm:"uniqueIndex;size:255;default:''"`
}
```

### 3. Backend — Handler

Create a new handler file or add methods to `auth.go`:

```go
func (h *AuthHandler) AppleLogin(c *gin.Context) { ... }
func (h *AuthHandler) AppleCallback(c *gin.Context) { ... }
```

The pattern is always:
1. Redirect to provider's OAuth URL
2. Provider redirects back with a code
3. Exchange code for access token
4. Fetch user info (email, name, avatar)
5. Find or create Creator by email + provider ID
6. Generate JWT
7. Redirect to frontend callback

### 4. Backend — Routes

Add to `main.go`:

```go
auth.GET("/apple/login", authHandler.AppleLogin)
auth.GET("/apple/callback", authHandler.AppleCallback)
```

### 5. Config

Add to `.env.example` and `docker-compose.yml`:

```env
APPLE_CLIENT_ID=your-apple-client-id
APPLE_CLIENT_SECRET=your-apple-client-secret
```

### 6. Frontend

Add a button to `Login.tsx`:

```tsx
<button onClick={() => window.location.href = '/api/auth/apple/login'}>
  Sign in with Apple
</button>
```

---

## Adding Email + Password Registration

### Prerequisites
- SMTP server (SendGrid, Mailgun, or self-hosted)

### Steps
1. Add SMTP config to `config.go`
2. Create a `Signup` handler that:
   - Validates email + password
   - Generates a verification code
   - Sends code via email (SMTP)
   - Stores pending verification in DB (or Redis with TTL)
3. Create a `VerifyEmail` handler that:
   - Checks code matches
   - Activates the account
4. Create a `ResendCode` handler (rate limited)
5. Frontend: signup page with email + password form + verification code input

---

## Environment Variables

### Current

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://blog:blogpassword@localhost:5432/blog?sslmode=disable` | PostgreSQL connection |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO server |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO username |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO password |
| `MINIO_BUCKET` | `blog-images` | MinIO bucket |
| `MINIO_USE_SSL` | `false` | MinIO SSL |
| `JWT_SECRET` | `dev-secret-change-in-production` | JWT signing key |
| `SERVER_PORT` | `8080` | Backend port |
| `GOOGLE_CLIENT_ID` | — | Google OAuth Client ID |
| `GOOGLE_CLIENT_SECRET` | — | Google OAuth Client Secret |
| `FRONTEND_URL` | `http://localhost` | Frontend URL for redirects |

### Future

| Variable | Planned For |
|----------|-------------|
| `APPLE_CLIENT_ID` | Apple SSO |
| `APPLE_CLIENT_SECRET` | Apple SSO |
| `GITHUB_CLIENT_ID` | GitHub SSO |
| `GITHUB_CLIENT_SECRET` | GitHub SSO |
| `SMTP_HOST` | Email registration |
| `SMTP_PORT` | Email registration |
| `SMTP_USER` | Email registration |
| `SMTP_PASS` | Email registration |

---

## Development

### Quick Start

```bash
# Copy env
cp .env.example .env

# Start services
docker compose up -d

# Backend runs on :8080
# Frontend runs on :80
# PostgreSQL on :5432
# Redis on :6379
# MinIO on :9000 (API) :9001 (Console)
```

### Test Account

```
Email:    creator@test.com
Password: Test123!
```

### Google SSO Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project or select existing
3. Go to **APIs & Services** → **Credentials**
4. Click **Create Credentials** → **OAuth Client ID**
5. Set **Application type**: Web application
6. Add **Authorized redirect URI**: `http://localhost:8080/api/auth/google/callback`
7. Copy the **Client ID** and **Client Secret**
8. Add to `.env`:
   ```env
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   FRONTEND_URL=http://localhost
   ```

### Docker Compose

When deploying with Docker Compose, add the Google OAuth variables to the `backend` service environment (or put them in `.env` and use `${VAR}` in compose).

---

## Deployment

### Build & Push

```bash
# The CI/CD pipeline handles this
./ci.sh docker-build-push
```

### Environment Variables

For production, you need to set:
- `GOOGLE_CLIENT_ID` — Production Google OAuth Client ID
- `GOOGLE_CLIENT_SECRET` — Production Google OAuth Client Secret
- `FRONTEND_URL` — Your production domain (e.g., `https://yourdomain.com`)
- `JWT_SECRET` — Random long string (change from default)
- `DB_PASSWORD` — Strong database password
- `MINIO_ROOT_PASSWORD` — Strong MinIO password

The Google OAuth redirect URI for production:
```
https://yourdomain.com/api/auth/google/callback
```

Add this to the Google Cloud Console OAuth settings alongside the local one.