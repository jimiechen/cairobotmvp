package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// mockRepository 实现 Repository 接口的 Mock，用于隔离数据库
type mockRepository struct {
	users  map[string]*User
	blocks map[string]*MemberBlock // key: blockerID:blockedID
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:  make(map[string]*User),
		blocks: make(map[string]*MemberBlock),
	}
}

func (m *mockRepository) CreateUser(ctx context.Context, user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, nil
}

// 以下为满足 Repository 接口的存根方法（Register 流程不使用）
func (m *mockRepository) GetUserByUID(ctx context.Context, uid string) (*User, error) { return nil, nil }
func (m *mockRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}
func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error)   { return nil, nil }
func (m *mockRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error)  { return nil, nil }
func (m *mockRepository) UpdateUser(ctx context.Context, user *User) error {
	if _, ok := m.users[user.ID]; ok {
		m.users[user.ID] = user
	}
	return nil
}
func (m *mockRepository) BatchGetUsersByID(ctx context.Context, ids []string) ([]*User, error) {
	var result []*User
	for _, id := range ids {
		if u, ok := m.users[id]; ok {
			result = append(result, u)
		}
	}
	return result, nil
}
func (m *mockRepository) CreateBlock(ctx context.Context, block *MemberBlock) error {
	key := block.BlockerID + ":" + block.BlockedID
	m.blocks[key] = block
	return nil
}
func (m *mockRepository) DeleteBlock(ctx context.Context, blockerID, blockedID string) error {
	key := blockerID + ":" + blockedID
	delete(m.blocks, key)
	return nil
}
func (m *mockRepository) ListBlocks(ctx context.Context, blockerID string, page, size int) ([]*MemberBlock, int64, error) {
	var result []*MemberBlock
	for key, b := range m.blocks {
		// 解析 key 获取 blockerID
		if len(key) > 0 {
			prefix := blockerID + ":"
			if len(key) > len(prefix) && key[:len(prefix)] == prefix {
				result = append(result, b)
			}
		}
	}
	total := int64(len(result))
	// 简单分页
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(result) {
		return []*MemberBlock{}, total, nil
	}
	end := start + size
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}
func (m *mockRepository) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	key := blockerID + ":" + blockedID
	_, ok := m.blocks[key]
	return ok, nil
}
func (m *mockRepository) GetBlockCount(ctx context.Context, blockerID string) (int64, error) { return 0, nil }
func (m *mockRepository) GetOrCreateStats(ctx context.Context, userID string) (*MemberStats, error) {
	return &MemberStats{UserID: userID}, nil
}
func (m *mockRepository) UpdateStats(ctx context.Context, stats *MemberStats) error { return nil }

// TestSvcRegister_正常注册 当用户名和密码合法时_应返回成功响应包含用户ID
func TestSvcRegister_正常注册(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcRegister(mockRepo, nil)

	req := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:   "test@example.com",
		Nickname: "测试用户",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp == nil {
		t.Fatal("期望返回 Response，实际为 nil")
	}
	if resp.Result == nil {
		t.Fatal("期望 Result 不为 nil")
	}
	if resp.Result.Code != 10200 { // ERROR_CODE_SUCCESS
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.UserId == "" {
		t.Error("期望 UserId 不为空")
	}
}

// TestSvcRegister_用户名重复 当用户名已存在时_应返回名称已被占用错误
func TestSvcRegister_用户名重复(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	// 预先创建一个用户占用用户名
	existingUser := &User{
		ID:       "user-001",
		Username: "testuser",
	}
	mockRepo.users[existingUser.ID] = existingUser

	svc := NewSvcRegister(mockRepo, nil)

	req := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "password123",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != 10612 { // USER_ERROR_NAME_ALREADY_TAKEN
		t.Errorf("期望名称已占用错误码 10612，实际得到 %d", resp.Result.Code)
	}
}

// TestSvcRegister_缺少必填字段 当用户名为空时_应返回参数校验错误
func TestSvcRegister_缺少必填字段(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcRegister(mockRepo, nil)

	req := &pb.UserRegisterRequest{
		Username: "",
		Password: "password123",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code >= 10200 && resp.Result.Code < 10300 {
		t.Errorf("期望参数校验错误（< 10200 或 >= 10400），实际得到成功类码 %d", resp.Result.Code)
	}
}
