// Package archive 提供音频文件处理和转换的功能，包括音频文件的检测、转换和存档
package archive

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 音频处理相关的常量定义
const (
	// AudioBookType 有声书类型标识
	AudioBookType = "audiobook"
	// RapMusicType 说唱音乐类型标识
	RapMusicType = "rap"
	// Speed        = "1.54" //等效audition的65%
	// Speed = "1.43" 音频播放速度，等效audition的70%
	// Volume 音频音量增益值
	Volume = "2.7"
)

/*
ConvertAudio 转换音频文件
src 为源文件路径
mytype 为音频类型，决定处理方式
*/
func ConvertAudio(src, mytype string) {
	// 生成临时文件路径
	dst := strings.Replace(src, filepath.Ext(src), "_tmp.mp3", 1)

	// 构建ffmpeg命令参数
	args := []string{"-i", src}
	ff := audition2ffmpeg("65")
	atempo := strings.Join([]string{"atempo", ff}, "=")
	volume := strings.Join([]string{"volume", Volume}, "=")
	filter := strings.Join([]string{atempo, volume}, ",")

	args = append(args, "-ac", "1")
	args = append(args, "-map_metadata", "-1")
	args = append(args, "-ar", "44100")
	args = append(args, "-ab", "128k")
	// 根据音频类型设置不同的处理参数
	switch mytype {
	case AudioBookType:
		// 有声书加速65% 电平增加
		args = append(args, "-filter:a", filter)
	// 歌曲类只增加电平
	case RapMusicType:
		args = append(args, "-filter:a", volume)
	default:
		// 其他类型
		args = append(args, "-c:a", "aac")
	}
	args = append(args, dst)
	cmd := exec.Command("ffmpeg", args...)

	// 获取输出和错误管道

	// 等待命令完成并处理结果
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("转换失败：%v\n", err)
	} else {
		fmt.Printf("转换成功：%s\n", string(out))
		// 先尝试删除源文件
		if err := os.Remove(src); err != nil {
			log.Fatalf("删除源文件失败：%v\n", err)
		}
		// 源文件删除成功后，等待短暂时间确保文件句柄完全释放
		time.Sleep(100 * time.Millisecond)
		// 尝试重命名
		if err := os.Rename(dst, src); err != nil {
			log.Fatalf("重命名文件失败：%v\n", err)
		}
	}
}

// audition2ffmpeg 将Adobe Audition的速度参数转换为ffmpeg的速度参数
// speed 为输入的速度参数
// 返回转换后的ffmpeg速度参数
func audition2ffmpeg(speed string) string {
	audition, err := strconv.ParseFloat(speed, 64)
	if err != nil {
		log.Fatalf("解析加速参数错误:%v,退出程序", err)
	}
	param := 100 / audition
	log.Printf("转换后的原始参数:%v\n", param)
	final := fmt.Sprintf("%.2f", param)
	log.Printf("保留两位小数的原始参数:%v\n", final)
	return final
}

func Convert2Mp3(src string) {
	if strings.HasSuffix(src, ".mp3") {
		log.Printf("文件已经是MP3格式,无需转换:%v\n", src)
		return
	}
	//转换给定的文件为高品质MP3编码的MP3文件（320kbps，人耳无损）
	mp3 := strings.Replace(src, filepath.Ext(src), ".mp3", 1)
	var (
		args []string
		cmd  *exec.Cmd
	)
	args = append(args, "-i", src)
	args = append(args, "-c:a", "libmp3lame")
	args = append(args, "-b:a", "320k") // 比特率320kbps，最高音质
	args = append(args, "-q:a", "0")    // LAME编码器最高质量等级
	args = append(args, "-ar", "44100") // 采样率44.1kHz（CD音质）
	args = append(args, mp3)
	cmd = exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("转换失败：%v\n", err)
	} else {
		fmt.Printf("转换成功：%s\n", string(out))
		if err := os.Remove(src); err != nil {
			log.Fatalf("删除源文件失败：%v\n", err)
		}
	}
}

func Convert2Aac(src string) {
	if strings.HasSuffix(src, ".aac") {
		log.Printf("文件已经是AAC格式,无需转换:%v\n", src)
		return
	}
	//转换给定的文件为高品质AAC编码的M4A文件（256kbps，接近无损）
	aac := strings.Replace(src, filepath.Ext(src), ".m4a", 1)
	var (
		args []string
		cmd  *exec.Cmd
	)
	args = append(args, "-i", src)
	args = append(args, "-c:a", "aac")             // 使用AAC编码器
	args = append(args, "-b:a", "256k")            // 比特率256kbps，高品质
	args = append(args, "-ar", "44100")            // 采样率44.1kHz（CD音质）
	args = append(args, "-profile:a", "aac_he_v2") // 使用HE-AAC v2配置，更高效率
	args = append(args, aac)
	cmd = exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("转换失败：%v\n", err)
	} else {
		fmt.Printf("转换成功：%s\n", string(out))
		if err := os.Remove(src); err != nil {
			log.Fatalf("删除源文件失败：%v\n", err)
		}
	}
}
