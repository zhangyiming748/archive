package util

import (
	"os"
	"os/exec"
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
