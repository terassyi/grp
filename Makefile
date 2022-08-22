PROJECT := github.com/terassyi/grp

GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean

GRPD_BINARY := grpd
GRP_BINARY := grp
GRPD_DIR := ./cmd/grpd
GRP_DIR := ./cmd/grp

COMPOSE := docker compose

BGP_TEST_COMPOSE := ./scenario/bgp-test/docker-compose.dev.yml
BGP_TEST_GRP_CONTAINER_NAME := pe1

TINET_SPEC := ./scenario/bgp/tinet-spec.yml
TINET_FRR_SPEC := ./scenario/bgp/tinet-frr-spec.yml

.PHONY: build
build:
	$(GOBUILD) -o $(GRP_BINARY) $(GRP_DIR)
	$(GOBUILD) -o $(GRPD_BINARY) $(GRPD_DIR)

.PHONY: protogen
protogen:
	protoc -Ipb --go_out=module=$(PROJECT)/pb:pb grp.proto
	protoc -Ipb --go-grpc_out=module=$(PROJECT)/pb:pb grp.proto
	protoc -Ipb --go_out=module=$(PROJECT)/pb:pb bgp.proto
	protoc -Ipb --go-grpc_out=module=$(PROJECT)/pb:pb bgp.proto
	protoc -Ipb --go_out=module=$(PROJECT)/pb:pb route.proto
	protoc -Ipb --go-grpc_out=module=$(PROJECT)/pb:pb route.proto
	protoc -Ipb --go_out=module=$(PROJECT)/pb:pb rip.proto
	protoc -Ipb --go-grpc_out=module=$(PROJECT)/pb:pb rip.proto

.PHONY: up
up:
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) up -d 

.PHONY: down
down:
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) down 

.PHONY: test
test:
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) up -d 
	@echo "----- Unit Test -----"
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) exec $(BGP_TEST_GRP_CONTAINER_NAME) $(GOTEST) -cover ./...

	$(COMPOSE) -f $(BGP_TEST_COMPOSE) down 

.PHONY: bgp_test
bgp_test: up
	@echo "----- Unit Test -----"
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) exec $(BGP_TESt_GRP_CONTAINER_NAME) $(GOTEST) -v ./...
	@echo "----- Integration Test -----"

.PHONY: bgp_test_down
bgp_test_down:
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) down 

.PHONY:

.PHONY: clean
clean: down
	$(GOCLEAN)
	rm $(GRP_BINARY)
	rm $(GRPD_BINARY)

tinet.%:
	@echo "# tinet"
	if [ ${@:tinet.%=%} = "upconf" ]; then \
		make; \
	elif [ ${@:tinet.%=%} = "up" ]; then \
		make; \
	fi
	tinet ${@:tinet.%=%} -c $(TINET_SPEC) | sudo sh -x

tinet-exec.%:
	docker exec -it ${@:tinet-exec.%=%} bash

tinet-frr.%:
	tinet ${@:tinet-frr.%=%} -c $(TINET_FRR_SPEC) | sudo sh -x

