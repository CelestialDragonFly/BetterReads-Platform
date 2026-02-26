# Local BetterReads Testing with Docker Compose

This directory contains a Docker Compose configuration designed to spin up the latest version of the BetterReads server alongside a local PostgreSQL database. This setup is ideal for local testing, frontend integration, or general development.

## Prerequisites

1. **Docker & Docker Compose** installed on your system.
2. **Firebase Secrets**: Ensure your `firebase-serviceaccount.json` is located in a `secrets` directory at the root of the project (i.e., `../secrets/firebase-serviceaccount.json` relative to this compose file).

## Configuration & Security

While safe defaults are provided natively in the `docker-compose.yaml` file, for security best practices and to avoid hardcoding credentials we recommend using a `.env` file.

1. Create a `.env` file in the `compose` directory. You can copy the example file:
```bash
cp .env.example .env
```
2. Update the values in the `.env` file as you see fit.

## Usage

**Start the stack (in the background):**
```bash
docker compose up -d
```

This will:
- Pull the latest `betterreads:latest` image.
- Start the PostgreSQL database and wait for it to become healthy.
- Start the BetterReads server and connect it to Postgres.

**Access the Application:**
- HTTP API Gateway: `http://localhost:8080`
- gRPC Server: `localhost:9090`
- PostgreSQL DB: `localhost:5432` *(bound to localhost only for security)*

**View Logs:**
```bash
docker compose logs -f
```

**Stop the stack:**
```bash
docker compose down
```

**Stop the stack and remove database volumes (clean slate):**
```bash
docker compose down -v
```

## Security Features Implemented
- **No Hardcoded Passwords**: Uses variable interpolation (`${VAR:-default}`) and an `.env` file prioritizing local configuration.
- **Port Binding**: The database port is bound to `127.0.0.1` exclusively to ensure the DB isn't accidentally exposed on your local network.
- **Read-Only Volumes**: The Firebase secrets are mounted as read-only (`:ro`) to prevent the container from modifying them.
- **No-New-Privileges**: The application container starts with `--security-opt no-new-privileges:true` to prevent privilege escalation within the container.
- **Isolated Network**: Uses an isolated bridge network (`betterreads_net`) for the services to communicate.
- **Alpine Image**: The Postgres container uses an alpine image to minimize the internal attack surface.
- **Healthchecks**: Ensures the database is fully up and ready before the BetterReads application begins taking traffic.

## FAQ / Troubleshooting

**Q: I get a "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?" error.**
A: This usually means the Docker service is not running or your user doesn't have permission to access the Docker socket.
1. Ensure Docker is installed: `docker --version`
2. Ensure Docker is running: `sudo systemctl status docker` (or start it with `sudo systemctl start docker`)
3. Ensure your user has access to the docker group: `sudo usermod -aG docker $USER` (you may need to log out and back in)
4. Ensure your `~/.bashrc` (or `~/.zshrc`) is configured to export the correct docker host. Add the following line to your RC file and restart your shell (e.g. `source ~/.bashrc`):
```bash
export DOCKER_HOST=unix:///var/run/docker.sock
```

**Q: I get an "Error response from daemon: Head "...": unauthorized" or "denied" error when pulling the BetterReads image.**
A: The `betterreads` image is hosted on the GitHub Container Registry (GHCR) and requires authentication to pull.
1. Go to your GitHub Settings -> Developer settings -> Personal access tokens -> Tokens (classic).
2. Generate a new token with at least the `read:packages` scope.
3. Log in to the GitHub Container Registry via Docker:
```bash
docker login ghcr.io -u <YOUR_GITHUB_USER_NAME>
```
4. Paste the Personal Access Token (PAT) you just generated when prompted for the password.
