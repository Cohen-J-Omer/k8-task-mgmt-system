ifndef DOCKER_USER
$(error DOCKER_USER environment variable is not set)
endif

DOCKER_USER = $(DOCKER_USER)
API_IMAGE=$(DOCKER_USER)/task-api:latest
BACKEND_IMAGE=$(DOCKER_USER)/task-backend:latest

build:
	docker build -f Dockerfile.api -t $(API_IMAGE) .
	docker build -f Dockerfile.backend -t $(BACKEND_IMAGE) .

push:
	docker push $(API_IMAGE)
	docker push $(BACKEND_IMAGE)

# Render YAMLs with DOCKER_USER substituted
render-yamls:
	rm -rf $(RENDERED_DIR)
	mkdir -p $(RENDERED_DIR)
	for f in k8s/*.yaml; do \
		envsubst < $$f > $(RENDERED_DIR)/$$(basename $$f); \
	done

deploy: render-yamls
	kubectl apply -f $(RENDERED_DIR)

minikube:
	minikube start --driver=docker

port-forward:
	kubectl port-forward svc/api 8080:80 -n task-mgmt
