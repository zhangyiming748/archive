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

	// 获取命令的标准输出和标准错误管道
	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Printf("连接Stdout产生错误:%v\n", err)
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		log.Printf("连接Stderr产生错误:%v\n", err)
		return err
	}

	// 用于捕获所有输出内容
	var stdoutBuffer strings.Builder
	var stderrBuffer strings.Builder

	// 启动命令
	if err = c.Start(); err != nil {
		log.Printf("启动cmd命令产生错误:%v\n", err)
		return err
	}
	log.Printf("命令已启动，开始读取输出...\n")

	// 启动 goroutine 读取 stdout
	doneStdout := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, e := stdout.Read(buf)
			if n > 0 {
				stdoutBuffer.Write(buf[:n])
				os.Stdout.Write(buf[:n]) // 同时输出到控制台
			}
			if e != nil {
				break
			}
		}
		doneStdout <- true
	}()

	// 使用缓冲区累积读取输出，避免截断帧数信息
	buf := make([]byte, 4096)
	var leftover string
	lineCount := 0
	frameFoundCount := 0

	// 循环读取 stderr 输出并更新进度条
	for {
		n, e := stderr.Read(buf)
		if n > 0 {
			// 保存到缓冲区
			stderrBuffer.Write(buf[:n])

			// 将新读取的数据与之前的剩余数据合并
			output := leftover + string(buf[:n])

			// 按行分割处理
			lines := strings.Split(output, "\n")

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

	// 等待 stdout 读取完成
	<-doneStdout

	// 等待命令执行完成
	waitErr := c.Wait()
	if waitErr != nil {
		log.Printf("命令执行中产生错误:%v\n", waitErr)
	}
	bar.Finish()

	// 等待一小段时间，确保文件完全写入磁盘
	time.Sleep(200 * time.Millisecond)

	// 检查输出文件是否为0字节（通过命令参数获取输出文件路径）
	if len(c.Args) > 0 {
		outputFile := c.Args[len(c.Args)-1]

		// 多次检查文件大小，避免因文件系统延迟导致的误判
		var fileSize int64 = -1
		for i := 0; i < 3; i++ {
			if fileInfo, statErr := os.Stat(outputFile); statErr == nil {
				fileSize = fileInfo.Size()
				log.Printf("[文件检查] 第%d次检查，文件大小: %d 字节\n", i+1, fileSize)
				if fileSize > 0 {
					break // 文件正常，退出检查循环
				}
				if i < 2 {
					log.Printf("[文件检查] 文件大小为0，等待500ms后重试...\n")
					time.Sleep(500 * time.Millisecond)
				}
			} else {
				log.Printf("[文件检查] 无法获取文件信息: %v\n", statErr)
				break
			}
		}

		if fileSize == 0 {
			log.Printf("\n========== FFmpeg 执行失败检测 ==========\n")
			log.Printf("错误：输出文件大小为0字节，说明FFmpeg实际执行失败\n")
			log.Printf("命令: %s\n", c.String())
			log.Printf("输出文件: %s\n", outputFile)
			log.Printf("文件大小: 0 字节\n")
			log.Printf("Wait() 返回错误: %v\n", waitErr)
			log.Printf("\n--- 标准输出 (stdout) ---\n%s\n", stdoutBuffer.String())
			log.Printf("\n--- 标准错误 (stderr) ---\n%s\n", stderrBuffer.String())
			log.Printf("=========================================\n")
			panic(fmt.Sprintf("FFmpeg转换失败：输出文件 %s 大小为0字节", outputFile))
		}
	}

	log.Printf("命令结束:%v\n", c.String())
	if waitErr != nil {
		return waitErr
	}
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
