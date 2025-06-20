ifndef DOCKER_USER
$(error DOCKER_USER environment variable is not set)
endif

API_IMAGE=$(DOCKER_USER)/task-api:latest
BACKEND_IMAGE=$(DOCKER_USER)/task-backend:latest
RENDERED_DIR = k8s-rendered

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
	# deployment of namespace is done separately to ensure the namespace exists before applying other resources
	minikube kubectl -- apply -f $(RENDERED_DIR)/namespace.yaml
	minikube kubectl -- apply -f $(RENDERED_DIR)

minikube:
	minikube start --driver=docker

port-forward:
	minikube kubectl -- port-forward svc/api 8080:80 -n task-mgmt
