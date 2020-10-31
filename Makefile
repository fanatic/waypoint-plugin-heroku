all: protos build

protos:
	@echo ""
	@echo "Build Protos"

	protoc -I . --go_opt=plugins=grpc --go_out=../../../ ./output.proto

build:
	@echo ""
	@echo "Compile Plugin"

	go build -o ./bin/waypoint-plugin-heroku ./*.go 

install: build
	@echo ""
	@echo "Installing Plugin"

	cp ./bin/waypoint-plugin-heroku ${HOME}/.config/waypoint/plugins/