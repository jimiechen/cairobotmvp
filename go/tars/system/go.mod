module github.com/jimiechen/mineplanet/go/tars/system

go 1.23

require (
	github.com/jimiechen/mineplanet/go/common-lib v0.0.0
	github.com/jimiechen/mineplanet/go/services/hello v0.0.0
	github.com/jimiechen/mineplanet/go/services/health v0.0.0
	google.golang.org/protobuf v1.36.11
)

replace github.com/jimiechen/mineplanet/protocols/generated/go/proto/base => ../../../proto/generated/go/proto/base
replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib
replace github.com/jimiechen/mineplanet/go/services/hello => ../../services/hello
replace github.com/jimiechen/mineplanet/go/services/health => ../../services/health
