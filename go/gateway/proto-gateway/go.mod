module github.com/jimiechen/mineplanet/go/gateway/proto-gateway

go 1.23

require (
	github.com/google/uuid v1.6.0
	google.golang.org/protobuf v1.36.11
)

require gopkg.in/yaml.v3 v3.0.1

replace github.com/jimiechen/mineplanet/protocols/generated/go/proto/base => ../../../proto/generated/go/proto/base
