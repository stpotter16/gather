shell:
	nix develop -c $$SHELL


server/build:
	./dev-scripts/build-server.sh

server/run: server/build
	./tmp/server

server/live:
	./dev-scripts/serve.sh

server/deploy:
	./dev-scripts/deploy.sh

db/migrate:
	go run ./cmd/migrate

user/hash:
	@go run ./cmd/hash


css/build:
	npx @tailwindcss/cli -i style/input.css -o internal/handlers/static/css/style.css --minify


secrets/hmac:
	xxd -l32 /dev/urandom | xxd -r -ps | base64 | tr -d = | tr + - | tr / _


lint/go:
	./dev-scripts/check-go.sh

lint/shell:
	./dev-scripts/check-shell.sh


test/go:
	go test ./...
