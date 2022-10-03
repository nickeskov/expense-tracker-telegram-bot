package tg

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Client struct {
	bot    *telebot.Bot
	uc     expense.UseCase
	logger *log.Logger
}

type Options struct {
	Logger     *log.Logger
	LogUpdates bool
	BlackList  []int64
	WhiteList  []int64
}

func NewWithOptions(token string, uc expense.UseCase, opts Options) (*Client, error) {
	return newWithOfflineOption(token, uc, opts, false)
}

func newWithOfflineOption(token string, uc expense.UseCase, opts Options, offline bool) (*Client, error) {
	pref := telebot.Settings{
		Token:   token,
		Poller:  &telebot.LongPoller{Timeout: 5 * time.Second},
		Offline: offline, // used for testing purposes
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
		uc:     uc,
		logger: logger,
	}
	client.initHandlers()
	return client, nil
}

const (
	helloMsg = "Hello!"
	helpMsg  = "List of supported commands:\n" +
		"/start - sends hello\n" +
		"/hello - sends hello\n" +
		"/help - prints this help\n" +
		"/expense - creates new expense. Usage: /expense <category - one word> <amount - float> <date - format 'yyyy.mm.dd'> <comment, optional>\n" +
		"/report - summary report by categories since and till some dates. Usage: /report <since - format 'yyyy.mm.dd'> <till - format 'yyyy.mm.dd'>\n" +
		"/list - list expenses since and till some dates. Usage: /list <since - format 'yyyy.mm.dd'> <till - format 'yyyy.mm.dd'>\n"
)

func requireArgsCountMiddleware(minArgsCount, maxArgsCount int) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(teleCtx telebot.Context) error {
			args := teleCtx.Args()
			l := len(args)
			if l < minArgsCount {
				return teleCtx.Send(fmt.Sprintf("Not enough arguments: minumum required %d, provided %d", minArgsCount, l))
			}
			if l > maxArgsCount {
				return teleCtx.Send(fmt.Sprintf("Too many arguments: maximum allowed %d, provided %d", maxArgsCount, l))
			}
			return next(teleCtx)
		}
	}
}

func (c *Client) initHandlers() {
	wrap := func(handler func(reducedCtx telebotReducedContext) error) func(telebot.Context) error {
		return func(teleCtx telebot.Context) error {
			return handler(teleCtx)
		}
	}
	c.bot.Handle("/start", func(teleCtx telebot.Context) error {
		return teleCtx.Send(helloMsg)
	})
	c.bot.Handle("/hello", func(teleCtx telebot.Context) error {
		return teleCtx.Send(helloMsg)
	})
	c.bot.Handle("/help", func(teleCtx telebot.Context) error {
		return teleCtx.Send(helpMsg)
	})
	c.bot.Handle("/expense", wrap(c.handleExpenseCmd), requireArgsCountMiddleware(3, 258))
	c.bot.Handle("/report", wrap(c.handleExpensesReportCmd), requireArgsCountMiddleware(2, 2))
	c.bot.Handle("/list", wrap(c.handleExpensesListCmd), requireArgsCountMiddleware(2, 2))
	c.bot.Handle(telebot.OnText, func(teleCtx telebot.Context) error {
		msg := "Unsupported action or command\n\n" + helpMsg
		return teleCtx.Send(msg)
	})
}

type telebotReducedContext interface {
	Args() []string
	Send(what interface{}, opts ...interface{}) error
	Message() *telebot.Message
}

const dateLayout = "2006.01.02"

func (c *Client) handleExpenseCmd(teleCtx telebotReducedContext) error {
	args := teleCtx.Args()
	if len(args) < 3 {
		return errors.New("not enough arguments to create expense")
	}
	category, strAmount, date, commentWords := args[0], args[1], args[2], args[3:]
	amount, err := strconv.ParseFloat(strAmount, 64)
	if err != nil {
		return teleCtx.Send(fmt.Sprint("Failed to parse amount:", err))
	}
	day, err := time.Parse(dateLayout, date)
	if err != nil {
		return teleCtx.Send(fmt.Sprint("Failed to parse date:", err))
	}
	comment := strings.Join(commentWords, " ")
	teleMsg := teleCtx.Message()
	exp := models.Expense{
		ID:       models.ExpenseID(teleMsg.ID),
		Category: models.ExpenseCategory(category),
		Amount:   amount,
		Date:     day,
		Comment:  comment,
	}
	userID := models.UserID(teleMsg.Sender.ID)
	if _, err := c.uc.AddExpense(userID, exp); err != nil {
		return errors.Wrapf(err, "failed to create expense for userID=%d", userID)
	}
	return teleCtx.Send("Expense successfully created")
}

func (c *Client) handleExpensesReportCmd(teleCtx telebotReducedContext) error {
	args := teleCtx.Args()
	if len(args) < 2 {
		return errors.New("not enough arguments to create expenses report")
	}
	sinceStr, tillStr := args[0], args[1]
	since, err := time.Parse(dateLayout, sinceStr)
	if err != nil {
		return teleCtx.Send(fmt.Sprint("Failed to parse since date:", err))
	}
	till, err := time.Parse(dateLayout, tillStr)
	if err != nil {
		return teleCtx.Send(fmt.Sprint("Failed to parse till date:", err))
	}
	userID := models.UserID(teleCtx.Message().Sender.ID)
	report, err := c.uc.ExpensesSummaryByCategorySince(userID, since, till)
	if err != nil {
		return errors.Wrapf(err, "failed to create expenses report for userID=%d", userID)
	}
	return teleCtx.Send(report.Text())
}

const maxExpensesList = 100

func printExpense(exp models.Expense) string {
	return fmt.Sprintf("%s %f %s %s", exp.Category, exp.Amount, exp.Date.Format(dateLayout), exp.Comment)
}

func (c *Client) handleExpensesListCmd(teleCtx telebotReducedContext) error {
	args := teleCtx.Args()
	if len(args) < 2 {
		return errors.New("not enough arguments to create expenses list")
	}
	sinceStr, tillStr := args[0], args[1]
	since, err := time.Parse(dateLayout, sinceStr)
	if err != nil {
		return teleCtx.Send(fmt.Sprint("Failed to parse since date:", err))
	}
	till, err := time.Parse(dateLayout, tillStr)
	if err != nil {
		return teleCtx.Send(fmt.Sprint("Failed to parse till date:", err))
	}
	userID := models.UserID(teleCtx.Message().Sender.ID)
	expenses, err := c.uc.ExpensesAscendSinceTill(userID, since, till, maxExpensesList)
	if err != nil {
		return errors.Wrapf(err, "failed to create expenses report for userID=%d", userID)
	}
	for _, exp := range expenses {
		msg := printExpense(exp)
		if err := teleCtx.Send(msg); err != nil {
			return errors.Wrapf(err, "failed to send for userID=%d category info '%s'", userID, msg)
		}
	}
	return nil
}

func (c *Client) Start() {
	c.bot.Start()
}

func (c *Client) Stop() {
	c.bot.Stop()
}
