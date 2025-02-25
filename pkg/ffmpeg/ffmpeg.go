package ffmpeg

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/dev3mike/video-stream-api/config"
)

// Validator validates FFmpeg installation
type Validator struct{}

// NewValidator creates a new FFmpeg validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks if FFmpeg is installed and available
func (v *Validator) Validate() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg is not installed or not accessible: %w", err)
	}
	return nil
}

// GetVideoDuration returns the duration of a video file in milliseconds
func GetVideoDuration(filePath string) (int64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return int64(duration * 1000), nil
}

// TranscodeOptions contains options for video transcoding
type TranscodeOptions struct {
	VideoID    string
	InputPath  string
	OutputDir  string
	Config     config.FFmpegConfig
	Logger     *logrus.Logger
	OnProgress func(videoID string, percent int)
	DurationMs int64
}

// Transcoder handles video transcoding operations
type Transcoder struct{}

// NewTranscoder creates a new Transcoder instance
func NewTranscoder() *Transcoder {
	return &Transcoder{}
}

// Transcode transcodes a video file to HLS format with multiple renditions
func (t *Transcoder) Transcode(opts TranscodeOptions) error {
	// Build FFmpeg command arguments based on configuration
	cmdArgs := []string{
		"-y", // Overwrite output files
		"-i", opts.InputPath,
		"-stats",
		"-stats_period", "0.1",
		"-loglevel", "info", // Add this to ensure we get progress output
		"-preset", opts.Config.PresetSpeed,
		"-g", "48",
		"-sc_threshold", "0",
	}

	// Add filter complex for multiple renditions
	filterComplex := make([]string, 0, len(opts.Config.Renditions))
	for i := range opts.Config.Renditions {
		filterComplex = append(filterComplex, fmt.Sprintf("[v%d]", i))
	}
	cmdArgs = append(cmdArgs, "-filter_complex",
		fmt.Sprintf("[0:v]split=%d%s",
			len(opts.Config.Renditions),
			strings.Join(filterComplex, "")))

	// Add mapping and encoding parameters for each rendition
	for i, rendition := range opts.Config.Renditions {
		cmdArgs = append(cmdArgs,
			"-map", fmt.Sprintf("[v%d]", i),
			"-map", "0:a?",
			"-b:v:"+strconv.Itoa(i), rendition.Bitrate,
			"-s:v:"+strconv.Itoa(i), fmt.Sprintf("%dx%d", rendition.Width, rendition.Height))
	}

	// Add HLS specific parameters
	cmdArgs = append(cmdArgs,
		"-f", "hls",
		"-hls_time", strconv.Itoa(opts.Config.SegmentDuration),
		"-hls_playlist_type", "vod",
		"-master_pl_name", "master.m3u8",
		filepath.Join(opts.OutputDir, "output_%v.m3u8"))

	opts.Logger.WithFields(logrus.Fields{
		"videoID":     opts.VideoID,
		"command":     "ffmpeg " + strings.Join(cmdArgs, " "),
		"duration_ms": opts.DurationMs,
	}).Info("Starting FFmpeg command")

	// Create and start the FFmpeg command
	cmd := exec.Command("ffmpeg", cmdArgs...)

	// We'll read progress from stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Monitor stderr for progress
	scanner := bufio.NewScanner(stderr)
	var lastProgress int
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Log every line for debugging
		opts.Logger.WithFields(logrus.Fields{
			"videoID": opts.VideoID,
			"line":    line,
			"lineNum": lineCount,
		}).Debug("FFmpeg output line")

		// FFmpeg outputs progress like: frame=  168 fps= 56 q=28.0 size=     256kB time=00:00:06.72 bitrate= 311.3kbits/s speed=2.24x
		if strings.Contains(line, "time=") {
			// Extract the time part
			timeIndex := strings.Index(line, "time=")
			if timeIndex == -1 {
				opts.Logger.WithField("line", line).Debug("Found 'time=' but couldn't extract index")
				continue
			}

			// Get the time string (format: HH:MM:SS.ms)
			timeStr := line[timeIndex+5:]
			timeStr = strings.Fields(timeStr)[0] // Get first part before space

			opts.Logger.WithFields(logrus.Fields{
				"videoID": opts.VideoID,
				"timeStr": timeStr,
				"line":    line,
			}).Debug("Found time string")

			// Parse the time
			timeParts := strings.Split(timeStr, ":")
			if len(timeParts) != 3 {
				opts.Logger.WithField("timeParts", timeParts).Debug("Invalid time parts length")
				continue
			}

			hours, errH := strconv.ParseFloat(timeParts[0], 64)
			minutes, errM := strconv.ParseFloat(timeParts[1], 64)
			secondsParts := strings.Split(timeParts[2], ".")
			seconds, errS := strconv.ParseFloat(secondsParts[0], 64)
			var milliseconds float64
			if len(secondsParts) > 1 {
				milliseconds, _ = strconv.ParseFloat(secondsParts[1], 64)
			}

			if errH != nil || errM != nil || errS != nil {
				opts.Logger.WithFields(logrus.Fields{
					"hours":   timeParts[0],
					"minutes": timeParts[1],
					"seconds": secondsParts[0],
					"errH":    errH,
					"errM":    errM,
					"errS":    errS,
				}).Debug("Error parsing time components")
				continue
			}

			totalSeconds := hours*3600 + minutes*60 + seconds + milliseconds/100
			currentMs := int64(totalSeconds * 1000)

			opts.Logger.WithFields(logrus.Fields{
				"videoID":      opts.VideoID,
				"hours":        hours,
				"minutes":      minutes,
				"seconds":      seconds,
				"milliseconds": milliseconds,
				"totalSeconds": totalSeconds,
				"currentMs":    currentMs,
				"durationMs":   opts.DurationMs,
			}).Debug("Time calculation")

			percentage := int((float64(currentMs) / float64(opts.DurationMs)) * 100)
			if percentage > 100 {
				percentage = 100
			}

			// Only update if progress has changed
			if percentage > lastProgress {
				opts.Logger.WithFields(logrus.Fields{
					"videoID":      opts.VideoID,
					"timeStr":      timeStr,
					"currentMs":    currentMs,
					"durationMs":   opts.DurationMs,
					"percentage":   percentage,
					"lastProgress": lastProgress,
				}).Info("Progress update")

				opts.OnProgress(opts.VideoID, percentage)
				lastProgress = percentage
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading FFmpeg output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("FFmpeg process failed: %w", err)
	}

	return nil
}
