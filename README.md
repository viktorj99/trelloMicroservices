iz Backend foldera pozvati ovu funkciju kad god hocete da kreirate GRPC api pozive

protoc --proto_path=proto proto/\*.proto --go_out=pb --go-grpc_out=pb