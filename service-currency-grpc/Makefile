.PHONY: protos

protos:
	protoc -I=protos/ --go_out=protos protos/currency.proto --go-grpc_out=require_unimplemented_servers=false:protos
