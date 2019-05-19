package lemon

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/monochromegane/conflag"
)

var confFile = "~/.config/lemonade.toml"

func (c *CLI) FlagParse(args []string, skip bool) error {
	style := c.getCommandType(args)
	if style == SUBCOMMAND {
		args = args[:len(args)-1]
	}

	return c.parse(args, skip)
}

func (c *CLI) getCommandType(args []string) (s CommandStyle) {
	s = ALIAS
	switch {
	case regexp.MustCompile(`/?xdg-open$`).MatchString(args[0]):
		c.Type = OPEN
		return
	case regexp.MustCompile(`/?pbpaste$`).MatchString(args[0]):
		c.Type = PASTE
		return
	case regexp.MustCompile(`/?pbcopy$`).MatchString(args[0]):
		c.Type = COPY
		return
	}

	del := func(i int) {
		copy(args[i+1:], args[i+2:])
		args[len(args)-1] = ""
	}

	s = SUBCOMMAND
	for i, v := range args[1:] {
		switch v {
		case "open":
			c.Type = OPEN
			del(i)
			return
		case "paste":
			c.Type = PASTE
			del(i)
			return
		case "copy":
			c.Type = COPY
			del(i)
			return
		case "server":
			c.Type = SERVER
			del(i)
			return
		}
	}

	s = NULL
	return s
}

func (c *CLI) flags() *flag.FlagSet {
	flags := flag.NewFlagSet("lemonade", flag.ContinueOnError)
	flags.IntVar(&c.Port, "port", 2489, "TCP port number")
	flags.StringVar(&c.Allow, "allow", "0.0.0.0/0,::/0", "Allow IP range")
	flags.StringVar(&c.Host, "host", "localhost", "Destination host name.")
	flags.BoolVar(&c.Help, "help", false, "Show this message")
	flags.BoolVar(&c.TransLoopback, "trans-loopback", true, "Translate loopback address")
	flags.BoolVar(&c.TransLocalfile, "trans-localfile", true, "Translate local file")
	flags.StringVar(&c.LineEnding, "line-ending", "", "Convert Line Endings (CR/CRLF)")
	flags.BoolVar(&c.NoFallbackMessages, "no-fallback-messages", false, "Do not show fallback messages")
	flags.IntVar(&c.LogLevel, "log-level", 1, "Log level")
	return flags
}

func (c *CLI) GetConfPath() (string, error) {
	confPath, err := homedir.Expand(confFile)
	if err != nil {
		return confPath, err
	}

	file, err := os.Open(confPath)
	defer func() {
		file.Close()
	}()

	if os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(confPath), os.ModePerm)
		s := "#port=2489\n#allow=\"127.0.0.1,::1\"\n#line-ending=\"lf\""
		if runtime.GOOS == "windows" {
			s = "#port=2489\r\n#allow=\"127.0.0.1,::1\"\r\n#line-ending=\"crlf\""
		}
		ioutil.WriteFile(confPath, []byte(s), os.ModePerm)
	}

	return confPath, nil
}

func (c *CLI) parse(args []string, skip bool) error {
	flags := c.flags()

	confPath, err := c.GetConfPath()
	if err == nil && !skip {
		if confArgs, err := conflag.ArgsFrom(confPath); err == nil {
			flags.Parse(confArgs)
		}
	}

	if len(args) == 1 {
		return nil
	}

	var arg string
	err = flags.Parse(args[1:])
	if err != nil {
		return err
	}
	if c.Type == PASTE || c.Type == SERVER {
		return nil
	}

	for 0 < flags.NArg() {
		arg = flags.Arg(0)
		err := flags.Parse(flags.Args()[1:])
		if err != nil {
			return err
		}

	}

	if c.Help {
		return nil
	}

	if arg != "" {
		c.DataSource = arg
	} else {
		b, err := ioutil.ReadAll(c.In)
		if err != nil {
			return err
		}
		c.DataSource = string(b)
	}

	return nil
}
