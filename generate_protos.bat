@rem Generate the go code for .proto files
.\protoc.exe -I ./protos Ares.proto --go_out=plugins=grpc:protos
