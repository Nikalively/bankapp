package main

import (
	"bankapp/internal/config"
	"bankapp/internal/handlers"
	"bankapp/internal/repo"
	"bankapp/internal/services"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// логируем в текстовом формате
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	// Загружаем конфиг
	cfg := config.Load()

	// Подключаем БД
	db := repo.NewDB(cfg)
	defer db.Close()

	// Инициализация репозиториев
	userRepo := repo.NewUserRepo(db)
	accRepo := repo.NewAccountRepo(db)
	cardRepo := repo.NewCardRepo(db)
	txRepo := repo.NewTransactionRepo(db)
	credRepo := repo.NewCreditRepo(db)
	schedRepo := repo.NewScheduleRepo(db)

	// Сервис
	svc := services.NewBankService(
		userRepo, accRepo, cardRepo, txRepo, credRepo, schedRepo, cfg,
	)

	// Запускаем шедулер: проверяет каждые сутки все просроченные платежи
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			<-ticker.C
			if err := svc.ProcessScheduledPayments(); err != nil {
				logrus.Errorf("scheduler error: %v", err)
			}
		}
	}()

	// HTTP
	h := handlers.NewHandler(svc)
	r := mux.NewRouter()

	r.HandleFunc("/register", h.Register).Methods("POST")
	r.HandleFunc("/login", h.Login).Methods("POST")

	auth := r.PathPrefix("/").Subrouter()
	auth.Use(h.AuthMiddleware)

	auth.HandleFunc("/accounts", h.CreateAccount).Methods("POST")
	auth.HandleFunc("/accounts", h.GetAccounts).Methods("GET")
	auth.HandleFunc("/accounts/{id}/cards", h.GenerateCard).Methods("POST")
	auth.HandleFunc("/accounts/{id}/cards", h.GetCards).Methods("GET")
	auth.HandleFunc("/payments", h.PayWithCard).Methods("POST")
	auth.HandleFunc("/transfers", h.Transfer).Methods("POST")
	auth.HandleFunc("/deposits", h.Deposit).Methods("POST")
	auth.HandleFunc("/credits", h.ApplyCredit).Methods("POST")
	auth.HandleFunc("/credits", h.GetCredits).Methods("GET")
	auth.HandleFunc("/schedule/{credit_id}", h.GetSchedule).Methods("GET")

	addr := fmt.Sprintf(":%d", cfg.Port)
	logrus.Infof("starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logrus.Fatalf("listen: %v", err)
	}
}
