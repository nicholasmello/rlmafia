package MafiaGame

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Command int

const (
	Clear Command = iota
	Winner
	Score
	Leaderboard
	Join
	Vote
	NumMafia
	Start
	Help
	SetPrefix
	err
)

type MafiaGame struct {
	Guild      string
	Prefix     string
	Players    []MafiaPlayer
	NumMafia   int
	InProgress bool
}

type MafiaPlayer struct {
	ID     string
	score  int
	mafia  bool
	active bool
	vote   string
}

func (m *MafiaGame) Winner(args []string) string {
	if !m.InProgress {
		return "Game has not started, !start"
	}

	for i := 0; i < len(m.Players); i++ {
		if m.Players[i].vote == "" {
			return "Not everyone has voted"
		}
	}

	for i := 0; i < 3; i++ {
		args[i] = args[i][3:21]
		contains := false
		for j := 0; j < len(m.Players); j++ {
			if args[i] == m.Players[j].ID {
				contains = true
			}
		}
		if contains == false {
			return "<@" + args[i] + "> is not in the game"
		}
	}

	var mafiaIs []int
	var votesForMafia []int
	for i := 0; i < len(m.Players); i++ {
		if m.Players[i].mafia {
			mafiaIs = append(mafiaIs, i)
			votesForMafia = append(votesForMafia, 0)
		} else {
			for j := 0; j < 3; j++ {
				if m.Players[i].ID == args[j] {
					m.Players[i].score++
				}
			}
		}
	}

	for i := 0; i < len(m.Players); i++ {
		for j := 0; j < m.NumMafia; j++ {
			if m.Players[i].vote == m.Players[mafiaIs[j]].ID {
				m.Players[i].score++
				votesForMafia[j]++
			}
		}
	}

	for i := 0; i < len(mafiaIs); i++ {
		if votesForMafia[i] < 3 {
			m.Players[i].score += 3
		}
	}

	for i := 0; i < len(m.Players); i++ {
		m.Players[i].vote = ""
		m.Players[i].mafia = false
	}
	m.InProgress = false

	return m.LeaderBoard()
}

func (m *MafiaGame) Start(s *discordgo.Session) string {
	if len(m.Players) < m.NumMafia {
		return "Too many mafia, add more players or reduce the number of mafia"
	} else if len(m.Players) == 0 {
		return "Nobody is in the game, type !join to play"
	} else if m.InProgress {
		return "The game is already started"
	}

	for i := 0; i < m.NumMafia; i++ {
		for {
			rand.Seed(time.Now().UnixNano())
			randomnum := rand.Intn(len(m.Players))
			if !m.Players[randomnum].mafia {
				m.Players[randomnum].mafia = true
				break
			}
		}
	}

	for i := 0; i < len(m.Players); i++ {
		channel, err := s.UserChannelCreate(m.Players[i].ID)
		if err != nil {
			fmt.Println("error creating channel: ", err)
		}
		if m.Players[i].mafia {
			_, err = s.ChannelMessageSend(channel.ID, "You are mafia")
			if err != nil {
				fmt.Println("error sending DM message: ", err)
			}
		} else {
			_, err = s.ChannelMessageSend(channel.ID, "You are innocent")
			if err != nil {
				fmt.Println("error sending DM message: ", err)
			}
		}
	}

	m.InProgress = true
	return "Players may join their team, after the game is completed !vote"
}

func (m *MafiaGame) Vote(playerID string, args []string) string {
	if !m.InProgress {
		return "Game has not started, !start"
	}

	// Find self in list given ID
	playerI := -1
	for i := 0; i < len(m.Players); i++ {
		if m.Players[i].ID == playerID {
			playerI = i
		}
	}
	if playerI == -1 {
		return "<@" + playerID + ">, you must be in the game to vote"
	}

	if len(args[0]) != 22 {
		return "Invalid player"
	}
	opp := args[0][3:21]

	// Find opponent in list given ID
	oppI := -1
	for i := 0; i < len(m.Players); i++ {
		if m.Players[i].ID == opp {
			oppI = i
		}
	}
	if oppI == -1 {
		return "<@" + opp + ">, is not in the game"
	}

	m.Players[playerI].vote = m.Players[oppI].ID
	return "<@" + playerID + "> has voted for <@" + opp + ">"
}

