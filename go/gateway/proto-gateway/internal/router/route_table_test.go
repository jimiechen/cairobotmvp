package router

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/config"
)

func TestFindRoute(t *testing.T) {
	cfg := &config.RoutesConfig{
		Routes: []config.Route{
			{
				RequestMax:  2100,
				RequestMin:  2097,
				RouteKey:    "2100:2097",
				CommandName: "ServiceHealthCheck",
				TarsMethod:  "HealthCheck",
			},
			{
				RequestMax:  2100,
				RequestMin:  2101,
				RouteKey:    "2100:2101",
				CommandName: "HelloWorld",
				TarsMethod:  "HelloWorld",
			},
		},
	}

	r := NewRouteTable(cfg)

	tests := []struct {
		name        string
		maxType     int32
		minType     int32
		wantFound   bool
		wantCommand string
	}{
		{
			name:        "HealthCheck route",
			maxType:     2100,
			minType:     2097,
			wantFound:   true,
			wantCommand: "ServiceHealthCheck",
		},
		{
			name:        "HelloWorld route",
			maxType:     2100,
			minType:     2101,
			wantFound:   true,
			wantCommand: "HelloWorld",
		},
		{
			name:      "unknown route",
			maxType:   9999,
			minType:   9999,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, found := r.FindRoute(tt.maxType, tt.minType)
			if found != tt.wantFound {
				t.Fatalf("FindRoute(%d, %d) found = %v, want %v", tt.maxType, tt.minType, found, tt.wantFound)
			}
			if tt.wantFound && route.CommandName != tt.wantCommand {
				t.Fatalf("FindRoute(%d, %d) command = %q, want %q", tt.maxType, tt.minType, route.CommandName, tt.wantCommand)
			}
		})
	}
}
