package member

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcRegister 用户注册服务（minType=1021 UserRegister）
// 负责新用户注册流程：参数校验 → 唯一性检查 → 创建用户 → 返回令牌
// 不负责登录认证（由 SvcLogin 负责）
type SvcRegister struct {
	repo Repository
}

// NewSvcRegister 创建注册服务实例
func NewSvcRegister(repo Repository) *SvcRegister {
	return &SvcRegister{repo: repo}
}

// Handle 处理用户注册请求，遵循 DevGuide §7 五步模式
func (s *SvcRegister) Handle(ctx context.Context, req *pb.UserRegisterRequest) (*pb.UserRegisterResponse, error) {
	// Step 1: 参数校验 — 必填字段非空、用户名长度合法
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr // 系统级错误（不应发生）
	} else if errResp != nil {
		return errResp, nil // 业务校验错误，通过 Result 表达
	}

	// Step 2: 权限校验 — 注册为公开操作，无需权限检查

	// Step 3: 1级数据读写 — 检查用户名唯一性 + 创建用户记录
	existing, _ := s.repo.GetUserByUsername(ctx, req.Username)
	if existing != nil {
		return &pb.UserRegisterResponse{
			Result: &base.Result{
				Code:    int32(base.UserErrorCode_USER_ERROR_NAME_ALREADY_TAKEN),
				Message: "用户名已被占用",
			},
		}, nil
	}

	now := time.Now().UnixMilli()
	hashedPwd, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希处理失败: %w", err)
	}
	user := &User{
		ID:              generateUserID(),
		Username:        req.Username,
		Password:        hashedPwd,
		Email:           req.Email,
		Nickname:        req.Nickname,
		Status:          UserStatusActive, // active
		MembershipLevel: "normal",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// Step 4: 领域事件 — 注册成功后初始化统计记录（MVP-P0 可延迟到首次查询时懒初始化）

	// Step 5: 返回响应
	return &pb.UserRegisterResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "注册成功",
		},
		UserId: user.ID,
	}, nil
}

// validateRequest 校验注册请求必填字段
func (s *SvcRegister) validateRequest(req *pb.UserRegisterRequest) (*pb.UserRegisterResponse, error) {
	if req.Username == "" {
		return &pb.UserRegisterResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "用户名不能为空",
			},
		}, nil
	}
	if len(req.Username) < 4 || len(req.Username) > 50 {
		return &pb.UserRegisterResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "用户名长度必须在4-50个字符之间",
			},
		}, nil
	}
	if req.Password == "" {
		return &pb.UserRegisterResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "密码不能为空",
			},
		}, nil
	}
	return nil, nil
}

// generateUserID 生成用户内部 ID（MVP-P0 使用简单时间戳+随机数，生产环境用 ULID）
func generateUserID() string {
	return fmt.Sprintf("user-%d", time.Now().UnixNano())
}

// hashPassword 使用 bcrypt 对密码进行单向哈希（cost=10）
// bcrypt 内置 salt，无需额外 Salt 字段
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}
	return string(hash), nil
}
