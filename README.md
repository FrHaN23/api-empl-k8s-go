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

Apply:

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

## Deploying  to Google Kubernetes Engine (GKE)

This shows step-by-step instructions to deploy your **api-empl-k8s-go** project to **GKE (Autopilot)**, push the image to **Artifact Registry**, connect to **Cloud SQL (Postgres)** using **Workload Identity + Cloud SQL Auth Proxy**, and expose the app with a LoadBalancer.

> Region used in these instructions: **asia-southeast2 (Jakarta)**
> Adjust `PROJECT_ID` and `CLUSTER` names to match your environment.


## 1. Prerequisites (local machine)
- `gcloud` (Google Cloud SDK) installed and initialized.
- `kubectl` installed and configured (runs via `gcloud container clusters get-credentials`).  
- `docker` installed (for building images) or use `cloud build`.
- `psql` client (optional, for direct DB work).

Example (openSUSE): install gcloud (tarball or package) and psql:
```bash
# gcloud (if not installed) — recommended install to $HOME
curl -LO https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-linux-x86_64.tar.gz
tar -xzf google-cloud-cli-linux-x86_64.tar.gz
./google-cloud-sdk/install.sh --quiet
source $HOME/google-cloud-sdk/path.bash.inc

# psql (match your Cloud SQL major version; Cloud SQL is Postgres 18 in these steps)
sudo zypper install postgresql18 # i'm using openSUSE
```

Login and set project:
```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
gcloud config set compute/region asia-southeast2
gcloud config set compute/zone asia-southeast2-a
```

---

## 2. Enable required GCP APIs
```bash
gcloud services enable \
  artifactregistry.googleapis.com \
  container.googleapis.com \
  compute.googleapis.com \
  sqladmin.googleapis.com
```

---

## 3. Create Artifact Registry (Docker repo)
```bash
gcloud artifacts repositories create api-images \
  --repository-format=docker \
  --location=asia-southeast2 \
  --description="docker repo for api-empl"
gcloud auth configure-docker asia-southeast2-docker.pkg.dev
```

---

## 4. Build & push Docker image
From your repository root (where the `Dockerfile` is):
```bash
docker build -t api-empl:latest .

PROJECT_ID=$(gcloud config get-value project)
docker tag api-empl:latest asia-southeast2-docker.pkg.dev/${PROJECT_ID}/api-images/api-empl:latest
docker push asia-southeast2-docker.pkg.dev/${PROJECT_ID}/api-images/api-empl:latest
```

---

## 5. Create Cloud SQL (Postgres) instance
Create the instance (example config):
```bash
gcloud sql instances create api-postgres \
  --database-version=POSTGRES_15 \ # client 18, instance 15 = good
  --cpu=1 --memory=4GiB \
  --region=asia-southeast2
gcloud sql databases create employees_db --instance=api-postgres
gcloud sql users set-password frhan --instance=api-postgres --password='db_postgres'
```
Get connection name (you will need this for the proxy):
```bash
gcloud sql instances describe api-postgres --format="value(connectionName)"
# example: api-empl-go:asia-southeast2:api-postgres
```

---

## 6. Create GKE cluster (Autopilot) and configure Workload Identity
Create Autopilot cluster:
```bash
gcloud container clusters create-auto my-cluster --region=asia-southeast2
gcloud container clusters get-credentials my-cluster --region=asia-southeast2
```

Enable Workload Identity on the cluster (safe to run even if already enabled):
```bash
PROJECT_ID=$(gcloud config get-value project)
gcloud container clusters update my-cluster --region=asia-southeast2 \
  --workload-pool=${PROJECT_ID}.svc.id.goog
```

Create a Google Service Account (GSA) and grant Cloud SQL role:
```bash
GSA=cloudsql-gsa
gcloud iam service-accounts create $GSA --display-name="Cloud SQL client for api-empl"
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${GSA}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"
```

Create Kubernetes Service Account (KSA) and bind to GSA (Workload Identity):
```bash
KSA=api-empl-ksa
NAMESPACE=default
kubectl create serviceaccount $KSA --namespace $NAMESPACE || true

# allow the KSA to impersonate the GSA
gcloud iam service-accounts add-iam-policy-binding ${GSA}@${PROJECT_ID}.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/${KSA}]"

# annotate the KSA
kubectl annotate serviceaccount -n ${NAMESPACE} $KSA \
  iam.gke.io/gcp-service-account=${GSA}@${PROJECT_ID}.iam.gserviceaccount.com --overwrite
```

