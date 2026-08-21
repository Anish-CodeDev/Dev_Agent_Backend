agent:
	@protoc \
		--go_out=common --go_opt=paths=source_relative \
		--go-grpc_out=common --go-grpc_opt=paths=source_relative \
		config.proto

grpc-server:
	@go run ./agentservice

grpc-client:
	@go run .