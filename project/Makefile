AUTH_BINARY=authApp
TRANS_BINARY=transactionApp
LIS_BINARY=listenerApp

export DB_USER=renan
export DB_PASSWORD=admin123
export DB_ROOT_PASSWORD=admin123
export TOKEN_SECRET=and0X3NlY3JldA==
export TRANSACTION_DB_USER=renan
export TRANSACTION_DB_PASSWORD=admin123
export TRANSACTION_DB_ROOT_PASSWORD=admin123

up:
	@echo "Starting Docker images..."
	docker compose up -d
	@echo "Docker images started!"

down:
	@echo "Stopping docker compose..."
	docker compose down
	@echo "Done!"

up_build: build_auth build_listener build_transaction
	@echo "Stopping docker images (if running...)"
	docker compose down
	@echo "Building (when required) and starting docker images..."
	docker compose up --build -d
	@echo "Docker images built and started!"

build_auth:
	@echo "Building auth service image.."
	cd ../auth && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BINARY} ./cmd/api
	@echo "Done!"

build_transaction:
	@echo "Building transaction service image.."
	cd ../transaction && env GOOS=linux CGO_ENABLED=0 go build -o ${TRANS_BINARY} ./cmd/api
	@echo "Done!"

build_listener:
	@echo "Building listener service image.."
	cd ../listener && env GOOS=linux CGO_ENABLED=0 go build -o ${LIS_BINARY} ./cmd/api
	@echo "Done!"
