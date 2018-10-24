/* 	
	Lolabot v0.1
	Author(s): Daniel Thorpe, Rev. Taylor R. Rainwater

	- Since Dan cannot document then I will. 
	
	- This is the official Discord bot for the GNU/Fruitsalad
	server. Her name is Lola, named after Pyro's kitty. 
	Currently she just manages roles but is open (source) to 
	changes and additions.

	- If you work on the code: comment changes, update the 
	changelog, and make sure the code runs before making a 
	pull. Bad code will be rejected, no matter what. 

	Vocabulary Used In Comments:
		- User: 
			the server side stdin/stdout person.
		- Caller: 
			the Discord side io person who calls a command. 
*/


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

	"github.com/bwmarrin/discordgo"
)

// Why are the MAJOR and PATCH version numbers the same?
const (
	VERSION_MAJOR = 1
	VERSION_MINOR = 0
	VERSION_PATCH = 1
)

// Can you store all the variables in one var block?
var (
	versionString = fmt.Sprintf("%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
)

var (
	flagDiscordToken string
)

// init function, stores the token from stdin. 
func init() {
	flag.StringVar(&flagDiscordToken, "t", "", "Discord token")
	flag.Parse()
}

// struct for allowed roles.
// stored in a json file. 
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

// An array of allowedRoles consisting of data of the type allowedRole, defined above
type allowedRoles []*allowedRole

// Main function does main things. 
func main() {
	// Error out if user does not supply a token. 
	if flagDiscordToken == "" {
		log.Fatal("No Discord token specified.")
	}
	// Create a Discord session object, with the error handler defined, the type of Bot, and the supplied API token.
	dsession, err := discordgo.New("Bot " + flagDiscordToken)
	// Print error message to stdout.
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}
	// Create a handler for when a message is sent to a channel the bot is in. 
	dsession.AddHandler(messageCreate)
	// Opens the Discord session,store errors in err variable, sets Discord status. 
	err = dsession.Open()
	dsession.UpdateStatus(0, "with the GPL")
	// Checks to see if there was an error, logs if there was.
	if err != nil {
		log.Fatal("Error opening discord ws conn:", err)
	}
	// No problems, client connected and bot running.
	log.Println("Ready received! Ctrl-c to stop.")
	// Kill bot process if the process recieves any type of kill signal, else do nothing.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	// Close the Discord session upon kill signal received.
	dsession.Close()
}

/* 
	messageCreate
	Args: 
		s, Discord session pointer.
		m, Discord message creator pointer. 
	- Dan, for the love of God, use better variable names.
	- Need to break this up into smaller functions. 
*/
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// This is the bot trigger.
	const prefix = "Lola, please"
	// Create an array called loadedRoles with the same properties as allowedRoles - these are used differently
	loadedRoles := allowedRoles{}
	// Check to see if the json file is there for roles.
	if _, err := os.Stat("./allowedRoles.json"); err == nil {
		// Read the json file in.
		readFile, err := ioutil.ReadFile("./allowedRoles.json")
		// Let the user know the role list is read.
		if err != nil {
			log.Println("Opening roles file", err.Error())
		}
		// Decode the file.
		decode := json.NewDecoder(bytes.NewReader(readFile))
		// Store error variable.
		err = decode.Decode(&loadedRoles)
		// Tell the user if the was an error.
		if err != nil {
			log.Println(err)
		}
	} else {
		// Ain't no got dang allowed roles in the file.
		log.Println("No allowed roles defined, please add some roles to be added.")
	}
	// What the fuck does this do?
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Store the ID information.
	channelID, _ := s.State.Channel(m.ChannelID)
	guildID, _ := s.Guild(channelID.GuildID)
	guildAdmin := guildID.OwnerID
	// This prints out the available roles in the json file.
	// Note: this is manually populated.
	if strings.HasPrefix(m.Content, prefix+" list") {
		availableRoles := "Roles available on this server: \n"
		availableRoles = availableRoles + "```\n"
		for i := 0; i < len(loadedRoles); i++ {
			availableRoles = availableRoles + loadedRoles[i].Name + "\n"
		}
		availableRoles = availableRoles + "```"
		s.ChannelMessageSend(m.ChannelID, availableRoles)
	}
	// This checks to see if the command is trying to add a role to the json file
	// and if the caller is the admin. If so, add the role to the json file.
	if strings.HasPrefix(m.Content, prefix+" add") && (m.Author.ID == guildAdmin) {
		roleName := m.Content[17:len(m.Content)]
		// Check to see if the role already exists.
		for i := 0; i < len(loadedRoles); i++ {
			if loadedRoles[i].Name == roleName {
				s.ChannelMessageSend(m.ChannelID, "Role "+roleName+" is already in the list, and has not been added.")
				return
			}
		}
		// Load the channel variable.
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
		}
		// Load the guildRoles variable.
		guildRoles, err := s.GuildRoles(channel.GuildID)
		if err != nil {
		}
		// Default out the roleID variable.
		roleID := ""
		// Check if the role is on the server.
		for i := 0; i < len(guildRoles); i++ {
			if guildRoles[i].Name == roleName {
				roleID = strconv.Itoa(i)
			}
		}
		// If the role does not exist on the server, then let the caller know.
		if roleID == "" {
			s.ChannelMessageSend(m.ChannelID, roleName+" was not found on this server.")
		} else {
			// Role exists on the server, now time to store in json.
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
	// Add role to caller. 
	if strings.HasPrefix(m.Content, prefix+" apply") {
		roleName := m.Content[19:len(m.Content)]
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
	// Remove role from caller.
	if strings.HasPrefix(m.Content, prefix+" remove") {
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
			s.ChannelMessageSend(m.ChannelID, roleName+" was not found on this      server.")
		} else {
			s.GuildMemberRoleRemove(channel.GuildID, m.Author.ID, roleID)
			s.ChannelMessageSend(m.ChannelID, "Removed user <@"+m.Author.ID+"> from "+roleName)
		}
	}

}
