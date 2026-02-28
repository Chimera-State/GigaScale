PROTO_DIR=api/proto
MODULE=github.com/Chimera-State/GigaScale
# Windows için ters slash kullanıyoruz
OUT_DIR=internal\backend

.PHONY: proto clean

proto:
	@if not exist $(OUT_DIR) mkdir $(OUT_DIR)
	protoc --proto_path=$(PROTO_DIR) --go_out=. --go_opt=module=$(MODULE) --go-grpc_out=. --go-grpc_opt=module=$(MODULE) $(PROTO_DIR)/reservation.proto

clean:
	@if exist $(OUT_DIR) rmdir /s /q $(OUT_DIR)