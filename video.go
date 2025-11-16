package archive

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/zhangyiming748/FastMediaInfo"
)

// 转换mkv文件为h265格式,但保留全部的音频轨、字幕轨
// ffmpeg -i .\天将雄狮.Dragon.Blade.2015.BluRay.1080p.x265.10bit.MNHD-FRDS.mkv -map 0 -c:v libx265 -c:a aac -tag:v hvc1 -c:s copy 天将雄狮.mkv
func ConvertMKV2H265(src string) {
	mi := FastMediaInfo.GetStandMediaInfo(src)
	vInfo := mi.Video
	var cmd *exec.Cmd
	args := []string{"-i", src}
	if runtime.GOARCH == "arm64" && runtime.GOOS == "linux" {
		args = append(args, "-threads", "1")
	}
	purgePath := filepath.Dir(src)
	seed := rand.New(rand.NewSource(time.Now().Unix()))
	b := seed.Intn(2000) + 1000
	tmp := strconv.Itoa(b)
	tmp = strings.Join([]string{tmp, ".mkv"}, "")
	dst := filepath.Join(purgePath, tmp)

	log.Printf("处理视频文件:%s\n", src)
	args = append(args, "-map", "0")
	args = append(args, "-c:v", "libx265")
	args = append(args, "-tag:v", "hvc1")
	args = append(args, "-c:a", "aac")
	args = append(args, "-c:s", "copy")
	if overFHD(vInfo) {
		args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw)")
	}
	args = append(args, dst)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("开始执行命令:%s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("转换失败：%v\n输出内容%s\n", err, string(out))
		return
	}
	fmt.Printf("转换成功：%s\n", string(out))

	//在这里添加一个功能，判断源文件和转换后的文件大小，源文件通常会大于转换后的文件所以用源文件的大小减去目标文件大小，之后用fmt.Sprintf打印出差值，单位为MB，保留三位小数
	diffSize(src, dst)
	// 先尝试删除源文件
	if err := os.Remove(src); err != nil {
		log.Printf("删除源文件失败：%v\t尝试重命名源文件，添加 should_be_deleted\n", err)
		//尝试重命名源文件，添加 should_be_deleted
		nName := strings.Replace(src, filepath.Ext(src), ".should_be_deleted", 1)
		if err := os.Rename(src, nName); err != nil {
			log.Fatalf("重命名文件失败：%v\n", err)
		}
	}
	// 源文件删除成功后，等待短暂时间确保文件句柄完全释放
	time.Sleep(100 * time.Millisecond)
	// 尝试重命名
	src = strings.Replace(src, filepath.Ext(src), ".mkv", 1)
	if err := os.Rename(dst, src); err != nil {
		log.Fatalf("重命名文件失败：%v\n", err)
	}
}

/*
最终转换视频文件为带hvc1标签的MP4文件
*/
func Convert2H265(src string) {
	mi := FastMediaInfo.GetStandMediaInfo(src)
	vInfo := mi.Video
	var cmd *exec.Cmd
	args := []string{"-i", src}
	if runtime.GOARCH == "arm64" && runtime.GOOS == "linux" {
		args = append(args, "-threads", "1")
	}
	purgePath := filepath.Dir(src)
	seed := rand.New(rand.NewSource(time.Now().Unix()))
	b := seed.Intn(2000) + 1000
	tmp := strconv.Itoa(b)
	tmp = strings.Join([]string{tmp, ".mp4"}, "")
	dst := filepath.Join(purgePath, tmp)
	if isH265(vInfo) && filepath.Ext(src) == ".mp4" {
		if hasTag(vInfo) {
			log.Printf("跳过已经是h265编码并且带有hvc1标签的视频文件:%s\n", src)
			return
		}
		log.Printf("处理HEVC编码但是不带有hvc1标签的视频文件:%s\n", src)
		args = append(args, "-c:v", "copy", "-c:a", "copy", "-tag:v", "hvc1")

	} else {
		log.Printf("处理不是HEVC编码的视频文件:%s\n", src)
		args = append(args, "-c:v", "libx265", "-c:a", "aac", "-tag:v", "hvc1")
		if overFHD(vInfo) {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw)")
		}
	}
	args = append(args, "-c:a", "aac")
	args = append(args, "-map_chapters", "-1")
	args = append(args, dst)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("开始执行命令:%s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("转换失败：%v\n输出内容%s\n", err, string(out))
		return
	}
	fmt.Printf("转换成功：%s\n", string(out))

	//在这里添加一个功能，判断源文件和转换后的文件大小，源文件通常会大于转换后的文件所以用源文件的大小减去目标文件大小，之后用fmt.Sprintf打印出差值，单位为MB，保留三位小数
	diffSize(src, dst)
	// 先尝试删除源文件
	if err := os.Remove(src); err != nil {
		log.Printf("删除源文件失败：%v\t尝试重命名源文件，添加 should_be_deleted\n", err)
		//尝试重命名源文件，添加 should_be_deleted
		nName := strings.Replace(src, filepath.Ext(src), ".should_be_deleted", 1)
		if err := os.Rename(src, nName); err != nil {
			log.Fatalf("重命名文件失败：%v\n", err)
		}
	}
	// 源文件删除成功后，等待短暂时间确保文件句柄完全释放
	time.Sleep(100 * time.Millisecond)
	// 尝试重命名
	src = strings.Replace(src, filepath.Ext(src), ".mp4", 1)
	if err := os.Rename(dst, src); err != nil {
		log.Fatalf("重命名文件失败：%v\n", err)
	}
}
func isH265(vInfo FastMediaInfo.Video) bool {
	if vInfo.Format == "HEVC" {
		return true
	}
	return false

}
func hasTag(vInfo FastMediaInfo.Video) bool {
	if vInfo.CodecID == "hvc1" {
		return true
	}
	return false

}
func overFHD(vInfo FastMediaInfo.Video) bool {
	height, _ := strconv.Atoi(vInfo.Height)
	width, _ := strconv.Atoi(vInfo.Width)
	if height > 1920 || width > 1920 {
		log.Printf("高度为%s,宽度为%s\n", vInfo.Height, vInfo.Width)
		return true
	}
	return false

}
