package tg

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Client struct {
	bot                *telebot.Bot
	baseCurr           models.CurrencyCode
	supportedCurr      map[models.CurrencyCode]struct{}
	supportedCurrSlice []models.CurrencyCode
	expUC              expense.UseCase
	userUC             user.UseCase
	logger             *log.Logger
}

type Options struct {
	Logger     *log.Logger
	LogUpdates bool
	BlackList  []int64
	WhiteList  []int64
	offline    bool
}

func NewWithOptions(
	token string,
	baseCurr models.CurrencyCode, supported []models.CurrencyCode,
	expUC expense.UseCase, userUC user.UseCase,
	opts Options,
) (*Client, error) {
	pref := telebot.Settings{
		Token:   token,
		Poller:  &telebot.LongPoller{Timeout: 5 * time.Second},
		Offline: opts.offline, // used for testing purposes
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
	supportedCurr := make(map[models.CurrencyCode]struct{})
	for _, code := range supported {
		supportedCurr[code] = struct{}{}
	}
	supportedCurr[baseCurr] = struct{}{}
	supported = supported[:0]
	for code := range supportedCurr {
		supported = append(supported, code)
	}
	sort.Slice(supported, func(i, j int) bool {
		return supported[i] < supported[j]
	})
	client := &Client{
		bot:                bot,
		baseCurr:           baseCurr,
		supportedCurr:      supportedCurr,
		supportedCurrSlice: supported,
		expUC:              expUC,
		userUC:             userUC,
		logger:             logger,
	}
	return client, nil
}

const (
	helloMsg           = "Hello!"
	startAlreadyWeKnow = "We already know each other ;)"
	startNowWeKnow     = "Hello! Now we know each other!"
	helpMsg            = "" +
		"List of supported commands:\n" +
		"/start - send hello and register new user with default selected currency\n" +
		"/hello - send hello\n" +
		"/help - print this help\n" +
		"/currency - show selected currency or change it to the new one. Usage: /currency <currency - optional>\n" +
		"/expense - create new expense. Usage: /expense <category - one word> <amount - float> <date - format 'yyyy.mm.dd'> <comment, optional>\n" +
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

func (c *Client) initHandlers(ctx context.Context) {
	wrap := func(handler func(_ context.Context, reducedCtx telebotReducedContext) error) func(telebot.Context) error {
		return func(teleCtx telebot.Context) error {
			return handler(ctx, teleCtx)
		}
	}
	c.bot.Handle("/start", wrap(c.handleStartCmd), requireArgsCountMiddleware(0, 0))
	c.bot.Handle("/hello", func(teleCtx telebot.Context) error {
		return teleCtx.Send(helloMsg)
	})
	c.bot.Handle("/help", func(teleCtx telebot.Context) error {
		return teleCtx.Send(helpMsg)
	})
	c.bot.Handle("/currency", wrap(c.handleCurrencyCmd), requireArgsCountMiddleware(0, 1))
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

func (c *Client) handleExpenseCmd(ctx context.Context, teleCtx telebotReducedContext) error {
	args := teleCtx.Args()
	if len(args) < 3 {
		return errors.New("not enough arguments to create expense")
	}
	category, strAmount, date, commentWords := args[0], args[1], args[2], args[3:]

	amount, err := decimal.NewFromString(strAmount)
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
	if _, err := c.expUC.AddExpense(ctx, userID, exp); err != nil {
		return errors.Wrapf(err, "failed to create expense for userID=%d", userID)
	}
	return teleCtx.Send("Expense successfully created")
}

func (c *Client) handleExpensesReportCmd(ctx context.Context, teleCtx telebotReducedContext) error {
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
	report, err := c.expUC.GetExpensesSummaryByCategorySince(ctx, userID, since, till)
	if err != nil {
		return errors.Wrapf(err, "failed to create expenses report for userID=%d", userID)
	}
	msg, err := report.Text()
	if err != nil {
		return errors.Wrapf(err, "failed to convert expenses report to text message for userID=%d", userID)
	}
	return teleCtx.Send(msg)
}

const maxExpensesList = 100

func printExpense(exp models.Expense) string {
	return fmt.Sprintf("%s %v %s %s", exp.Category, exp.Amount, exp.Date.Format(dateLayout), exp.Comment)
}

func (c *Client) handleExpensesListCmd(ctx context.Context, teleCtx telebotReducedContext) error {
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
	expenses, err := c.expUC.GetExpensesAscendSinceTill(ctx, userID, since, till, maxExpensesList)
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

func (c *Client) handleStartCmd(ctx context.Context, teleCtx telebotReducedContext) error {
	userID := models.UserID(teleCtx.Message().Sender.ID)
	u := models.NewUser(userID, c.baseCurr)
	exists, err := c.userUC.IsUserExists(ctx, userID)
	if err != nil {
		return errors.Wrapf(err, "failed to check whether user with ID=%d exists or not", userID)
	}
	if exists {
		return teleCtx.Send(startAlreadyWeKnow)
	}
	if _, err := c.userUC.CreateUser(ctx, u); err != nil {
		return errors.Wrapf(err, "failed to create user with ID=%d if not exist", userID)
	}
	return teleCtx.Send(startNowWeKnow)
}

func (c *Client) handleCurrencyCmd(ctx context.Context, teleCtx telebotReducedContext) error {
	args := teleCtx.Args()
	userID := models.UserID(teleCtx.Message().Sender.ID)
	if len(args) == 0 {
		curr, err := c.userUC.GetUserCurrency(ctx, userID)
		if err != nil {
			return errors.Wrapf(err, "failed to get currency for userID=%d", userID)
		}
		return teleCtx.Send(fmt.Sprintf("Your selected currency is %q", curr))
	}
	currency := models.CurrencyCode(args[0])
	if _, ok := c.supportedCurr[currency]; !ok {
		msg := fmt.Sprintf("Currency %q is not supported. Supported currencies: %v", currency, c.supportedCurrSlice)
		return teleCtx.Send(msg)
	}
	if err := c.userUC.ChangeUserCurrency(ctx, userID, currency); err != nil {
		return errors.Wrapf(err, "failed to change currency to %q for user with ID=%d)", currency, userID)
	}
	return teleCtx.Send(fmt.Sprintf("Currency successfully changed to %q", currency))
}

func (c *Client) Start(ctx context.Context) {
	c.initHandlers(ctx)
	c.bot.Start()
}

func (c *Client) Stop() {
	c.bot.Stop()
}
