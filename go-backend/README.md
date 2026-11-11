# Portfolio Go backend

This is a Go implementation of the Spring Boot portfolio API using the Gin web framework. It uses the same PostgreSQL
tables, response fields, URLs, seeded content, and admin password hash format as the Java backend. Gin route groups
apply the existing admin session and CSRF checks. The React frontend and Caddy configuration do not need changes.

It requires Go 1.26.10 or later and PostgreSQL. Gin and its networking dependencies use a newer Go version than the
default Ubuntu 24.04 package. On Ubuntu, install the Go and PostgreSQL client packages with:

```bash
sudo apt update
sudo apt install -y golang-go postgresql-client
cd ~/portfolio/Portfolio/go-backend
go version
```

With the default `GOTOOLCHAIN=auto`, the Go command downloads the version required by `go.mod` on first use from this
directory. This needs internet access and disk space. Alternatively install a current Go release using the
[official installation instructions](https://go.dev/doc/install). Use an up-to-date Go toolchain when building for
production.

To install Docker, see the existing [Java backend setup instructions](../backend/README.md). The existing
`backend/docker-compose.yml` starts the same PostgreSQL database used by Java; keep its volume when switching backends.

## Run locally

From this directory, start the existing database container if it is not already running:

```bash
docker compose -f ../backend/docker-compose.yml up -d db
go mod download
go run .
```

The default database is `localhost:5432/portfolio` with the same `portfolio` username and password as the Java
development configuration. The default listener is `127.0.0.1:8080`. Stop the Java service first, or run the Go version
on a different port for testing:

```bash
SERVER_PORT=8081 go run .
```

For a separate port, test it directly with `curl http://127.0.0.1:8081/api/portfolio`. The frontend's existing proxy
continues to use port 8080, so using a different port alone does not switch the website to Go.

On first launch, the service creates missing tables using the existing Hibernate table layout and seeds the same
placeholder data if there is no profile. Existing portfolio rows are preserved. It does not alter incompatible existing
tables; take a database backup before switching a production deployment.

## Admin login

If the database already contains the Java backend's admin credential, the same key works. Otherwise generate a strong
key with `openssl rand -hex 32`, set it as `ADMIN_BOOTSTRAP_KEY` for the first startup, and save it somewhere secure.
The service stores a salted PBKDF2-SHA256 hash in PostgreSQL. Remove the bootstrap variable after successful login;
changing it later does not replace an existing hash.

Open `/admin` on the same origin as the API. The browser receives an HttpOnly, SameSite=Strict session cookie. Sessions
expire after 30 minutes by default, require a CSRF token for changes, and are lost when the Go process restarts. On an
HTTPS deployment set `SESSION_COOKIE_SECURE=true`; then logins require HTTPS. Keep port 8080 private and let Caddy proxy
`/api/*` to the Go listener.

## Configuration

Under **Education** in `/admin`, upload a JPEG or PNG to replace the campus placeholder.
The Go service writes the image to `UPLOAD_DIR/college/` and stores its public path in
`education_photo.image_path`. Each distinct image has its own filename, so existing links
keep working after a replacement. Back up the upload directory alongside PostgreSQL.

Under **Experience**, save the company name and use **Upload logo** to choose a JPEG or PNG.
Each upload updates only the selected Experience entry.
The image lives in `UPLOAD_DIR/company-logos/`, and `experience.company_logo_path` holds its public path.

| Variable                | Default                                                 | Purpose                                                                                                                    |
|-------------------------|---------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|
| `DB_URL`                | `postgresql://localhost:5432/portfolio?sslmode=disable` | PostgreSQL URL; a Java-style `jdbc:postgresql://...` URL also works. For remote servers specify the appropriate `sslmode`. |
| `DB_USER`               | `portfolio`                                             | Database user; overrides credentials in `DB_URL` when set.                                                                 |
| `DB_PASSWORD`           | `portfolio`                                             | Database password; overrides credentials in `DB_URL` when set.                                                             |
| `SERVER_ADDRESS`        | `127.0.0.1`                                             | API bind address.                                                                                                          |
| `SERVER_PORT`           | `8080`                                                  | API port.                                                                                                                  |
| `UPLOAD_DIR`            | `uploads`                                               | Directory for static uploads, relative to the service working directory or absolute. |
| `ADMIN_BOOTSTRAP_KEY`   | unset                                                   | Initial key, at least 32 characters; used only when no admin credential exists.                                            |
| `ADMIN_SESSION_MINUTES` | `30`                                                    | Absolute session lifetime, from 1 to 1440 minutes.                                                                         |
| `SESSION_COOKIE_SECURE` | `false`                                                 | Set to `true` for HTTPS in production.                                                                                     |
| `CORS_ORIGINS`          | `http://localhost:5173`                                 | Comma-separated allowed origins for public GET requests. Admin sessions require a same-origin frontend.                    |

For an existing Java installation, `DB_URL`, `DB_USER`, `DB_PASSWORD`, `ADMIN_BOOTSTRAP_KEY`, `ADMIN_SESSION_MINUTES`,
and `SESSION_COOKIE_SECURE` retain the same meaning. Do not commit secrets or place them in frontend `VITE_*` variables.

## Build and switch the Lightsail service

Copy the new code to the VM, then build:

```bash
cd ~/portfolio/Portfolio/go-backend
go mod download
go build -o portfolio-api .
```

If your existing `portfolio-api` systemd service already points to the Java JAR, run `sudo systemctl edit portfolio-api`
and add this override, adjusting the paths if your checkout is elsewhere:

```ini
[Service]
StateDirectory=portfolio-api
Environment=UPLOAD_DIR=/var/lib/portfolio-api/uploads
WorkingDirectory=/home/ubuntu/portfolio/Portfolio/go-backend
ExecStart=
ExecStart=/home/ubuntu/portfolio/Portfolio/go-backend/portfolio-api
```

Keep its existing user and environment file settings. The environment file should contain `SESSION_COOKIE_SECURE=true`
for HTTPS, the existing `DB_*` values if needed, and `ADMIN_BOOTSTRAP_KEY` only if the database has no admin credential.
If the Java service used a JDBC URL, it can be retained. Apply the change:

```bash
sudo systemctl daemon-reload
sudo systemctl restart portfolio-api
sudo systemctl status portfolio-api
curl http://127.0.0.1:8080/api/portfolio
```

Then test the website and `/admin` over HTTPS. Existing Java login sessions do not transfer, so sign in again. Roll back
by removing this override and restarting the service; neither implementation requires deleting the database.

## Tests

```bash
go test ./...
go vet ./...
```

The PostgreSQL integration test is skipped unless `TEST_DATABASE_URL` points to a **disposable** database whose name
ends in `_test`, for example:

```bash
TEST_DATABASE_URL='postgresql://portfolio:portfolio@localhost:5432/portfolio_go_test?sslmode=disable' go test -v ./...
```

The integration test creates tables, seeds data and performs CRUD operations. Never point it at a database containing
data you want to keep. Unlike the Java H2 profile, the Go backend uses PostgreSQL for both development and production.
