package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var Token string

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Unable to load .env:\n\t %v", err)
	}
	flag.StringVar(&Token, "t", os.Getenv("TOKEN"), "Bot token")
	flag.Parse()
}

func main() {
	var startTime = time.Now()
	fmt.Printf("\n\n---- START %s ----\n", startTime.Format(time.TimeOnly))

	var bot, err = discordgo.New("Bot " + Token)

	bot.Identify.Intents = discordgo.IntentsGuildMessages

	if err != nil {
		log.Fatalf("Unable to create discord bot instance:\n\t %v", err)
	}

	err = bot.Open()
	if err != nil {
		log.Fatalf("Unable to open discord bot connection:\n\t %v", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	bot.Close()
}
