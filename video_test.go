package archive

import (
	"testing"

	"github.com/zhangyiming748/finder"
)

// go test -v -timeout 1h -run TestVideo
func TestVideo(t *testing.T) {
	t.Log("Video Test")
	// 初始化SQLite数据库
	InitSqlte()
	videos := finder.FindAllVideos("/Users/zen/Movies/done/无敌大喵子")
	for _, video := range videos {
		t.Log(video)
		Convert2H265(video, false)
	}
}
