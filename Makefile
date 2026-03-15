API_PROTO_DIR=api/proto
MODULE=github.com/Chimera-State/GigaScale

.PHONY: proto clean

proto:
	protoc -I . \
	  --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  $(API_PROTO_DIR)/reservation/v1/reservation.proto
	protoc -I . \
	  --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  $(API_PROTO_DIR)/payment/v1/payment.proto

clean:
	rm -f $(API_PROTO_DIR)/reservation/v1/*.pb.go $(API_PROTO_DIR)/reservation/v1/*.gw.go
	rm -f $(API_PROTO_DIR)/payment/v1/*.pb.go $(API_PROTO_DIR)/payment/v1/*.gw.go