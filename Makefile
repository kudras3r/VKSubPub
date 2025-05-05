PROTO_DIR=proto
PROTO_FILE=$(PROTO_DIR)/sp.proto
GO_OUT_DIR=$(PROTO_DIR)

GENERATE_PROTO=protoc \
	--go_out=$(GO_OUT_DIR) \
	--go-grpc_out=$(GO_OUT_DIR) \
	--proto_path=$(PROTO_DIR) \
	$(PROTO_FILE)

.PHONY: proto clean

proto:
	$(GENERATE_PROTO)

clean:
	rm -f $(PROTO_DIR)/*.pb.go
