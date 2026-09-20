package helper

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func Beep() {
	fmt.Print("\a")
}

// OpenFolder پوشه دانلود رو باز میکنه - با مسیر absolute
func OpenFolder(folderPath string) {
	absPath, err := filepath.Abs(folderPath)
	if err != nil {
		absPath = folderPath
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", absPath)
	case "darwin":
		cmd = exec.Command("open", absPath)
	case "linux":
		cmd = exec.Command("xdg-open", absPath)
	default:
		return
	}
	_ = cmd.Start()
}

// RevealInFinder فایل را در Finder/Explorer نشان می‌دهد
func RevealInFinder(filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-R", absPath)
	case "windows":
		cmd = exec.Command("explorer", "/select,", absPath)
	default:
		OpenFolder(filepath.Dir(absPath))
		return
	}
	_ = cmd.Start()
}

// PreviewFile Quick Look / باز کردن فایل
func PreviewFile(filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("qlmanage", "-p", absPath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", absPath)
	default:
		cmd = exec.Command("xdg-open", absPath)
	}
	_ = cmd.Start()
}

// Notify اعلان سیستم‌عامل
func Notify(title, message string) {
	title = strings.TrimSpace(title)
	message = strings.TrimSpace(message)
	if title == "" {
		title = "Filimo Downloader"
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		script := "display notification " + strconv.Quote(message) + " with title " + strconv.Quote(title) + " sound name \"Glass\""
		cmd = exec.Command("osascript", "-e", script)
	case "linux":
		cmd = exec.Command("notify-send", title, message)
	case "windows":
		ps := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null
$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$texts = $template.GetElementsByTagName('text')
$texts.Item(0).AppendChild($template.CreateTextNode(%s)) | Out-Null
$texts.Item(1).AppendChild($template.CreateTextNode(%s)) | Out-Null
$toast = [Windows.UI.Notifications.ToastNotification]::new($template)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('FilimoDownloader').Show($toast)
`, strconv.Quote(title), strconv.Quote(message))
		cmd = exec.Command("powershell", "-NoProfile", "-Command", ps)
	default:
		return
	}
	_ = cmd.Start()
}
