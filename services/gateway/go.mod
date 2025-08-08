module blog-system/services/gateway

go 1.24.2

require (
	blog-system/common v0.0.0
	github.com/CoucouMonEcho/go-framework v0.1.4
	github.com/redis/go-redis/v9 v9.11.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace blog-system/common => ../../common