func (m *MafiaGame) Clear() string {
	finalLeaderboard := m.LeaderBoard()
	m.Players = []MafiaPlayer{}
	m.NumMafia = 1
	return "Game has been cleared\n" + finalLeaderboard
}

func (m *MafiaGame) LeaderBoard() string {
	retVal := "Leaderboard: \n"
	for i := 0; i < len(m.Players); i++ {
		retVal += "<@" + m.Players[i].ID + ">: " + strconv.Itoa(m.Players[i].score) + "\n"
	}
	return retVal
}

func (m *MafiaGame) SetNumMafia(args []string) string {
	num, err := strconv.Atoi(args[0])
	if err != nil {
		return "Invalid number of mafia"
	}
	m.NumMafia = num
	return "Number of Mafia set to: " + strconv.Itoa(m.NumMafia)
}

func (m *MafiaGame) Score(playerID string) string {
	score := 0
	for i := 0; i < len(m.Players); i++ {
		if m.Players[i].ID == playerID {
			score = m.Players[i].score
		}
	}
	return "<@" + playerID + "> score: " + strconv.Itoa(score)
}

func (m *MafiaGame) Join(playerID string) string {
	for i := 0; i < len(m.Players); i++ {
		if m.Players[i].ID == playerID {
			return "<@" + playerID + "> you are already in the game"
		}
	}

	if m.InProgress == true {
		return "<@" + playerID + "> Game is in progress, please wait until after the round"
	}
	m.Players = append(m.Players, MafiaPlayer{
		ID:     playerID,
		score:  0,
		mafia:  false,
		active: true,
		vote:   "",
	})
	return "<@" + playerID + "> Has joined the game"
}

func (m *MafiaGame) SetPrefix(args []string) string {
	m.Prefix = args[0]
	return "The Prefix is now " + m.Prefix
}

func (m *MafiaGame) Help() string {
	return "```!clear\n" +
		"\"Restarts\" the game, clears all the players and scores out of memory.\n```" +
		"```!winner {Player} {Player} {Player}\n" +
		"Lists the 3 players who won the rocket league game, must be @messages\n```" +
		"```!score\n" +
		"Replays with the current score of the player who used the command.\n```" +
		"```!leaderboard\n" +
		"Replays with the scores of all players in the game.\n```" +
		"```!join\n" +
		"Adds the user of the command to the game.\n```" +
		"```!vote {Player}\n" +
		"Casts a vote for a player, used at the end of the Rocket League game. Must be @mentions to work properly.\n```" +
		"```!num-mafia {Number}\n" +
		"Sets the number of mafia, the game cannot be start with a higher number of mafia than number of players.\n```" +
		"```!start\n" +
		"Starts the game by sending a direct message to everyone with their role.\n```" +
		"```!setprefix {New Prefix}\n" +
		"Sets the prefix for the server to use before commands```" +
		"```!help\n" +
		"Lists commands and descriptions.```"
}

func Parse(input string) (Command, []string) {
	s := strings.Split(input, " ")
	switch strings.ToLower(s[0][1:]) {
	case "clear":
		return Clear, s[1:]
	case "winner":
		return Winner, s[1:]
	case "score":
		return Score, s[1:]
	case "leaderboard":
		return Leaderboard, s[1:]
	case "join":
		return Join, s[1:]
	case "vote":
		return Vote, s[1:]
	case "nummafia":
		return NumMafia, s[1:]
	case "start":
		return Start, s[1:]
	case "help":
		return Help, s[1:]
	case "setprefix":
		return SetPrefix, s[1:]
	}
	return err, nil
}

func New(guild string) *MafiaGame {
	return &MafiaGame{
		guild,
		"!",
		[]MafiaPlayer{},
		1,
		false,
	}
}
