// repository_memory_test.go — MemoryRepository 单元测试
// 覆盖核心 CRUD 路径：帖子创建/查询/更新/删除、评论操作、点赞/收藏/阅读记录
// 使用纯内存 map 实现，每个测试独立实例，无外部依赖
//
// 相关文档：
// - PRD 社交域 MVP-P0 Step 8：Topic 域 Repository 接口内存实现

package topic

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== 辅助函数 ==========

// newTestRepo 创建独立的 MemoryRepository 实例，避免测试间数据干扰
func newTestRepo() *MemoryRepository {
	return NewMemoryRepository()
}

// makeMemTestTopic 构建默认活跃状态的测试帖子（MemoryRepository 测试专用）
func makeMemTestTopic(id, groupID, authorID string) *Topic {
	now := time.Now().UnixMilli()
	return &Topic{
		ID:          id,
		Title:       fmt.Sprintf("测试帖子_%s", id),
		Content:     "这是测试内容",
		AuthorID:    authorID,
		AuthorName:  "测试作者",
		GroupID:     groupID,
		TopicType:   TopicTypeNormal,
		Status:      TopicStatusActive,
		Visibility:  TopicVisibilityPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// makeTestReply 构建默认活跃状态的测试评论
func makeTestReply(id, topicID, authorID string) *TopicReply {
	now := time.Now().UnixMilli()
	return &TopicReply{
		ID:        id,
		TopicID:   topicID,
		Content:   "这是测试评论内容",
		AuthorID:  authorID,
		AuthorName: "测试评论者",
		Status:    ReplyStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// makeTestLike 构建测试点赞记录
func makeTestLike(id, userID, topicID string) *TopicLike {
	return &TopicLike{
		ID:        id,
		UserID:    userID,
		TopicID:   topicID,
		CreatedAt: time.Now().UnixMilli(),
	}
}

// makeTestFavorite 构建测试收藏记录
func makeTestFavorite(id, userID, topicID string) *TopicFavorite {
	return &TopicFavorite{
		ID:        id,
		UserID:    userID,
		TopicID:   topicID,
		CreatedAt: time.Now().UnixMilli(),
	}
}

// makeTestRead 构建测试阅读记录
func makeTestRead(id, userID, topicID string) *TopicRead {
	return &TopicRead{
		ID:           id,
		UserID:       userID,
		TopicID:      topicID,
		ReadAt:       time.Now().UnixMilli(),
		ReadDuration: 60,
	}
}

// ========== 初始化与边界测试 ==========

func Test_NewMemoryRepository_初始化非nil(t *testing.T) {
	repo := NewMemoryRepository()

	require.NotNil(t, repo)
	assert.NotNil(t, repo.topics, "topics map 应初始化为非 nil")
	assert.NotNil(t, repo.replies, "replies map 应初始化为非 nil")
	assert.NotNil(t, repo.likes, "likes map 应初始化为非 nil")
	assert.NotNil(t, repo.favorites, "favorites map 应初始化为非 nil")
	assert.NotNil(t, repo.replyLikes, "replyLikes map 应初始化为非 nil")
	assert.NotNil(t, repo.reads, "reads map 应初始化为非 nil")
}

// ========== Topic CRUD 测试 ==========

func Test_CreateTopic_和_GetTopicByID_正常流程(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	topic := makeMemTestTopic("topic-001", "group-001", "user-001")

	err := repo.CreateTopic(ctx, topic)
	require.NoError(t, err)

	got, err := repo.GetTopicByID(ctx, "topic-001")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "topic-001", got.ID)
	assert.Equal(t, "测试帖子_topic-001", got.Title)
	assert.Equal(t, "user-001", got.AuthorID)
	assert.Equal(t, "group-001", got.GroupID)
}

func Test_GetTopicByID_查询不存在返回nil不报错(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	got, err := repo.GetTopicByID(ctx, "non-existent-id")
	require.NoError(t, err)
	assert.Nil(t, got, "查询不存在的帖子应返回 nil 而非错误")
}

func Test_UpdateTopic_更新已存在帖子内容(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	original := makeMemTestTopic("topic-002", "group-001", "user-001")
	err := repo.CreateTopic(ctx, original)
	require.NoError(t, err)

	updated := makeMemTestTopic("topic-002", "group-001", "user-001")
	updated.Title = "修改后的标题"
	updated.Content = "修改后的内容"
	updated.UpdatedAt = time.Now().UnixMilli()

	err = repo.UpdateTopic(ctx, updated)
	require.NoError(t, err)

	got, err := repo.GetTopicByID(ctx, "topic-002")
	require.NoError(t, err)
	assert.Equal(t, "修改后的标题", got.Title)
	assert.Equal(t, "修改后的内容", got.Content)
}

func Test_DeleteTopic_删除后查询返回nil(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	topic := makeMemTestTopic("topic-003", "group-001", "user-001")
	err := repo.CreateTopic(ctx, topic)
	require.NoError(t, err)

	err = repo.DeleteTopic(ctx, "topic-003")
	require.NoError(t, err)

	got, err := repo.GetTopicByID(ctx, "topic-003")
	require.NoError(t, err)
	assert.Nil(t, got, "删除后查询应返回 nil")
}

func Test_ListTopics_分页列出活跃帖子(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		topic := makeMemTestTopic(fmt.Sprintf("topic-list-%d", i), "group-001", "user-001")
		err := repo.CreateTopic(ctx, topic)
		require.NoError(t, err)
	}

	list, total, err := repo.ListTopics(ctx, 1, 3, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total, "总数应为全部活跃帖子数")
	assert.Len(t, list, 3, "第一页应返回 3 条")

	list2, _, err := repo.ListTopics(ctx, 2, 3, nil, "")
	require.NoError(t, err)
	assert.Len(t, list2, 2, "第二页应返回剩余 2 条")
}

func Test_ListTopicsByGroupID_按群组过滤分页(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	// 群组 A 的帖子
	for i := 1; i <= 3; i++ {
		topic := makeMemTestTopic(fmt.Sprintf("topic-gA-%d", i), "group-A", "user-001")
		repo.CreateTopic(ctx, topic)
	}
	// 群组 B 的帖子
	for i := 1; i <= 2; i++ {
		topic := makeMemTestTopic(fmt.Sprintf("topic-gB-%d", i), "group-B", "user-002")
		repo.CreateTopic(ctx, topic)
	}

	list, total, err := repo.ListTopicsByGroupID(ctx, "group-A", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 3)
	for _, tpc := range list {
		assert.Equal(t, "group-A", tpc.GroupID)
	}
}

func Test_ListTopicsByAuthorID_按作者过滤分页(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	topic1 := makeMemTestTopic("topic-auth-1", "group-001", "author-alice")
	topic2 := makeMemTestTopic("topic-auth-2", "group-001", "author-bob")
	topic3 := makeMemTestTopic("topic-auth-3", "group-001", "author-alice")

	repo.CreateTopic(ctx, topic1)
	repo.CreateTopic(ctx, topic2)
	repo.CreateTopic(ctx, topic3)

	list, total, err := repo.ListTopicsByAuthorID(ctx, "author-alice", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

// ========== Reply 评论操作测试 ==========

func Test_CreateReply_和_ListReplies_正常流程(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	reply1 := makeTestReply("reply-001", "topic-100", "user-001")
	reply2 := makeTestReply("reply-002", "topic-100", "user-002")

	err := repo.CreateReply(ctx, reply1)
	require.NoError(t, err)
	err = repo.CreateReply(ctx, reply2)
	require.NoError(t, err)

	list, total, err := repo.ListReplies(ctx, "topic-100", 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func Test_DeleteReply_删除后列表不再包含该评论(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	reply := makeTestReply("reply-del-001", "topic-100", "user-001")
	err := repo.CreateReply(ctx, reply)
	require.NoError(t, err)

	err = repo.DeleteReply(ctx, "reply-del-001")
	require.NoError(t, err)

	list, total, err := repo.ListReplies(ctx, "topic-100", 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
}

// ========== Like 点赞操作测试（复合键重点）==========

func Test_CreateLike_幂等创建同一用户对同一帖子的重复点赞不报错(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	like1 := makeTestLike("like-001", "user-001", "topic-100")
	err := repo.CreateLike(ctx, like1)
	require.NoError(t, err)

	like2 := makeTestLike("like-002", "user-001", "topic-100") // 同一 user→topic
	err = repo.CreateLike(ctx, like2)
	require.NoError(t, err, "幂等语义：重复点赞不应报错")

	liked, err := repo.IsTopicLiked(ctx, "topic-100", "user-001")
	require.NoError(t, err)
	assert.True(t, liked, "用户仍处于已点赞状态")
}

func Test_DeleteLike_取消点赞后状态变为未点赞(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	like := makeTestLike("like-001", "user-001", "topic-100")
	repo.CreateTopic(ctx, makeMemTestTopic("topic-100", "g1", "u1"))
	repo.CreateLike(ctx, like)

	err := repo.DeleteLike(ctx, "topic-100", "user-001")
	require.NoError(t, err)

	liked, err := repo.IsTopicLiked(ctx, "topic-100", "user-001")
	require.NoError(t, err)
	assert.False(t, liked, "取消点赞后应返回 false")
}

func Test_IsTopicLiked_检查是否已点赞(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	// 未点赞时
	liked, err := repo.IsTopicLiked(ctx, "topic-100", "user-001")
	require.NoError(t, err)
	assert.False(t, liked)

	// 点赞后
	like := makeTestLike("like-001", "user-001", "topic-100")
	repo.CreateLike(ctx, like)

	liked, err = repo.IsTopicLiked(ctx, "topic-100", "user-001")
	require.NoError(t, err)
	assert.True(t, liked)
}

func Test_CountLikesByTopicID_统计某帖子点赞数(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	users := []string{"user-001", "user-002", "user-003"}
	for i, uid := range users {
		like := makeTestLike(fmt.Sprintf("like-%d", i+1), uid, "topic-100")
		repo.CreateLike(ctx, like)
	}

	count, err := repo.CountLikesByTopicID(ctx, "topic-100")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// 无点赞的帖子
	countZero, err := repo.CountLikesByTopicID(ctx, "topic-no-likes")
	require.NoError(t, err)
	assert.Equal(t, int64(0), countZero)
}

// ========== Favorite 收藏操作测试 ==========

func Test_CreateFavorite_幂等收藏(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	fav := makeTestFavorite("fav-001", "user-001", "topic-200")
	err := repo.CreateFavorite(ctx, fav)
	require.NoError(t, err)

	favorited, err := repo.IsTopicFavorited(ctx, "topic-200", "user-001")
	require.NoError(t, err)
	assert.True(t, favorited)
}

func Test_DeleteFavorite_取消收藏后状态变为未收藏(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	fav := makeTestFavorite("fav-001", "user-001", "topic-200")
	repo.CreateFavorite(ctx, fav)

	err := repo.DeleteFavorite(ctx, "topic-200", "user-001")
	require.NoError(t, err)

	favorited, err := repo.IsTopicFavorited(ctx, "topic-200", "user-001")
	require.NoError(t, err)
	assert.False(t, favorited)
}

func Test_IsTopicFavorited_检查是否已收藏(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	// 未收藏时
	fav, err := repo.IsTopicFavorited(ctx, "topic-200", "user-001")
	require.NoError(t, err)
	assert.False(t, fav)

	// 收藏后
	f := makeTestFavorite("fav-001", "user-001", "topic-200")
	repo.CreateFavorite(ctx, f)

	fav, err = repo.IsTopicFavorited(ctx, "topic-200", "user-001")
	require.NoError(t, err)
	assert.True(t, fav)
}

func Test_CountFavoritesByTopicID_统计某帖子收藏数(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	for i := 1; i <= 4; i++ {
		fav := makeTestFavorite(fmt.Sprintf("fav-%d", i), fmt.Sprintf("user-%03d", i), "topic-200")
		repo.CreateFavorite(ctx, fav)
	}

	count, err := repo.CountFavoritesByTopicID(ctx, "topic-200")
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)
}

// ========== Read 阅读记录操作测试（UPSERT 语义）==========

func Test_UpsertReadRecord_首次写入创建记录(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	read := makeTestRead("read-001", "user-001", "topic-300")
	read.ReadDuration = 120

	err := repo.UpsertReadRecord(ctx, read)
	require.NoError(t, err)

	got, err := repo.GetReadRecord(ctx, "topic-300", "user-001")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 120, got.ReadDuration)
}

func Test_UpsertReadRecord_重复写入累加时长(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	first := makeTestRead("read-001", "user-001", "topic-300")
	first.ReadDuration = 60
	repo.UpsertReadRecord(ctx, first)

	second := makeTestRead("read-002", "user-001", "topic-300")
	second.ReadDuration = 90
	err := repo.UpsertReadRecord(ctx, second)
	require.NoError(t, err)

	got, err := repo.GetReadRecord(ctx, "topic-300", "user-001")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 150, got.ReadDuration, "UPSERT 语义：第二次写入应累加 duration")
}

func Test_GetReadRecord_查询不存在返回nil(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	got, err := repo.GetReadRecord(ctx, "topic-nonexist", "user-999")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func Test_CountDistinctReaders_统计独立阅读人数去重(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	// user-001 读了两次（同一个人）
	r1 := makeTestRead("r-1", "user-001", "topic-400")
	r1.ReadDuration = 30
	repo.UpsertReadRecord(ctx, r1)

	r2 := makeTestRead("r-2", "user-001", "topic-400")
	r2.ReadDuration = 40
	repo.UpsertReadRecord(ctx, r2)

	// user-002 读了一次
	r3 := makeTestRead("r-3", "user-002", "topic-400")
	r3.ReadDuration = 50
	repo.UpsertReadRecord(ctx, r3)

	// user-003 读了一次
	r4 := makeTestRead("r-4", "user-003", "topic-400")
	r4.ReadDuration = 20
	repo.UpsertReadRecord(ctx, r4)

	count, err := repo.CountDistinctReaders(ctx, "topic-400")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count, "user-001 读了两次但只算一个独立读者")
}

// ========== CountTopicsByGroupID 统计测试 ==========

func Test_CountTopicsByGroupID_统计群组内帖子数(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	for i := 1; i <= 7; i++ {
		topic := makeMemTestTopic(fmt.Sprintf("t-count-%d", i), "group-stats", "user-001")
		repo.CreateTopic(ctx, topic)
	}
	// 其他群组的帖子
	repo.CreateTopic(ctx, makeMemTestTopic("t-other", "other-group", "user-002"))

	count, err := repo.CountTopicsByGroupID(ctx, "group-stats")
	require.NoError(t, err)
	assert.Equal(t, int64(7), count)
}

// ========== CountRepliesByTopicID 统计测试 ==========

func Test_CountRepliesByTopicID_统计帖子有效评论数(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	// 活跃评论
	for i := 1; i <= 5; i++ {
		reply := makeTestReply(fmt.Sprintf("r-cnt-%d", i), "topic-cnt", "user-001")
		repo.CreateReply(ctx, reply)
	}
	// 已删除评论不计入
	deletedReply := makeTestReply("r-cnt-del", "topic-cnt", "user-002")
	deletedReply.Status = ReplyStatusDeleted
	repo.CreateReply(ctx, deletedReply)

	count, err := repo.CountRepliesByTopicID(ctx, "topic-cnt")
	require.NoError(t, err)
	assert.Equal(t, int64(5), count, "仅统计活跃状态的评论")
}

// ========== ListReplies 分页与 parentReplyID 过滤测试 ==========

func Test_ListReplies_按ParentReplyID过滤子回复(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	parent := makeTestReply("parent-001", "topic-filt", "user-001")
	repo.CreateReply(ctx, parent)

	child1 := makeTestReply("child-001", "topic-filt", "user-002")
	child1.ParentReplyID = "parent-001"
	repo.CreateReply(ctx, child1)

	child2 := makeTestReply("child-002", "topic-filt", "user-003")
	child2.ParentReplyID = "parent-001"
	repo.CreateReply(ctx, child2)

	// 不指定 parentReplyID，返回全部评论
	_, totalAll, _ := repo.ListReplies(ctx, "topic-filt", 1, 10, nil)
	assert.Equal(t, int64(3), totalAll)

	// 指定 parentReplyID 过滤子回复
	parentID := "parent-001"
	children, totalChildren, _ := repo.ListReplies(ctx, "topic-filt", 1, 10, &parentID)
	assert.Equal(t, int64(2), totalChildren)
	assert.Len(t, children, 2)
}

// ========== ListReadsByUserID 分页测试 ==========

func Test_ListReadsByUserID_分页查询用户阅读历史(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	for i := 1; i <= 6; i++ {
		read := makeTestRead(fmt.Sprintf("rd-hist-%d", i), "user-hist", fmt.Sprintf("topic-%03d", i))
		repo.UpsertReadRecord(ctx, read)
	}

	page1, total, err := repo.ListReadsByUserID(ctx, "user-hist", 1, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(6), total)
	assert.Len(t, page1, 3)

	page2, _, err := repo.ListReadsByUserID(ctx, "user-hist", 2, 3)
	require.NoError(t, err)
	assert.Len(t, page2, 3)
}

// ========== ListFavoritesByUserID 分页测试 ==========

func Test_ListFavoritesByUserID_分页查询用户收藏列表(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	for i := 1; i <= 4; i++ {
		fav := makeTestFavorite(fmt.Sprintf("f-lst-%d", i), "user-fav", fmt.Sprintf("topic-%03d", i))
		repo.CreateFavorite(ctx, fav)
	}

	list, total, err := repo.ListFavoritesByUserID(ctx, "user-fav", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Len(t, list, 4)
}

// ========== 复合键隔离性验证 ==========

func Test_Like复合键不同用户对同一帖子互不影响(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	repo.CreateLike(ctx, makeTestLike("l-1", "user-a", "topic-x"))
	repo.CreateLike(ctx, makeTestLike("l-2", "user-b", "topic-x"))

	count, err := repo.CountLikesByTopicID(ctx, "topic-x")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// 取消 user-a 的点赞不影响 user-b
	repo.DeleteLike(ctx, "topic-x", "user-a")

	likedB, _ := repo.IsTopicLiked(ctx, "topic-x", "user-b")
	assert.True(t, likedB, "user-b 的点赞不受影响")

	likedA, _ := repo.IsTopicLiked(ctx, "topic-x", "user-a")
	assert.False(t, likedA, "user-a 已取消点赞")
}

func Test_Favorite复合键不同用户对同一帖子互不影响(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	repo.CreateFavorite(ctx, makeTestFavorite("f-1", "user-a", "topic-y"))
	repo.CreateFavorite(ctx, makeTestFavorite("f-2", "user-b", "topic-y"))

	count, _ := repo.CountFavoritesByTopicID(ctx, "topic-y")
	assert.Equal(t, int64(2), count)

	repo.DeleteFavorite(ctx, "topic-y", "user-a")

	favB, _ := repo.IsTopicFavorited(ctx, "topic-y", "user-b")
	assert.True(t, favB)

	favA, _ := repo.IsTopicFavorited(ctx, "topic-y", "user-a")
	assert.False(t, favA)
}

func Test_Read复合键不同用户对同一帖子独立记录(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	rA := makeTestRead("ra-1", "user-a", "topic-z")
	rA.ReadDuration = 30
	repo.UpsertReadRecord(ctx, rA)

	rB := makeTestRead("rb-1", "user-b", "topic-z")
	rB.ReadDuration = 45
	repo.UpsertReadRecord(ctx, rB)

	count, _ := repo.CountDistinctReaders(ctx, "topic-z")
	assert.Equal(t, int64(2), count)

	gotA, _ := repo.GetReadRecord(ctx, "topic-z", "user-a")
	assert.Equal(t, 30, gotA.ReadDuration)

	gotB, _ := repo.GetReadRecord(ctx, "topic-z", "user-b")
	assert.Equal(t, 45, gotB.ReadDuration)
}
