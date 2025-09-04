.PHONY: generate
generate:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --include_imports --descriptor_set_out=./pb/task.protoset pb/task.proto