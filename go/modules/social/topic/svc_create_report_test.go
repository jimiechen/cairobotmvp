package topic

import (
	"context"
	"strings"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcCreateReport_完整参数_应返回成功且report_id非空
func TestSvcCreateReport_完整参数_成功(t *testing.T) {
	svc := NewSvcCreateReport(newMockRepository())
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.CreateReportRequest{
		TargetType: "topic",
		TargetId:   "topic-001",
		GroupId:    "grp-001",
		ReportType: "spam",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.ReportId == "" {
		t.Error("期望 ReportId 非空")
	}
	if !strings.HasPrefix(resp.ReportId, "rpt-") {
		t.Errorf("期望 ReportId 以 rpt- 开头，实际得到: %s", resp.ReportId)
	}
	if resp.Status != "pending" {
		t.Errorf("期望 Status = pending，实际得到: %s", resp.Status)
	}
}

// TestSvcCreateReport_带补充说明和截图_应正常处理
func TestSvcCreateReport_带可选字段_成功(t *testing.T) {
	svc := NewSvcCreateReport(newMockRepository())
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.CreateReportRequest{
		TargetType:    "reply",
		TargetId:      "reply-001",
		GroupId:       "grp-001",
		ReportType:    "inappropriate",
		ReasonText:    "该回复包含不当内容",
		ScreenshotUrls: []string{"https://example.com/img1.png"},
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.ReportId == "" {
		t.Error("期望 ReportId 非空")
	}
}

// TestSvcCreateReport_target_type为空_应返回参数校验错误
func TestSvcCreateReport_缺少target_type(t *testing.T) {
	svc := NewSvcCreateReport(newMockRepository())

	req := &pb.CreateReportRequest{
		TargetId:   "topic-001",
		GroupId:    "grp-001",
		ReportType: "spam",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败（target_type 为空），实际返回成功")
	}
}

// TestSvcCreateReport_target_id为空_应返回参数校验错误
func TestSvcCreateReport_缺少target_id(t *testing.T) {
	svc := NewSvcCreateReport(newMockRepository())

	req := &pb.CreateReportRequest{
		TargetType: "topic",
		GroupId:    "grp-001",
		ReportType: "spam",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败（target_id 为空），实际返回成功")
	}
}

// TestSvcCreateReport_group_id为空_应返回参数校验错误
func TestSvcCreateReport_缺少group_id(t *testing.T) {
	svc := NewSvcCreateReport(newMockRepository())

	req := &pb.CreateReportRequest{
		TargetType: "topic",
		TargetId:   "topic-001",
		ReportType: "spam",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败（group_id 为空），实际返回成功")
	}
}

// TestSvcCreateReport_report_type为空_应返回参数校验错误
func TestSvcCreateReport_缺少report_type(t *testing.T) {
	svc := NewSvcCreateReport(newMockRepository())

	req := &pb.CreateReportRequest{
		TargetType: "topic",
		TargetId:   "topic-001",
		GroupId:    "grp-001",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败（report_type 为空），实际返回成功")
	}
}

// TestSvcCreateReport_所有必填字段为空_应返回第一个校验错误
func TestSvcCreateReport_全部为空_校验失败(t *testing.T) {
	svc := NewSvcCreateReport(newMockRepository())

	req := &pb.CreateReportRequest{}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}
