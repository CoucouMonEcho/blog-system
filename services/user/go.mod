module blog-system/services/user

go 1.24.2

require (
	github.com/CoucouMonEcho/go-framework v0.1.4
	golang.org/x/crypto v0.36.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	golang.org/x/sync v0.16.0 // indirect
)

require (
	blog-system/common v0.0.0
	github.com/redis/go-redis/v9 v9.11.0
	gopkg.in/yaml.v2 v2.4.0
)

replace blog-system/common => ../../common
