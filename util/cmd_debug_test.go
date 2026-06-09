package util

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// TestExecCommandWithBarDebug 测试并打印ffmpeg输出以验证帧数解析
func TestExecCommandWithBarDebug(t *testing.T) {
	// 创建一个测试视频
	testVideo := "test_debug_input.mp4"
	cmd := exec.Command("ffmpeg", "-y",
		"-f", "lavfi",
		"-i", "color=c=red:s=640x480:d=2:r=25",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		testVideo)

	if err := cmd.Run(); err != nil {
		t.Skipf("无法创建测试视频: %v", err)
	}
	defer os.Remove(testVideo)

	outputVideo := "test_debug_output.mp4"
	defer os.Remove(outputVideo)

	// 使用普通方式执行,查看ffmpeg的实际输出
	cmd = exec.Command("ffmpeg", "-i", testVideo,
		"-c:v", "libx265",
		"-tag:v", "hvc1",
		"-c:a", "aac",
		"-preset", "ultrafast", // 最快预设
		outputVideo)

	t.Log("执行ffmpeg命令,捕获输出来验证frame=字段...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("转换失败: %v", err)
	}

	output := string(out)
	t.Logf("ffmpeg完整输出:\n%s", output)

	// 查找frame=字段
	frame, err := GetFrameNum(output)
	if err != nil {
		t.Logf("未能从输出中提取帧数: %v", err)
	} else {
		t.Logf("✓ 成功提取帧数: %d", frame)
	}

	// 测试多行输出
	lines := []string{
		"frame=   10 fps= 15 q=28.0 size=     100kB time=00:00:00.40 bitrate=2000.0kbits/s speed=0.5x",
		"frame=   25 fps= 18 q=28.0 size=     250kB time=00:00:01.00 bitrate=2000.0kbits/s speed=0.8x",
		"frame=   50 fps= 20 q=28.0 size=     500kB time=00:00:02.00 bitrate=2000.0kbits/s speed=1.0x",
	}

	t.Log("\n测试多行帧数提取:")
	for i, line := range lines {
		frame, err := GetFrameNum(line)
		if err != nil {
			t.Errorf("第%d行提取失败: %v", i+1, err)
		} else {
			t.Logf("第%d行: 提取帧数 = %d ✓", i+1, frame)
		}
	}

	fmt.Println("\n===== 帧数提取测试完成 =====")
}
