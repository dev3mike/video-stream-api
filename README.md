# Video Streaming API 🎥

A video streaming api that converts videos into HLS format for adaptive bitrate streaming.

## Features 🌟

- Upload videos in various formats (MP4, MOV, AVI, MKV)
- Convert videos to HLS format for streaming
- Multiple quality renditions for adaptive streaming
- Real-time transcoding progress tracking
- Support for video seeking
- RESTful API endpoints

## Requirements 🛠️

- Go 1.16 or higher
- FFmpeg installed on your system
- At least 1GB of free disk space for video processing

## Installation 📥

1. Install FFmpeg:
   ```bash
   # For macOS
   brew install ffmpeg

   # For Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install ffmpeg

   # For Windows
   # Download from https://ffmpeg.org/download.html
   ```

2. Clone the repository:
   ```bash
   git clone https://github.com/dev3mike/video-stream-api.git
   cd video-stream-api
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Run the server:
   ```bash
   go run main.go
   ```

## API Endpoints 🔌

### 1. Upload Video
- **URL**: `/upload`
- **Method**: `POST`
- **Content-Type**: `multipart/form-data`
- **Form Field**: `video`
```bash
curl -X POST -F "video=@your-video.mp4" http://localhost:8080/upload
```

### 2. Check Progress
- **URL**: `/video/{id}/progress`
- **Method**: `GET`
```bash
curl http://localhost:8080/video/123/progress
```

### 3. Stream Video File
- **URL**: `/video/{id}/streamfile`
- **Method**: `GET`
- Supports range requests for seeking

### 4. HLS Streaming
- **URL**: `/video/{id}/stream/master.m3u8`
- **Method**: `GET`
- Access the HLS manifest file for adaptive streaming

## Video Processing 🎞️

The service automatically creates multiple renditions of your video:
- 360p (640x360) - Good for mobile data
- 480p (854x480) - Better quality
- 720p (1280x720) - HD quality

## Directory Structure 📁

```
video-stream-api/
├── api/          # API handlers
├── config/       # Configuration
├── internal/     # Internal packages
├── pkg/         # Public packages
├── videos/      # Uploaded videos
└── output/      # Processed HLS files
```

## Configuration ⚙️

Default settings:
- Max file size: 500MB
- Allowed formats: MP4, MOV, AVI, MKV
- HLS segment duration: 4 seconds
- FFmpeg preset: fast
