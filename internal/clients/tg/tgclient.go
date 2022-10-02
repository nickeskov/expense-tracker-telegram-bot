package tg

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Client struct {
	bot    *telebot.Bot
	logger *log.Logger
}

type Options struct {
	Logger     *log.Logger
	LogUpdates bool
	BlackList  []int64
	WhiteList  []int64
}

func NewWithOptions(token string, opts Options) (*Client, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 5 * time.Second},
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, errors.Wrap(err, "creating bot")
	}
	logger := opts.Logger
	if logger == nil {
		logger = log.Default()
	}
	if opts.LogUpdates {
		bot.Use(middleware.Logger(logger))
	}
	if len(opts.WhiteList) != 0 {
		bot.Use(middleware.Whitelist(opts.WhiteList...))
	}
	if len(opts.BlackList) != 0 {
		bot.Use(middleware.Blacklist(opts.BlackList...))
	}
	client := &Client{
		bot:    bot,
		logger: logger,
	}
	initHandlers(client)
	return client, nil
}

const (
	helloMsg = "Hello!"
	stubMsg  = "STUB"
	helpMsg  = "List of supported commands:\n" +
		"/start - sends hello\n" +
		"/hello - sends hello\n" +
		"/help - prints this help\n" +
		"/expense - creates new expense. Usage: /expense <category - one word> <amount> <date - format 'yyyy:mm:dd'> <comment>\n" +
		"/report - summary report by category since some date. Usage: /report <category - one word> <date - format 'yyyy:mm:dd'>"
)

func initHandlers(c *Client) {
	c.bot.Handle("/start", func(teleCtx telebot.Context) error {
		return teleCtx.Send(helloMsg)
	})
	c.bot.Handle("/hello", func(teleCtx telebot.Context) error {
		return teleCtx.Send(helloMsg)
	})
	c.bot.Handle("/help", func(teleCtx telebot.Context) error {
		return teleCtx.Send(helpMsg)
	})
	c.bot.Handle("/expense", func(teleCtx telebot.Context) error {
		return teleCtx.Send(stubMsg)
	})
	c.bot.Handle("/report", func(teleCtx telebot.Context) error {
		return teleCtx.Send(stubMsg)
	})
	c.bot.Handle(telebot.OnText, func(teleCtx telebot.Context) error {
		msg := "Unsupported action or command\n\n" + helpMsg
		return teleCtx.Send(msg)
	})
}

func (c *Client) Start() {
	c.bot.Start()
}

func (c *Client) Stop() {
	c.bot.Stop()
}
