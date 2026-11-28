# Common targets
.PHONY: all check format format-whitespace check-format test test-rbe lint fix install generate-pkl tag

# Stable PATH for Bazel tests under --incompatible_strict_action_env.
BAZEL_TEST_PATH := $(shell \
	extra=$$(for t in go lsof pkl; do \
		p=$$(command -v "$$t" 2>/dev/null || true); \
		[ -n "$$p" ] && dirname "$$p"; \
	done | grep -vxE '/usr/local/bin|/usr/bin|/bin|/usr/sbin|/sbin' | LC_ALL=C sort -u | paste -sd: -); \
	if [ -n "$$extra" ]; then \
		printf '%s:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin' "$$extra"; \
	else \
		printf '/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin'; \
	fi)
BAZEL_TEST_ENV := --test_env=PATH=$(BAZEL_TEST_PATH)

all:
	$(MAKE) generate-pkl
	$(MAKE) format
	$(MAKE) fix
	$(MAKE) test

check: test

format: format-whitespace
	bazel run //:format

format-whitespace:
	bazel run //tools/format/whitespace -- -w

check-format:
	bazel test --test_summary=terse $(BAZEL_TEST_ENV) //tools/format/...

test:
	bazel test --test_summary=terse $(BAZEL_TEST_ENV) $(BAZEL_FLAGS) //...

test-rbe:
	bazel test --config=rbe --test_summary=terse --build_tag_filters=-local,-no-remote --test_tag_filters=-local,-no-remote $(BAZEL_FLAGS) //...
	bazel test --test_summary=terse $(BAZEL_TEST_ENV) --build_tag_filters=local,no-remote --test_tag_filters=local,no-remote $(BAZEL_FLAGS) //...

lint:
	bazel test --test_summary=terse $(BAZEL_TEST_ENV) //tools/lint/...

fix: fix-go

fix-go:
	GOPACKAGESDRIVER= go tool oserrorsgodernize --fix ./...

generate-pkl:
	bazel run //gh-nudge/internal/config/pkl:generate
	bazel run //orgctl/pkl:generate

INSTALL_CMDS := \
	./agentbus/cmd/agentbus \
	./agentbus/cmd/agentbusd \
	./decide/cmd/decide \
	./docsync/cmd/docsync \
	./factory/cmd/factory \
	./genfiles/cmd/genfiles \
	./gh-ask/cmd/gh-ask \
	./gh-nudge/cmd/gh-approve \
	./gh-nudge/cmd/gh-codeowners \
	./gh-nudge/cmd/gh-merge \
	./gh-nudge/cmd/gh-nudge \
	./gh-nudge/cmd/gh-slack \
	./gh-nudge/cmd/gh-storage \
	./gh-nudge/cmd/gh-throttle \
	./orgctl/cmd/orgctl \
	./prsync/cmd/prsync

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
VERSION_LDFLAGS := -ldflags "-X github.com/chorong-sh/chorong/pkg/version.Version=$(VERSION)"

install:
	GOPACKAGESDRIVER= go install $(VERSION_LDFLAGS) $(INSTALL_CMDS)

tag:
	@if [ -z "$(TAG)" ]; then echo "Usage: make tag TAG=vX.Y.Z"; exit 1; fi
	@echo "$(TAG)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$' || { echo "Error: TAG must match vMAJOR.MINOR.PATCH"; exit 1; }
	@if [ -f .release-please-manifest.json ] && [ "$(FORCE)" != "1" ]; then \
		manifest_ver=$$(python3 -c 'import json; print(json.load(open(".release-please-manifest.json")).get(".", ""))'); \
		if [ "v$$manifest_ver" != "$(TAG)" ]; then \
			echo "Error: .release-please-manifest.json has $$manifest_ver (expected $(TAG)). Normal releases are cut by squash-merging the Release PR; pass FORCE=1 to override."; \
			exit 1; \
		fi; \
	fi
	git tag -a "$(TAG)" -m "Release $(TAG)"
	@echo "Created tag $(TAG). Push with: git push origin $(TAG)"
