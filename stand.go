package archive

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zhangyiming748/archive/sqlite"
)

func InitSqlte() {
	sqlite.SetSqlite()
	// 这里同步表结构
	s := new(sqlite.Save)
	s.Sync()
}

// checkDependencies 检测系统中是否存在必要的依赖命令
func CheckVideoDependencies() {
	commands := map[string]string{
		"ffmpeg":    "FFmpeg is a complete, cross-platform solution to record, convert and stream audio and video.",
		"mediainfo": "MediaInfo is a convenient unified display of the most relevant technical and tag data for video and audio files.",
	}

	missingDeps := []string{}

	for command, description := range commands {
		if !isCommandAvailable(command) {
			missingDeps = append(missingDeps, command)
			log.Printf("警告: 未找到依赖命令 '%s' - %s\n", command, description)
		} else {
			log.Printf("成功: 找到依赖命令 '%s'\n", command)
		}
	}

	// 检查是否有缺失的依赖
	if len(missingDeps) > 0 {
		log.Fatalf("缺少以下必要依赖: %v,请安装后再运行程序.\n", missingDeps)
	} else {
		log.Println("所有必要依赖均已找到,程序可以正常运行.")
	}
}
func CheckImageDependencies() {
	commands := map[string]string{
		"avifenc": "AVIF encoder (libavif) is used for encoding images to AVIF format.",
	}

	missingDeps := []string{}

	for command, description := range commands {
		if !isCommandAvailable(command) {
			missingDeps = append(missingDeps, command)
			log.Printf("警告: 未找到依赖命令 '%s' - %s\n", command, description)
		} else {
			log.Printf("成功: 找到依赖命令 '%s'\n", command)
		}
	}

	// 检查是否有缺失的依赖
	if len(missingDeps) > 0 {
		log.Fatalf("缺少以下必要依赖: %v,请安装后再运行程序.\n", missingDeps)
	} else {
		log.Println("所有必要依赖均已找到,程序可以正常运行.")
	}
}

// isCommandAvailable 检查指定的命令是否在系统中可用
func isCommandAvailable(name string) bool {
	// 在Windows系统上,需要添加.exe扩展名
	if _, err := exec.LookPath(name); err != nil {
		return false
	}
	return true
}

func diffSize(src, dst string) {
	s := new(sqlite.Save)
	s.Sync()
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
	if err := s.Insert(); err != nil {
		log.Printf("记录文件大小信息到数据库失败: %v\n", err)
	}
}
