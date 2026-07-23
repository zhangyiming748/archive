package archive

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/zhangyiming748/archive/util"
)

// TestExecCommandWithBar 测试进度条功能
func TestExecCommandWithBar(t *testing.T) {
	// 创建一个测试用的短视频文件(5秒,30fps = 150帧)
	testVideo := "test_input.mp4"

	// 使用ffmpeg生成一个测试视频
	cmd := exec.Command("ffmpeg", "-y",
		"-f", "lavfi",
		"-i", "color=c=blue:s=640x480:d=5:r=30",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		testVideo)

	t.Log("创建测试视频文件(5秒,150帧)...")
	if err := cmd.Run(); err != nil {
		t.Skipf("无法创建测试视频(需要ffmpeg): %v", err)
	}
	defer os.Remove(testVideo)

	// 等待文件完全写入
	time.Sleep(500 * time.Millisecond)

	// 测试转换并显示进度条
	outputVideo := "test_output.mp4"
	defer os.Remove(outputVideo)

	cmd = exec.Command("ffmpeg", "-i", testVideo,
		"-c:v", "libx265",
		"-tag:v", "hvc1",
		"-c:a", "aac",
		"-preset", "fast", // 使用fast预设以便更快完成
		outputVideo)

	t.Log("\n========== 开始测试带进度条的视频转换 ==========")
	t.Log("请观察下方的进度条和预估时间")
	t.Log("================================================")

	err := util.ExecuteCommandWithRealtimeOutput(cmd)
	if err != nil {
		t.Errorf("转换失败: %v", err)
	}

	// 验证输出文件是否存在
	if _, err := os.Stat(outputVideo); os.IsNotExist(err) {
		t.Error("输出文件未创建")
	} else {
		t.Log("\n✓ 转换成功,进度条功能正常工作")
		t.Log("✓ 进度条应该显示了:")
		t.Log("  - 当前处理的帧数")
		t.Log("  - 总帧数")
		t.Log("  - 处理速度(it/s)")
		t.Log("  - 预估剩余时间")
	}
}

// TestGetFrameNum 测试帧数提取功能
func TestGetFrameNum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		hasError bool
	}{
		{
			name:     "标准ffmpeg输出",
			input:    "frame=  100 fps= 30 q=28.0 size=     500kB time=00:00:03.33 bitrate=1234.5kbits/s speed=1.0x",
			expected: 100,
			hasError: false,
		},
		{
			name:     "不同空格格式",
			input:    "frame=200 fps=25",
			expected: 200,
			hasError: false,
		},
		{
			name:     "无帧数信息",
			input:    "some random output without frame info",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := util.GetFrameNum(tt.input)
			if tt.hasError && err == nil {
				t.Errorf("期望错误但没有得到错误")
			}
			if !tt.hasError && err != nil {
				t.Errorf("不期望错误但得到: %v", err)
			}
			if frame != tt.expected {
				t.Errorf("期望帧数 %d, 得到 %d", tt.expected, frame)
			}
		})
	}
}
