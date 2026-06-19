package member

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// ========== SvcRefresh 测试 ==========

func newTestRefreshDependencies(t testing.TB) (*SvcRefresh, *mockRepository) {
	t.Helper()
	repo := newMockRepository()
	jwtMgr, _ := NewJWTManager(DefaultJWTConfig().SetSecretKey("test-secret-key-for-jwt-refresh-tests-32!"))
	ts := NewMemoryTokenStore()
	svc := NewSvcRefresh(ts, jwtMgr, repo)
	return svc, repo
}

func TestSvcRefresh_刷新成功(t *testing.T) {
	svc, repo := newTestRefreshDependencies(t)
	repo.users["u1"] = &User{ID: "u1", Status: UserStatusActive}

	// 先获取 refresh_token
	jwtMgr, _ := NewJWTManager(DefaultJWTConfig().SetSecretKey("test-secret-key-for-jwt-refresh-tests-32!"))
	refreshToken, _, _ := jwtMgr.GenerateRefreshToken("u1")

	req := &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	rsp, err := svc.Handle(nil, req)

	require.NoError(t, err)
	assert.Equal(t, int32(base.ErrorCode_ERROR_CODE_SUCCESS), rsp.Result.Code)
	assert.NotEmpty(t, rsp.AccessToken)
	assert.NotEmpty(t, rsp.RefreshToken)
	assert.Equal(t, "u1", rsp.UserId)
	assert.Greater(t, rsp.ExpiresAt, int64(0))
}

func TestSvcRefresh_空的refreshToken(t *testing.T) {
	svc, _ := newTestRefreshDependencies(t)

	rsp, err := svc.Handle(nil, &pb.RefreshTokenRequest{RefreshToken: ""})

	require.NoError(t, err)
	assert.Equal(t, int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), rsp.Result.Code)
	assert.Contains(t, rsp.Result.Message, "不能为空")
}

func TestSvcRefresh_无效的refreshToken格式(t *testing.T) {
	svc, _ := newTestRefreshDependencies(t)

	rsp, err := svc.Handle(nil, &pb.RefreshTokenRequest{RefreshToken: "garbage"})

	require.NoError(t, err)
	assert.Equal(t, int32(base.ErrorCode_ERROR_CODE_UNAUTHORIZED), rsp.Result.Code)
}

func TestSvcRefresh_access类型的token被拒绝(t *testing.T) {
	svc, _ := newTestRefreshDependencies(t)
	jwtMgr, _ := NewJWTManager(DefaultJWTConfig().SetSecretKey("test-secret-key-for-jwt-refresh-tests-32!"))
	accessToken, _, _ := jwtMgr.GenerateAccessToken("u1")

	rsp, err := svc.Handle(nil, &pb.RefreshTokenRequest{RefreshToken: accessToken})

	require.NoError(t, err)
	assert.Equal(t, int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), rsp.Result.Code)
	assert.Contains(t, rsp.Result.Message, "令牌类型错误")
}

func TestSvcRefresh_用户不存在(t *testing.T) {
	svc, _ := newTestRefreshDependencies(t)
	// 不插入用户到 repo
	jwtMgr, _ := NewJWTManager(DefaultJWTConfig().SetSecretKey("test-secret-key-for-jwt-refresh-tests-32!"))
	refreshToken, _, _ := jwtMgr.GenerateRefreshToken("ghost-user")

	rsp, err := svc.Handle(nil, &pb.RefreshTokenRequest{RefreshToken: refreshToken})

	require.NoError(t, err)
	assert.Equal(t, int32(base.ErrorCode_ERROR_CODE_NOT_FOUND), rsp.Result.Code)
}

func TestSvcRefresh_用户被封禁(t *testing.T) {
	svc, repo := newTestRefreshDependencies(t)
	repo.users["banned"] = &User{ID: "banned", Status: UserStatusSuspended}
	jwtMgr, _ := NewJWTManager(DefaultJWTConfig().SetSecretKey("test-secret-key-for-jwt-refresh-tests-32!"))
	refreshToken, _, _ := jwtMgr.GenerateRefreshToken("banned")

	rsp, err := svc.Handle(nil, &pb.RefreshTokenRequest{RefreshToken: refreshToken})

	require.NoError(t, err)
	assert.Equal(t, int32(base.ErrorCode_ERROR_CODE_FORBIDDEN), rsp.Result.Code)
}
