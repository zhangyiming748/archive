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

	"github.com/zhangyiming748/FastMediaInfo"
	"github.com/zhangyiming748/stand"
)

// 转换mkv文件为h265格式,但保留全部的音频轨、字幕轨
// ffmpeg -i .\天将雄狮.Dragon.Blade.2015.BluRay.1080p.x265.10bit.MNHD-FRDS.mkv -map 0 -c:v libx265 -c:a aac -tag:v hvc1 -c:s copy 天将雄狮.mkv
func ConvertMKV2H265(src string, fhd bool) {
	mi := FastMediaInfo.GetStandMediaInfo(src)
	vInfo := mi.Video
	var cmd *exec.Cmd
	args := []string{"-i", src}
	dst := strings.Replace(src, filepath.Ext(src), "_tmp.mkv", 1)

	log.Printf("处理视频文件:%s\n", src)
	args = append(args, "-map", "0")
	args = append(args, "-c:v", "libx265")
	args = append(args, "-tag:v", "hvc1")
	args = append(args, "-c:a", "aac")
	args = append(args, "-c:s", "copy")
	if fhd {
		if overFHD(vInfo) {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	}
	args = append(args, dst)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("开始执行命令:%s\n", cmd.String())
	err := stand.ExecCommandWithBar(cmd, src)
	if err != nil {
		log.Printf("转换失败：%v\n", err)
		return
	}

	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(dst)
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
最终转换视频文件为带hvc1标签的MP4文件 mkv除外
*/
func Convert2H265(src string, fhd, force bool) {
	if strings.ToLower(filepath.Ext(src)) == ".mkv" {
		log.Printf("检测到mkv文件:%s,使用mkv逻辑单独处理", src)
		CloneMkv2H265(src)
		return
	}
	mi := FastMediaInfo.GetStandMediaInfo(src)
	vInfo := mi.Video
	var cmd *exec.Cmd
	args := []string{"-i", src}
	dst := strings.Replace(src, filepath.Ext(src), "_tmp.mp4", 1)

	// 优先检查分辨率是否需要转换
	needsResize := fhd && overFHD(vInfo)

	if isH265(vInfo) && filepath.Ext(src) == ".mp4" {
		if hasTag(vInfo) && !needsResize {
			log.Printf("跳过已经是h265编码并且带有hvc1标签且分辨率符合要求的视频文件:%s\n", src)
			return
		}
		if needsResize {
			log.Printf("处理HEVC编码带有hvc1标签但分辨率超标的视频文件:%s\n", src)
		} else {
			log.Printf("处理HEVC编码但是不带有hvc1标签的视频文件:%s\n", src)
		}
		args = append(args, "-c:v", "copy", "-c:a", "copy", "-tag:v", "hvc1")
		if needsResize {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	} else {
		log.Printf("处理不是HEVC编码的视频文件:%s\n", src)
		args = append(args, "-c:v", "libx265")
		args = append(args, "-tag:v", "hvc1")
		args = append(args, "-c:a", "aac")
		args = append(args, "-preset", "slow")
		args = append(args, "-crf", "24") // H.265的CRF 24在画质和大小之间取得良好平衡
		// 根据源视频位深自动选择像素格式，避免不必要的10-bit转换
		args = append(args, "-pix_fmt", "yuv420p")
		args = append(args, "-x265-params", "aq-mode=3:aq-strength=1.2:psy-rd=2.0:psy-rdoq=2.0:rdoq-level=1") // 优化的心理视觉参数
		if needsResize {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	}
	args = append(args, "-map_chapters", "-1")
	if force {
		args = append(args, "-y")
		log.Printf("强制覆盖输出文件:%s\n", dst)
	}
	args = append(args, dst)
	cmd = exec.Command("ffmpeg", args...)

	// 打印视频信息用于调试
	log.Printf("视频信息 - 格式:%s, 编码:%s, 分辨率:%sx%s, 帧率:%s, 帧数:%s\n",
		vInfo.Format, vInfo.CodecID, vInfo.Width, vInfo.Height, vInfo.FrameRate, vInfo.FrameCount)

	// 直接使用 ExecCommandWithBar 执行命令
	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("转换失败：%v\n", err)
		return
	}

	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(dst)

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

/*
最终转换视频文件为带hvc1标签的MP4文件
*/
func Convert2H265MP4(src string, fhd, force bool) {
	mi := FastMediaInfo.GetStandMediaInfo(src)
	vInfo := mi.Video
	var cmd *exec.Cmd
	args := []string{"-i", src}
	dst := strings.Replace(src, filepath.Ext(src), "_tmp.mp4", 1)

	// 优先检查分辨率是否需要转换
	needsResize := fhd && overFHD(vInfo)

	if isH265(vInfo) && filepath.Ext(src) == ".mp4" {
		if hasTag(vInfo) && !needsResize {
			log.Printf("跳过已经是h265编码并且带有hvc1标签且分辨率符合要求的视频文件:%s\n", src)
			return
		}
		if needsResize {
			log.Printf("处理HEVC编码带有hvc1标签但分辨率超标的视频文件:%s\n", src)
		} else {
			log.Printf("处理HEVC编码但是不带有hvc1标签的视频文件:%s\n", src)
		}
		args = append(args, "-c:v", "copy", "-c:a", "copy", "-tag:v", "hvc1")
		if needsResize {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	} else {
		log.Printf("处理不是HEVC编码的视频文件:%s\n", src)
		args = append(args, "-c:v", "libx265")
		args = append(args, "-tag:v", "hvc1")
		args = append(args, "-c:a", "aac")
		args = append(args, "-preset", "slow")
		args = append(args, "-crf", "24") // H.265的CRF 24在画质和大小之间取得良好平衡
		// 根据源视频位深自动选择像素格式，避免不必要的10-bit转换
		args = append(args, "-pix_fmt", "yuv420p")
		args = append(args, "-x265-params", "aq-mode=3:aq-strength=1.2:psy-rd=2.0:psy-rdoq=2.0:rdoq-level=1") // 优化的心理视觉参数
		if needsResize {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	}
	args = append(args, "-map_chapters", "-1")
	if force {
		args = append(args, "-y")
		log.Printf("强制覆盖输出文件:%s\n", dst)
	}
	args = append(args, dst)
	cmd = exec.Command("ffmpeg", args...)

	// 打印视频信息用于调试
	log.Printf("视频信息 - 格式:%s, 编码:%s, 分辨率:%sx%s, 帧率:%s, 帧数:%s\n",
		vInfo.Format, vInfo.CodecID, vInfo.Width, vInfo.Height, vInfo.FrameRate, vInfo.FrameCount)

	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("转换失败：%v\n", err)
		return
	}

	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(dst)

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
	} else {
		log.Printf("视频格式为:%s\n", vInfo.Format)
	}
	return false

}
func hasTag(vInfo FastMediaInfo.Video) bool {
	if vInfo.CodecID == "hvc1" {
		return true
	}
	return false

}

/*
最终转换视频文件为带hvc1标签的MP4文件
*/
func Convert2SmallerH265MP4(src string, fhd, force bool) {
	mi := FastMediaInfo.GetStandMediaInfo(src)
	vInfo := mi.Video
	var cmd *exec.Cmd
	args := []string{"-i", src}
	dst := strings.Replace(src, filepath.Ext(src), "_tmp.mp4", 1)

	// 优先检查分辨率是否需要转换
	needsResize := fhd && overFHD(vInfo)

	if isH265(vInfo) && filepath.Ext(src) == ".mp4" {
		if hasTag(vInfo) && !needsResize {
			log.Printf("跳过已经是h265编码并且带有hvc1标签且分辨率符合要求的视频文件:%s\n", src)
			return
		}
		if needsResize {
			log.Printf("处理HEVC编码带有hvc1标签但分辨率超标的视频文件:%s\n", src)
		} else {
			log.Printf("处理HEVC编码但是不带有hvc1标签的视频文件:%s\n", src)
		}
		args = append(args, "-c:v", "copy", "-c:a", "copy", "-tag:v", "hvc1")
		if needsResize {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	} else {
		log.Printf("处理不是HEVC编码的视频文件:%s\n", src)
		args = append(args, "-c:v", "libx265")
		args = append(args, "-tag:v", "hvc1")
		args = append(args, "-c:a", "aac")
		args = append(args, "-preset", "fast")
		args = append(args, "-crf", "28") // H.265的CRF 24在画质和大小之间取得良好平衡
		// 根据源视频位深自动选择像素格式，避免不必要的10-bit转换
		args = append(args, "-pix_fmt", "yuv420p")
		if needsResize {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	}
	args = append(args, "-map_chapters", "-1")
	if force {
		args = append(args, "-y")
		log.Printf("强制覆盖输出文件:%s\n", dst)
	}
	args = append(args, dst)
	cmd = exec.Command("ffmpeg", args...)

	// 打印视频信息用于调试
	log.Printf("视频信息 - 格式:%s, 编码:%s, 分辨率:%sx%s, 帧率:%s, 帧数:%s\n",
		vInfo.Format, vInfo.CodecID, vInfo.Width, vInfo.Height, vInfo.FrameRate, vInfo.FrameCount)

	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("转换失败：%v\n", err)
		return
	}

	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(dst)

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

func overFHD(vInfo FastMediaInfo.Video) bool {
	height, err := strconv.Atoi(vInfo.Height)
	if err != nil {
		return false
	}
	width, err := strconv.Atoi(vInfo.Width)
	if err != nil {
		return false
	}

	if height > 1920 || width > 1920 {
		log.Printf("高度为%s,宽度为%s\n", vInfo.Height, vInfo.Width)
		return true
	}
	return false
}

// checkOutputFileValid 检查输出文件是否有效（非0字节）
// 如果文件大小为0，说明FFmpeg实际执行失败但未返回错误码
func checkOutputFileValid(dst string) {
	dstInfo, err := os.Stat(dst)
	if err != nil {
		log.Fatalf("无法获取目标文件信息：%v\n", err)
	}
	if dstInfo.Size() == 0 {
		log.Printf("错误：转换后的文件大小为0字节，说明FFmpeg执行失败但未返回错误码\n")
		log.Printf("目标文件: %s\n", dst)
		panic(fmt.Sprintf("FFmpeg转换失败：输出文件 %s 大小为0字节", dst))
	}
}

/*
最终转换视频文件为aac音频全部字幕流带hvc1标签的MKV文件
*/
func CloneMkv2H265(src string) {
	if strings.ToLower(filepath.Ext(src)) != ".mkv" {
		log.Printf("文件格式不是mkv,请检查文件:%s\n", src)
		return
	}
	mi := FastMediaInfo.GetStandMediaInfo(src)
	vInfo := mi.Video
	var cmd *exec.Cmd
	args := []string{"-i", src}
	dst := strings.Replace(src, filepath.Ext(src), "_tmp.mkv", 1)

	// 优先检查分辨率是否需要转换（MKV默认启用FHD检查）
	needsResize := overFHD(vInfo)

	if isH265(vInfo) && strings.ToLower(filepath.Ext(src)) == ".mkv" {
		if hasTag(vInfo) && !needsResize {
			log.Printf("跳过已经是h265编码并且带有hvc1标签且分辨率符合要求的mkv视频文件:%s\n", src)
			return
		}
		if needsResize {
			log.Printf("处理HEVC编码带有hvc1标签但分辨率超标的mkv视频文件:%s\n", src)
		} else {
			log.Printf("处理HEVC编码但是不带有hvc1标签的视频文件:%s\n", src)
		}
		args = append(args, "-map", "0", "-c:v", "copy", "-c:a", "copy", "-c:s", "copy", "-tag:v", "hvc1")
		if needsResize {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	} else {
		log.Printf("处理不是HEVC编码的视频文件:%s\n", src)
		args = append(args, "-map", "0", "-c:v", "libx265", "-c:a", "aac", "-tag:v", "hvc1")
		if needsResize {
			args = append(args, "-vf", "scale=if(gt(iw\\,ih)\\,iw*1080/ih\\,1920):if(gt(iw\\,ih)\\,1080\\,ih*1920/iw):-2")
		}
	}
	args = append(args, "-c:a", "aac")
	args = append(args, "-map_chapters", "-1")
	args = append(args, dst)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("开始执行命令:%s\n", cmd.String())
	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("转换失败：%v\n", err)
		return
	}

	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(dst)

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
*
快速转换视频文件为标准H264(avc)视频
*/
func FastConvertVideo2StandAvc(src string) {
	mi := FastMediaInfo.GetStandMediaInfo(src)
	if mi.Video.Format == "AVC" || mi.Video.Format == "HEVC" {
		if strings.ToLower(filepath.Ext(src)) != ".mp4" {
			log.Printf("文件已经转换过:%s\n", src)
			return
		}
	}

	tmp_name := strings.Replace(src, filepath.Ext(src), "_fast.mp4", 1)
	cmd := exec.Command("ffmpeg", "-i", src, "-c:v", "libx264", "-c:a", "aac", "-map_chapters", "-1", tmp_name)
	log.Printf("开始执行命令:%s\n", cmd.String())
	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("转换失败：%v\n", err)
		return
	}
	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(tmp_name)
	if err := os.Remove(src); err != nil {
		log.Fatalf("删除源文件失败：%v\n", err)
	}
}

/*
快速转换错误(包括不是电影 不带特效字幕的视频)的mkv文件
*/
func FastConvertMkv(src string) {
	if strings.ToLower(filepath.Ext(src)) != ".mkv" {
		log.Printf("文件格式不是mkv,请检查文件:%s\n", src)
		return
	}
	var cmd *exec.Cmd
	mp4 := strings.Replace(src, filepath.Ext(src), ".mp4", 1)
	args := []string{"-i", src}
	args = append(args, "-c:v", "copy")
	args = append(args, "-c:a", "copy")
	args = append(args, "-map_chapters", "-1")
	args = append(args, mp4)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("开始执行命令:%s\n", cmd.String())
	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("转换失败：%v\n", err)
		return
	}
	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(mp4)
	if err := os.Remove(src); err != nil {
		log.Fatalf("删除源文件失败：%v\n", err)
	}
}
func MergeMp4WithSameNameSrt(video, srt string) error {
	//ffmpeg -i input.mp4 -vf "subtitles=subtitle.srt" output.mp4
	var cmd *exec.Cmd
	args := []string{"-i", video}
	args = append(args, "-vf", "subtitles="+srt)
	args = append(args, "-c:v", "libx265")
	args = append(args, "-tag:v", "hvc1")
	args = append(args, "-c:a", "aac")
	output := strings.Replace(srt, filepath.Ext(srt), "_subInside.mp4", 1)
	args = append(args, output)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("当前生成的内嵌字幕的命令是:%v\n", cmd.String())
	if err := stand.ExecCommandWithBar(cmd, video); err != nil {
		log.Printf("内嵌字幕失败：%v\n", err)
		return err
	}
	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(output)
	return nil
}

const (
	ToRight = "ClockWise90"
	ToLeft  = "ClockWise270"
)

func RotateVideo(src string, direction string) {
	var (
		cmd  *exec.Cmd
		args []string
	)
	tmp_name := strings.Replace(src, filepath.Ext(src), "_rotate.mp4", 1)
	args = append(args, "-i", src)
	switch direction {
	case ToRight:
		args = append(args, "-vf", "transpose=1")
	case ToLeft:
		args = append(args, "-vf", "transpose=2")
	default:
		log.Printf("请输入正确的旋转方向:%s\n", direction)
		return
	}
	if hasNvidia() {
		// NVIDIA NVENC: 使用高质量预设，CRF模式保持画质
		args = append(args, "-c:v", "h264_nvenc")
		args = append(args, "-preset", "p5") // slow预设，质量与速度平衡
		args = append(args, "-rc", "vbr")    // 可变比特率
		args = append(args, "-cq", "18")     // 恒定质量等级，18为高质量
		args = append(args, "-b:v", "0")     // 不限制最大比特率
	} else if hasIntel() {
		// Intel QSV: 使用ICQ模式获得最佳质量
		args = append(args, "-c:v", "h264_qsv")
		args = append(args, "-global_quality", "18")   // ICQ质量等级，18为高质量
		args = append(args, "-look_ahead", "1")        // 启用前瞻分析
		args = append(args, "-look_ahead_depth", "40") // 前瞻深度
	} else if hasAMD() {
		// AMD AMF: 使用质量优先预设
		args = append(args, "-c:v", "h264_amf")
		args = append(args, "-quality", "quality") // 质量优先模式
		args = append(args, "-qp_i", "18")         // I帧量化参数
		args = append(args, "-qp_p", "20")         // P帧量化参数
		args = append(args, "-qp_b", "22")         // B帧量化参数
	} else {
		// CPU软件编码 libx264: 使用slow预设和CRF 18
		args = append(args, "-c:v", "libx264")
		args = append(args, "-preset", "slow")     // slow预设，质量与压缩率平衡
		args = append(args, "-crf", "18")          // CRF 18，视觉无损级别
		args = append(args, "-pix_fmt", "yuv420p") // 标准像素格式
	}
	args = append(args, "-tag:v", "avc1")
	args = append(args, "-c:a", "aac")
	args = append(args, "-map_chapters", "-1")
	args = append(args, tmp_name)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("开始执行命令:%s\n", cmd.String())
	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("旋转失败：%v\n", err)
		return
	}
	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(tmp_name)
	/*
		1. 删除旧文件
		2. 临时文件改为旧文件的文件名
	*/
	if err := os.Remove(src); err != nil {
		log.Printf("删除源文件失败：%v\n", err)
	} else {
		if err := os.Rename(tmp_name, strings.Replace(src, filepath.Ext(src), ".mp4", 1)); err != nil {
			log.Printf("重命名文件失败：%v\n", err)
		}
	}
}

/*
视频文件提取为同名aac音频
*/func ExtractAudioFromVideo(src string) {
	var cmd *exec.Cmd
	args := []string{"-i", src}
	args = append(args, "-c:a", "aac")
	args = append(args, "-vn")
	aacFile := strings.Replace(src, filepath.Ext(src), ".aac", 1)
	args = append(args, aacFile)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("开始执行命令:%s\n", cmd.String())
	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("提取音频失败：%v\n", err)
		return
	}
	// 检查提取的音频文件是否为0字节
	checkOutputFileValid(aacFile)
	os.Remove(src)
}

/*
专门为大疆直接转换tf卡上的视频设计的方法
*/
func DjiVideoConvert(src, dst string) {
	var (
		cmd  *exec.Cmd
		args []string
	)
	args = append(args, "-i", src)
	args = append(args, "-c:v", "libx265")
	args = append(args, "-tag:v", "hvc1")
	args = append(args, "-c:a", "aac")
	args = append(args, "-map_chapters", "-1")
	args = append(args, dst)
	cmd = exec.Command("ffmpeg", args...)
	log.Printf("开始执行命令:%s\n", cmd.String())
	if err := stand.ExecCommandWithBar(cmd, src); err != nil {
		log.Printf("转换失败：%v\n", err)
		return
	}
	// 检查转换后的文件是否为0字节，如果是则说明FFmpeg实际执行失败
	checkOutputFileValid(dst)
}
func hasNvidia() bool {
	// 检查FFmpeg是否支持NVIDIA NVENC H.264编码器
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// 查找h264_nvenc编码器
	return strings.Contains(string(output), "h264_nvenc")
}

func hasIntel() bool {
	// 检查FFmpeg是否支持Intel QSV H.264编码器
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// 查找h264_qsv编码器
	return strings.Contains(string(output), "h264_qsv")
}

func hasAMD() bool {
	// 检查FFmpeg是否支持AMD VCE H.264编码器
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// 查找h264_amf编码器
	return strings.Contains(string(output), "h264_amf")
}