---

## 7. Create Kubernetes secrets & config (DB creds & app config)
Create k8s secret for DB user/password (used by the app):
```bash
kubectl delete secret api-empl-secret --ignore-not-found # if secret already exist
kubectl create secret generic api-empl-secret \
  --from-literal=POSTGRES_USER='frhan' \
  --from-literal=POSTGRES_PASSWORD='db_postgres'
kubectl create configmap api-empl-config --from-literal=DB_NAME=employees_db --from-literal=DB_PORT=5432 --from-literal=DB_TZ=UTC
```

(If you prefer JSON key method instead of Workload Identity, create `cloudsql-instance-credentials` secret from a service account key — not recommended for production.)

---

## 8. Deploy manifest (Deployment + LoadBalancer Service)
Below is a ready-to-apply `deploy-app-gke.yml`. It assumes:
- image: `asia-southeast2-docker.pkg.dev/${PROJECT_ID}/api-images/api-empl:latest`
- Cloud SQL connection name: `api-empl-go:asia-southeast2:api-postgres`
- using Workload Identity via `serviceAccountName: api-empl-ksa`

Save this as `deploy-app-gke.yml` and `kubectl apply -f deploy-app-gke.yml`.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-empl
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api-empl
  template:
    metadata:
      labels:
        app: api-empl
    spec:
      serviceAccountName: api-empl-ksa
      containers:
        - name: api-empl
          image: asia-southeast2-docker.pkg.dev/PROJECT_ID/api-images/api-empl:latest
          imagePullPolicy: IfNotPresent
          resources:
            requests:
              cpu: "100m"
              memory: "512Mi"
            limits:
              cpu: "500m"
              memory: "1Gi"
          env:
            - name: DB_HOST
              value: "127.0.0.1"
            - name: DB_PORT
              value: "5432"
            - name: DB_NAME
              valueFrom:
                configMapKeyRef:
                  name: api-empl-config
                  key: DB_NAME
            - name: DB_USERNAME
              valueFrom:
                secretKeyRef:
                  name: api-empl-secret
                  key: POSTGRES_USER
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: api-empl-secret
                  key: POSTGRES_PASSWORD
            - name: DB_TZ
              valueFrom:
                configMapKeyRef:
                  name: api-empl-config
                  key: DB_TZ
            - name: DB_SSLMODE
              value: "disable"
          ports:
            - containerPort: 5000
          readinessProbe:
            httpGet:
              path: /health
              port: 5000
            initialDelaySeconds: 10
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          livenessProbe:
            httpGet:
              path: /health
              port: 5000
            initialDelaySeconds: 30
            periodSeconds: 15
        - name: cloud-sql-proxy
          image: gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.19.0
          imagePullPolicy: IfNotPresent
          resources:
            requests:
              cpu: "20m"
              memory: "128Mi"
            limits:
              cpu: "200m"
              memory: "256Mi"
          args:
            - "api-empl-go:asia-southeast2:api-postgres"
            - "--address=0.0.0.0"
            - "--port=5432"
          securityContext:
            runAsNonRoot: true
            allowPrivilegeEscalation: false
          livenessProbe:
            tcpSocket:
              port: 5432
            initialDelaySeconds: 60
            periodSeconds: 10
            failureThreshold: 6
          readinessProbe:
            tcpSocket:
              port: 5432
            initialDelaySeconds: 40
            periodSeconds: 5
            failureThreshold: 6
---
apiVersion: v1
kind: Service
metadata:
  name: empl-api
spec:
  type: LoadBalancer
  selector:
    app: api-empl
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: 5000
```

After apply:
```bash
kubectl apply -f deploy-app-gke.yml
kubectl rollout status deployment/api-empl
kubectl get pods -l app=api-empl -o wide
kubectl get svc empl-api
```

---

## 9. Create DB schema
Connect and create the `employees` table if not present:
```bash
gcloud sql connect api-postgres --user=frhan --database=employees_db
# then in psql:
CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    position TEXT NOT NULL,
    salary INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_employees_deleted_at ON employees (deleted_at);
```
---

## 10. Access your app
Once the `empl-api` service has an `EXTERNAL-IP`, open:
```
http://<EXTERNAL-IP>/health
http://<EXTERNAL-IP>/your-endpoint
```
You can check the ip:

```bash
kubectl get svc empl-api
```
It will show the external IP

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
