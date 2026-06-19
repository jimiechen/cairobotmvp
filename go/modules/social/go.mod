module github.com/jimiechen/mineplanet/go/modules/social

go 1.23

require (
	github.com/jimiechen/mineplanet/protocols/generated/go/base v0.0.0
	github.com/jimiechen/mineplanet/protocols/generated/go/social v0.0.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/crypto v0.33.0
	google.golang.org/protobuf v1.36.11
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/base

replace github.com/jimiechen/mineplanet/protocols/generated/go/social => ../../../proto/generated/go/social

replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib
