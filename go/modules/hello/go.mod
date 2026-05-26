module github.com/jimiechen/mineplanet/go/modules/hello

go 1.23

require (
	github.com/TarsCloud/TarsGo/contrib/log v0.0.0
	google.golang.org/protobuf v1.36.11
)

replace github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/base

replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib

replace github.com/TarsCloud/TarsGo/contrib/log => ../../third_party/TarsGo/TarsGo-1.4.6/contrib/log
