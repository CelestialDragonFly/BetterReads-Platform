# BetterReads-Platform

## Local Developement
### Running Locally
Make sure you clone the [Betterreads-Platform-Release](https://github.com/CelestialDragonFly/BetterReads-Platform-Release) Repository and setup the .env file.

Start a local Postgres Database.
```sh
docker compose -f '../betterreads-platform-release/docker-compose.yaml' --env-file ../betterreads-platform-release/.env up --build 'postgres'
```

You may now run the program locally.

Command line example:
```sh
go run ./...
```

### Locally building a Docker Image
```sh
 docker build . \
    --build-arg FIREBASE_CONFIG="$(cat secrets/firebase-serviceaccount.json)" \
    --tag betterreads:plat-10
 ```

### Running BetterReads on Machine (PaaS)
Update the Betterreads-Platform-Release `docker-compose.yaml` with either a locally built docker image or a remote image from GitHub Container Registry (GHCR).

Run
```sh
docker compose up
```
