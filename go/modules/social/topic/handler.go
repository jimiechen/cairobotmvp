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
	likeTopicSvc     *SvcLikeTopic
	unlikeTopicSvc   *SvcUnlikeTopic
	favoriteTopicSvc *SvcFavoriteTopic
	getTopicDetailSvc *SvcGetTopicDetail
	updateTopicSvc     *SvcUpdateTopic
	readTopicSvc       *SvcReadTopic
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
		unlikeTopicSvc:   NewSvcUnlikeTopic(repo),
		favoriteTopicSvc: NewSvcFavoriteTopic(repo, publisher),
		getTopicDetailSvc: NewSvcGetTopicDetail(repo),
		updateTopicSvc:   NewSvcUpdateTopic(repo),
		readTopicSvc:     NewSvcReadTopic(repo),
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

	// 点赞主题（minType=3061, is_like=true）
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

	// 取消点赞主题（minType=3062, 复用 LikeTopicRequest.is_like=false）
	case "3062":
		var req pb.LikeTopicRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal LikeTopicRequest(Unlike) failed: %w", err)
		}
		rsp, err := h.unlikeTopicSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 收藏主题（minType=3063, is_favorite=true）
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

	// 取消收藏主题（minType=3064, 复用 FavoriteTopicRequest.is_favorite=false）
	// TODO: Phase 2 需要独立的 SvcUnfavorite，当前使用 favorite svc 的反向逻辑
	case "3064":
		var req pb.FavoriteTopicRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal FavoriteTopicRequest(Unfavorite) failed: %w", err)
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

	// 提交举报（minType=3095，对齐 topic_base.proto CreateReport 定义）
	// TODO(impl): 当前复用 updateTopicSvc 作为 stub，Phase 2 需实现独立的 CreateReport 逻辑
	case "3095":
		var req pb.CreateReportRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal CreateReportRequest failed: %w", err)
		}
		// stub: 暂用 updateTopicSvc 处理，返回基础响应
		rsp, err := h.updateTopicSvc.Handle(ctx, &pb.CreateTopicRequest{TopicId: req.TargetId})
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

	default:
		return nil, fmt.Errorf("unsupported minType: %s", minType)
	}
}
