# Adinusa Challenge: Deploy the Guest Book Application

> \*\*You are given a working microservices application. Your job is to containerize it, push images to a registry, and deploy it to Kubernetes with production-grade configuration.\*\*

\---

## Application Overview

This is a GuestBook app with 3 services:

```
┌─────────────────┐       ┌────────────────────┐      ┌────────────────────┐
│   Frontend      │       │   Auth Service      │      │  Booking Service   │
│   (Vue + Nginx) │─────▶│   (Go - port 8081)  │      │  (Go - port 8082)  │
│   port 80       │─────▶│                     │◀─────│                    │
│                 │       │   File Sessions     │      │   SQLite Database  │
└─────────────────┘       │   /tmp/sessions/    │      │   /data/guestbook  │
                          └────────────────────┘      └────────────────────┘
```

|Service|Role|Port|Storage|
|-|-|-|-|
|**frontend**|Vue SPA served by Nginx, proxies API calls|80|—|
|**auth-service**|Login / logout / session check|8081|File-based sessions at `/tmp/sessions/`|
|**booking-service**|CRUD guestbook entries, verifies auth via auth-service|8082|SQLite database at `/data/guestbook.db`|

**Default admin credentials** (printed to auth-service stdout on startup):

* Username: `admin`
* Password: `admin123`

### API Endpoints

**Auth Service:**

|Method|Path|Auth|Description|
|-|-|-|-|
|POST|`/api/auth/login`|No|Login, returns `{ token, username }`|
|POST|`/api/auth/logout`|Yes|Destroys session|
|GET|`/api/auth/me`|Yes|Returns current user|
|GET|`/healthz`|No|Health check|

**Booking Service:**

|Method|Path|Auth|Description|
|-|-|-|-|
|GET|`/api/guestbook`|No|List all entries|
|POST|`/api/guestbook`|Yes|Create entry `{ name, address, message }`|
|DELETE|`/api/guestbook/{id}`|Yes|Delete entry|
|GET|`/healthz`|No|Health check|

### Environment Variables

**Auth Service:**

|Variable|Default|Description|
|-|-|-|
|`SESSION\_DIR`|`/tmp/sessions`|Directory for session files|
|`ADMIN\_USER`|`admin`|Default admin username|
|`ADMIN\_PASS`|`admin123`|Default admin password|
|`PORT`|`8081`|Listen port|

**Booking Service:**

|Variable|Default|Description|
|-|-|-|
|`DB\_PATH`|`/data/guestbook.db`|SQLite database file path|
|`AUTH\_SERVICE\_URL`|`http://auth-service:8081`|Auth service internal URL|
|`PORT`|`8082`|Listen port|

### Nginx Proxy Rules (frontend)

The `nginx.conf` in the frontend routes:

* `/api/auth/\*` → `http://auth-service:8081`
* `/api/guestbook\*` → `http://booking-service:8082`

These service names must match your Kubernetes Service names.

\---

## Your Tasks

### Part 1 — Build \& Push Docker Images

The Dockerfiles are already provided. You need to build and push them.

* \[ ] Review the 3 Dockerfiles and understand the multi-stage builds
* \[ ] Build the **auth-service** image

```bash
  docker build -t <your-registry>/guestbook-auth:v1 ./auth-service
  ```

* \[ ] Build the **booking-service** image

```bash
  docker build -t <your-registry>/guestbook-booking:v1 ./booking-service
  ```

* \[ ] Build the **frontend** image

```bash
  docker build -t <your-registry>/guestbook-frontend:v1 ./frontend
  ```

* \[ ] Test locally with `docker run` — verify each container starts without errors
* \[ ] Push all 3 images to your container registry (DockerHub or private)

```bash
  docker push <your-registry>/guestbook-auth:v1
  docker push <your-registry>/guestbook-booking:v1
  docker push <your-registry>/guestbook-frontend:v1
  ```

* \[ ] **(If private registry)** Prepare a `docker-registry` secret for Kubernetes:

```bash
  kubectl create secret docker-registry regcred \\
    --docker-server=<your-registry> \\
    --docker-username=<user> \\
    --docker-password=<password> \\
    --namespace=guestbook
  ```

\---

### Part 2 — Deploy Database (SQLite via PersistentVolume)

This app uses SQLite — there is no external database server to deploy. However, the data must survive pod restarts, so you need persistent storage.

* \[ ] Create a `PersistentVolumeClaim` for the booking-service SQLite data (`/data`)

  * Storage: `1Gi`
  * Access mode: `ReadWriteOnce`
* \[ ] Create a `PersistentVolumeClaim` for the auth-service session files (`/tmp/sessions`)

  * Storage: `512Mi`
  * Access mode: `ReadWriteOnce`
* \[ ] Apply and verify PVCs are `Bound`: `kubectl get pvc -n guestbook`



\---

### Part 3 — Write Kubernetes YAML Manifests

Create all manifests yourself. Every YAML file must be written by you from scratch.

#### 3.1 Namespace

* \[ ] Create `namespace.yaml` — namespace named `guestbook`
* \[ ] Apply it

#### 3.2 Auth Service

* \[ ] Create `auth-service.yaml` containing:

  * \[ ] **Deployment** with:

    * \[ ] Correct container image from your registry
    * \[ ] Container port `8081`
    * \[ ] Environment variable `SESSION\_DIR` = `/tmp/sessions`
    * \[ ] Resource **requests**: CPU `50m`, Memory `64Mi`
    * \[ ] Resource **limits**: CPU `200m`, Memory `128Mi`
    * \[ ] **Liveness probe** on `/healthz` port `8081`
    * \[ ] **Readiness probe** on `/healthz` port `8081`
    * \[ ] Volume mount for sessions PVC at `/tmp/sessions`
    * \[ ] (If private registry) `imagePullSecrets`
  * \[ ] **Service** (ClusterIP) exposing port `8081`

    * Service name **must be** `auth-service` (the nginx config and booking-service rely on this)
  * \[ ] **HorizontalPodAutoscaler**

    * \[ ] Min replicas: `1`
    * \[ ] Max replicas: `3`
    * \[ ] Target CPU utilization: `70%`

