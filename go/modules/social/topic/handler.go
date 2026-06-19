package topic

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// Handler 协议分发器，按 minType 路由到对应的 svc
// switch case 以外禁止业务逻辑
type Handler struct {
	publisher          event.Publisher
	createTopicSvc    *SvcCreateTopic
	listTopicSvc      *SvcListTopic
	deleteTopicSvc     *SvcDeleteTopic
	replyTopicSvc      *SvcReplyTopic
	likeTopicSvc       *SvcLikeTopic
	favoriteTopicSvc   *SvcFavoriteTopic
	getTopicDetailSvc  *SvcGetTopicDetail
	updateTopicSvc     *SvcUpdateTopic
	createReportSvc    *SvcCreateReport
	readTopicSvc       *SvcReadTopic
	getReplyListSvc    *SvcGetReplyList
}

// NewHandler 创建 Handler 实例，注入所有 svc 依赖
func NewHandler(repo Repository, publisher event.Publisher) *Handler {
	return &Handler{
		publisher:        publisher,
		createTopicSvc:   NewSvcCreateTopic(repo, publisher),
		listTopicSvc:     NewSvcListTopic(repo),
		deleteTopicSvc:   NewSvcDeleteTopic(repo),
		replyTopicSvc:    NewSvcReplyTopic(repo, publisher),
		likeTopicSvc:     NewSvcLikeTopic(repo, publisher),
		favoriteTopicSvc: NewSvcFavoriteTopic(repo, publisher),
		getTopicDetailSvc: NewSvcGetTopicDetail(repo),
		updateTopicSvc:   NewSvcUpdateTopic(repo),
		createReportSvc:  NewSvcCreateReport(repo),
		readTopicSvc:     NewSvcReadTopic(repo),
		getReplyListSvc:  NewSvcGetReplyList(repo),
	}
}

// Dispatch 根据 minType 分发到对应的 svc 处理
// 每个 case 统一：Unmarshal → svc.Handle → Marshal
func (h *Handler) Dispatch(ctx context.Context, minType string, reqBytes []byte) ([]byte, error) {
	switch minType {
	// 创建主题（minType=3001）
	case "3001":
		var req pb.CreateTopicRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal CreateTopicRequest failed: %w", err)
		}
		rsp, err := h.createTopicSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 主题列表（minType=3005）
	case "3005":
		var req pb.GetTopicListRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal GetTopicListRequest failed: %w", err)
		}
		rsp, err := h.listTopicSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 删除主题（minType=3009）
	case "3009":
		var req pb.DeleteTopicRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal DeleteTopicRequest failed: %w", err)
		}
		rsp, err := h.deleteTopicSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 回复主题（minType=3043）
	case "3043":
		var req pb.AddTopicReplyRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal AddTopicReplyRequest failed: %w", err)
		}
		rsp, err := h.replyTopicSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 点赞/取消点赞主题（minType=3061, 通过 is_like 区分）
	case "3061":
		var req pb.LikeTopicRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal LikeTopicRequest failed: %w", err)
		}
		rsp, err := h.likeTopicSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 收藏/取消收藏主题（minType=3063, 通过 is_favorite 区分）
	case "3063":
		var req pb.FavoriteTopicRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal FavoriteTopicRequest failed: %w", err)
		}
		rsp, err := h.favoriteTopicSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 主题详情/批量查询（minType=3057）
	case "3057":
		var req pb.BatchGetTopicInfoRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal BatchGetTopicInfoRequest(GetDetail) failed: %w", err)
		}
		rsp, err := h.getTopicDetailSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 提交举报（minType=3095）
	case "3095":
		var req pb.CreateReportRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal CreateReportRequest failed: %w", err)
		}
		rsp, err := h.createReportSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 已读主题（minType=3099）
	case "3099":
		var req pb.CheckTopicActionsRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal CheckTopicActionsRequest(Read) failed: %w", err)
		}
		rsp, err := h.readTopicSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 获取评论列表（minType=3065）
	case "3065":
		var req pb.GetReplyListRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal GetReplyListRequest failed: %w", err)
		}
		rsp, err := h.getReplyListSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	default:
		return nil, fmt.Errorf("unsupported minType: %s", minType)
	}
}
