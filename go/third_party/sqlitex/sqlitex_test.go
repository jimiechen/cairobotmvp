package sqlitex

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultMemoryConfig(t *testing.T) {
	cfg := DefaultMemoryConfig()
	assert.Equal(t, ModeMemory, cfg.Mode)
	assert.Equal(t, ":memory:", cfg.DSN)
	assert.True(t, cfg.WAL)
	assert.Equal(t, 5000, cfg.BusyTimeout)
}

func TestDefaultFileConfig(t *testing.T) {
	cfg := DefaultFileConfig("/tmp/test.db")
	assert.Equal(t, ModeFile, cfg.Mode)
	assert.Equal(t, "/tmp/test.db", cfg.DSN)
}

func TestOpen_NilConfig(t *testing.T) {
	_, err := Open(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestOpen_MemoryMode(t *testing.T) {
	cfg := DefaultMemoryConfig()
	db, err := Open(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	err = db.Ping(context.Background())
	require.NoError(t, err)
}

func TestOpen_ExecAndQuery(t *testing.T) {
	cfg := DefaultMemoryConfig()
	db, err := Open(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	_, err = db.Exec(ctx, `CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO test_table (name) VALUES (?)`, "hello")
	require.NoError(t, err)

	rows, err := db.Query(ctx, `SELECT name FROM test_table`)
	require.NoError(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, "hello", name)
		count++
	}
	assert.Equal(t, 1, count)
}

func TestBeginTx(t *testing.T) {
	cfg := DefaultMemoryConfig()
	db, err := Open(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	_, err = db.Exec(ctx, `CREATE TABLE tx_test (id INTEGER PRIMARY KEY, val INTEGER)`)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx, `INSERT INTO tx_test (val) VALUES (?)`, 42)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	row := db.QueryRow(ctx, `SELECT val FROM tx_test`)
	var val int
	err = row.Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 42, val)
}

func TestSchemaAdapter_ConvertDDL(t *testing.T) {
	a := NewSchemaAdapter()

	mysqlDDL := `
		CREATE TABLE users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			is_active TINYINT(1) DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			meta JSON,
			Engine=InnoDB
		) COMMENT='用户表';
	`

	sqliteDDL := a.ConvertDDL(mysqlDDL)

	assert.NotContains(t, sqliteDDL, "BIGINT AUTO_INCREMENT")
	assert.Contains(t, sqliteDDL, "INTEGER PRIMARY KEY AUTOINCREMENT")
	assert.NotContains(t, sqliteDDL, "TINYINT(1)")
	assert.Contains(t, sqliteDDL, "DATETIME")
	assert.NotContains(t, sqliteDDL, "ENGINE=InnoDB")
	assert.NotContains(t, sqliteDDL, "JSON")
}
