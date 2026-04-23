.PHONY: build test run-server run-tester docker-build docker-test docker-run docker-logs docker-stop clean

# Default target: build the project
build:
	cargo build --release

# Run unit tests
test:
	cargo test

# Run the server locally (requires OpenCV installed)
run-server:
	cargo run --release --bin server

# Run the tester locally (requires OpenCV installed)
# Usage: make run-tester URL="http://example.com/image.jpg"
run-tester:
	cargo run --release --bin tester -- --url "$(URL)"

# Build the docker image
docker-build:
	docker build -t watermark-remover .

# Run the integration test using the docker image
# Usage: make docker-test URL="http://example.com/image.jpg"
docker-test:
	./run_test.sh "$(URL)" --build

# Run the test with a sample car image URL
test-sample:
	./run_test.sh "https://img.nomadocars.com/unsafe/rs:fit:0:0/plain/s3://sorted-tote-m1bwf4wyssq-2/carpicture09/pic4189/41895714_001.jpg"

# Run the server in docker (detached)
docker-run:
	docker compose up -d

# Watch logs of the running docker container
docker-logs:
	docker compose logs -f

# Stop and remove the docker containers
docker-stop:
	docker compose down

# Clean build artifacts
clean:
	cargo clean
	rm -rf output/
