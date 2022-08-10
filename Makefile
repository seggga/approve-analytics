PROTO_FILE := pkg/proto/task-msg-v1.proto

run_app:
	ANALYTICS_REST_PORT=3001 AUTH_PORT_4000_TCP_PORT=40533 go run ./cmd/main.go

compose/up:
	docker-compose -f stack_postgres.yaml up -d

compose/down:
	docker-compose -f stack_postgres.yaml down

gen_proto:
	mkdir -p pkg/proto && \
	protoc  proto/*.proto --go-grpc_out=pkg --go_out=pkg

swag:
	swag init \
		--parseDependency \
		--parseInternal \
		--dir ./internal/adapters/rest \
		--generalInfo swagger.go \
		--output ./api/swagger/public

kafka/compose/up:
	docker-compose -f stack_kafka.yaml up -d

kafka/compose/down:
	docker-compose -f stack_kafka.yaml down