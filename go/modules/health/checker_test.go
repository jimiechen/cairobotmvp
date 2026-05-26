package health

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib/health"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/configsdk"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/i18nsdk"
)

func TestConfigChecker_Name(t *testing.T) {
	checker := NewConfigChecker(configsdk.NewFakeClient())
	if checker.Name() != "config" {
		t.Fatalf("期望 name='config'，实际 '%s'", checker.Name())
	}
}

func TestConfigChecker_Check_Success(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	checker := NewConfigChecker(cfg)

	status := checker.Check(context.Background())

	if !status.Healthy {
		t.Fatalf("期望健康，实际不健康: %s", status.Error)
	}
	if status.Name != "config" {
		t.Fatalf("期望 name='config'，实际 '%s'", status.Name)
	}
	if status.LatencyMs < 0 {
		t.Fatal("期望 latencyMs >= 0")
	}
}

func TestI18nChecker_Name(t *testing.T) {
	checker := NewI18nChecker(i18nsdk.NewFakeClient())
	if checker.Name() != "i18n" {
		t.Fatalf("期望 name='i18n'，实际 '%s'", checker.Name())
	}
}

func TestI18nChecker_Check_WithClient(t *testing.T) {
	i18n := i18nsdk.NewFakeClient()
	checker := NewI18nChecker(i18n)

	status := checker.Check(context.Background())

	if !status.Healthy {
		t.Fatalf("期望健康，实际不健康: %s", status.Error)
	}
}

func TestI18nChecker_Check_NilClient(t *testing.T) {
	checker := NewI18nChecker(nil)

	status := checker.Check(context.Background())

	if !status.Healthy {
		t.Fatalf("期望 nil client 时仍健康，实际不健康: %s", status.Error)
	}
}

func TestMySQLChecker_Name(t *testing.T) {
	checker := NewMySQLChecker(nil)
	if checker.Name() != "mysql" {
		t.Fatalf("期望 name='mysql'，实际 '%s'", checker.Name())
	}
}

func TestMySQLChecker_Check_NilDB(t *testing.T) {
	checker := NewMySQLChecker(nil)

	status := checker.Check(context.Background())

	if status.Healthy {
		t.Fatal("期望 nil db 时不健康")
	}
	if status.Error != ErrDBNil.Error() {
		t.Fatalf("期望错误 '%s'，实际 '%s'", ErrDBNil.Error(), status.Error)
	}
}

func TestRedisChecker_Name(t *testing.T) {
	checker := NewRedisChecker(nil)
	if checker.Name() != "redis" {
		t.Fatalf("期望 name='redis'，实际 '%s'", checker.Name())
	}
}

func TestRedisChecker_Check_NilClient(t *testing.T) {
	checker := NewRedisChecker(nil)

	status := checker.Check(context.Background())

	if status.Healthy {
		t.Fatal("期望 nil client 时不健康")
	}
	if status.Error != ErrClientNil.Error() {
		t.Fatalf("期望错误 '%s'，实际 '%s'", ErrClientNil.Error(), status.Error)
	}
}

func TestNewComponentStatusHelpers(t *testing.T) {
	t.Run("健康状态", func(t *testing.T) {
		status := health.NewComponentStatus("test", 10)
		if !status.Healthy {
			t.Fatal("期望 Healthy=true")
		}
		if status.LatencyMs != 10 {
			t.Fatalf("期望 LatencyMs=10，实际 %d", status.LatencyMs)
		}
		if status.Error != "" {
			t.Fatalf("期望 Error=''，实际 '%s'", status.Error)
		}
	})

	t.Run("不健康状态", func(t *testing.T) {
		err := errors.New("connection refused")
		status := health.NewUnhealthyComponentStatus("test", 20, err)
		if status.Healthy {
			t.Fatal("期望 Healthy=false")
		}
		if status.LatencyMs != 20 {
			t.Fatalf("期望 LatencyMs=20，实际 %d", status.LatencyMs)
		}
		if status.Error == "" {
			t.Fatal("期望 Error 非空")
		}
	})
}

func TestCheckerTimeout_Constant(t *testing.T) {
	if checkerTimeout != 1*time.Second {
		t.Fatalf("期望 checkerTimeout=1s，实际 %v", checkerTimeout)
	}
}

var _ = fmt.Sprintf
