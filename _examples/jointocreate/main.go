package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/wumpkit-go/wumpkit"
)

// Config

const (
	GeneratorChannelID wumpkit.Snowflake = 1449557097187512370
)

// Database

type TempChannelData struct {
	ChannelID wumpkit.Snowflake
	OwnerID   wumpkit.Snowflake
}

type DataBase struct {
	tempChannelDataMu sync.RWMutex
	tempChannelData   map[wumpkit.Snowflake]TempChannelData
}

func (d *DataBase) PutTempChannel(data TempChannelData) {
	d.tempChannelDataMu.Lock()
	d.tempChannelData[data.ChannelID] = data
	d.tempChannelDataMu.Unlock()
}

func (d *DataBase) GetTempChannel(channelID wumpkit.Snowflake) (data TempChannelData, ok bool) {
	d.tempChannelDataMu.RLock()
	data, ok = d.tempChannelData[channelID]
	d.tempChannelDataMu.RUnlock()
	return
}

func (d *DataBase) DelTempChannel(channelID wumpkit.Snowflake) {
	d.tempChannelDataMu.Lock()
	delete(d.tempChannelData, channelID)
	d.tempChannelDataMu.Unlock()
}

func NewDataBase() DataBase {
	return DataBase{
		tempChannelData: make(map[wumpkit.Snowflake]TempChannelData),
	}
}

func main() {
	database := NewDataBase()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Error("DISCORD_TOKEN environment variable is required")
		os.Exit(1)
	}

	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent |
				wumpkit.GatewayIntentGuildVoiceStates,
		}),
		wumpkit.WithLogger(logger),
	)

	bot.OnVoiceStateUpdate(func(ctx context.Context, event wumpkit.VoiceStateUpdateEvent) {
		// create Temp Channel when user joins generator channel
		if event.NewState.ChannelID == GeneratorChannelID && event.NewState.Member.ID != 0 {
			displayName := event.NewState.Member.DisplayName()

			tempChannel, err := bot.CreateChannel(ctx, event.NewState.GuildID, wumpkit.CreateChannelOptions{
				Name:     displayName,
				Type:     wumpkit.ChannelTypeGuildVoice,
				ParentID: 1396267977074016429,
			})
			if err != nil {
				logger.Error("failed to create temporary channel", "error", err)
				return
			}

			// save to Database
			database.PutTempChannel(TempChannelData{
				ChannelID: tempChannel.GetID(),
				OwnerID:   event.NewState.Member.ID,
			})

			// move User to Created Channel
			_, _ = bot.ModifyMember(ctx, event.NewState.GuildID, event.NewState.Member.ID, wumpkit.ModifyMemberOptions{
				ChannelID: tempChannel.GetID(),
			})
			return
		}

		// delete Temp Channel if Empty
		if data, ok := database.GetTempChannel(event.OldState.ChannelID); ok {
			members, _ := bot.GetVoiceChannelMembers(ctx, data.ChannelID)
			if len(members) == 0 {
				_, _ = bot.DeleteChannel(ctx, data.ChannelID, wumpkit.DeleteChannelOptions{})
				database.DelTempChannel(data.ChannelID)
			}
		}
	})

	if err := bot.StartGatewayWithSignals(); err != nil {
		logger.Error("bot terminated", "error", err)
	}
}
