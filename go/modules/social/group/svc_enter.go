package group

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcEnter 用户进入圈子服务（minType=2087 GroupUserEnter）
// 负责返回圈子详情和当前用户的成员信息
// 不负责权限判断（由客户端控制展示内容）
type SvcEnter struct {
	repo Repository
}

// NewSvcEnter 创建服务实例
func NewSvcEnter(repo Repository) *SvcEnter {
	return &SvcEnter{repo: repo}
}

// Handle 处理用户进入圈子请求，遵循 DevGuide §7 五步模式
func (s *SvcEnter) Handle(ctx context.Context, req *pb.GroupUserEnterRequest) (*pb.GroupUserEnterResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 进入圈子为公开操作（私密圈子在客户端层拦截）

	// Step 3: 1级数据读写 — 查询圈子详情 + 查询成员信息
	group, err := s.repo.GetGroupByID(ctx, req.GroupId)
	if err != nil {
		return nil, fmt.Errorf("查询圈子失败: %w", err)
	}
	if group == nil {
		return &pb.GroupUserEnterResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_FOUND),
				Message: "圈子不存在",
			},
			GroupId: req.GroupId,
		}, nil
	}

	// 查询当前用户成员信息（可能为 nil，表示未加入）
	var userMemberInfo *pb.UserMemberInfo
	if req.UserId != "" {
		member, err := s.repo.GetMember(ctx, req.GroupId, req.UserId)
		if err != nil {
			return nil, fmt.Errorf("查询成员信息失败: %w", err)
		}
		if member != nil {
			userMemberInfo = convertToProtoUserMemberInfo(member)
		}
	}

	// Step 4: 领域事件 — 更新最后活跃时间（MVP-P0 可延迟）

	// Step 5: 返回响应
	return &pb.GroupUserEnterResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "成功",
		},
		GroupInfo:       convertToProtoGroupInfo(group),
		UserMemberInfo: userMemberInfo,
		GroupId:         req.GroupId,
	}, nil
}

// validateRequest 校验进入圈子请求必填字段
func (s *SvcEnter) validateRequest(req *pb.GroupUserEnterRequest) (*pb.GroupUserEnterResponse, error) {
	if req.GroupId == "" {
		return &pb.GroupUserEnterResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
