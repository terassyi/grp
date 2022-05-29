GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean

GRPD_BINARY := grpd
GRP_BINARY := grp
GRPD_DIR := ./cmd/grpd
GRP_DIR := ./cmd/grp

COMPOSE := docker compose

BGP_TEST_COMPOSE := ./scenario/bgp-test/docker-compose.yml
BGP_TESt_GRP_CONTAINER_NAME := pe1

BGP_DEV_COMPOSE := ./scenario/bgp-test/docker-compose.dev.yml

.PHONY: build
build:
	$(GOBUILD) -o $(GRP_BINARY) $(GRP_DIR)
	$(GOBUILD) -o $(GRPD_BINARY) $(GRPD_DIR)

.PHONY: up
up:
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) up -d 

.PHONY: dev-up
dev-up:
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) up -d 

.PHONY: down
down:
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) down 

.PHONY: dev-down
dev-down:
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) down 

.PHONY: dev-test
dev-test: dev-up
	@echo "----- Unit Test -----"
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec $(BGP_TESt_GRP_CONTAINER_NAME) $(GOTEST) -v ./...
	@echo "----- Integration Test -----"


.PHONY: bgp_test
bgp_test: up
	@echo "----- Unit Test -----"
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) exec $(BGP_TESt_GRP_CONTAINER_NAME) $(GOTEST) -v ./...
	@echo "----- Integration Test -----"

exec.%:
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec ${@:exec.%=%} bash



.PHONY: test
test: up

.PHONY:

.PHONY: clean
clean: down
	$(GOCLEAN)
	rm $(GRP_BINARY)
	rm $(GRPD_BINARY)
