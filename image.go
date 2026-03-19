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
	if strings.ToLower(filepath.Ext(src)) == ".gif" {
		log.Printf("跳过gif文件:%v\n", src)
		return
	}
	if strings.ToLower(filepath.Ext(src)) == ".avif" {
		log.Printf("跳过avif文件:%v\n", src)
		return
	}
	dst := strings.Replace(src, filepath.Ext(src), ".avif", 1)
	// avifenc --codec aom --min 20 --max 30 --speed 6
	args := []string{"--codec", "aom"}
	args = append(args, "--min", "20")
	args = append(args, "--max", "30")
	args = append(args, "--speed", "0")
	args = append(args, src)
	args = append(args, dst)
	cmd := exec.Command("avifenc", args...)
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
