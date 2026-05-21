module github.com/jimiechen/mineplanet/go/gateway/proto-gateway

go 1.23

require (
	github.com/TarsCloud/TarsGo v1.4.6
	github.com/google/uuid v1.6.0
	github.com/jimiechen/mineplanet/go/common-lib v0.0.0
	github.com/jimiechen/mineplanet/go/modules/hello v0.0.0
	github.com/jimiechen/mineplanet/go/modules/health v0.0.0
	google.golang.org/protobuf v1.36.11
)

require gopkg.in/yaml.v3 v3.0.1

replace github.com/jimiechen/mineplanet/protocols/generated/go/proto/base => ../../../proto/generated/go/proto/base
replace github.com/TarsCloud/TarsGo => ../../third_party/TarsGo/TarsGo-1.4.6
replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib
replace github.com/jimiechen/mineplanet/go/modules/hello => ../../modules/hello
replace github.com/jimiechen/mineplanet/go/modules/health => ../../modules/health
