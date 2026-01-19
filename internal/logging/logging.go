package logging

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/wcx0206/hermes/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logMaxSizeMB  = 50
	logMaxBackups = 7
	logMaxAgeDays = 14
)

var (
	global *zap.Logger
	once   sync.Once
)

func Init(cfg config.Logging) error {
	var err error
	once.Do(func() {
		err = os.MkdirAll(filepath.Dir(cfg.Path), 0o755)
		if err != nil {
			return
		}

		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    logMaxSizeMB,  // 最大日志文件大小（MB）
			MaxBackups: logMaxBackups, // 最大备份文件数量
			MaxAge:     logMaxAgeDays, // 最大备份文件保存天数
			Compress:   true,
		})

		encCfg := zap.NewProductionEncoderConfig()
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encCfg),
			zapcore.NewMultiWriteSyncer(writer, zapcore.AddSync(os.Stdout)),
			zap.InfoLevel,
		)

		global = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	})
	if err != nil {
		return nil
	}
	return nil
}

func L() *zap.Logger {
	if global == nil {
		panic("logger not initialized: call logging.Init first")
	}
	return global
}

func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}
