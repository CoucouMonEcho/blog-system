module blog-system/services/admin

go 1.24.2

require (
	blog-system/common v0.0.0
	github.com/CoucouMonEcho/go-framework v0.1.4
)

require github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect

replace blog-system/common => ../../common
