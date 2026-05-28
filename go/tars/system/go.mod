module github.com/jimiechen/mineplanet/go/tars/system

go 1.25.5

require (
	github.com/jimiechen/mineplanet/go/common-lib v0.0.0
	github.com/jimiechen/mineplanet/go/modules/health v0.0.0
	github.com/jimiechen/mineplanet/go/modules/hello v0.0.0
	github.com/jimiechen/mineplanet/protocols/generated/go/base v0.0.0
	google.golang.org/protobuf v1.36.11
)

replace github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/base

replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib

replace github.com/jimiechen/mineplanet/go/modules/hello => ../../modules/hello

replace github.com/jimiechen/mineplanet/go/modules/health => ../../modules/health
