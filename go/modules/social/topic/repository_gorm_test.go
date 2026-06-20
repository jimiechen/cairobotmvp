// repository_gorm_test.go — GormRepository 冒烟测试
// 覆盖核心 CRUD 路径：帖子创建/查询/列表、评论操作、点赞/收藏/阅读记录
// 使用真实 MySQL 数据库，每个测试通过 ID 前缀隔离数据
// 环境变量未配置时自动 Skip
//
// 相关文档：
// - PRD 社交域 MVP-P0 Step 8：Topic 域 Repository 接口 GORM 实现

package topic

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const testIDPrefix = "gt_tpc_"

// getEnv 读取环境变量，未设置时返回 fallback 值
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// setupTestDB 连接真实 MySQL 数据库（从环境变量读取连接参数）
// 未配置 MYSQL_HOST 时自动 t.Skip
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	host := getEnv("MYSQL_HOST", "")
	if host == "" {
		t.Skip("跳过：未设置 MYSQL_HOST 环境变量")
	}
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "root")
	pass := getEnv("MYSQL_PASSWORD", "")
	dbname := getEnv("MYSQL_DATABASE", "cairobot_test")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// cleanupTopicData 清理当前测试创建的帖子相关数据（按 ID 前缀匹配）
func cleanupTopicData(t *testing.T, db *gorm.DB) {
	t.Helper()
	prefix := testIDPrefix + "%"
	db.Exec("DELETE FROM reply_likes WHERE id LIKE ?", prefix)
	db.Exec("DELETE FROM topic_favorites WHERE id LIKE ?", prefix)
	db.Exec("DELETE FROM topic_likes WHERE id LIKE ?", prefix)
	db.Exec("DELETE FROM topic_reads WHERE id LIKE ?", prefix)
	db.Exec("DELETE FROM topic_replies WHERE id LIKE ?", prefix)
	db.Exec("DELETE FROM topics WHERE id LIKE ?", prefix)
}

// 辅助函数：构建一个默认的 Topic 测试数据
func makeTestTopic(id, groupID, authorID string) *Topic {
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

// ========== Topic 帖子 CRUD 测试 ==========

func TestCreateAndGetTopic(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	topic := makeTestTopic(testIDPrefix+"t001", testIDPrefix+"g001", "u001")

	// 创建
	err := repo.CreateTopic(ctx, topic)
	require.NoError(t, err)

	// 按 ID 查回，验证关键字段
	got, err := repo.GetTopicByID(ctx, testIDPrefix+"t001")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, testIDPrefix+"t001", got.ID)
	assert.Equal(t, testIDPrefix+"g001", got.GroupID)
	assert.Equal(t, "u001", got.AuthorID)
	assert.Equal(t, "测试帖子_"+testIDPrefix+"t001", got.Title)
	assert.Equal(t, "这是测试内容", got.Content)
	assert.Equal(t, TopicStatusActive, got.Status)
	assert.Equal(t, TopicVisibilityPublic, got.Visibility)
}

func TestGetTopicNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	got, err := repo.GetTopicByID(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestUpdateTopic(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	topic := makeTestTopic(testIDPrefix+"t002", testIDPrefix+"g001", "u001")
	require.NoError(t, repo.CreateTopic(ctx, topic))

	// 更新标题和状态
	topic.Title = "更新后的标题"
	topic.Status = TopicStatusLocked
	topic.UpdatedAt = time.Now().UnixMilli()
	err := repo.UpdateTopic(ctx, topic)
	require.NoError(t, err)

	// 验证更新结果
	got, err := repo.GetTopicByID(ctx, testIDPrefix+"t002")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "更新后的标题", got.Title)
	assert.Equal(t, TopicStatusLocked, got.Status)
}

