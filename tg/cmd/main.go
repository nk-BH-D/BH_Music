package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	mux "github.com/gorilla/mux"
	com "github.com/nk-BH-D/BH_Music/tg/communication"
	config "github.com/nk-BH-D/BH_Music/tg/internal/config"
	mus "github.com/nk-BH-D/BH_Music/tg/internal/music_db"
	us "github.com/nk-BH-D/BH_Music/tg/internal/users_db"
	tg "github.com/nk-BH-D/BH_Music/tg/service"
)

func main() {
	conf, err := config.Loader()
	if err != nil {
		log.Fatalf("error whem loading config: %v", err)
	}

	// connect to Postgres lib
	pg_mus_db, err := mus.NewMusPostgres(conf.DB_MUS_URL, conf.POSTGRES_MUS_SMOC, conf.POSTGRES_MUS_SMIC)
	if err != nil {
		log.Fatalf("failed to connect postgres_mus_db: %v", err)
	}
	defer pg_mus_db.CloseMus()

	// connect to Postgres users
	pg_us_db, err := us.NewUsPostgres(conf.DB_US_URL, conf.POSTGRES_US_SMOC, conf.POSTGRES_US_SMIC)
	if err != nil {
		log.Fatalf("failed to connect postgres_us_db: %v", err)
	}
	defer pg_us_db.CloseUs()

	// func for initialization of variables
	tg.Init(conf, pg_mus_db, pg_us_db)

	// handler for health-check
	muxHealth := http.NewServeMux()
	muxHealth.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		if err := pg_mus_db.DB_mus.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("db_lib not ready, status: %d, error: %v",
				http.StatusInternalServerError,
				err,
			)
			return
		}
		if err := pg_us_db.DB_us.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("db_us not ready, status: %d, error: %v",
				http.StatusInternalServerError,
				err,
			)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// struct for start health-check handler and greceful shutdown this handler
	HC := &http.Server{
		Addr:    fmt.Sprintf(":%s", conf.HEALTH_CHECK_PORT),
		Handler: muxHealth,
	}

	// start health-check handler inside gorutine
	go func() {
		log.Printf("health-check handler listening port :%s", HC.Addr)
		if err := HC.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve health-check handler: %v", err)
		}
	}()

	// start Tg bot with graceful shutdown
	bot, err := tgbotapi.NewBotAPI(conf.BOT_TOKEN)
	if err != nil {
		log.Panicf("error create bot: %v", err)
	}
	log.Printf("Logged in as %s", bot.Self.UserName)

	up := tgbotapi.NewUpdate(0)
	up.Timeout = 60
	updates := bot.GetUpdatesChan(up)
	botDone := make(chan struct{})

	go func() {
		defer close(botDone)
		for update := range updates {
			if update.Message != nil {
				go tg.HandleMessage(bot, update.Message)
			} else if update.CallbackQuery != nil {
				go tg.HandleCallback(bot, update.CallbackQuery)
			}
		}
	}()

	// start and handler com-server
	rout := mux.NewRouter()

	rout.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Перехвачена паника: %v", r)
				http.Error(w, "panic", http.StatusInternalServerError)
				return
			}
		}()
		log.Println("Запрос: /home")
		com.HandlerHelloParser(w, r)
		log.Println("Запрос /home обработан")
	})

	CS := &http.Server{
		Addr:    fmt.Sprintf(":%s", conf.COMMUN_PORT),
		Handler: rout,
	}

	go func() {
		log.Printf("com-server listening on port %s", CS.Addr)
		if err := CS.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve com-server: %v", err)
		}
	}()

	// graceful shutdown global
	signa_chan := make(chan os.Signal, 1)
	signal.Notify(signa_chan, syscall.SIGINT, syscall.SIGTERM)

	<-signa_chan
	// health-check handler
	ctxHC, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Graceful shutdown: stopping health-check")
	HC.Shutdown(ctxHC)
	log.Println("health-check http handler stopped")

	// TG bot
	log.Printf("Graceful shutdown: stopping %s", bot.Self.UserName)
	// oficial method for stopping receiving updates
	bot.StopReceivingUpdates()
	<-botDone
	log.Printf("%s stopped", bot.Self.UserName)
}
