.PHONY: dev build frontend go-test go-race typecheck install-desktop desktop desktop-dev

## Frontend

dev:
	nub run dev

frontend:
	nub run build

typecheck:
	nub run -r --if-present typecheck

## Go core

go-test:
	go test ./...

go-race:
	go test -race ./...

go-bench:
	go test -bench=. ./...

## Desktop (requires Wails v3 + system WebKit deps)

install-desktop:
	cd apps/desktop/backend && wails3 generate bindings -d frontend/bindings -i -ts

desktop:
	cd apps/desktop/backend && wails3 build

desktop-dev:
	cd apps/desktop/backend && wails3 dev