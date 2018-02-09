package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/dannyt66/discordgo"
)

const (
	VERSION_MAJOR = 1
	VERSION_MINOR = 0
	VERSION_PATCH = 1
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

type allowedRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Managed     bool   `json:"managed"`
	Mentionable bool   `json:"mentionable"`
	Hoist       bool   `json:"hoist"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Permissions int    `json:"permissions"`
}

type allowedRoles []*allowedRole

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
	dsession.UpdateStatus(0, "with the GPL")
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
	const prefix = "Lola, please"

	loadedRoles := allowedRoles{}
	if _, err := os.Stat("./allowedRoles.json"); err == nil {
		readFile, err := ioutil.ReadFile("./allowedRoles.json")
		if err != nil {
			log.Println("Opening roles file", err.Error())
		}
		decode := json.NewDecoder(bytes.NewReader(readFile))
		err = decode.Decode(&loadedRoles)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println("No allowed roles defined, please add some roles to be added.")
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	channelID, _ := s.State.Channel(m.ChannelID)
	guildID, _ := s.Guild(channelID.GuildID)
	guildAdmin := guildID.OwnerID

	if strings.HasPrefix(m.Content, prefix+" list") {
		availableRoles := "Roles available on this server: \n"
		availableRoles = availableRoles + "```\n"
		for i := 0; i < len(loadedRoles); i++ {
			availableRoles = availableRoles + loadedRoles[i].Name + "\n"
		}
		availableRoles = availableRoles + "```"
		s.ChannelMessageSend(m.ChannelID, availableRoles)
	}

	if strings.HasPrefix(m.Content, prefix+" add") && (m.Author.ID == guildAdmin) {
		roleName := m.Content[11:len(m.Content)]
		for i := 0; i < len(loadedRoles); i++ {
			if loadedRoles[i].Name == roleName {
				s.ChannelMessageSend(m.ChannelID, "Role "+roleName+" is already in the list, and has not been added.")
				return
			}
		}
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
			s.ChannelMessageSend(m.ChannelID, roleName+" was not found on this server.")
		} else {
			roleIDInt, err := strconv.Atoi(roleID)
			data, _ := json.Marshal(guildRoles[roleIDInt])
			log.Println("Begin file write")
			err = ioutil.WriteFile("./allowedRoles.json", []byte("["), 0644)
			if err != nil {
			}
			for i := 0; i < len(loadedRoles); i++ {
				data, _ := json.Marshal(loadedRoles[i])
				f, err := os.OpenFile("./allowedRoles.json", os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					panic(err)
				}

				defer f.Close()

				if _, err = f.WriteString(string(data[:]) + ","); err != nil {
					panic(err)
				}

			}
			f, err := os.OpenFile("./allowedRoles.json", os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}

			defer f.Close()

			if _, err = f.WriteString(string(data[:]) + "]"); err != nil {
				panic(err)
			}
			s.ChannelMessageSend(m.ChannelID, "Added "+roleName+" to available roles.")

		}
	}

	if strings.HasPrefix(m.Content, prefix+" apply") {
		roleName := m.Content[20:len(m.Content)]
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
		}
		if err != nil {
		}
		roleID := ""
		for i := 0; i < len(loadedRoles); i++ {
			if loadedRoles[i].Name == roleName {
				roleID = loadedRoles[i].ID
			}
		}
		if roleID == "" {
			s.ChannelMessageSend(m.ChannelID, roleName+" was not found on this server.")
		} else {
			s.GuildMemberRoleAdd(channel.GuildID, m.Author.ID, roleID)
			s.ChannelMessageSend(m.ChannelID, "Given user <@"+m.Author.ID+"> "+roleName)
		}
	}

	if strings.HasPrefix(m.Content, prefix+" remove") {
		roleName := m.Content[21:len(m.Content)]
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
		}
		if err != nil {
		}
		roleID := ""
		for i := 0; i < len(loadedRoles); i++ {
			if loadedRoles[i].Name == roleName {
				roleID = loadedRoles[i].ID
			}
		}
		if roleID == "" {
			s.ChannelMessageSend(m.ChannelID, roleName+" was not found on this      server.")
		} else {
			s.GuildMemberRoleRemove(channel.GuildID, m.Author.ID, roleID)
			s.ChannelMessageSend(m.ChannelID, "Removed user <@"+m.Author.ID+"> from "+roleName)
		}
	}

}
