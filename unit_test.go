package archive

import (
	"github.com/zhangyiming748/archive/sqlite"
	"testing"
)

func TestConvertOneVideo(t *testing.T) {
	Convert2H265("G:\\pikpak\\My Pack222\\a.avi")
}
func TestInsert2Sqlite(t *testing.T) {
	diffSize("G:\\pikpak\\My Pack222\\a.avi", "G:\\pikpak\\My Pack222\\1001.mp4")
}
func TestHistoryModel(t *testing.T) {

	// 同步表结构
	s := new(sqlite.Save)
	s.Sync()
	// 插入测试数据
	s.SaveSize = 1.0
	err := s.Insert()
	if err != nil {
		t.Errorf("插入数据失败: %v", err)
	}
}
