//Package jeff - A helper library for writing discord bots
package jeff

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

//VERSION is jeff's version
const VERSION = "0.0.1"

var sessions []*Session

//New - Create a new session
func New(token string) (s *Session, err error) {
	ses, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}

	s = &Session{Session: ses, Prefix: "!", hasCmdHandler: false}

	sessions = append(sessions, s)

	return
}

//Run - Open the session and hang until interrupt
func (ses *Session) Run() (err error) {
	err = ses.Open()

	if err != nil {
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	<-sc

	ses.Close()
	return
}

//NewCommand - Define a new command
func (ses *Session) NewCommand(cmd string, flags []string, subcmds []string, handler CommandHandler) error {
	if !ses.hasCmdHandler {
		ses.AddHandler(commandParser)
		ses.hasCmdHandler = true
	}

	for _, v := range ses.commands {
		if v.cmd == cmd {
			return ErrCmdExists
		}
	}

	if strings.Contains(cmd, " ") || cmd == "" {
		return ErrCmdContainsIllegalChars
	}

	for _, v := range flags {
		if strings.Contains(v, " ") || v == "" {
			return ErrCmdContainsIllegalChars
		}
	}

	for _, v := range subcmds {
		if strings.Contains(v, " ") || v == "" {
			return ErrCmdContainsIllegalChars
		}

		for _, w := range flags {
			if w == v {
				return ErrCmdArgsNotUnique
			}
		}
	}

	var ncmd command

	ncmd.cmd = cmd
	ncmd.subcmds = subcmds
	ncmd.flags = flags
	ncmd.handler = handler

	ses.commands = append(ses.commands, ncmd)

	return nil
}

func commandParser(s *discordgo.Session, m *discordgo.MessageCreate) {
	var ses *Session

	for _, v := range sessions {
		if v.Session == s {
			ses = v
			break
		}
	}

	if ses == nil {
		return
	}

	var prefix = ses.Prefix
	for _, v := range ses.guildPrefixOverrides {
		if v.guild == m.GuildID {
			prefix = v.prefix
			break
		}
	}

	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	content := m.Content
	content = strings.TrimPrefix(content, prefix)

	split := strings.Split(content, " ")

	var pcmd ParsedCommand
	var test bool

	var cmd command

	for _, v := range ses.commands {
		if v.cmd == split[0] {
			cmd = v
			test = true
			break
		}
	}

	split = append(split[:0], split[1:]...)

	if !test {
		return
	}

	skip := false

	var usedFlags []string
	var usedSubs []string

	for i, v := range split {
		if skip {
			skip = false
			continue
		}
		f := false
		for _, w := range cmd.flags {
			if w == v {
				used := false
				for _, use := range usedFlags {
					if use == w {
						used = true
					}
				}
				if used {
					break
				}
				pcmd.Flags = append(pcmd.Flags, v)
				f = true
				usedFlags = append(usedFlags, w)
				break
			}
		}

		if f {
			continue
		}

		for _, w := range cmd.subcmds {
			if w == v {
				used := false
				for _, use := range usedSubs {
					if use == w {
						used = true
					}
				}
				if used {
					break
				}
				f = true
				skip = true
				usedSubs = append(usedSubs, w)
				var subcmd struct {
					Cmd string
					Arg string
				}
				subcmd.Cmd = v
				if i+1 != len(split) {
					subcmd.Arg = split[i+1]
				} else {
					subcmd.Arg = ""
				}
				pcmd.Subcmds = append(pcmd.Subcmds, subcmd)
				break
			}
		}

		if f {
			continue
		}

		l := len(split[i:]) - 1
		for j, w := range split[i:] {
			pcmd.Arg += w
			if !(j == l) {
				pcmd.Arg += " "
			}
		}
		break
	}

	cmd.handler(ses, m.Message, pcmd)
}
