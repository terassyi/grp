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

BGP_DEV_COMPOSE := ./scenario/bgp/docker-compose.dev.yml

TINET_SPEC := ./scenario/bgp/tinet-spec.yml
TINET_FRR_SPEC := ./scenario/bgp/tinet-frr-spec.yml

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
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec r0 /tmp/r0/bgp-100.sh
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec r1 /tmp/r1/bgp-200.sh
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec r2 /tmp/r2/bgp-300.sh 
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec r3 /tmp/r3/bgp-400.sh
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec c0 /tmp/c0/setup.sh
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec c1 /tmp/c1/setup.sh

.PHONY: down
down:
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) down 

.PHONY: dev-down
dev-down:
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) down 

.PHONY: test
test:
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) up -d 
	@echo "----- Unit Test -----"
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) exec $(BGP_TEST_GRP_CONTAINER_NAME) $(GOTEST) -v ./...
	@echo "----- Integration Test -----"
	$(COMPOSE) -f $(BGP_TEST_COMPOSE) exec $(BGP_TEST_GRP_CONTAINER_NAME) $(GOTEST) -tags=integration -v ./...

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

exec.%:
	$(COMPOSE) -f $(BGP_DEV_COMPOSE) exec ${@:exec.%=%} bash

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

