package stream

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"FilimoDownloader-GholamTaksir/internal/helper"
)

type Builder struct {
	directory     string
	temporary     []string
	Input         string
	Output        string
	Video         string
	Audio         []string
	Subtitle      []string
	FileExtension string
	Compress      bool
}

func (b *Builder) outputFile(dir string) string {
	fileName := path.Base(dir)
	if fileName == "." {
		fileName = "output"
	}
	if !strings.HasSuffix(fileName, b.FileExtension) {
		fileName += b.FileExtension
	}
	return path.Join(b.directory, fileName)
}

func (b *Builder) buildPlaylist(dir string) error {
	playlistFile := PlaylistFile(dir)
	tempOutput := b.outputFile(dir)
	b.temporary = append(b.temporary, tempOutput)

	fmt.Printf("  Converting: %s\n", path.Base(tempOutput))

	// 1) کپی مستقیم با تحمل بسته‌های خراب (رایج در HLS فیلیمو)
	argsCopy := hlsInputArgs(playlistFile)
	if b.Compress {
		argsCopy = append(argsCopy,
			"-c:v", "libx264", "-preset", "fast", "-crf", "23",
			"-c:a", "aac", "-b:a", "128k",
			"-movflags", "+faststart",
		)
	} else {
		argsCopy = append(argsCopy, "-c", "copy", "-movflags", "+faststart")
	}
	argsCopy = append(argsCopy, "-y", tempOutput)

	err := runFFmpeg(dir, argsCopy)
	if err == nil || usableMediaFile(tempOutput) {
		if err != nil {
			fmt.Printf("  Warning: FFmpeg خطا داد ولی خروجی قابل‌استفاده است (%s)\n", path.Base(tempOutput))
		}
		return nil
	}

	// 2) ویدیو کپی + صدای دوباره انکود (رفع خطای ADTS/AAC خراب)
	fmt.Println("  Retry: re-encoding audio (corrupt AAC packets)...")
	_ = os.Remove(tempOutput)
	argsAudio := hlsInputArgs(playlistFile)
	argsAudio = append(argsAudio,
		"-c:v", "copy",
		"-c:a", "aac", "-b:a", "160k",
		"-movflags", "+faststart",
		"-y", tempOutput,
	)
	err2 := runFFmpeg(dir, argsAudio)
	if err2 == nil || usableMediaFile(tempOutput) {
		if err2 != nil {
			fmt.Printf("  Warning: audio re-encode reported errors but output is usable\n")
		}
		return nil
	}

	// 3) آخرین تلاش: انکود کامل
	fmt.Println("  Retry: full re-encode...")
	_ = os.Remove(tempOutput)
	argsFull := hlsInputArgs(playlistFile)
	argsFull = append(argsFull,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20",
		"-c:a", "aac", "-b:a", "160k",
		"-movflags", "+faststart",
		"-y", tempOutput,
	)
	err3 := runFFmpeg(dir, argsFull)
	if err3 == nil || usableMediaFile(tempOutput) {
		return nil
	}
	if err2 != nil {
		return err2
	}
	return err
}

func hlsInputArgs(playlistFile string) []string {
	return []string{
		"-fflags", "+genpts+discardcorrupt+igndts",
		"-err_detect", "ignore_err",
		"-allowed_extensions", "ALL",
		"-protocol_whitelist", "file,crypto,data,http,https,tcp,tls",
		"-i", path.Base(playlistFile),
	}
}

func usableMediaFile(p string) bool {
	info, err := os.Stat(p)
	if err != nil || info.IsDir() {
		return false
	}
	// کمتر از ۱ مگابایت تقریباً حتماً خراب/ناقص است
	return info.Size() >= 1*1024*1024
}