#### 3.3 Booking Service

* \[ ] Create `booking-service.yaml` containing:

  * \[ ] **Deployment** with:

    * \[ ] Correct container image from your registry
    * \[ ] Container port `8082`
    * \[ ] Environment variables:

      * `DB\_PATH` = `/data/guestbook.db`
      * `AUTH\_SERVICE\_URL` = `http://auth-service:8081`
    * \[ ] Resource **requests**: CPU `50m`, Memory `64Mi`
    * \[ ] Resource **limits**: CPU `200m`, Memory `128Mi`
    * \[ ] **Liveness probe** on `/healthz` port `8082`
    * \[ ] **Readiness probe** on `/healthz` port `8082`
    * \[ ] Volume mount for data PVC at `/data`
    * \[ ] (If private registry) `imagePullSecrets`
  * \[ ] **Service** (ClusterIP) exposing port `8082`

    * Service name **must be** `booking-service`
  * \[ ] **HorizontalPodAutoscaler**

    * \[ ] Min replicas: `1`
    * \[ ] Max replicas: `5`
    * \[ ] Target CPU utilization: `70%`

#### 3.4 Frontend

* \[ ] Create `frontend.yaml` containing:

  * \[ ] **Deployment** with:

    * \[ ] Correct container image from your registry
    * \[ ] Container port `80`
    * \[ ] Resource **requests**: CPU `30m`, Memory `32Mi`
    * \[ ] Resource **limits**: CPU `100m`, Memory `64Mi`
    * \[ ] **Liveness probe** on `/` port `80`
    * \[ ] **Readiness probe** on `/` port `80`
    * \[ ] (If private registry) `imagePullSecrets`
  * \[ ] **Service** (ClusterIP) exposing port `80`
  * \[ ] **HorizontalPodAutoscaler**

    * \[ ] Min replicas: `2`
    * \[ ] Max replicas: `10`
    * \[ ] Target CPU utilization: `70%`

#### 3.5 Ingress (or alternative access)

* \[ ] **Option A — Ingress:** Create an Ingress resource pointing to the frontend service on port 80
* \[ ] **Option B — NodePort:** Change frontend Service type to `NodePort`
* \[ ] **Option C — Port-forward** (for testing only):

```bash
  kubectl port-forward svc/frontend 8080:80 -n guestbook
  ```

\---

### Part 4 — Deploy \& Verify

#### Deployment Order

Deploy in this order (dependencies matter):

```
1. namespace
2. PVCs
3. auth-service    (no dependencies)
4. booking-service (depends on auth-service)
5. frontend        (depends on both backend services)
```

* \[ ] Apply all manifests in the correct order
* \[ ] All pods are `Running`: `kubectl get pods -n guestbook`
* \[ ] All services created: `kubectl get svc -n guestbook`
* \[ ] PVCs are `Bound`: `kubectl get pvc -n guestbook`
* \[ ] HPAs registered: `kubectl get hpa -n guestbook`

#### Functional Testing

* \[ ] Open the frontend in browser — homepage loads
* \[ ] Homepage shows "No guestbook entries yet" (empty state)
* \[ ] Go to `/login`, log in with `admin` / `admin123`
* \[ ] After login, the Add Entry form appears
* \[ ] Submit a guestbook entry — it appears in the list
* \[ ] Submit 2-3 more entries — all display correctly
* \[ ] Delete an entry — it disappears
* \[ ] Logout — form and delete buttons disappear, entries remain visible
* \[ ] Check auth-service logs: `kubectl logs -l app=auth-service -n guestbook`
* \[ ] Check booking-service logs: `kubectl logs -l app=booking-service -n guestbook`

#### Persistence Testing

* \[ ] Delete the booking-service pod: `kubectl delete pod -l app=booking-service -n guestbook`
* \[ ] Wait for it to restart, verify guestbook entries are still there (PVC worked)
* \[ ] Delete the auth-service pod, verify you need to log in again (session lost or preserved depending on PVC)

\---

### Part 5 — Troubleshooting Commands (Reference)

```bash
# Check events on a failing pod
kubectl describe pod <pod-name> -n guestbook

# Stream logs
kubectl logs -f <pod-name> -n guestbook

# Exec into a container
kubectl exec -it <pod-name> -n guestbook -- /bin/sh

# Check service endpoints resolve
kubectl get endpoints -n guestbook

# Watch HPA scaling
kubectl get hpa -n guestbook -w

# DNS test from inside the cluster
kubectl run dns-test --rm -it --namespace guestbook \\
  --image=alpine -- nslookup auth-service.guestbook.svc.cluster.local

# Check resource usage (requires metrics-server)
kubectl top pods -n guestbook
```



\---

### Bonus Challenges

* \[ ] Create a `ConfigMap` for environment variables instead of hardcoding in Deployment
* \[ ] Create a `Secret` for the admin credentials (`ADMIN\_USER`, `ADMIN\_PASS`)
* \[ ] Add a `NetworkPolicy` that restricts traffic:

  * Frontend → auth-service and booking-service only
  * Booking-service → auth-service only
  * No other inter-pod communication
* \[ ] Add `PodDisruptionBudget` for each service
* \[ ] Enable TLS on the Ingress (self-signed or cert-manager)
* \[ ] Increase auth-service replicas — observe what breaks with file sessions on `ReadWriteOnce` PVC, and propose a solution

\---

