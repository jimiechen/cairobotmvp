package member

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcBlock 拉黑用户服务（minType=1039 BlockUser）
// 负责创建用户之间的拉黑关系（单向）
// 不负责权限校验（由上层确保 blocked_by 为当前登录用户），不负责通知被拉黑用户
type SvcBlock struct {
	repo Repository
}

// NewSvcBlock 创建拉黑服务实例
func NewSvcBlock(repo Repository) *SvcBlock {
	return &SvcBlock{repo: repo}
}

// Handle 处理拉黑用户请求，遵循 DevGuide §7 五步模式
func (s *SvcBlock) Handle(ctx context.Context, req *pb.BlockUserRequest) (*pb.BlockUserResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 由上层保证 blocked_by 为当前登录用户

	// Step 3: 1级数据读写 — 检查是否已拉黑 → 创建拉黑记录
	now := time.Now().UnixMilli()

	// 幂等检查：如果已存在拉黑关系，直接返回已有记录
	isBlocked, err := s.repo.IsBlocked(ctx, req.BlockedBy, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("检查拉黑状态失败: %w", err)
	}
	if isBlocked {
		// 已拉黑：查询已有记录并返回（幂等语义）
		existingBlocks, _, err := s.repo.ListBlocks(ctx, req.BlockedBy, 1, 1000)
		if err != nil {
			return nil, fmt.Errorf("查询拉黑记录失败: %w", err)
		}
		for _, b := range existingBlocks {
			if b.BlockedID == req.UserId {
				return &pb.BlockUserResponse{
					Result: &base.Result{
						Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
						Message: "已拉黑该用户",
					},
					BlockInfo: &pb.UserBlock{
						UserId:    b.BlockedID,
						BlockedBy: b.BlockerID,
						CreatedAt: b.CreatedAt,
						UpdatedAt: now,
						GroupId:   req.GroupId,
					},
				}, nil
			}
		}
	}

	// 创建新的拉黑记录
	block := &MemberBlock{
		ID:        generateUserID(),
		BlockerID: req.BlockedBy,
		BlockedID: req.UserId,
		CreatedAt: now,
	}
	if err := s.repo.CreateBlock(ctx, block); err != nil {
		return nil, fmt.Errorf("创建拉黑记录失败: %w", err)
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.BlockUserResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "拉黑成功",
		},
		BlockInfo: &pb.UserBlock{
			UserId:    req.UserId,
			BlockedBy: req.BlockedBy,
			CreatedAt: now,
			UpdatedAt: now,
			GroupId:   req.GroupId,
		},
	}, nil
}

// validateRequest 校验拉黑请求必填字段
func (s *SvcBlock) validateRequest(req *pb.BlockUserRequest) (*pb.BlockUserResponse, error) {
	if req.UserId == "" {
		return &pb.BlockUserResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "被拉黑用户ID不能为空",
			},
		}, nil
	}
	if req.GroupId == "" {
		return &pb.BlockUserResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
