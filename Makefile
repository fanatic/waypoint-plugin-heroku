all: protos build

protos:
	@echo ""
	@echo "Build Protos"

	protoc -I . --go_opt=plugins=grpc --go_out=../../../ ./builder/output.proto
	protoc -I . --go_opt=plugins=grpc --go_out=../../../ ./registry/output.proto
	protoc -I . --go_opt=plugins=grpc --go_out=../../../ ./platform/output.proto
	protoc -I . --go_opt=plugins=grpc --go_out=../../../ ./release/output.proto

build:
	@echo ""
	@echo "Compile Plugin"

	go build -o ./bin/waypoint-plugin-heroku ./main.go 

install:
	@echo ""
	@echo "Installing Plugin"

	cp ./bin/waypoint-plugin-heroku ${HOME}/.config/waypoint/plugins/