# Kubernetes Task Management System

A containerized task management system with a REST API (built using Go and Gin), a gRPC backend and MongoDB, all orchestrated on Kubernetes.

---

## Quick Start

### Prerequisites

- Go 1.23+
- Docker
- Minikube
- Python 3 (for load testing)
- `protoc` (for development)

---

### 1. Clone and Prepare

```
git clone https://github.com/Cohen-J-Omer/k8-task-mgmt-system.git
cd k8-task-mgmt-system
```

### 2. Start Kubernetes Locally

```
make minikube
```

---

### 3. Deploy the Cluster (Choose One Route)

**3.a. Recommended: Deploy with Latest Images from GitHub Actions**

This is a one-stop shop that deploys the cluster using the latest commit hash in the project's repo, utilizing the GitHub Actions CD pipeline.

```
make deploy
```

**3.b. Manual: Build, Push, and Deploy Local Images (for dev/debugging purposes)**

Run these commands if you want to build, push, and deploy using your own local images, instead of latest commit-based image:

```
export DOCKER_USER=<your-dockerhub-username>
make build-local
make push-manual
make deploy-manual
```

---

### 4. Access the REST API

Forward the API service:

```
make port-forward
```

REST API will be available at [http://localhost:8080](http://localhost:8080).

---

## CRUD Examples (cURL)

### Create a Task

```
curl -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer hardcoded-token" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Task","description":"A test","completed":false}'
```

### Get All Tasks

```
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer hardcoded-token"
```

### Get a Single Task by ID

```
curl -X GET http://localhost:8080/tasks/{id} \
  -H "Authorization: Bearer hardcoded-token"
```

### Update a Task

```
curl -X PUT http://localhost:8080/tasks/{id} \
  -H "Authorization: Bearer hardcoded-token" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title","description":"Updated description","completed":true}'
```

### Delete a Task

```
curl -X DELETE http://localhost:8080/tasks/{id} \
  -H "Authorization: Bearer hardcoded-token"
```

---

## Load Testing

To simulate load and trigger autoscaling:
1. Enable the metrics addon for hpa pods: `minikube addons enable metrics-server`.
2. Wait for metrics server deployment to become ready: `minikube kubectl -- get deployment metrics-server -n kube-system`.
Verify deployment is integrated successfully via: `minikube kubectl -- top pods -n task-mgmt`. 
3. Run `python tests/load_test.py` and inspect scaling via either: `minikube kubectl get pods -n task-mgmt`, or `minikube kubectl get hpa -n task-mgmt`

---

## Local Testing & Debugging

1. Set Your Environment Variables in `.env`:
```bash
MONGO_USERNAME=mongo-user
MONGO_PASSWORD=mongo-password
BEARER_TOKEN=hardcoded-token
DEBUG_TASK_MGMT=true
BACKEND_GRPC_ADDR=localhost:50051
```
2. Start MongoDB with docker compose: `docker compose -f devtools/docker-compose.mongodb.yml up -d`
3. Run the Backend service locally: `go run taskmgmt/cmd/backend/main.go`
4. Run the API service locally: `go run taskmgmt/cmd/api/main.go`
5. Test the local stack by running CRUD operations listed in the above segment.

## Notes

- The REST API uses the [Gin](https://gin-gonic.com/) framework for fast HTTP routing and middleware.
- All secrets are managed via Kubernetes Secrets.
- MongoDB is only accessible from the backend service.
- HPA (Horizontal Pod Autoscaling) is enabled for both API and backend deployments.
- Continuous Deployment (CD) is incorporated via GitHub Actions, building and pushing images based on the latest commit.
- Images are built for `linux/amd64` and are pushed to Docker Hub.
- Designed to run locally on macOS/Linux with Minikube.
