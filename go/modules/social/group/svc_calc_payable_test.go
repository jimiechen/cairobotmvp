package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcCalcPayableAmount_正常计算 当参数合法时_应返回计算结果
func TestSvcCalcPayableAmount_正常计算(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcCalcPayableAmount(mockRepo)

	group := &Group{ID: "group-001", Slug: "test-pay"}
	mockRepo.groups[group.ID] = group
	mockRepo.groups[group.Slug] = group

	req := &pb.CalcPayableAmountRequest{
		GroupId:        "group-001",
		DiscountType:   pb.DiscountType_DISCOUNT_TYPE_JOIN,
		OriginalAmount: 10000, // 100元
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.FinalAmount != 10000 {
		t.Errorf("期望最终金额为原始金额 10000（无折扣），实际得到 %d", resp.FinalAmount)
	}
}

// TestSvcCalcPayableAmount_圈子不存在 当groupId无效时_应返回圈子不存在错误
func TestSvcCalcPayableAmount_圈子不存在(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcCalcPayableAmount(mockRepo)

	req := &pb.CalcPayableAmountRequest{
		GroupId:        "nonexistent",
		OriginalAmount: 10000,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NOT_FOUND) {
		t.Errorf("期望圈子不存在错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NOT_FOUND, resp.Result.Code)
	}
}

// TestSvcCalcPayableAmount_金额非正 当原始金额<=0时_应返回参数校验错误
func TestSvcCalcPayableAmount_金额非正(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcCalcPayableAmount(mockRepo)

	req := &pb.CalcPayableAmountRequest{
		GroupId:        "group-001",
		OriginalAmount: 0, // 非法金额
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST) {
		t.Errorf("期望参数校验错误码 %d，实际得到 %d", base.ErrorCode_ERROR_CODE_INVALID_REQUEST, resp.Result.Code)
	}
}
