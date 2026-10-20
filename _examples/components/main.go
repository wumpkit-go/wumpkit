package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wumpkit-go/wumpkit"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	bot := wumpkit.New(
		context.Background(),
		wumpkit.WithToken(token),
		wumpkit.WithGatewayConfig(wumpkit.ShardManagerConfig{
			Intents: wumpkit.GatewayIntentGuilds |
				wumpkit.GatewayIntentGuildMessages |
				wumpkit.GatewayIntentMessageContent,
		}),
	)

	bot.OnMessageCreate(func(ctx context.Context, event wumpkit.MessageCreateEvent) {
		if event.Message.Author.Bot {
			return
		}

		if event.Message.Content == "!menu" {
			// row 1: Interactive Buttons
			btnSuccess := wumpkit.NewButtonBuilder().
				SetCustomID("btn_accept").
				SetLabel("Accept").
				SetStyle(wumpkit.ButtonStyleSuccess).
				Build()

			btnDanger := wumpkit.NewButtonBuilder().
				SetCustomID("btn_decline").
				SetLabel("Decline").
				SetStyle(wumpkit.ButtonStyleDanger).
				Build()

			btnLink := wumpkit.NewButtonBuilder().
				SetURL("https://github.com/wumpkit-go/wumpkit").
				SetLabel("GitHub Docs").
				SetStyle(wumpkit.ButtonStyleLink).
				Build()

			buttonRow := wumpkit.NewActionRowBuilder().
				SetComponent(btnSuccess).
				SetComponent(btnDanger).
				SetComponent(btnLink).
				Build()

			// row 2: String Select Menu
			selectMenu := wumpkit.NewStringSelectMenuBuilder().
				SetCustomID("select_role_color").
				SetPlaceholder("Choose your favorite role color...").
				SetOptions([]wumpkit.SelectOptionStructure{
					{
						Label:       "Summer Sky Blue",
						Value:       "Summer Sky Blue",
						Description: "Classic light blue theme",
					},
					{
						Label:       "Emerald Green",
						Value:       "Emerald Green",
						Description: "Vibrant and fresh",
					},
					{
						Label:       "Ruby Red",
						Value:       "Ruby Red",
						Description: "Bold and fierce",
					},
				}).
				Build()

			selectRow := wumpkit.NewActionRowBuilder().
				SetComponent(selectMenu).
				Build()

			embed := wumpkit.NewEmbedBuilder().
				SetTitle("Interactive Component Showcase").
				SetDescription("Click the buttons below or select a theme from the dropdown menu.").
				SetColor(wumpkit.ColorSummerSky).
				SetTimestamp(time.Now()).
				Build()

			_, err := bot.CreateMessage(ctx, event.Message.ChannelID, wumpkit.CreateMessageOptions{
				Embeds:     []wumpkit.Embed{embed},
				Components: []wumpkit.LayoutComponent{buttonRow, selectRow},
			})
			if err != nil {
				log.Printf("Failed to send message: %v", err)
			}
		}
	})

	// handle component interactions
	bot.OnInteractionCreate(func(ctx context.Context, event wumpkit.InteractionCreateEvent) {
		in, ok := event.Interaction.(*wumpkit.ComponentInteraction)
		if !ok {
			return
		}

		user := in.Author()

		switch in.Data.CustomID {
		case "btn_accept":
			_ = bot.RespondMessage(ctx, in, fmt.Sprintf("Thank you %s, you have accepted the terms.", user.DisplayName()), wumpkit.RespondMessageOptions{
				Ephemeral: true,
			})

		case "btn_decline":
			_ = bot.RespondMessage(ctx, in, fmt.Sprintf("%s, you have declined the terms.", user.DisplayName()), wumpkit.RespondMessageOptions{
				Ephemeral: true,
			})

		case "select_role_color":
			selected := strings.Join(in.Data.Values, ", ")
			msg := fmt.Sprintf("%s, you selected: %s", user.DisplayName(), selected)
			_ = bot.RespondMessage(ctx, in, msg, wumpkit.RespondMessageOptions{
				Ephemeral: true,
			})
		}
	})

	log.Println("Starting Components Bot...")
	if err := bot.StartGatewayWithSignals(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
