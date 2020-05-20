package jeff

import "github.com/bwmarrin/discordgo"

//Session - Session object for a given connection, implements *discordgo.Session
type Session struct {
	*discordgo.Session
	Prefix               string
	hasCmdHandler        bool
	commands             []command
	guildPrefixOverrides []struct {
		guild  string
		prefix string
	}
}

//Message - A message object that implements *discordgo.Message, but with jeff methods
type Message struct {
	*discordgo.Message
	session *Session
}

//ParsedCommand - A parsed command
type ParsedCommand struct {
	Flags   []string
	Subcmds []struct {
		Cmd string
		Arg string
	}
	Arg string
}

//CommandHandler - A function to handle a parsed command
type CommandHandler func(s *Session, m *Message, c ParsedCommand)

type command struct {
	cmd     string
	subcmds []string
	flags   []string
	handler CommandHandler
}

//InternalError - A jeff internal error
type InternalError struct {
	ErrorString string
}

func (err *InternalError) Error() string { return err.ErrorString }

//ErrCmdExists is the error for preexisting commands
var ErrCmdExists = &InternalError{"A command with that name already exists for this session."}

//ErrCmdContainsIllegalChars is the error for commands, subcommands and flags that contain illegal characters
var ErrCmdContainsIllegalChars = &InternalError{"The command, subcommand or flag contains illegal character."}

//ErrCmdArgsNotUnique is the error for when a flag and subcommand share the same name
var ErrCmdArgsNotUnique = &InternalError{"The flags and subcommands must all be unique."}

//ErrMessageSessionNil is the error when the session linked to a message is a nil pointer
var ErrMessageSessionNil = &InternalError{"The session linked to the message is a nil pointer."}
