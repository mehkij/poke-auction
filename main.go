package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/mehkij/poke-auction/internal/cmd"
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

	if botToken == "" || appID == "" {
		log.Fatal("Required environment variables not set!")
	}

	// Setup a simple HTTP server to keep the Repl alive
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Bot is running!")
		})
		http.ListenAndServe(":8080", nil)
	}()

	log.Println("Creating new session...")
	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("error creating new session: %s", err)
	}

	session.AddHandler(cmd.HandleInteraction)
	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	log.Println("Opening new session...")
	err = session.Open()
	if err != nil {
		log.Fatalf("error opening session: %s", err)
	}
	defer session.Close()
	log.Println("Session successfully opened!")

	log.Println("Registering commands...")
	cmds := cmd.RegisterAll(session, appID, "")
	log.Println("Commands successfully registered!")

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
