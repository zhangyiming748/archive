package util

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

/*
执行命令并实时输出命令的输出
逐行打印命令本身的输出(包括标准输出和错误输出)
*/
func ExecuteCommandWithRealtimeOutput(cmd *exec.Cmd) error {
	// 获取标准输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// 获取标准错误管道
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return err
	}

	// 用于缓冲区累积读取输出
	var stdoutLeftover string
	var stderrLeftover string

	// 读取标准输出并逐行打印
	doneStdout := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, e := stdout.Read(buf)
			if n > 0 {
				output := stdoutLeftover + string(buf[:n])
				lines := strings.Split(output, "\n")
				stdoutLeftover = lines[len(lines)-1]

				for i := 0; i < len(lines)-1; i++ {
					line := strings.Replace(lines[i], "\u0000", "", -1)
					fmt.Printf("标准输出:%s\n", line)
				}
			}
			if e != nil {
				break
			}
		}
		doneStdout <- true
	}()

	// 读取标准错误并逐行打印
	doneStderr := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, e := stderr.Read(buf)
			if n > 0 {
				output := stderrLeftover + string(buf[:n])
				lines := strings.Split(output, "\n")
				stderrLeftover = lines[len(lines)-1]

				for i := 0; i < len(lines)-1; i++ {
					line := strings.Replace(lines[i], "\u0000", "", -1)
					fmt.Printf("标准错误:%s\n", line)
				}
			}
			if e != nil {
				break
			}
		}
		doneStderr <- true
	}()

	// 等待两个 goroutine 完成
	<-doneStdout
	<-doneStderr

	// 等待命令执行完成
	return cmd.Wait()
}

// GetFrameNum 从ffmpeg输出中提取当前处理的帧数
// 使用正则表达式匹配输出中的frame字段
// 参数:
//   - s: ffmpeg命令的输出字符串
//
// 返回值:
//   - int: 提取到的帧数
//   - error: 提取过程中的错误,如果提取成功则返回nil
func GetFrameNum(s string) (int, error) {
	// 使用正则表达式匹配frame=后面的数字
	re := regexp.MustCompile(`frame=\s*(\d+)`)
	matches := re.FindStringSubmatch(s)
	// 如果匹配成功,将匹配到的数字转换为整数返回
	if len(matches) > 1 {
		frameNumber := matches[1]
		frame, _ := strconv.Atoi(frameNumber)
		return frame, nil
	} else {
		return 0, errors.New("not found")
	}
}
