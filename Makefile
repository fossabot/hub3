.PHONY: package

NAME:=rapid
MAINTAINER:="Sjoerd Siebinga <sjoerd@delving.eu>"
DESCRIPTION:="RAPID Linked Open Data Platform"
MODULE:=github.com/delving/rapid-saas

GO ?= go
TEMPDIR:=$(shell mktemp -d)
VERSION:=$(shell sh -c 'grep "Version = \"" cmd/root.go  | cut -d\" -f2')
GOVERSION:=$(shell sh -c 'go version | cut -d " " -f3')

LDFLAGS:=-X main.Version=$(VERSION) -X main.BuildStamp=`date '+%Y-%m-%d_%I:%M:%S%p'` -X main.GitHash=`git rev-parse HEAD` -X main.BuildAgent=`git config user.email`

# var print rule
print-%  : ; @echo $* = $($*)

clean:
	rm -rf $(NAME) build report gin-bin result.bin *.coverprofile */*.coverprofile hub3/rapid.db hub3/models/rapid.db dist server/assets/assets_vfsdata.go rapidctl/rapidctl rapidctl/*.xml rapidctl/build target rapid-saas results.bin *.log coverage.out webresource

clean-harvesting:
	rm -rf *_ids.txt *_records.xml

clean-build:
	@make clean
	mkdir -p build

create-assets:
	@go run -tags=dev server/assets/assets_generate.go
	mv assets_vfsdata.go server/assets/

run:
	@go run main.go

build:
	@make clean-build
	@make create-assets
	@go build -a -o build/$(NAME) -ldflags=$(LDFLAGS) $(MODULE)

gox-build:
	@make clean-build
	@make create-assets
	cd build 
	@make build 
	gox -os="linux" -os="darwin" -os="windows" -arch="amd64" -ldflags=$(LDFLAGS) -output="build/$(NAME)-{{.OS}}-{{.Arch}}" $(MODULE) 
	ls -la ./build/

run-dev:
	gin -buildArgs "-i -tags=dev -ldflags '${LDFLAGS}'" run http


test:
	@go test  ./...

benchmark:
	@go test --bench=. -benchmem ./...

ginkgo:
	@ginkgo -r  -skipPackage go_tests

twatch:
	@ginkgo watch -r -skipPackage go_tests

docker-image:
	docker build -t $(NAME) .

docker-start:
	docker run -p 3001:3001 -d $(NAME)

docker-stop:
	@sh -c "docker ps -a -q --filter ancestor=$(NAME) | xargs docker stop "

docker-remove:
	@make docker-stop
	@sh -c "docker image list -q rapid | xargs docker image rm -f"

docker-clean-build:
	@make docker-remove; 
	@make docker-image; 
	@make docker-start; 
	docker ps -all

compose-up:
	@docker-compose up

compose-down:
	@docker-compose down

compose-clean:
	@docker-compose down --volumes

goreport:
	@mkdir -p report
	@rm -rf report/*
	@goreporter -p ../rapid-saas -r report -e vendor,rapidctl,cmd,utils -f html

setup-npm:
	# used for getting dependencies to render swagger specifications
	@npm install

release:
	@make create-assets
	@goreleaser --rm-dist --skip-publish
	@rpm --addsign dist/*.rpm
	@debsigs --sign=origin -k E2D6BD239452B1ED15CB99A66C417F6E7521731E dist/*.deb

release-dirty:
	@make create-assets
	@goreleaser --rm-dist --skip-publish --snapshot --skip-validate
	@rpm --addsign dist/*.rpm

release-snapshot:
	@make create-assets
	@goreleaser --rm-dist --skip-publish --snapshot
	@rpm --addsign dist/*.rpm

release-public:
	@make create-assets
	@goreleaser --rm-dist --skip-publish

protobuffer:
	@make pb.api
	@make pb.viewconfig
	@make pb.ead
	@make pb.webresource

pb.webresource:
	@protoc --go_out=. hub3/mediamanager/webresource.proto

pb.api:
	@protoc --go_out=. hub3/fragments/api.proto

pb.viewconfig:
	@protoc --go_out=. hub3/fragments/viewconfig.proto

pb.ead:
	@protoc --go_out=. hub3/experimental/ead/ead.proto

cqlsh:
	@docker exec -it cassandra0 cqlsh

pprof-dev:
	@pprof --http localhost:6060 -seconds 30 http://localhost:3000/debug/pprof/profile
