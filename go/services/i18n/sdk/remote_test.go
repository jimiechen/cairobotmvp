package sdk

import (
	"context"
	"errors"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

// mockRemoteClient 用于测试的模拟远程客户端
type mockRemoteClient struct {
	response []byte
	err      error
}

func (m *mockRemoteClient) Invoke(ctx context.Context, method string, request []byte) ([]byte, error) {
	return m.response, m.err
}

func TestRemoteClient_GetLangPack_NoClient(t *testing.T) {
	rc := newRemoteClient(&Options{Mode: ModeRemote})

	_, err := rc.getLangPack(context.Background(), "zh-CN")
	if err == nil {
		t.Error("expected error when remote client is nil")
	}
}

func TestRemoteClient_GetLangDifference_NoClient(t *testing.T) {
	rc := newRemoteClient(&Options{Mode: ModeRemote})

	_, err := rc.getLangDifference(context.Background(), "zh-CN", 1)
	if err == nil {
		t.Error("expected error when remote client is nil")
	}
}

func TestRemoteClient_Ping_NoClient(t *testing.T) {
	rc := newRemoteClient(&Options{Mode: ModeRemote})

	err := rc.ping(context.Background())
	if err == nil {
		t.Error("expected error when remote client is nil")
	}
}

func TestNewRemoteClient(t *testing.T) {
	opts := &Options{
		Mode:        ModeRemote,
		TarsServant: "TestObj",
		Env:         "test",
	}
	rc := newRemoteClient(opts)

	if rc.options.Mode != ModeRemote {
		t.Errorf("expected mode %s, got %s", ModeRemote, rc.options.Mode)
	}
	if rc.options.TarsServant != "TestObj" {
		t.Errorf("expected TarsServant 'TestObj', got '%s'", rc.options.TarsServant)
	}
}

func TestErrRemoteClientRequired_Message(t *testing.T) {
	expectedMsg := "remote client required for remote mode"
	if ErrRemoteClientRequired.Error() != expectedMsg {
		t.Errorf("expected message '%s', got '%s'", expectedMsg, ErrRemoteClientRequired.Error())
	}
}

func TestRemoteClient_ReturnTypes(t *testing.T) {
	rc := newRemoteClient(&Options{})

	packResp, packErr := rc.getLangPack(context.Background(), "test")
	if packResp != nil {
		t.Error("expected nil response on error")
	}
	if packErr == nil {
		t.Error("expected non-nil error")
	}

	diffResp, diffErr := rc.getLangDifference(context.Background(), "test", 0)
	if diffResp != nil {
		t.Error("expected nil diff response on error")
	}
	if diffErr == nil {
		t.Error("expected non-nil diff error")
	}

	var zeroDiff service.LangDiffResponse
	if diffResp != &zeroDiff && diffResp != nil {
		t.Error("diff response should be nil or empty")
	}
}

func TestRemoteClient_GetLangPack_WithMock(t *testing.T) {
	mockClient := &mockRemoteClient{err: errors.New("network error")}
	rc := newRemoteClient(&Options{Mode: ModeRemote, RemoteClient: mockClient})

	_, err := rc.getLangPack(context.Background(), "zh-CN")
	if err == nil {
		t.Error("expected error from mock client")
	}
}

func TestRemoteClient_GetLangDifference_WithMock(t *testing.T) {
	mockClient := &mockRemoteClient{err: errors.New("network error")}
	rc := newRemoteClient(&Options{Mode: ModeRemote, RemoteClient: mockClient})

	_, err := rc.getLangDifference(context.Background(), "zh-CN", 1)
	if err == nil {
		t.Error("expected error from mock client")
	}
}

func TestRemoteClient_Ping_WithMock(t *testing.T) {
	mockClient := &mockRemoteClient{err: errors.New("network error")}
	rc := newRemoteClient(&Options{Mode: ModeRemote, RemoteClient: mockClient})

	err := rc.ping(context.Background())
	if err == nil {
		t.Error("expected error from mock client")
	}
}
