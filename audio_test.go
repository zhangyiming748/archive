package archive

import (
	"github.com/zhangyiming748/finder"
	"testing"
)

// go test -v -timeout 72h -run TestAudio
func TestAudio(t *testing.T) {
	fs := finder.FindAllAudios("/Users/zen/Downloads/第四季_副本")
	for _, f := range fs {
		ConvertAudio(f, AudioBookType)
	}
}
