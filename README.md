# Portfolio Website

Java (Spring Boot) backend + React (Vite) frontend.

```
portfolio/
├── backend/          Spring Boot 3, JPA, PostgreSQL REST API
├── frontend/         React + Vite single-page site
└── docker-compose.yml  PostgreSQL for local development
```

## Run it

**1. Database** (from this folder)

```bash
docker compose up -d
```

No Docker? Skip this and use the in-memory database in step 2 instead.

**2. Backend** (port 8080, needs Java 17+ and Maven)

```bash
cd backend
mvn spring-boot:run
# or without PostgreSQL:
mvn spring-boot:run -Dspring-boot.run.profiles=h2
```

On first start it seeds placeholder content. Check: http://localhost:8080/api/portfolio

**3. Frontend** (port 5173, needs Node 18+)

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:5173

## Put in your real details

- **Text content** (name, description, projects, education, jobs): edit
  `backend/src/main/java/com/portfolio/config/DataSeeder.java`, then reset the database
  (`docker compose down -v && docker compose up -d`) and restart the backend.
  The seeder only runs when the database is empty.
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

## Deploying

- Frontend: `npm run build` produces `frontend/dist` (host on Netlify, Vercel, S3, etc.). Set
  `VITE_API_URL` to your API's URL at build time.
- Backend: `mvn package` produces an executable jar in `backend/target`. Set `DB_URL`, `DB_USER`,
  `DB_PASSWORD` and `CORS_ORIGINS` (your frontend's URL) as environment variables.
