.PHONY: generate
generate:
	protoc -I./pb --go_out=plugins=grpc:$(GOPATH)/src --include_imports --descriptor_set_out=./pb/task.protoset pb/task.proto