func (b *Builder) make() (string, error) {
	inputIndex := 0
	inputs := []string{"-i", b.outputFile(b.Video)}
	mapping := []string{"-map", fmt.Sprintf("%d:v", inputIndex)}
	actions := []string{"-c:v", "copy", "-c:a", "copy"}
	meta := []string{}

	var outputFile string
	if b.Output == "" {
		outputFile = b.outputFile(b.Input)
	} else {
		outputFile = b.outputFile(b.Output)
	}
	if !strings.HasSuffix(outputFile, b.FileExtension) {
		outputFile = strings.TrimSuffix(outputFile, path.Ext(outputFile)) + b.FileExtension
	}

	// اگه audio جداگانه نداره از ویدیو audio بگیر
	if len(b.Audio) == 0 {
		mapping = append(mapping, "-map", fmt.Sprintf("%d:a?", inputIndex))
	}

	// audio tracks - همه track ها با هم
	for idx, audio := range b.Audio {
		inputIndex++
		inputs = append(inputs, "-i", b.outputFile(audio))
		mapping = append(mapping, "-map", fmt.Sprintf("%d:a", inputIndex))
		meta = append(meta, buildMeta("a", idx, audio)...)
	}

	// subtitle tracks - همه زیرنویس‌ها embed بشن
	subIndex := 0
	for _, subtitle := range b.Subtitle {
		srtFile := SrtFile(subtitle)
		if !helper.IsFileExists(srtFile) {
			fmt.Printf("  Warning: subtitle file not found: %s\n", srtFile)
			continue
		}
		content := helper.ReadFile(srtFile)
		if len(strings.TrimSpace(content)) < 10 {
			fmt.Printf("  Warning: subtitle %s is empty, skipping\n", path.Base(subtitle))
			continue
		}
		inputIndex++
		inputs = append(inputs, "-i", srtFile)
		mapping = append(mapping, "-map", fmt.Sprintf("%d:s", inputIndex))
		if b.FileExtension == ".mkv" {
			actions = append(actions, "-c:s", "srt")
		} else {
			actions = append(actions, "-c:s", "mov_text")
		}
		meta = append(meta, buildMeta("s", subIndex, subtitle)...)
		subIndex++
		fmt.Printf("  Embedding subtitle: %s\n", path.Base(subtitle))
	}

	args := []string{}
	args = append(args, inputs...)
	args = append(args, mapping...)
	args = append(args, actions...)
	args = append(args, meta...)
	args = append(args, "-y", outputFile)

	fmt.Println("  Merging all tracks...")
	err := runFFmpeg("", args)
	if err != nil && !usableMediaFile(outputFile) {
		fmt.Println("  Merge failed, trying simple copy fallback...")
		if err2 := runFFmpeg("", []string{
			"-fflags", "+genpts+discardcorrupt",
			"-i", b.outputFile(b.Video),
			"-c", "copy", "-movflags", "+faststart",
			"-y", outputFile,
		}); err2 != nil && !usableMediaFile(outputFile) {
			return "", fmt.Errorf("ساخت فایل نهایی ممکن نشد: %w", err)
		}
	} else if !usableMediaFile(outputFile) {
		fmt.Println("  Merge failed, trying simple copy fallback...")
		if err := runFFmpeg("", []string{
			"-fflags", "+genpts+discardcorrupt",
			"-i", b.outputFile(b.Video),
			"-c", "copy", "-movflags", "+faststart",
			"-y", outputFile,
		}); err != nil && !usableMediaFile(outputFile) {
			return "", fmt.Errorf("ساخت فایل نهایی ممکن نشد: %w", err)
		}
	} else {
		if err != nil {
			fmt.Printf("  Warning: merge reported errors but output is usable\n")
		}
		fmt.Printf("  Output: %s\n", path.Base(outputFile))
	}
	if !usableMediaFile(outputFile) {
		return "", fmt.Errorf("فایل نهایی ساخته نشد")
	}
	return outputFile, nil
}

func (b *Builder) cleanup() {
	for _, tmp := range b.temporary {
		if helper.IsFileExists(tmp) {
			helper.DeleteFile(tmp)
		}
	}
}

func (b *Builder) Build() {
	if _, err := b.Run(); err != nil {
		helper.ShowErrorAndExit(err.Error())
	}
}

