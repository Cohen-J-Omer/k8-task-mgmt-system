ifndef DOCKER_USER
$(error DOCKER_USER environment variable is not set)
endif

TAG = $(shell git ls-remote https://github.com/Cohen-J-Omer/k8-task-mgmt-system.git HEAD | awk '{print $$1}')
TAG_MANUAL := $(if $(TAG_MANUAL),$(TAG_MANUAL),latest) # Use TAG_MANUAL env var if set, else 'latest'
API_IMAGE=$(DOCKER_USER)/task-api:$(TAG_MANUAL)
BACKEND_IMAGE=$(DOCKER_USER)/task-backend:$(TAG_MANUAL)
RENDERED_DIR = k8s-rendered

# deploy cluster using images built by GitHub Actions
deploy:
	rm -rf $(RENDERED_DIR)
	mkdir -p $(RENDERED_DIR)
	cp k8s/* $(RENDERED_DIR)
	cd $(RENDERED_DIR) && \
		kustomize edit set image "DOCKER_USER/task-api=${DOCKER_USER}/task-api:${TAG}" && \
		kustomize edit set image "DOCKER_USER/task-backend=${DOCKER_USER}/task-backend:${TAG}" && \
		kustomize build . >  all.yaml && \
		kubectl apply -f all.yaml

# allows access to REST API by Forwarding the API service to http://localhost:8080
port-forward:
	minikube kubectl -- port-forward svc/api 8080:80 -n task-mgmt

# start minikube with kvm2 driver
minikube:
	minikube start --driver=kvm2

# dev cmd: update gRPC service definitions when changes to proto/* files are made
build-grpc:
	protoc --proto_path=taskmgmt/proto --go_out=taskmgmt/proto --go_opt=paths=source_relative \
	--go-grpc_out=taskmgmt/proto --go-grpc_opt=paths=source_relative taskmgmt/proto/task.proto

# build docker images locally, instead of using images pushed by GitHub Actions
build-local:
	docker buildx build --platform linux/amd64 -f Dockerfile.api -t $(API_IMAGE) .
	docker buildx build --platform linux/amd64 -f Dockerfile.backend -t $(BACKEND_IMAGE) .

# push locally built images instead of using images pushed by GitHub Actions
push-manual:
	docker push $(API_IMAGE)
	docker push $(BACKEND_IMAGE)

# render YAMLs with DOCKER_USER substituted
render-yamls:
	rm -rf $(RENDERED_DIR)
	mkdir -p $(RENDERED_DIR)
	for f in k8s/*.yaml; do \
		envsubst < $$f > $(RENDERED_DIR)/$$(basename $$f); \
	done

# deploy cluster using manually picked image tags
deploy-manual: render-yamls
	# deployment of namespace is done separately to ensure the namespace exists before applying other resources
	minikube kubectl -- apply -f $(RENDERED_DIR)/namespace.yaml
	minikube kubectl -- apply -f $(RENDERED_DIR)

