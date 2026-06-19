package member

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcGetBlockList 查询拉黑列表服务（minType=1043 GetBlockList）
// 负责分页查询当前用户的拉黑用户列表
// 不负责权限校验（由认证中间件保证 userId 为当前登录用户）
type SvcGetBlockList struct {
	repo Repository
}

// NewSvcGetBlockList 创建查询拉黑列表服务实例
func NewSvcGetBlockList(repo Repository) *SvcGetBlockList {
	return &SvcGetBlockList{repo: repo}
}

// Handle 处理查询拉黑列表请求，遵循 DevGuide §7 五步模式
func (s *SvcGetBlockList) Handle(ctx context.Context, req *pb.GetBlockListRequest) (*pb.GetBlockListResponse, error) {
	// Step 1: 参数校验 — 从上下文提取 userId + 分页参数默认值
	userID := ctx.Value(CtxKeyUserID)
	if userID == nil || userID.(string) == "" {
		return &pb.GetBlockListResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "缺少用户身份信息",
			},
		}, nil
	}
	uidStr := userID.(string)

	// 分页参数默认值
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20 // 默认每页 20 条
	}
	if pageSize > 100 {
		pageSize = 100 // 上限 100 条
	}

	// Step 2: 权限校验 — 查询自身拉黑列表为已认证操作

	// Step 3: 1级数据读写 — 分页查询拉黑记录
	blocks, total, err := s.repo.ListBlocks(ctx, uidStr, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询拉黑列表失败: %w", err)
	}

	// 转换为 proto UserBlock 列表
	pbBlocks := make([]*pb.UserBlock, 0, len(blocks))
	for _, b := range blocks {
		pbBlocks = append(pbBlocks, &pb.UserBlock{
			UserId:    b.BlockedID,
			BlockedBy: b.BlockerID,
			CreatedAt: b.CreatedAt,
			GroupId:   req.GroupId,
		})
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.GetBlockListResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "查询成功",
		},
		Blocks:   pbBlocks,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
