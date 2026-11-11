# Portfolio Website

Java (Spring Boot) backend + React (Vite) frontend.

```
portfolio/
├── backend/          Spring Boot 3, JPA, PostgreSQL REST API
├── frontend/         React + Vite single-page site
└── backend/docker-compose.yml  PostgreSQL for local development
```

## Run it

**1. Database** (from the `backend` folder)

```bash
cd backend
docker compose up -d
```

No Docker? Skip this and use the in-memory database in step 2 instead.

**2. Backend** (port 8080, needs Java 17+ and Maven)

```bash
mvn spring-boot:run
# or without PostgreSQL:
mvn spring-boot:run -Dspring-boot.run.profiles=h2
```

On first start it seeds placeholder content. Check: http://localhost:8080/api/portfolio

**3. Frontend** (port 5173, needs Node 18+)

```bash
cd ../frontend
npm install
npm run dev
```

Open http://localhost:5173

## Edit your content

Set a unique admin key of at least 32 characters to bootstrap the admin account, then open `/admin` on the frontend. On first start the backend stores only a salted PBKDF2 hash in PostgreSQL. Login compares the entered key with that hash and creates a server-side session valid for 30 minutes. The browser gets an HttpOnly session cookie; refreshing the page keeps the session until it expires, and logging out invalidates it.

For local development, generate a key with `openssl rand -hex 32`, then start the backend once with `ADMIN_BOOTSTRAP_KEY='your-generated-key' mvn spring-boot:run`. After the hash has been saved, you can remove the environment variable; subsequent logins use the database hash. Never put the key in frontend environment variables or Git.

The admin page can edit the profile and create, edit, or delete projects, education, and experience. Image fields accept URLs; put local images in `frontend/public/` and use paths such as `/my-photo.jpg`.

## Initial content

- The first backend start seeds placeholder content from `backend/src/main/java/com/portfolio/config/DataSeeder.java`. Update it safely through `/admin`; there is no need to reset the database.
- **Cover photo and profile picture**: drop your images into `frontend/public/`
  (e.g. `cover.jpg`, `me.jpg`) and change `setCoverPhotoUrl` / `setProfilePhotoUrl`
  in the seeder to `/cover.jpg` and `/me.jpg`. Add `setImageUrl("/my-project.png")` on a project to show a thumbnail.

## API

| Endpoint | Returns |
|---|---|
| `GET /api/portfolio` | profile + projects + education + experience |
| `GET /api/profile` | profile |
| `GET /api/projects` | projects |
| `GET /api/education` | education |
| `GET /api/experience` | experience |

Admin endpoints under `/api/admin/*` require a valid session. Write requests also require a session-bound CSRF token. The login endpoint returns 503 until an admin credential has been bootstrapped. The public API remains available.

## Deploying

- Frontend: `npm run build` produces `frontend/dist`. For admin sessions, serve the frontend and API on the same HTTPS origin (as with your Caddy setup), leaving `VITE_API_URL` unset.
- Backend: `mvn package` produces an executable jar in `backend/target`. Set `DB_URL`, `DB_USER`,
  `DB_PASSWORD` and `CORS_ORIGINS` (your frontend's URL) as environment variables. Set `ADMIN_BOOTSTRAP_KEY` only for the first launch, and `SESSION_COOKIE_SECURE=true` on the HTTPS production site. Remove the bootstrap key after the hash is stored.
