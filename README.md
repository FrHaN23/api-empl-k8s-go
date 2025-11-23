# Employee CRUD API — Go + Postgres + SQLX + Docker + Kubernetes (kind)

A simple **Employee CRUD REST API** implemented in Go, using PostgreSQL and sqlx.  
Includes Docker, Docker Compose, and Kubernetes (kind) manifests for local development and demo deployment.

---

## Project Overview

This project demonstrates a minimal but production-like backend service using:

- Go 1.25+
- PostgreSQL with sqlx
- PostgreSQL DB (schema in migrations/)
- Gorilla Mux router
- Clean structure: handler → service → repository
- Docker multi-stage build
- Docker Compose for local development
- Kubernetes manifests for running in kind (local Kubernetes)
- Kubernetes (kind) deployment with:
  - Postgres Deployment + PVC
  - API Deployment + NodePort
  - Secrets / ConfigMaps

---

## Repo Structure

```
api-empl-k8s-go/
├── handler/
├── models/
├── repos/
├── service/
├── utils/
├── routes/
├── modules/
├── migrations/
│   └── 001_create_employees_table.sql
├── k8s/
│   ├── deploy-app.yml
│   ├── deploy-postgres.yml
│   ├── secret-postgres.yml
│   ├── pvc-postgres.yml
│   ├── migrate-job.yml
├── main.go
├── Dockerfile
├── docker-compose.yml
├── .env
└── README.md
```

---

## Prerequisites

Ensure you have installed:

- Go 1.25+
- Docker + docker compose
- kind
- kubectl

---

## Environment Variables

Create `.env`:

```
DB_NAME=employees_db
DB_HOST=db
DB_USERNAME=frhan
DB_PASSWORD=db_postgres
DB_PORT=5432
DB_TZ=Asia/Jakarta
```

---

## Run Locally (Without Docker)

```bash
go mod tidy
go run .
# API listens on :5000 by default
```

---

## Docker & Docker Compose

### Build image

```bash
docker build -t api-empl:latest .
```

### Compose up

```bash
docker compose up --build
```
This will start app and db services. The DB init scripts can be placed under migrations/ and mounted into /docker-entrypoint-initdb.d by the compose file if desired.

---

## Database Migrations (Manual)

Migration SQL files live in migrations/ (e.g. 001_create_employees_table.sql). For a quick one-off run:
```bash
psql -h localhost -U frhan -d employees_db -f migrations/001_create_employees_table.sql
```
(Adjust connection info to match your environment.)

---

## Kubernetes (kind) — Local Cluster

### Create kind cluster

```bash
kind create cluster
```

### Load local docker image

```bash
docker build -t api-empl:latest .
kind load docker-image api-empl:latest
```

---

## PVC / Storage Note

Kind requires a local path provisioner.

Install:

```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
```

Create PVC:

```bash
kubectl apply -f k8s/pvc-postgres.yml
```

---

## Deploy

### 1. Create postgres secrets

```bash
kubectl create secret generic secret-postgres   --from-literal=POSTGRES_USER=frhan   --from-literal=POSTGRES_PASSWORD=db_postgres   --from-literal=POSTGRES_DB=employees_db
```
or from the file
```bash
kubectl apply -f k8s/secret-postgres.yml
```

### 2. Deploy postgres server

```bash
kubectl apply -f k8s/deploy-postgres.yml
kubectl rollout status deployment/postgres
```

### 3. Deploy API

Load image (required):

```bash
kind load docker-image api-empl:latest
```

Apply configMap and secret app:
```bash
kubectl apply -f k8s/configmap-app.yml
kubectl apply -f k8s/secret-app.yml
```

Apply the app:

```bash
kubectl apply -f k8s/deploy-app.yml
kubectl rollout status deployment/api-empl
```

---

## Apply Migration

Copy SQL:

```bash
PODPG=$(kubectl get pod -l app=postgres -o jsonpath='{.items[0].metadata.name}')
kubectl cp migrations/001_create_employees_table.sql $PODPG:/tmp/migrate.sql
```

Execute:

```bash
PGUSER=$(kubectl get secret secret-postgres -o jsonpath='{.data.POSTGRES_USER}' | base64 -d)
PGPASS=$(kubectl.get secret secret-postgres -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)
PGDB=$(kubectl.get secret secret-postgres -o jsonpath='{.data.POSTGRES_DB}' | base64 -d)

kubectl exec -it $PODPG -- sh -c "export PGPASSWORD='$PGPASS'; psql -U '$PGUSER' -d '$PGDB' -f /tmp/migrate.sql"
```
---

## API Endpoints

| Method | Endpoint          | Description        |
|--------|-------------------|--------------------|
| GET    | /employees        | List employees     |
| POST   | /employees        | Create employee    |
| GET    | /employees/{id}   | Get employee       |
| PUT    | /employees/{id}   | Update employee    |
| DELETE | /employees/{id}   | Soft delete        |

---

## Troubleshooting

### ImagePullBackOff

```
kind load docker-image api-empl:latest
```

### Postgres PVC Corrupt

```bash
kubectl delete pvc pvc-postgres
kubectl apply -f k8s/pvc-postgres.yml
```

### Missing secrets/configmaps

```bash
kubectl get secrets
kubectl get configmap
```

### Postgres cannot start

Delete PV + PVC:

```bash
kubectl delete pvc pvc-postgres
kubectl delete pv $(kubectl get pv -o jsonpath='{.items[0].metadata.name}')
```

---

## Future Improvements

- Real migration tool
- Health checks
- Helm chart
- CI/CD pipeline
- Logging + metrics

## Useful Commands

```bash
# K8s
kubectl get pods -A
kubectl describe pod <pod-name>
kubectl logs deployment/postgres -f

# Load image into kind
kind load docker-image api-empl:latest

# Create secret (example)
kubectl create secret generic secret-postgres \
  --from-literal=POSTGRES_USER=frhan \
  --from-literal=POSTGRES_PASSWORD=db_postgres \
  --from-literal=POSTGRES_DB=employees_db

```
