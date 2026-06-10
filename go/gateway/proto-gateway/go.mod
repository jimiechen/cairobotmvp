module github.com/jimiechen/mineplanet/go/gateway/proto-gateway

go 1.25.5

require (
	github.com/TarsCloud/TarsGo v1.4.6
	github.com/google/uuid v1.6.0
	github.com/jimiechen/mineplanet/go/common-lib v0.0.0
	github.com/jimiechen/mineplanet/go/modules/health v0.0.0
	github.com/jimiechen/mineplanet/go/modules/hello v0.0.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/jimiechen/mineplanet/protocols/generated/go/base v0.0.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.50.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.uber.org/automaxprocs v1.5.2 // indirect
	golang.org/x/sys v0.42.0 // indirect
	modernc.org/libc v1.72.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

replace github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/base

replace github.com/TarsCloud/TarsGo => ../../third_party/TarsGo/TarsGo-1.4.6

replace github.com/jimiechen/mineplanet/go/common-lib => ../../common-lib

replace github.com/jimiechen/mineplanet/go/modules/hello => ../../modules/hello

replace github.com/jimiechen/mineplanet/go/modules/health => ../../modules/health
