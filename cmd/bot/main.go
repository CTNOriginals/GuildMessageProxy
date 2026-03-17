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
	"github.com/CTNOriginals/GuildMessageProxy/internal/commands"
	"github.com/CTNOriginals/GuildMessageProxy/internal/events"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

var (
	Token   string
	GuildID string
	Global  bool
	NoSync  bool
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Unable to load .env:\n\t %v", err)
	}
	flag.StringVar(&Token, "t", os.Getenv("TOKEN"), "Bot token")
	flag.StringVar(&GuildID, "guild", os.Getenv("DEV_GUILD_ID"), "Guild ID for command registration (dev mode)")
	flag.BoolVar(&Global, "global", false, "Register commands globally (prod mode)")
	flag.BoolVar(&NoSync, "no-sync", false, "Skip command sync for faster restarts")
	flag.Parse()
}

func main() {
	var startTime = time.Now()
	fmt.Printf("\n\n---- START %s ----\n", startTime.Format(time.TimeOnly))

	// Initialize storage
	var store storage.Store = storage.NewMemoryStore()

	var bot, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Unable to create discord bot instance:\n\t %v", err)
	}

	// Update intents: guild messages and guilds (for guild lifecycle events)
	bot.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	// Wire event handlers
	bot.AddHandler(events.HandleReady)
	bot.AddHandler(events.HandleInteractionCreate)
	bot.AddHandler(events.HandleGuildCreate(store))
	bot.AddHandler(events.HandleGuildDelete(store))

	err = bot.Open()
	if err != nil {
		log.Fatalf("Unable to open discord bot connection:\n\t %v", err)
	}

	// Sync commands if not disabled
	if !NoSync {
		var targetGuild string = ""
		if !Global {
			targetGuild = GuildID
		}

		err = commands.SyncCommands(bot, targetGuild)
		if err != nil {
			log.Printf("Warning: Command sync failed: %v", err)
		}
	} else {
		log.Println("Command sync skipped (--no-sync flag)")
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Graceful shutdown
	log.Println("Shutting down gracefully...")
	var shutdownTime = time.Now()
	fmt.Printf("---- END %s (Runtime: %s) ----\n\n", shutdownTime.Format(time.TimeOnly), shutdownTime.Sub(startTime))

	// Cleanly close down the Discord session.
	bot.Close()
}
