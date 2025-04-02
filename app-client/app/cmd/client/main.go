package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-client/app/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	// Уведомляем о сигналах Interrupt (Ctrl+C) и SIGTERM
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	logging.L(ctx).Info("Starting application")

	newApp, err := app.NewApp(ctx)
	if err != nil {
		logging.L(ctx).Error("Failed to initialize application", logging.ErrAttr(err))
		os.Exit(1)
	}

	// Запускаем Run синхронно в основной горутине
	runErr := newApp.Run(ctx)
	if runErr != nil {
		logging.L(ctx).Error("app run failed", logging.ErrAttr(runErr))
		// Можно решить, нужно ли выходить с ошибкой
		// os.Exit(1)
	} else {
		logging.L(ctx).Info("Application finished task successfully")
	}

	// Можно добавить ожидание сигнала, если нужно, чтобы клиент
	// не завершался сразу, а ждал, например, Ctrl+C.
	// Если клиент должен просто выполнить задачу и выйти,
	// то этот блок можно убрать.
	// go func() {
	// 	<-sigs
	// 	logging.L(ctx).Info("Received termination signal, shutting down...")
	// 	cancel() // Отменяем контекст, если нужно прервать длительные операции
	// }()
	//
	// <-ctx.Done() // Ждем сигнала или завершения Run (если cancel вызван там)
	// logging.L(ctx).Info("Application shutting down.")
}
