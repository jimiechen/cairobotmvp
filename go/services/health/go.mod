module github.com/jimiechen/mineplanet/go/services/health

go 1.23

require google.golang.org/protobuf v1.36.11

replace github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/proto/base

replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib
