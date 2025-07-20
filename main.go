package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mehkij/poke-auction/internal/database"
	"github.com/mehkij/poke-auction/internal/dispatcher"
	"github.com/mehkij/poke-auction/internal/types"
)

func main() {
	log.Println("Loading environment variables...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %s", err)
	}
	log.Println("Environment variables loaded!")

	botToken := os.Getenv("BOT_TOKEN")
	appID := os.Getenv("APP_ID")
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("error connecting to database: %s", err)
	}

	dbQueries := database.New(db)
	cfg := &types.GlobalConfig{
		GlobalDispatcher: dispatcher.NewDispatcher(),
		Queries:          dbQueries,
	}

	if botToken == "" || appID == "" {
		log.Fatal("Required environment variables not set!")
	}

	log.Println("Creating new session...")
	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("error creating new session: %s", err)
	}

	session.AddHandler(NewInteractionHandler(cfg))
	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	log.Println("Opening new session...")
	err = session.Open()
	if err != nil {
		log.Fatalf("error opening session: %s", err)
	}
	defer session.Close()
	log.Println("Session successfully opened!")

	// Setup a simple HTTP server to serve the status of the bot
	go func() {
		http.HandleFunc("/api/status", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			guilds := len(session.State.Guilds)

			res := fmt.Sprintf(`{
				"status": "online",
				"uptime": "%s",
				"guilds": %d,
				"ping": %d,
				"version": "1.4.0"
			}`, getUptime(), guilds, session.HeartbeatLatency().Milliseconds())

			_, err := w.Write([]byte(res))
			if err != nil {
				log.Printf("error setting up HTTP server %s", err)
				return
			}
		}))

		log.Println("Starting HTTP server on port 8080")
		server := &http.Server{
			Addr:              ":8080",
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Println("Registering commands...")
	cmds := RegisterAll(session, appID, "")
	log.Println("Commands successfully registered!")

	log.Println("Loading global dispatcher...")
	cfg.GlobalDispatcher.Start()
	defer cfg.GlobalDispatcher.Stop()
	log.Println("Global dispatcher loaded!")

	log.Println("The bot is online!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Clean up by deleting the command
	for _, cmd := range cmds {
		err = session.ApplicationCommandDelete(appID, "", cmd.ID)
		if err != nil {
			fmt.Println("Cannot delete slash command:", err)
		}
	}
}

var startTime = time.Now()

func getUptime() string {
	uptime := time.Since(startTime)

	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
}
