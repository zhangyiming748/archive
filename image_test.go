package archive

import (
	"testing"
)

// go test -v -timeout 0 -run TestConvertImage
func TestConvertImage(t *testing.T) {
	Convert2AVIF("/Users/zen/gitea/FastTdl/discussions/面饼仙儿 皮裤连体红毛衣 [20P-289MB]/TG@coserTG (15).jpg")
}
