.PHONY: help build clean
.DEFAULT_GOAL := build

help: ## This help-screen
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build: goimports -> go fmt -> go build
	# goimports will "go fmt" after adding/removing import-statements
	find ./../ -path ./vendor -prune -o -name "*.go" -print | xargs goimports -w
	rm simg2img || true
	go build -o simg2img .

clean: ## Delete executables
	rm simg2img || true
