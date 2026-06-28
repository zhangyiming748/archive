package util

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

/*
执行命令并实时输出命令的输出
实时打印命令本身的输出(包括标准输出和错误输出)
*/
func ExecuteCommandWithRealtimeOutput(cmd *exec.Cmd) error {
	// 将标准输出和标准错误直接连接到当前进程的输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 启动并等待命令执行完成
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// ExecCommand 执行命令并实时输出命令执行结果
// 通过管道获取命令的标准输出，并在执行过程中实时打印输出内容
// 参数:
//   - c: 要执行的命令对象
//
// 返回值:
//   - error: 执行过程中的错误，如果执行成功则返回nil
func ExecCommand(c *exec.Cmd) (e error) {
	log.Printf("开始执行命令:%v\n", c.String())
	// 获取命令的标准输出管道
	stdout, err := c.StdoutPipe()
	// 将标准错误重定向到标准输出
	c.Stderr = c.Stdout
	if err != nil {
		log.Fatalf("连接Stdout产生错误:%v\n", err)
		return err
	}
	// 启动命令
	if err = c.Start(); err != nil {
		log.Fatalf("启动cmd命令产生错误:%v\n", err)
		return err
	}
	// 循环读取并打印命令输出
	for {
		tmp := make([]byte, 1024)
		_, e := stdout.Read(tmp)
		t := string(tmp)
		// 清理输出中的空字符
		t = strings.Replace(t, "\u0000", "", -1)
		fmt.Println(t)
		if e != nil {
			break
		}
	}
	// 等待命令执行完成
	if err = c.Wait(); err != nil {
		log.Fatalf("命令执行中产生错误:%v\n", err)
		return err
	}
	return nil
}

// ExecCommandWithBar 执行命令并显示进度条
// 在执行过程中通过解析输出获取处理帧数，并更新进度条显示
// 参数:
//   - c: 要执行的命令对象
//   - totalFrame: 总帧数，用于初始化进度条
//
// 返回值:
//   - error: 执行过程中的错误，如果执行成功则返回nil
func ExecCommandWithBar(c *exec.Cmd, totalFrame int) (e error) {
	log.Printf("开始执行命令:%v\n", c.String())
	// 将总帧数转换为整数并创建进度条
	total := totalFrame
	if total <= 0 {
		log.Printf("总帧数为%d，无法显示进度条，使用普通执行方式\n", total)
		return ExecuteCommandWithRealtimeOutput(c)
	}
	log.Printf("总帧数为: %d，开始创建进度条\n", total)

	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription("处理进度"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true), // 启用预估时间
		progressbar.OptionFullWidth(),
		progressbar.OptionThrottle(65*time.Millisecond),
	)
	defer bar.Finish()

	// 启动一个后台 goroutine 定期刷新进度条
	stopRefresh := make(chan bool)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopRefresh:
				return
			case <-ticker.C:
				bar.RenderBlank()
			}
		}
	}()
	defer close(stopRefresh)

	// 获取命令的标准错误管道（ffmpeg的进度信息输出到stderr）
	stderr, err := c.StderrPipe()
	if err != nil {
		log.Printf("连接Stderr产生错误:%v\n", err)
		return err
	}
	// 将标准输出重定向到当前进程输出
	c.Stdout = os.Stdout

	// 启动命令
	if err = c.Start(); err != nil {
		log.Printf("启动cmd命令产生错误:%v\n", err)
		return err
	}
	log.Printf("命令已启动，开始读取输出...\n")

	// 使用缓冲区累积读取输出，避免截断帧数信息
	buf := make([]byte, 4096)
	var leftover string
	lineCount := 0
	frameFoundCount := 0

	// 循环读取输出并更新进度条
	for {
		n, e := stderr.Read(buf)
		if n > 0 {
			// 将新读取的数据与之前的剩余数据合并
			output := leftover + string(buf[:n])
			log.Printf("[调试] 读取到 %d 字节数据\n", n)

			// 按行分割处理
			lines := strings.Split(output, "\n")
			log.Printf("[调试] 分割为 %d 行\n", len(lines))

			// 最后一行可能不完整，留到下次处理
			leftover = lines[len(lines)-1]

			// 处理完整的行
			for i := 0; i < len(lines)-1; i++ {
				line := lines[i]
				lineCount++
				// 清理输出中的空字符
				line = strings.Replace(line, "\u0000", "", -1)

				// 调试：打印前20行和所有包含frame的行
				if lineCount <= 20 || strings.Contains(strings.ToLower(line), "frame") {
					log.Printf("[调试#%d] %s\n", lineCount, line)
				}

				// 从输出中提取当前帧数并更新进度条
				if frame, err := GetFrameNum(line); err == nil {
					frameFoundCount++
					log.Printf("[帧数匹配] 第%d次匹配到帧数: %d (总帧数: %d, 进度: %.2f%%)\n", frameFoundCount, frame, total, float64(frame)/float64(total)*100)
					bar.Set(frame)
					// 强制刷新进度条显示
					bar.RenderBlank()
				}
			}
		} else {
			log.Printf("[调试] 读取到 0 字节，错误: %v\n", e)
		}

		if e != nil {
			log.Printf("读取stderr结束: %v, 共读取%d行，匹配到%d次帧数\n", e, lineCount, frameFoundCount)
			break
		}
	}

	// 等待命令执行完成
	if err = c.Wait(); err != nil {
		log.Printf("命令执行中产生错误:%v\n", err)
		return err
	}
	bar.Finish()
	log.Printf("命令结束:%v\n", c.String())
	return nil
}

// GetFrameNum 从ffmpeg输出中提取当前处理的帧数
// 使用正则表达式匹配输出中的frame字段
// 参数:
//   - s: ffmpeg命令的输出字符串
//
// 返回值:
//   - int: 提取到的帧数
//   - error: 提取过程中的错误，如果提取成功则返回nil
func GetFrameNum(s string) (int, error) {
	// 使用正则表达式匹配frame=后面的数字
	re := regexp.MustCompile(`frame=\s*(\d+)`)
	matches := re.FindStringSubmatch(s)
	// 如果匹配成功，将匹配到的数字转换为整数返回
	if len(matches) > 1 {
		frameNumber := matches[1]
		frame, _ := strconv.Atoi(frameNumber)
		return frame, nil
	} else {
		return 0, errors.New("not found")
	}
}