func TestDeleteTopic(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	topic := makeTestTopic(testIDPrefix+"t003", testIDPrefix+"g001", "u001")
	require.NoError(t, repo.CreateTopic(ctx, topic))

	// 删除
	err := repo.DeleteTopic(ctx, testIDPrefix+"t003")
	require.NoError(t, err)

	// 验证已删除
	got, err := repo.GetTopicByID(ctx, testIDPrefix+"t003")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestListTopicsWithPagination(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 创建 5 条帖子（同一群组）
	for i := 0; i < 5; i++ {
		topic := makeTestTopic(fmt.Sprintf(testIDPrefix+"t_list_%d", i), testIDPrefix+"g_list", "u_list")
		require.NoError(t, repo.CreateTopic(ctx, topic))
	}

	// 第 1 页：每页 2 条
	topics, total, err := repo.ListTopics(ctx, 1, 2, map[string]interface{}{"group_id": testIDPrefix + "g_list"}, "")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, topics, 2)

	// 第 2 页：应剩余条数
	topics2, total2, err := repo.ListTopics(ctx, 2, 2, map[string]interface{}{"group_id": testIDPrefix + "g_list"}, "")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total2)
	assert.Len(t, topics2, 2)

	// 第 3 页：最后 1 条
	topics3, _, err := repo.ListTopics(ctx, 3, 2, map[string]interface{}{"group_id": testIDPrefix + "g_list"}, "")
	require.NoError(t, err)
	assert.Len(t, topics3, 1)
}

func TestListTopicsByGroupID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 群组 A 有 3 条，群组 B 有 1 条
	for i := 0; i < 3; i++ {
		topic := makeTestTopic(fmt.Sprintf(testIDPrefix+"tgA_%d", i), "group_A", "u001")
		require.NoError(t, repo.CreateTopic(ctx, topic))
	}
	topicB := makeTestTopic(testIDPrefix+"tgB_0", "group_B", "u001")
	require.NoError(t, repo.CreateTopic(ctx, topicB))

	topics, total, err := repo.ListTopicsByGroupID(ctx, "group_A", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, topics, 3)
}

func TestCountTopicsByGroupID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	count, err := repo.CountTopicsByGroupID(ctx, "empty_group")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 创建活跃帖子
	for i := 0; i < 3; i++ {
		topic := makeTestTopic(fmt.Sprintf(testIDPrefix+"tcnt_%d", i), testIDPrefix+"cnt_group", "u001")
		require.NoError(t, repo.CreateTopic(ctx, topic))
	}

	count, err = repo.CountTopicsByGroupID(ctx, testIDPrefix+"cnt_group")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

// ========== TopicReply 评论操作测试 ==========

func TestCreateAndGetReply(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	reply := &TopicReply{
		ID:        testIDPrefix + "r001",
		TopicID:   testIDPrefix + "t_topic",
		Content:   "这是一条评论",
		AuthorID:  "u001",
		AuthorName: "评论者",
		Status:    ReplyStatusActive,
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}
	err := repo.CreateReply(ctx, reply)
	require.NoError(t, err)

	got, err := repo.GetReplyByID(ctx, testIDPrefix+"r001")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, testIDPrefix+"r001", got.ID)
	assert.Equal(t, testIDPrefix+"t_topic", got.TopicID)
	assert.Equal(t, "这是一条评论", got.Content)
}

func TestGetReplyNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	got, err := repo.GetReplyByID(ctx, "no_reply")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestCreateAndListReplies(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 创建 4 条顶层评论
	for i := 0; i < 4; i++ {
		reply := &TopicReply{
			ID:        fmt.Sprintf(testIDPrefix + "r_list_%d", i),
			TopicID:   testIDPrefix + "t_for_replies",
			Content:   fmt.Sprintf("评论内容_%d", i),
			AuthorID:  "u001",
			AuthorName: "用户",
			Status:    ReplyStatusActive,
			CreatedAt: time.Now().UnixMilli() + int64(i*1000),
			UpdatedAt: time.Now().UnixMilli() + int64(i*1000),
		}
		require.NoError(t, repo.CreateReply(ctx, reply))
	}

	// 分页查询顶层评论
	replies, total, err := repo.ListReplies(ctx, testIDPrefix+"t_for_replies", 1, 2, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Len(t, replies, 2)
}

func TestDeleteReply(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	reply := &TopicReply{
		ID:        testIDPrefix + "r_del",
		TopicID:   testIDPrefix + "t_del",
		Content:   "待删除评论",
		AuthorID:  "u001",
		Status:    ReplyStatusActive,
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}
	require.NoError(t, repo.CreateReply(ctx, reply))

	err := repo.DeleteReply(ctx, testIDPrefix+"r_del")
	require.NoError(t, err)

	got, err := repo.GetReplyByID(ctx, testIDPrefix+"r_del")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestCountRepliesByTopicID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	count, err := repo.CountRepliesByTopicID(ctx, testIDPrefix+"t_empty")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	for i := 0; i < 3; i++ {
		reply := &TopicReply{
			ID:        fmt.Sprintf(testIDPrefix + "rcnt_%d", i),
			TopicID:   testIDPrefix + "t_cnt_reply",
			Content:   "有效评论",
			AuthorID:  "u001",
			Status:    ReplyStatusActive,
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
		}
		require.NoError(t, repo.CreateReply(ctx, reply))
	}

	count, err = repo.CountRepliesByTopicID(ctx, testIDPrefix+"t_cnt_reply")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

// ========== TopicLike 点赞操作测试 ==========

func TestLikeAndUnlike(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	topicID := testIDPrefix + "t_like"
	userID := "u_liker"

	// 初始未点赞
	liked, err := repo.IsTopicLiked(ctx, topicID, userID)
	require.NoError(t, err)
	assert.False(t, liked)

	// 点赞
	like := &TopicLike{
		ID:      testIDPrefix + "l001",
		TopicID: topicID,
		UserID:  userID,
		CreatedAt: time.Now().UnixMilli(),
	}
	err = repo.CreateLike(ctx, like)
	require.NoError(t, err)

	// 验证已点赞
	liked, err = repo.IsTopicLiked(ctx, topicID, userID)
	require.NoError(t, err)
	assert.True(t, liked)

	// 验证计数
	count, err := repo.CountLikesByTopicID(ctx, topicID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// 取消点赞
	err = repo.DeleteLike(ctx, topicID, userID)
	require.NoError(t, err)

	// 验证已取消
	liked, err = repo.IsTopicLiked(ctx, topicID, userID)
	require.NoError(t, err)
	assert.False(t, liked)

	// 验证计数归零
	count, err = repo.CountLikesByTopicID(ctx, topicID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestCreateLikeIdempotent(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	topicID := testIDPrefix + "t_idem"
	userID := "u_idem"

	like := &TopicLike{
		ID:      testIDPrefix + "l_idem",
		TopicID: topicID,
		UserID:  userID,
		CreatedAt: time.Now().UnixMilli(),
	}
	// 第一次点赞成功
	err := repo.CreateLike(ctx, like)
	require.NoError(t, err)

	// 重复点赞不报错（幂等）
	err = repo.CreateLike(ctx, like)
	require.NoError(t, err)

	// 计数仍为 1
	count, err := repo.CountLikesByTopicID(ctx, topicID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// ========== TopicFavorite 收藏操作测试 ==========

func TestFavoriteToggle(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	topicID := testIDPrefix + "t_fav"
	userID := "u_faver"

	// 初始未收藏
	fav, err := repo.IsTopicFavorited(ctx, topicID, userID)
	require.NoError(t, err)
	assert.False(t, fav)

	// 收藏
	favorite := &TopicFavorite{
		ID:      testIDPrefix + "f001",
		TopicID: topicID,
		UserID:  userID,
		CreatedAt: time.Now().UnixMilli(),
	}
	err = repo.CreateFavorite(ctx, favorite)
	require.NoError(t, err)

	// 验证已收藏
	fav, err = repo.IsTopicFavorited(ctx, topicID, userID)
	require.NoError(t, err)
	assert.True(t, fav)

	// 取消收藏
	err = repo.DeleteFavorite(ctx, topicID, userID)
	require.NoError(t, err)

	// 验证已取消
	fav, err = repo.IsTopicFavorited(ctx, topicID, userID)
	require.NoError(t, err)
	assert.False(t, fav)
}

func TestListFavoritesByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := "u_favlist"

	// 用户收藏了 3 个帖子
	for i := 0; i < 3; i++ {
		fav := &TopicFavorite{
			ID:       fmt.Sprintf(testIDPrefix+"f_list_%d", i),
			TopicID:  fmt.Sprintf(testIDPrefix+"t_fav_%d", i),
			UserID:   userID,
			CreatedAt: time.Now().UnixMilli() + int64(i*1000),
		}
		require.NoError(t, repo.CreateFavorite(ctx, fav))
	}

	favs, total, err := repo.ListFavoritesByUserID(ctx, userID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, favs, 3)
}

func TestCountFavoritesByTopicID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	count, err := repo.CountFavoritesByTopicID(ctx, testIDPrefix+"t_nofav")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	for i := 0; i < 2; i++ {
		fav := &TopicFavorite{
			ID:       fmt.Sprintf(testIDPrefix+"fcnt_%d", i),
			TopicID:  testIDPrefix + "t_favcnt",
			UserID:   fmt.Sprintf("u_favcnt_%d", i),
			CreatedAt: time.Now().UnixMilli(),
		}
		require.NoError(t, repo.CreateFavorite(ctx, fav))
	}

	count, err = repo.CountFavoritesByTopicID(ctx, testIDPrefix+"t_favcnt")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// ========== ReplyLike 评论点赞测试 ==========

func TestReplyLike(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	replyID := testIDPrefix + "r_like"
	userID := "u_rliker"

	// 初始未点赞
	liked, err := repo.IsReplyLiked(ctx, replyID, userID)
	require.NoError(t, err)
	assert.False(t, liked)

	// 点赞评论
	rl := &ReplyLike{
		ID:        testIDPrefix + "rl001",
		ReplyID:   replyID,
		UserID:    userID,
		CreatedAt: time.Now().UnixMilli(),
	}
	err = repo.CreateReplyLike(ctx, rl)
	require.NoError(t, err)

	// 验证已点赞
	liked, err = repo.IsReplyLiked(ctx, replyID, userID)
	require.NoError(t, err)
	assert.True(t, liked)

	// 验证计数
	count, err := repo.CountLikesByReplyID(ctx, replyID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// 取消点赞
	err = repo.DeleteReplyLike(ctx, replyID, userID)
	require.NoError(t, err)

	// 验证已取消
	liked, err = repo.IsReplyLiked(ctx, replyID, userID)
	require.NoError(t, err)
	assert.False(t, liked)

	// 计数归零
	count, err = repo.CountLikesByReplyID(ctx, replyID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ========== TopicRead 阅读记录测试 ==========

func TestUpsertReadRecord(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	topicID := testIDPrefix + "t_read"
	userID := "u_reader"

	// 首次阅读 — 应创建记录
	read := &TopicRead{
		ID:           testIDPrefix + "rd001",
		TopicID:      topicID,
		UserID:       userID,
		ReadAt:       time.Now().UnixMilli(),
		ReadDuration: 30,
	}
	err := repo.UpsertReadRecord(ctx, read)
	require.NoError(t, err)

	// 获取阅读记录验证
	got, err := repo.GetReadRecord(ctx, topicID, userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, topicID, got.TopicID)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, 30, got.ReadDuration)

	// 再次阅读 — 应更新（UPSERT）
	read.ReadAt = time.Now().UnixMilli() + 60000
	read.ReadDuration = 90 // 累加到 90 秒
	err = repo.UpsertReadRecord(ctx, read)
	require.NoError(t, err)

	got, err = repo.GetReadRecord(ctx, topicID, userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 90, got.ReadDuration)
}

func TestGetReadRecordNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	got, err := repo.GetReadRecord(ctx, "no_topic", "no_user")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestListReadsByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := "u_readlist"

	for i := 0; i < 3; i++ {
		read := &TopicRead{
			ID:           fmt.Sprintf(testIDPrefix+"rd_list_%d", i),
			TopicID:      fmt.Sprintf(testIDPrefix+"t_read_%d", i),
			UserID:       userID,
			ReadAt:       time.Now().UnixMilli() + int64(i*1000),
			ReadDuration: 10 * (i + 1),
		}
		require.NoError(t, repo.UpsertReadRecord(ctx, read))
	}

	reads, total, err := repo.ListReadsByUserID(ctx, userID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, reads, 3)
}

func TestCountDistinctReaders(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTopicData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	count, err := repo.CountDistinctReaders(ctx, testIDPrefix+"t_no_readers")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 3 个用户阅读同一帖子
	for i := 0; i < 3; i++ {
		read := &TopicRead{
			ID:           fmt.Sprintf(testIDPrefix+"rd_dr_%d", i),
			TopicID:      testIDPrefix + "t_has_readers",
			UserID:       fmt.Sprintf("u_reader_%d", i),
			ReadAt:       time.Now().UnixMilli(),
			ReadDuration: 10,
		}
		require.NoError(t, repo.UpsertReadRecord(ctx, read))
	}

	count, err = repo.CountDistinctReaders(ctx, testIDPrefix+"t_has_readers")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

// ========== 接口编译检查 ==========

// TestGormRepositoryImplementsInterface 编译期检查 GormRepository 是否实现 Repository 接口
func TestGormRepositoryImplementsInterface(t *testing.T) {
	var _ Repository = (*GormRepository)(nil)
}
