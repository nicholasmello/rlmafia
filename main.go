package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/xonmello/rlmafia/MafiaGame"
)

var (
	guildIDs []string
	games    []*MafiaGame.MafiaGame
)

func main() {
	// Get token from -t flag
	token := flag.String("t", "", "token for auth with discord")
	flag.Parse()

	// Create a new Discord session using the provided bot token
	bot, err := discordgo.New("Bot " + *token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	bot.AddHandler(eventHandler)

	bot.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening
	err = bot.Open()
	if err != nil {
		fmt.Println("Error opening connection: ", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	bot.Close()
	fmt.Println("")
}

func containsString(arr []string, val string) bool {
	for _, e := range arr {
		if e == val {
			return true
		}
	}
	return false
}

func eventHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore message if it is created by self
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !containsString(guildIDs, m.GuildID) {
		guildIDs = append(guildIDs, m.GuildID)
		games = append(games, MafiaGame.New(m.GuildID))
		fmt.Println("New Server Registered: " + m.GuildID)
	}

	gameID := 0
	for i := 0; i < len(games); i++ {
		if games[i].Guild == m.GuildID {
			gameID = i
		}
	}

	if !strings.HasPrefix(m.Message.Content, games[gameID].Prefix) {
		return
	}

	command, args := MafiaGame.Parse(m.Message.Content)

	switch command {
	case MafiaGame.Help:
		s.ChannelMessageSend(m.ChannelID, games[gameID].Help())
	case MafiaGame.SetPrefix:
		if len(args) == 0 {
			s.ChannelMessageSend(m.ChannelID, "```!setprefix {New Prefix}\n"+
				"Sets the prefix for the server to use before commands```")
		} else {
			s.ChannelMessageSend(m.ChannelID, games[gameID].SetPrefix(args))
		}
	case MafiaGame.Join:
		s.ChannelMessageSend(m.ChannelID, games[gameID].Join(m.Author.ID))
	case MafiaGame.Score:
		s.ChannelMessageSend(m.ChannelID, games[gameID].Score(m.Author.ID))
	case MafiaGame.NumMafia:
		if len(args) == 0 {
			s.ChannelMessageSend(m.ChannelID, "```!num-mafia {Number}\n"+
				"Sets the number of mafia, the game cannot be start with a higher number of mafia than number of players.\n```")
		} else {
			s.ChannelMessageSend(m.ChannelID, games[gameID].SetNumMafia(args))
		}
	case MafiaGame.Leaderboard:
		s.ChannelMessageSend(m.ChannelID, games[gameID].LeaderBoard())
	case MafiaGame.Clear:
		s.ChannelMessageSend(m.ChannelID, games[gameID].Clear())
	case MafiaGame.Vote:
		if len(args) == 0 {
			s.ChannelMessageSend(m.ChannelID, "```!vote {Player}\n"+
				"Casts a vote for a player, used at the end of the Rocket League game. Must be @mentions to work properly.\n```")
		} else {
			s.ChannelMessageSend(m.ChannelID, games[gameID].Vote(m.Author.ID, args))
		}
	case MafiaGame.Start:
		s.ChannelMessageSend(m.ChannelID, games[gameID].Start(s))
	case MafiaGame.Winner:
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "```!winner {Player} {Player} {Player}\n"+
				"Lists the 3 players who won the rocket league game, must be @messages\n```")
		} else {
			s.ChannelMessageSend(m.ChannelID, games[gameID].Winner(args))
		}
	}
}
