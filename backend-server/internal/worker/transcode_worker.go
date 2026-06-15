package worker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/internal/service"
	"backend-server/pkg/database"
	"backend-server/pkg/logger"
	"backend-server/pkg/storage"

	"go.uber.org/zap"
)

// TranscodeWorker 视频转码 Worker
type TranscodeWorker struct {
	mediaSvc      *service.MediaService
	fileAssetRepo *repository.FileAssetRepo
	ffmpegPath    string
	concurrency   int
}

func NewTranscodeWorker(ffmpegPath string) *TranscodeWorker {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	db := database.GetMySQL()
	return &TranscodeWorker{
		mediaSvc:      service.NewMediaService(),
		fileAssetRepo: repository.NewFileAssetRepo(db),
		ffmpegPath:    ffmpegPath,
		concurrency:   2,
	}
}

// Start 启动转码 Worker
func (w *TranscodeWorker) Start(ctx context.Context) {
	for i := 0; i < w.concurrency; i++ {
		go w.processLoop(ctx, i)
	}
}

func (w *TranscodeWorker) processLoop(ctx context.Context, workerID int) {
	logger.Info("转码 Worker 启动", zap.Int("worker_id", workerID))

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("转码 Worker 停止", zap.Int("worker_id", workerID))
			return
		case <-ticker.C:
			w.processPendingTasks(ctx)
		}
	}
}

func (w *TranscodeWorker) processPendingTasks(ctx context.Context) {
	// 查询待转码任务
	medias, err := w.mediaSvc.GetPendingTranscode(5)
	if err != nil {
		return
	}

	for _, media := range medias {
		w.processTask(ctx, &media)
	}
}

func (w *TranscodeWorker) processTask(ctx context.Context, media *model.MediaAsset) {
	logger.Info("开始转码", zap.Uint("file_asset_id", media.FileAssetID))

	// 更新状态为 processing
	w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "processing", "")

	// 获取文件资产（包含 ObjectKey 和 StorageType）
	fileAsset, err := w.fileAssetRepo.GetByID(media.FileAssetID)
	if err != nil {
		logger.Error("获取文件资产失败", zap.Error(err), zap.Uint("file_asset_id", media.FileAssetID))
		w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "failed", "")
		return
	}

	// 下载源文件到临时目录
	tmpDir, err := os.MkdirTemp("", "transcode-*")
	if err != nil {
		logger.Error("创建临时目录失败", zap.Error(err))
		w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "failed", "")
		return
	}
	defer os.RemoveAll(tmpDir)

	// 根据文件存储驱动获取存储实例
	st := storage.GetStorageByDriver(fileAsset.StorageType)
	if st == nil {
		logger.Error("获取存储实例失败", zap.String("driver", fileAsset.StorageType))
		w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "failed", "")
		return
	}

	// 确定输入文件扩展名
	inputExt := filepath.Ext(fileAsset.FileName)
	if inputExt == "" {
		inputExt = ".mp4"
	}
	inputPath := filepath.Join(tmpDir, "input"+inputExt)

	// 从存储下载文件到本地临时目录
	reader, err := st.Download(ctx, fileAsset.ObjectKey)
	if err != nil {
		logger.Error("下载源文件失败", zap.Error(err), zap.String("objectKey", fileAsset.ObjectKey))
		w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "failed", "")
		return
	}
	defer reader.Close()

	// 写入本地临时文件
	inputFile, err := os.Create(inputPath)
	if err != nil {
		logger.Error("创建临时文件失败", zap.Error(err))
		w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "failed", "")
		return
	}

	if _, err := io.Copy(inputFile, reader); err != nil {
		inputFile.Close()
		logger.Error("写入临时文件失败", zap.Error(err))
		w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "failed", "")
		return
	}
	inputFile.Close() // 立即关闭，确保 FFmpeg 能读取完整文件

	outputDir := filepath.Join(tmpDir, "hls")
	os.MkdirAll(outputDir, 0755)

	// 执行 FFmpeg 转码
	outputPath := filepath.Join(outputDir, "playlist.m3u8")
	err = w.runFFmpeg(inputPath, outputPath)
	if err != nil {
		logger.Error("FFmpeg 转码失败", zap.Error(err), zap.Uint("file_asset_id", media.FileAssetID))
		w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "failed", "")
		return
	}

	// 上传 HLS 文件到 storage
	hlsPrefix := fmt.Sprintf("hls/%d", media.FileAssetID)
	err = w.uploadHLSFiles(outputDir, hlsPrefix)
	if err != nil {
		logger.Error("上传 HLS 文件失败", zap.Error(err))
		w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "failed", "")
		return
	}

	// 更新转码状态
	hlsPath := hlsPrefix + "/playlist.m3u8"
	w.mediaSvc.UpdateTranscodeStatus(media.FileAssetID, "completed", hlsPath)

	logger.Info("转码完成", zap.Uint("file_asset_id", media.FileAssetID), zap.String("hls_path", hlsPath))
}

func (w *TranscodeWorker) runFFmpeg(inputPath, outputPath string) error {
	args := []string{
		"-i", inputPath,
		"-profile:v", "baseline",
		"-level", "3.0",
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		outputPath,
	}

	cmd := exec.Command(w.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg 执行失败: %w, output: %s", err, string(output))
	}
	return nil
}

func (w *TranscodeWorker) uploadHLSFiles(localDir, remotePrefix string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relPath, _ := filepath.Rel(localDir, path)
		remoteKey := remotePrefix + "/" + strings.ReplaceAll(relPath, "\\", "/")

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		contentType := "application/octet-stream"
		if strings.HasSuffix(path, ".m3u8") {
			contentType = "application/vnd.apple.mpegurl"
		} else if strings.HasSuffix(path, ".ts") {
			contentType = "video/mp2t"
		}

		_, err = storage.GetStorage().Upload(context.Background(), remoteKey, file, contentType)
		return err
	})
}
