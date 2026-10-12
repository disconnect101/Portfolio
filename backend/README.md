# Portfolio Backend

Spring Boot REST API for the portfolio application. It stores portfolio content in PostgreSQL and runs on port `8080` by default.

## Requirements

- Java 17 or newer
- Maven 3.9 or newer
- Docker Desktop, for the local PostgreSQL database

Check the command-line tools before starting:

```bash
java -version
mvn -version
docker --version
docker compose version
```

Start Docker Desktop before running any `docker compose` commands.

## Ubuntu 24.04 installation

Use these steps on an Ubuntu 24.04 VPS or server. They install Java 17, Maven, Git, Docker Engine, Docker Compose, and the optional PostgreSQL command-line client.

> Do **not** install the PostgreSQL server package on this machine for this project. Docker Compose starts the project's PostgreSQL server, which avoids managing a second database service and prevents port conflicts.

### 1. Install Java, Maven, Git, and the PostgreSQL client

```bash
sudo apt update
sudo apt install -y openjdk-17-jdk maven git postgresql-client ca-certificates curl
```

### 2. Install Docker and Docker Compose

The simplest Ubuntu installation is:

```bash
sudo apt update
sudo apt install -y docker.io docker-compose-v2
sudo systemctl enable --now docker
sudo usermod -aG docker "$USER"
newgrp docker
```

> Membership in the `docker` group grants root-equivalent access to the machine. Only add a trusted administrator account to it.

If `newgrp docker` does not take effect, sign out of the SSH session and sign in again.

### 3. Verify the installation

```bash
java -version
mvn -version
git --version
docker --version
docker compose version
psql --version
docker run hello-world
```

### 4. Start the application

Clone or upload this project, then use the normal local startup commands:

```bash
git clone <your-repository-url> portfolio
cd portfolio/backend
docker compose up -d
mvn spring-boot:run
```

For the newest Docker release instead of Ubuntu's packaged version, see the [official Docker Engine for Ubuntu guide](https://docs.docker.com/engine/install/ubuntu/).

## Run locally with PostgreSQL

From this directory:

```bash
cd /Users/aditya/IdeaProjects/portfolio/backend

# Start PostgreSQL in Docker.
docker compose up -d

# Confirm that the container is healthy.
docker compose ps
docker compose exec db pg_isready -U portfolio -d portfolio

# Start the Spring Boot API.
mvn spring-boot:run
```

When the application has started, verify it at:

```bash
curl http://localhost:8080/api/portfolio
```

The database container exposes PostgreSQL at `localhost:5433`. Port `5433` is used intentionally so it does not conflict with a PostgreSQL installation already running on the Mac at port `5432`.

### Local database credentials

```text
Host:     localhost
Port:     5433
Database: portfolio
Username: portfolio
Password: portfolio
```

Open a PostgreSQL prompt inside the Docker container:

```bash
docker compose exec db psql -U portfolio -d portfolio
```

Useful `psql` commands:

```sql
\conninfo
\dt
\du
\l
SELECT current_user, current_database();
\q
```

To test the connection from the Mac instead of from the container:

```bash
psql "postgresql://portfolio:portfolio@127.0.0.1:5433/portfolio" \
  -c "SELECT current_user, current_database();"
```

## Run without PostgreSQL

The project includes an in-memory H2 profile for quick local development. Data is discarded when the application stops.

```bash
mvn spring-boot:run -Dspring-boot.run.profiles=h2
```

## Run from IntelliJ IDEA

1. Open `backend/pom.xml` as a Maven project.
2. Start Docker Desktop and run `docker compose up -d` from `backend`.
3. Open **Run → Edit Configurations…** and add a **Maven** configuration.
4. Set the working directory to this `backend` folder.
5. Set the command line to `spring-boot:run`.
6. Click the green Run button.

To use the in-memory H2 database instead, set the Maven command line to:

```text
spring-boot:run -Dspring-boot.run.profiles=h2
```

## Configuration

The defaults are in `src/main/resources/application.properties`. They can be overridden with environment variables:

```text
DB_URL=jdbc:postgresql://localhost:5433/portfolio
DB_USER=portfolio
DB_PASSWORD=portfolio
CORS_ORIGINS=http://localhost:5173
```

For example:

```bash
DB_URL='jdbc:postgresql://db.example.com:5432/portfolio' \
DB_USER='portfolio_app' \
DB_PASSWORD='use-a-strong-secret' \
CORS_ORIGINS='https://portfolio.example.com' \
mvn spring-boot:run
```

Never commit production database passwords to Git.

## API endpoints

All endpoints are read-only and return JSON.

```text
GET /api/portfolio
GET /api/profile
GET /api/projects
GET /api/education
GET /api/experience
```

The initial portfolio content is seeded by `DataSeeder` when the database is empty.

## Build a JAR

```bash
mvn clean package
java -jar target/portfolio-backend-1.0.0.jar
```

## Common problems

### `role "portfolio" does not exist`

The Docker volume was created previously with different PostgreSQL credentials. If it contains no data you need, recreate it:

```bash
docker compose down -v
docker compose up -d
```

> Warning: `docker compose down -v` permanently deletes this project's local PostgreSQL data.

If the data must be kept, connect using the existing database administrator role and create the `portfolio` role instead.

### Spring reports that it cannot determine a Hibernate dialect

This normally means the backend could not connect to PostgreSQL. Confirm that the Docker database is ready:

```bash
docker compose exec db pg_isready -U portfolio -d portfolio
psql "postgresql://portfolio:portfolio@127.0.0.1:5433/portfolio" \
  -c "SELECT current_user, current_database();"
```

Also check that IntelliJ has not set stale `DB_URL`, `DB_USER`, or `DB_PASSWORD` environment variables in its Run Configuration.

### Port 5433 is already in use

Choose another unused host port in `docker-compose.yml`, for example:

```yaml
ports:
  - "5434:5432"
```

Then update `DB_URL` to use the matching port:

```text
DB_URL=jdbc:postgresql://localhost:5434/portfolio
```

Restart the database after changing the Compose file:

```bash
docker compose down
docker compose up -d
```
