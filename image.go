package archive

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
最终转换图片为 avif格式
*/
func Convert2AVIF(src string) {
	dst := strings.Replace(src, filepath.Ext(src), ".avif", 1)
	args := []string{"-i", src}
	args = append(args, "-c:v", "libaom-av1")
	args = append(args, "-still-picture", "1")
	args = append(args, dst)
	cmd := exec.Command("ffmpeg", args...)
	log.Printf("开始运行转换命令:%v\n", cmd.String())
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("转换失败：%v\n源文件%v\n", err, src)
		return
	} else {
		fmt.Printf("转换成功：%s\n", string(out))
		diffSize(src, dst)
		if e := os.Remove(src); e != nil {
			log.Fatalf("删除源文件失败：%v\n", err)
		}
	}
}
