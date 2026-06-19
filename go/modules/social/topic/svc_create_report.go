package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcCreateReport 提交举报服务（minType=3095 CreateReport）
//
// P1 范围：校验 + 生成 report_id + 返回 pending 状态
// 不包含：审核流程、通知、风控（Phase 2 补充）
type SvcCreateReport struct {
	repo Repository
}

// NewSvcCreateReport 创建服务实例
func NewSvcCreateReport(repo Repository) *SvcCreateReport {
	return &SvcCreateReport{repo: repo}
}

// Handle 处理举报提交请求
func (s *SvcCreateReport) Handle(ctx context.Context, req *pb.CreateReportRequest) (*pb.CreateReportResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// 生成举报记录 ID（P1: 仅生成 ID 返回，实际落库待 Phase 2 接入 topic_reports 表）
	reportID := fmt.Sprintf("rpt-%d", time.Now().UnixNano())

	return &pb.CreateReportResponse{
		Result:   &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "举报提交成功"},
		ReportId: reportID,
		Status:   "pending",
	}, nil
}

// validateRequest 校验举报请求必填字段
func (s *SvcCreateReport) validateRequest(req *pb.CreateReportRequest) (*pb.CreateReportResponse, error) {
	if req.TargetType == "" {
		return &pb.CreateReportResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "target_type 不能为空"},
		}, nil
	}
	if req.TargetId == "" {
		return &pb.CreateReportResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "target_id 不能为空"},
		}, nil
	}
	if req.GroupId == "" {
		return &pb.CreateReportResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "group_id 不能为空"},
		}, nil
	}
	if req.ReportType == "" {
		return &pb.CreateReportResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "report_type 不能为空"},
		}, nil
	}
	return nil, nil
}
