.PHONY: build run test integration load lint metrics

build:
	go build -o goboxd ./cmd/goboxd

run:
	docker-compose up --build

test:
	go test ./...

integration:
	go test ./tests/...

load:
	hey -n 1000 -c 50 -m POST \
		-H "Content-Type: application/json" \
		-d '{"language":"py3","source":"print(\"hello\")","tests":[{"stdin":"","expected_stdout":"hello"}]}' \
		http://localhost:8080/run

lint:
	go vet ./...

metrics:
	@echo "Prometheus metrics available at http://localhost:8080/metrics"
	@curl -s http://localhost:8080/metrics | grep goboxd | head -20