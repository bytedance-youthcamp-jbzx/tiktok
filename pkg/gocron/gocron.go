package gocron

import (
	"github.com/go-co-op/gocron"
	"time"
)

// Schedule 定时任务 gocron
type Schedule struct {
	*gocron.Scheduler
}

func NewSchedule() *gocron.Scheduler {
	return gocron.NewScheduler(time.Local)
}
