package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/dannyt66/discordgo"
)

const (
	VERSION_MAJOR = 0
	VERSION_MINOR = 0
	VERSION_PATCH = 2
)

var (
	versionString = fmt.Sprintf("%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
)

var (
	flagDiscordToken string
)

func init() {
	flag.StringVar(&flagDiscordToken, "t", "", "Discord token")
	flag.Parse()
}

func main() {
	if flagDiscordToken == "" {
		log.Fatal("No Discord token specified.")
	}

	dsession, err := discordgo.New("Bot " + flagDiscordToken)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	dsession.AddHandler(messageCreate)

	err = dsession.Open()
	if err != nil {
		log.Fatal("Error opening discord ws conn:", err)
	}

	log.Println("Ready received! Ctrl-c to stop.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dsession.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	const prefix = "Fruity"

	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, prefix+" add") {
		roleName := m.Content[11:len(m.Content)]
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
		}
		guildRoles, err := s.GuildRoles(channel.GuildID)
		if err != nil {
		}
		roleID := ""
		for i := 0; i < len(guildRoles); i++ {
			if guildRoles[i].Name == roleName {
				roleID = strconv.Itoa(i)
			}
		}
		if roleID == "" {
			s.ChannelMessageSend(m.ChannelID, roleName+"was not found on this server.")
		} else {
			roleIDInt, err := strconv.Atoi(roleID)
			data, _ := json.Marshal(guildRoles[roleIDInt])
			f, err := os.OpenFile("./allowedRoles.json", os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}

			defer f.Close()

			if _, err = f.WriteString(string(data[:])); err != nil {
				panic(err)
			}
		}
	}

	if strings.HasPrefix(m.Content, prefix+" apply") {
		roleName := m.Content[13:len(m.Content)]
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
		}
		guildRoles, err := s.GuildRoles(channel.GuildID)
		if err != nil {
		}
		roleID := ""
		for i := 0; i < len(guildRoles); i++ {
			if guildRoles[i].Name == roleName {
				roleID = guildRoles[i].ID
			}
		}
		if roleID == "" {
			s.ChannelMessageSend(m.ChannelID, roleName+" was not found on this server.")
		} else {
			s.GuildMemberRoleAdd(channel.GuildID, m.Author.ID, roleID)
			s.ChannelMessageSend(m.ChannelID, roleName)
		}
	}

	if strings.HasPrefix(m.Content, prefix+" remove") {
		roleName := m.Content[14:len(m.Content)]
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
		}
		guildRoles, err := s.GuildRoles(channel.GuildID)
		if err != nil {
		}
		roleID := ""
		for i := 0; i < len(guildRoles); i++ {
			if guildRoles[i].Name == roleName {
				roleID = guildRoles[i].ID
			}
		}
		if roleID == "" {
			s.ChannelMessageSend(m.ChannelID, roleName+" was not found on this      server.")
		} else {
			s.GuildMemberRoleRemove(channel.GuildID, m.Author.ID, roleID)
			s.ChannelMessageSend(m.ChannelID, roleName)
		}
	}

}
