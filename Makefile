grpc:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.15.2
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.15.2
	go install github.com/envoyproxy/protoc-gen-validate@v0.10.1
	go install github.com/zitadel/zitadel/internal/protoc/protoc-gen-authoption
	# This foreach is a workaround from a limitation of the authoption generator and only affects zitadel.
	# The authoption generator cannot work when passed *.proto but instead needs to have each file passed as {name}.proto
	for i in $$(find proto/zitadel -iname *.proto); do buf generate $${i}; done
	mv .artifacts/grpc/zitadel/auth.pb.authoptions.go .artifacts/grpc/github.com/zitadel/zitadel/pkg/grpc/auth
	mv .artifacts/grpc/zitadel/admin.pb.authoptions.go .artifacts/grpc/github.com/zitadel/zitadel/pkg/grpc/admin
	mv .artifacts/grpc/zitadel/management.pb.authoptions.go .artifacts/grpc/github.com/zitadel/zitadel/pkg/grpc/management
	mv .artifacts/grpc/zitadel/system.pb.authoptions.go .artifacts/grpc/github.com/zitadel/zitadel/pkg/grpc/system
	mv .artifacts/grpc/zitadel/session/v2alpha/*.pb.authoptions.go .artifacts/grpc/github.com/zitadel/zitadel/pkg/grpc/session/v2alpha
	mv .artifacts/grpc/zitadel/user/v2alpha/*.pb.authoptions.go .artifacts/grpc/github.com/zitadel/zitadel/pkg/grpc/user/v2alpha
	cp -rT .artifacts/grpc/github.com/zitadel/zitadel/pkg/grpc/ pkg/grpc/
	mkdir -p openapi/v2/zitadel
	cp .artifacts/grpc/zitadel/*.swagger.json openapi/v2/zitadel

static:
	go install github.com/rakyll/statik@v0.1.7
	go generate internal/api/ui/login/statik/generate.go
	go generate internal/api/ui/login/static/generate.go
	go generate internal/notification/statik/generate.go
	go generate internal/statik/generate.go

assets:
	go generate openapi/statik/generate.go && \
    mkdir -p docs/apis/assets/ && \
    go run internal/api/assets/generator/asset_generator.go -directory=internal/api/assets/generator/ -assets=docs/apis/assets/assets.md

generate: grpc static assets

clean:
	rm -rf .artifacts/grpc

test:
	go test -race -v -coverprofile=profile.cov ./...