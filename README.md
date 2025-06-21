# Kubernetes Task Management System

A containerized task management system with a REST API (built using Go and Gin), a gRPC backend and MongoDB, all orchestrated on Kubernetes.

---

## Quick Start

### Prerequisites

- Go 1.21+
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

---

### 2. Build and Push Docker Images

Set env var DOCKER_USER as your Docker Hub username: `export DOCKER_USER=<user-name>`.  
build and push images:

```
make build
make push
```

---

### 3. Start Kubernetes Locally

```
make minikube
```

---

### 4. Deploy All Components

```
make deploy
```

---

### 5. Access the REST API

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

```
python taskmgmt/load_test.py
```

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
2. Start MongoDB with docker compose: `docker compose -f taskmgmt/testlocal/docker-compose.mongodb.yml up -d`
3. Run the Backend service locally: `go run taskmgmt/cmd/backend/main.go`
4. Run the API service locally: `go run taskmgmt/cmd/api/main.go`
5. Test the local stack by running CRUD operations listed in the above segment.

## Notes

- The REST API uses the [Gin](https://gin-gonic.com/) framework for fast HTTP routing and middleware.
- All secrets are managed via Kubernetes Secrets.
- MongoDB is only accessible from the backend service.
- HPA is enabled for both API and backend deployments.
- Images are built for `linux/amd64` and should be pushed to Docker Hub.
- Designed to run locally on macOS/Linux with Minikube.
