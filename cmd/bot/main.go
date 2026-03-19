package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/CTNOriginals/GuildMessageProxy/internal/commands"
	"github.com/CTNOriginals/GuildMessageProxy/internal/events"
	"github.com/CTNOriginals/GuildMessageProxy/internal/health"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

var (
	Token        string
	GuildID      string
	Global       bool
	NoSync       bool
	DatabasePath string
	UseMemory    bool
)

func init() {
	if err := godotenv.Load(); err != nil {
		logging.Fatal("unable to load .env file", logging.Err("error", err))
	}
	flag.StringVar(&Token, "t", os.Getenv("TOKEN"), "Bot token")
	flag.StringVar(&GuildID, "guild", os.Getenv("DEV_GUILD_ID"), "Guild ID for command registration (dev mode)")
	flag.BoolVar(&Global, "global", false, "Register commands globally (prod mode)")
	flag.BoolVar(&NoSync, "no-sync", false, "Skip command sync for faster restarts")
	flag.StringVar(&DatabasePath, "db", os.Getenv("DATABASE_PATH"), "Path to SQLite database file")
	flag.BoolVar(&UseMemory, "memory", false, "Use in-memory storage instead of SQLite (for testing)")
	flag.Parse()
}

func main() {
	var startTime = time.Now()

	// Log bot startup with version information
	logging.Info("bot starting",
		logging.String("version", "dev"),
		logging.String("go_version", runtime.Version()),
	)

	fmt.Printf("\n\n---- START %s ----\n", startTime.Format(time.TimeOnly))

	// Initialize storage
	var store storage.Store
	var storageType string

	if UseMemory {
		store = storage.NewMemoryStore()
		storageType = "memory"
	} else {
		var dbPath = DatabasePath
		if dbPath == "" {
			dbPath = "guildmessageproxy.db"
		}
		var sqliteStore, err = storage.NewSQLiteStore(dbPath)
		if err != nil {
			logging.Fatal("failed to initialize sqlite storage", logging.Err("error", err))
		}
		store = sqliteStore
		storageType = "sqlite"
	}
	logging.Info("storage initialized", logging.String("type", storageType))

	// Initialize command handlers with storage
	commands.Store = store

	var bot, err = discordgo.New("Bot " + Token)
	if err != nil {
		logging.Fatal("unable to create discord bot instance", logging.Err("error", err))
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
		logging.Fatal("unable to open discord connection", logging.Err("error", err))
	}

	logging.Info("discord session opened")

	// Start health check server
	var healthServer *health.Server = health.NewServer(bot, startTime)
	err = healthServer.Start(":8080")
	if err != nil {
		logging.Error("failed to start health server", logging.Err("error", err))
		// Don't fail startup, just log the error
	}

	// Start draft cleanup goroutine
	go func() {
		var ticker = time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var cleaned = commands.CleanupExpiredDrafts()
				if cleaned > 0 {
					logging.Info("draft cleanup completed",
						logging.Int("cleaned_count", cleaned),
					)
				}
			}
		}
	}()

	// Sync commands if not disabled
	if !NoSync {
		var targetGuild string = ""
		if !Global {
			targetGuild = GuildID
		}

		var count = len(commands.CommandDefinitions)
		err = commands.SyncCommands(bot, targetGuild)
		if err != nil {
			logging.Warn("command sync failed", logging.Err("error", err))
		} else {
			var scope = "guild"
			if Global {
				scope = "global"
			}
			logging.Info("commands registered",
				logging.Int("count", count),
				logging.String("scope", scope),
			)
		}
	} else {
		logging.Info("command sync skipped", logging.String("reason", "--no-sync flag"))
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	var sig = <-sc

	// Graceful shutdown
	logging.Info("shutdown initiated", logging.String("signal", sig.String()))

	var shutdownTime = time.Now()
	fmt.Printf("---- END %s (Runtime: %s) ----\n\n", shutdownTime.Format(time.TimeOnly), shutdownTime.Sub(startTime))

	// Shutdown health server
	var shutdownCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = healthServer.Stop(shutdownCtx)
	if err != nil {
		logging.Warn("health server shutdown error", logging.Err("error", err))
	}

	// Cleanly close down the Discord session.
	bot.Close()

	var duration = time.Since(shutdownTime)
	logging.Info("shutdown complete", logging.Duration("duration", duration))
}
