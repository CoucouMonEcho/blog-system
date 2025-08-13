module blog-system/services/stat

go 1.24.2

require (
	blog-system/common v0.0.0
	github.com/CoucouMonEcho/go-framework v0.1.4
	github.com/go-sql-driver/mysql v1.9.3
	gopkg.in/yaml.v2 v2.4.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace blog-system/common => ../../common
