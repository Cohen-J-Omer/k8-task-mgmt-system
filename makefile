ifndef DOCKER_USER
$(error DOCKER_USER environment variable is not set)
endif

TAG ?= $(shell git ls-remote https://github.com/Cohen-J-Omer/k8-task-mgmt-system.git HEAD | awk '{print $$1}')
TAG-manual ?= latest
API_IMAGE=$(DOCKER_USER)/task-api:$(TAG-manual)
BACKEND_IMAGE=$(DOCKER_USER)/task-backend:$(TAG-manual)
RENDERED_DIR = k8s-rendered

build-grpc:
	protoc --proto_path=taskmgmt/proto --go_out=taskmgmt/proto --go_opt=paths=source_relative \
	--go-grpc_out=taskmgmt/proto --go-grpc_opt=paths=source_relative taskmgmt/proto/task.proto

build:
	docker buildx build --platform linux/amd64 -f Dockerfile.api -t $(API_IMAGE) .
	docker buildx build --platform linux/amd64 -f Dockerfile.backend -t $(BACKEND_IMAGE) .

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

deploy-manual: render-yamls
	# deployment of namespace is done separately to ensure the namespace exists before applying other resources
	minikube kubectl -- apply -f $(RENDERED_DIR)/namespace.yaml
	minikube kubectl -- apply -f $(RENDERED_DIR)

deploy:
	rm -rf $(RENDERED_DIR)
	mkdir -p $(RENDERED_DIR)
	kustomize build k8s/ \
	  --image $(DOCKER_USER)/task-api=$(DOCKER_USER)/task-api:$(TAG) \
	  --image $(DOCKER_USER)/task-backend=$(DOCKER_USER)/task-backend:$(TAG) \
	  --load-restrictor=LoadRestrictionsNone \
	  --reorder=legacy \
	  > $(RENDERED_DIR)/all.yaml
	kubectl apply -f $(RENDERED_DIR)/all.yaml

minikube:
	minikube start --driver=docker

port-forward:
	minikube kubectl -- port-forward svc/api 8080:80 -n task-mgmt
