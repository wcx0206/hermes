package backup

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/wcx0206/hermes/internal/config"
	"github.com/wcx0206/hermes/internal/logging"
)

type CronServer struct {
	cron   *cron.Cron
	cfg    *config.Config
	logger *zap.Logger
}

func NewCronServer(cfg *config.Config) *CronServer {
	return &CronServer{
		cfg:    cfg,
		cron:   cron.New(),
		logger: logging.L(),
	}
}

func (s *CronServer) Start(ctx context.Context) error {
	for _, project := range s.cfg.Projects {
		p := project
		_, err := s.cron.AddFunc(p.Cron, func() {
			s.logger.Info("backup started", zap.String("project", p.Name))
			now := time.Now()
			if err := RunProject(ctx, s.cfg, &p); err != nil {
				s.logger.Error("backup failed", zap.String("project", p.Name), zap.Error(err))
				return
			}
			cost := time.Since(now)
			s.logger.Info("backup done", zap.String("project", p.Name), zap.Duration("cost", cost))
		})
		if err != nil {
			return err
		}
	}
	s.cron.Start()

	s.logger.Info("Herme started")

	go func() {
		<-ctx.Done()
		s.cron.Stop()
	}()
	return nil
}
