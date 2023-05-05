NAME      := terraform-provider-warren
REPO_ROOT := $(shell dirname $(realpath $(lastword ${MAKEFILE_LIST})))
HACK_DIR  := ${REPO_ROOT}/hack
VERSION   := $(shell cat "${REPO_ROOT}/VERSION")
LD_FLAGS  := "-w $(shell $(HACK_DIR)/get-build-ld-flags.sh gitlab.com/warrenio/library/terraform-provider-warren $(REPO_ROOT)/VERSION $(NAME))"

#########################################
# Rules for local development scenarios #
#########################################

.PHONY: start
start:
	GO111MODULE=on \
	CSI_ENDPOINT=unix://@0 \
	ENABLE_METRICS=true \
	go run \
		-ldflags ${LD_FLAGS} \
		./cmd/${NAME} \
		--debug

.PHONY: debug
debug:
	CSI_ENDPOINT=unix://@0 \
	ENABLE_METRICS=true \
	dlv debug ./cmd/${NAME} -- \
		--debug

#########################################
# Rules for re-vendoring
#########################################

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod tidy -compat=1.19
	@GO111MODULE=on go mod vendor

.PHONY: update-dependencies
update-dependencies:
	@env GO111MODULE=on go get -u

#########################################
# Rules for testing
#########################################

.PHONY: test
test:
	@$(HACK_DIR)/test.sh

.PHONY: test-cov
test-cov:
	@$(HACK_DIR)/test.sh --coverage

.PHONY: test-clean
test-clean:
	@$(HACK_DIR)/test.sh --clean --coverage

#########################################
# Rules for build/release
#########################################

.PHONY: clean-examples
clean-examples:
	@env GO111MODULE=on terraform fmt -recursive $(REPO_ROOT)/examples/

.PHONY: generate-docs
generate-docs:
	@env GO111MODULE=on go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

.PHONY: build-local
build-local:
	@env LD_FLAGS=${LD_FLAGS} LOCAL_BUILD=1 $(HACK_DIR)/build.sh

.PHONY: build
build:
	@env LD_FLAGS=${LD_FLAGS} $(HACK_DIR)/build.sh