func (b *Builder) Run() (string, error) {
	if b.FileExtension == "" {
		b.FileExtension = ".mp4"
	}
	if b.Output != "" {
		b.directory = path.Dir(b.Output)
		helper.MakeDirectories(b.directory)
	}
	if b.directory == "" {
		b.directory = b.Input
	}

	fmt.Println("  [1/3] Converting video...")
	if err := b.buildPlaylist(b.Video); err != nil {
		return "", err
	}

	if len(b.Audio) > 0 {
		fmt.Printf("  [2/3] Converting %d audio track(s)...\n", len(b.Audio))
		for _, audio := range b.Audio {
			if err := b.buildPlaylist(audio); err != nil {
				return "", err
			}
		}
	} else {
		fmt.Println("  [2/3] No separate audio track.")
	}

	fmt.Println("  [3/3] Merging...")
	outputFile, err := b.make()
	if err != nil {
		return "", err
	}

	fmt.Println("  Cleaning up temp files...")
	b.cleanup()
	return outputFile, nil
}

func BuildAuto(input, output, ext string) (string, error) {
	if ext != ".mkv" {
		ext = ".mp4"
	}
	if existing := existingOutput(input, ext); existing != "" && len(OptionDirs(VideoDir(input), true)) == 0 {
		// Already built (e.g. previous attempt cleaned video/) — reuse it.
		return existing, nil
	}
	videos := OptionDirs(VideoDir(input), true)
	if len(videos) == 0 {
		if existing := existingOutput(input, ext); existing != "" {
			return existing, nil
		}
		return "", fmt.Errorf("پوشه ویدیو برای ساخت فایل پیدا نشد")
	}
	builder := Builder{
		Input:         input,
		Output:        output,
		Video:         videos[0],
		Audio:         OptionDirs(AudioDir(input), true),
		Subtitle:      OptionDirs(SubtitleDir(input), false),
		FileExtension: ext,
		Compress:      false,
	}
	out, err := builder.Run()
	if err != nil {
		if existing := existingOutput(input, ext); existing != "" {
			CleanupMediaDirs(input)
			return existing, nil
		}
		return "", err
	}
	CleanupMediaDirs(input)
	return out, nil
}

func existingOutput(input, ext string) string {
	base := path.Base(input)
	if base == "" || base == "." {
		return ""
	}
	candidate := path.Join(input, base+ext)
	if usableMediaFile(candidate) {
		return candidate
	}
	entries, err := os.ReadDir(input)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ext) && usableMediaFile(path.Join(input, name)) {
			return path.Join(input, name)
		}
	}
	return ""
}

func buildMeta(streamType string, idx int, dir string) []string {
	lang := path.Base(dir)
	return []string{
		fmt.Sprintf("-metadata:s:%s:%d", streamType, idx),
		fmt.Sprintf("language=%s", helper.ConvertISO6391ToISO6392(lang)),
	}
}

func runFFmpeg(workDir string, args []string) error {
	bin, err := lookFFmpeg()
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	// GUI apps on macOS often miss Homebrew in PATH; keep env but prepend common bins
	cmd.Env = append(os.Environ(), ffmpegPathEnv()...)

	var stderr bytes.Buffer
	cmd.Stdout = nil
	cmd.Stderr = &stderr

	done := make(chan struct{})
	go func() {
		sp := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Printf("\r  Done!                        \n")
				return
			default:
				fmt.Printf("\r  %s Processing...              ", sp[i%len(sp)])
				i++
				time.Sleep(120 * time.Millisecond)
			}
		}
	}()

	runErr := cmd.Run()
	close(done)
	if runErr != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 500 {
			msg = msg[len(msg)-500:]
		}
		if msg == "" {
			return fmt.Errorf("FFmpeg شکست خورد: %w", runErr)
		}
		return fmt.Errorf("FFmpeg شکست خورد: %s", msg)
	}
	return nil
}

func ffmpegPathEnv() []string {
	return []string{
		"PATH=/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:" + os.Getenv("PATH"),
	}
}

func lookFFmpeg() (string, error) {
	candidates := []string{
		"/opt/homebrew/bin/ffmpeg",
		"/usr/local/bin/ffmpeg",
		"/usr/bin/ffmpeg",
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		candidates = append([]string{p}, candidates...)
	}
	seen := map[string]bool{}
	for _, c := range candidates {
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("FFmpeg پیدا نشد. در ترمینال بنویس: brew install ffmpeg")
}

func ffmpegPath() string {
	app, err := lookFFmpeg()
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}
	return app
}
