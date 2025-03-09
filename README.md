# BetterReads-Platform

## Local Developement
### Running Locally

### Locally building a Docker Image and Deploying
```sh
> docker build . -t betterreads
> docker run -v $(pwd)/secrets/firebase-serviceaccount.json:/app/serviceAccount.jso
n -e FIREBASE_SERVICE_ACCOUNT=/app/serviceAccount.json -p 8080:8080 betterreads
```

### Deploying Docker Image from Github Container Registry
