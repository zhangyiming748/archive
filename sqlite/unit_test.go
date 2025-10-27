package sqlite

import (
	"testing"
)

func TestHistoryModel(t *testing.T) {
	SetSqlite()
	// 同步表结构
	s := new(Save)
	s.Sync()
	// 插入测试数据
	s.SaveSize = 1.0
	err := s.Insert()
	if err != nil {
		t.Errorf("插入数据失败: %v", err)
	}
}
