module github.com/jimiechen/mineplanet/go/services/hello

go 1.23

require google.golang.org/protobuf v1.36.11

replace github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/proto/base

replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib
