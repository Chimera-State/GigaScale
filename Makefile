PROTO_DIR=api/proto/reservation/v1
MODULE=github.com/Chimera-State/GigaScale

.PHONY: proto clean

proto:
	protoc -I . \
	  --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  $(PROTO_DIR)/reservation.proto

clean:
	rm -f $(PROTO_DIR)/*.pb.go $(PROTO_DIR)/*.gw.go