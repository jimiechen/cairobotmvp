module github.com/jimiechen/mineplanet/go/modules/hello

go 1.23

require google.golang.org/protobuf v1.36.11

replace github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/base

replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib
