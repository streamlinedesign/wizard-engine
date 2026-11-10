SHELL := bash
.DEFAULT_GOAL := help

# ---- Project ----------------------------------------------------------------
# wizard-engine: a desktop game (Go + Ebitengine), not a server. PORT is the
# block from PORTS.md in streamlinedesign/template, reserved for tool servers
# under tools/ only; the game itself opens no port. .env overrides.
NAME     ?= wizard
KIND     ?= go
PORT     ?= 4160
DROPLET  ?=
# -----------------------------------------------------------------------------

-include .env
export

VERSION := $(shell cat VERSION 2>/dev/null || echo 0.0.0)
LDFLAGS := -s -w -X main.version=$(VERSION)

help: ## List targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

# The game is a window, not a server: it does not listen on a port. The phone
# URL line is kept for any tool server under tools/ that binds 0.0.0.0:PORT.
dev: ## Run the game from source (go run ./cmd/sandbox)
	@echo "On your phone: http://$$(hostname -I 2>/dev/null | cut -d' ' -f1 || ipconfig getifaddr en0):$(PORT) (tool servers only; the game opens a window)"
	go run ./cmd/sandbox $(ARGS)

test: ## All tests (headless; no window is opened)
	go test ./...

build: ## One binary at ./wizard
	go build -trimpath -ldflags="$(LDFLAGS)" -o $(NAME) ./cmd/sandbox

lint: ## gofmt and go vet
	gofmt -l -w $$(git ls-files '*.go')
	go vet ./...

placeholders: ## Regenerate the stand-in art under assets/
	go run ./tools/mkplaceholders

art: ## Redraw the wizard, spell effects and campus tiles into assets/ (needs python3 + Pillow)
	python3 art/gen/wizard.py assets/sprites
	python3 art/gen/effects.py assets/effects
	python3 art/gen/tiles.py assets/tiles
	rm -f assets/sprites/wizard_preview.png assets/effects/effects_preview.png assets/tiles/tiles_preview.png
	python3 art/gen/preview.py assets docs/art

mockup: ## Render the campus scene in each art style to mockup/
	go run ./tools/mockup mockup/

screenshot: ## Render one frame of the campus to docs/campus.png
	go run ./cmd/sandbox -screenshot docs/campus.png -ticks 60

db-reset: ## Drop and recreate the local database, run migrations
	@echo "make $@: not used in this project"

db-seed: ## Load demo data from seed/
	@echo "make $@: not used in this project"

psql: ## Shell on the local database
	@echo "make $@: not used in this project"

deploy: ## Build for linux/amd64 and ship to DROPLET
	@echo "make $@: not used in this project"

release: ## Tag VERSION, move Unreleased in CHANGELOG.md, push the tag
	@test "$$(git branch --show-current)" = main || { echo "release from main only"; exit 1; }
	@git diff --quiet || { echo "commit first"; exit 1; }
	@grep -q "## \[Unreleased\]" CHANGELOG.md || { echo "CHANGELOG.md needs an Unreleased section"; exit 1; }
	sed -i.bak "s/## \[Unreleased\]/## [Unreleased]\n\n## [$(VERSION)] - $$(date +%Y-%m-%d)/" CHANGELOG.md && rm CHANGELOG.md.bak
	git commit -am "Release $(VERSION)"
	git tag -a v$(VERSION) -m "v$(VERSION)"
	git push && git push --tags

configure-repo: ## One-time GitHub settings: auto-merge, Actions may open PRs, main protected, hold label
	@repo=$$(gh repo view --json nameWithOwner --jq .nameWithOwner); \
	gh api -X PATCH repos/$$repo -F allow_auto_merge=true -F delete_branch_on_merge=true -F allow_squash_merge=true >/dev/null; \
	gh api -X PUT repos/$$repo/actions/permissions/workflow -F default_workflow_permissions=write -F can_approve_pull_request_reviews=true >/dev/null; \
	gh api -X PUT repos/$$repo/branches/main/protection --input - >/dev/null <<< '{"required_status_checks":{"strict":false,"contexts":["ci"]},"enforce_admins":false,"required_pull_request_reviews":null,"restrictions":null,"allow_force_pushes":false,"allow_deletions":false}'; \
	gh label create hold --color B60205 --description "Do not auto-merge" --force >/dev/null; \
	echo "$$repo configured"

clean: ## Remove build output
	rm -rf $(NAME) $(NAME)-linux-amd64 sandbox wizard-engine mockup *.test *.out

.PHONY: help dev test build lint placeholders art mockup screenshot db-reset db-seed psql deploy release configure-repo clean
