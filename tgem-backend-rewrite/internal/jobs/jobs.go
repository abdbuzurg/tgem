package jobs

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func Run() {
  location, err := time.LoadLocation("Asia/Dushanbe")
  if err != nil {
    panic(err)
  }
  fmt.Println(location, time.Now().In(location), time.Now())

	c := cron.New(cron.WithLocation(location))

  _, err = c.AddFunc("55 23 * * *", func() { 
    fmt.Printf("Началось ежедневное сохранение прогресса проектов - %v\n", time.Now().In(location))
    ProgressReportDaily() 
    fmt.Printf("Закончилось ежедневное сохранение прогресса проектов - %v\n", time.Now().In(location))
  })
  if err != nil {
    panic(err)
  }
  fmt.Println("CRON Прогресс Проекта запущен")

  c.Start()
}
