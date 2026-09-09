.PHONY: test backend-test frontend-build

backend-test:
	cd backend && go test ./...

frontend-build:
	cd frontend && npm run build

test: backend-test frontend-build
