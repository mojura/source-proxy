build:
	@go build;
install:
	@go install;
install_linux: install
	@sudo setcap 'cap_net_bind_service=+ep' ~/go/bin/bidder;
build_docker_image: 
	@docker build . --tag=source-proxy;