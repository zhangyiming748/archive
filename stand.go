package archive

import (
	"fmt"
	"github.com/zhangyiming748/archive/sqlite"
	"log"
	"os"
	"path/filepath"
)

func init() {
	sqlite.SetSqlite()
}
func diffSize(src, dst string) {
	s := new(sqlite.Save)
	s.FileName = filepath.Base(src)

	// 获取源文件和目标文件的大小并计算差值
	srcFileInfo, _ := os.Stat(src)
	s.Before = fmt.Sprintf("%.3f", float64(srcFileInfo.Size())/(1024*1024))
	dstFileInfo, _ := os.Stat(dst)
	s.After = fmt.Sprintf("%.3f", float64(dstFileInfo.Size())/(1024*1024))
	sizeDiff := float64(srcFileInfo.Size()-dstFileInfo.Size()) / (1024 * 1024)
	//s.Save = fmt.Sprintf("%.3f", sizeDiff)
	s.SaveSize = sizeDiff
	if sizeDiff > 0 {
		log.Printf("源文件%v比目标文件%v大%.3f MB\n", src, dst, sizeDiff)
	} else {
		log.Printf("源文件%v比目标文件%v小%.3f MB\n", src, dst, -sizeDiff)
	}
	log.Printf("源文件%v与目标文件%v大小差值为: %.3f MB\n", src, dst, sizeDiff)
	s.Insert()
}
