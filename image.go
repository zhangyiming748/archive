package archive

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

/*
最终转换图片为 avif格式
*/
func Convert2AVIF(src string, threads int) error {
	if threads == 0 {
		threads = runtime.NumCPU()
	}
	dst := replaceExt(src, ".avif")
	if strings.ToLower(filepath.Ext(src)) == ".gif" {
		log.Printf("跳过gif文件:%v\n", src)
		return nil
	}
	if strings.ToLower(filepath.Ext(src)) == ".bmp" {
		log.Printf("走bmp逻辑:%v\n", src)
		ConvertBMP2AVIF4JPG(src, dst, threads)
		return nil
	}
	if strings.ToLower(filepath.Ext(src)) == ".avif" {
		log.Printf("跳过avif文件:%v\n", src)
		return nil
	}
	//dst := strings.Replace(src, filepath.Ext(src), ".avif", 1)
	// avifenc --codec aom --min 20 --max 30 --speed 6
	args := []string{"--codec", "aom"}
	args = append(args, "--min", "20")
	args = append(args, "--max", "30")
	args = append(args, "--speed", "0")
	args = append(args, "--jobs", strconv.Itoa(threads))
	args = append(args, src)
	args = append(args, dst)
	cmd := exec.Command("avifenc", args...)
	log.Printf("开始运行转换命令:%v\n", cmd.String())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("转换失败：%v\n源文件%v\n", err, src)
	} else {
		fmt.Printf("转换成功：%s\n", string(out))
		diffSize(src, dst)
		if e := os.Remove(src); e != nil {
			return fmt.Errorf("删除源文件失败：%v\n", err)
		}
	}
	return nil
}

/*
bmp先使用ffmpeg转换成中间格式jpg,再使用avifenc转换成avif
需要安装用 ImageMagick(macOS 上 brew install imagemagick)
*/
func ConvertBMP2AVIF4JPG(src, dst string, threads int) {
	if strings.ToLower(filepath.Ext(src)) != ".bmp" {
		log.Printf("不是bmp文件:%v\n", src)
		return
	}
	middle := replaceExt(src, ".png")
	args := []string{}
	args = append(args, src)
	args = append(args, middle)
	cmd := exec.Command("magick", args...)
	log.Printf("开始运行转换命令:%v\n", cmd.String())
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("中间文件转换失败,保留原文件%v\n", src)
		return
	} else {
		log.Printf("中间文件转换成功：%s\n", string(out))
		if err := Convert2AVIF(middle, threads); err != nil {
			log.Printf("中间文件转换错误,保留中间文件：%v\t删除源文件%v\n", middle, src)
			os.Remove(src)
		} else {
			log.Printf("中间文件转换成功,删除中间文件：%v\t删除源文件%v\n", middle, src)
			os.Remove(middle)
			os.Remove(src)
		}
	}
}
