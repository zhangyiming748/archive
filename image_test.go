package archive

import (
	"testing"
)

// go test -v -timeout 0 -run TestConvertImage
func TestConvertImage(t *testing.T) {
	Convert2AVIF("/Volumes/T7/archive/办公室交接文件2020/照片2009-2013/图片汇总2005-2008/2005年检文件扫描版/7C97AB35.bmp")
}
