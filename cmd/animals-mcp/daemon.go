package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/AlexS8332/AnimalGuide_Task18/internal/daemon"
	"github.com/AlexS8332/AnimalGuide_Task18/internal/llm"
	"github.com/AlexS8332/AnimalGuide_Task18/internal/schedule"
)

// daemonFlags — флаги режима демона.
type daemonFlags struct {
	on        bool
	run       string
	every     time.Duration
	summaryAt string
	mddAt     string
	budget    float64
}

// runDaemon — режим 24/7 (-daemon) или одно задание (-run). Возвращает код
// выхода. Ключ модели — только из окружения: сервер не ищет .env сам, его
// запускают из консоли, службы или планировщика Windows, где переменная
// задаётся явно.
func runDaemon(f daemonFlags, dataDir, mddURL string, logger *slog.Logger) int {
	key := os.Getenv("DEEPSEEK_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "демону нужен ключ DeepSeek в переменной DEEPSEEK_API_KEY")
		return 2
	}
	// Журнал демона нужен человеку всегда, не только с -v.
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	d, err := daemon.Open(ctx, daemon.Config{
		DataDir: dataDir, Every: f.every, SummaryAt: f.summaryAt, MDDAt: f.mddAt, Budget: f.budget,
		LLM: llm.NewClient(key, os.Getenv("DEEPSEEK_BASE_URL")), Model: os.Getenv("DEEPSEEK_MODEL"),
		MDDURL: mddURL, Log: logger,
		OnRun: func(r schedule.Run) {
			logger.Info("запуск", "job", r.Job, "trigger", r.Trigger, "status", r.Status,
				"cost_usd", fmt.Sprintf("%.4f", r.CostUSD), "ref", r.Ref, "detail", r.Detail, "error", r.Error,
				"took", r.Finished.Sub(r.Started).Round(time.Millisecond))
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "демон:", err)
		return 1
	}
	defer d.Close()

	if f.run != "" {
		r, err := d.Sched.RunNow(ctx, f.run)
		if err != nil {
			fmt.Fprintln(os.Stderr, "задание:", err)
			return 1
		}
		if r.Status != schedule.RunOK {
			return 1
		}
		return 0
	}
	st, err := d.Sched.Status(ctx)
	if err == nil {
		for _, j := range st.Jobs {
			logger.Info("задание", "job", j.Name, "every", j.Every, "daily", j.Daily, "next", j.Next.Format("2006-01-02 15:04"))
		}
		logger.Info("демон запущен", "budget_usd", st.Budget, "spent_today_usd", fmt.Sprintf("%.4f", st.Spent), "location", st.Location)
	}
	if err := d.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "демон:", err)
		return 1
	}
	logger.Info("демон остановлен")
	return 0
}
