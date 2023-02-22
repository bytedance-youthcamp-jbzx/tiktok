package gocron_test

import (
	"testing"
	"time"

	"github.com/bytedance-youthcamp-jbzx/dousheng/pkg/gocron"
)

func task() {
	println("hello world")
	ch := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		ch <- true
	}()
	<-ch
	println("waibi waibi")
}
func TestGocron(t *testing.T) {
	s := gocron.NewSchedule()
	s.Every(2).Second().Do(task)
	s.StartAsync()
}

func TestZuSe(t *testing.T) {
	task()
}
