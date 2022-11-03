package tg

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
	"go.uber.org/zap"
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
	logger             *zap.Logger
}

type Options struct {
	Logger     *zap.Logger
	LogUpdates bool
	BlackList  []int64
	WhiteList  []int64
	Debug      bool
	offline    bool
}

func NewWithOptions(
	token string,
	baseCurr models.CurrencyCode, supported []models.CurrencyCode,
	expUC expense.UseCase, userUC user.UseCase,
	opts Options,
) (*Client, error) {
	logger := opts.Logger
	if logger == nil {
		logger = zap.L()
	}

	pref := telebot.Settings{
		Token:   token,
		Poller:  &telebot.LongPoller{Timeout: 5 * time.Second},
		Offline: opts.offline, // used for testing purposes
		OnError: func(err error, teleCtx telebot.Context) {
			var (
				updateID  = teleCtx.Update().ID
				messageID *int
				senderID  *int64
			)
			if msg := teleCtx.Message(); msg != nil {
				messageID = &msg.ID
			}
			if sender := teleCtx.Sender(); sender != nil {
				senderID = &sender.ID
			}
			logger.Error("Unknown error occurred in bot handler",
				zap.Int("update_id", updateID),
				zap.Intp("message_id", messageID),
				zap.Int64p("sender_id", senderID),
				zap.Error(err),
			)
		},
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, errors.Wrap(err, "creating bot")
	}
	if opts.LogUpdates {
		bot.Use(createIncomingUpdatesLoggerMiddleware(logger))
	}
	if len(opts.WhiteList) != 0 {
		bot.Use(middleware.Whitelist(opts.WhiteList...))
	}
	if len(opts.BlackList) != 0 {
		bot.Use(middleware.Blacklist(opts.BlackList...))
	}
	if opts.Debug {
		bot.Use(debugMiddleware)
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
	helloMsg                      = "Hello!"
	startAlreadyWeKnowMsg         = "We already know each other ;)"
	startNowWeKnowMsg             = "Hello! Now we know each other!"
	unknownUserMsg                = "Hello! Please, press /start to introduce yourself."
	noExpensesFoundMsg            = "No expenses found."
	expensesAmountExceededMsg     = "Can't add expense. Expenses amount exceeded."
	expenseAmountIsNotPositiveMsg = "Please, provide positive expense amount."
	expenseAmountIsTooBigMsg      = "Expense amount is too big"
	monthlyLimitIsNegativeMsg     = "Please, provide not negative limit amount or absense of amount."
	monthlyLimitIsTooBigMsg       = "Monthly limit is too big."
)

const (
	noneUserMonthlyLimitValue = "none"
)

func makeHelpMsg(baseCurr models.CurrencyCode) string {
	const helpMsgFormat = "" +
		"List of supported commands:\n" +
		"/start - send hello and register new user with default selected currency %q\n" +
		"/hello - send hello\n" +
		"/help - print this help\n" +
		"/currency - show selected currency or change it to the new one. Usage: /currency <currency - optional>\n" +
		"/expense - create new expense. Usage: /expense <category - one word> <amount - float> <date - format 'yyyy.mm.dd'> <comment, optional>\n" +
		"/report - summary report by categories since and till some dates. Usage: /report <since - format 'yyyy.mm.dd'> <till - format 'yyyy.mm.dd'>\n" +
		"/list - list expenses since and till some dates. Usage: /list <since - format 'yyyy.mm.dd'> <till - format 'yyyy.mm.dd'>\n" +
		"/limit - show expenses amount monthly limit in default currency %q or change it to another one. Usage: /limit <amount - float or '%s', optional>\n"
	return fmt.Sprintf(helpMsgFormat, baseCurr, baseCurr, noneUserMonthlyLimitValue)
}

func makeDefaultMsg(baseCurr models.CurrencyCode) string {
	return "Unsupported action or command\n\n" + makeHelpMsg(baseCurr)
}

func (c *Client) initHandlers(ctx context.Context) {
	checkUser := createIsUserExistsMiddleware(ctx, c.userUC)

	c.handle(ctx, "/hello", func(_ context.Context, teleCtx telebotReducedContext) error {
		return teleCtx.Send(helloMsg)
	})
	c.handle(ctx, "/help", func(_ context.Context, teleCtx telebotReducedContext) error {
		return teleCtx.Send(makeHelpMsg(c.baseCurr))
	})
	c.handle(ctx, telebot.OnText, func(_ context.Context, teleCtx telebotReducedContext) error {
		return teleCtx.Send(makeDefaultMsg(c.baseCurr))
	})
	c.handle(ctx, "/start", c.handleStartCmd, createRequireArgsCountMiddleware(0, 0))
	c.handle(ctx, "/currency", c.handleCurrencyCmd, checkUser, createRequireArgsCountMiddleware(0, 1))
	c.handle(ctx, "/expense", c.handleExpenseCmd, checkUser, createRequireArgsCountMiddleware(3, 258))
	c.handle(ctx, "/report", c.handleExpensesReportCmd, checkUser, createRequireArgsCountMiddleware(2, 2))
	c.handle(ctx, "/list", c.handleExpensesListCmd, checkUser, createRequireArgsCountMiddleware(2, 2))
	c.handle(ctx, "/limit", c.handleLimitCmd, checkUser, createRequireArgsCountMiddleware(0, 1))
}

type endpointHandler func(context.Context, telebotReducedContext) error

func (c *Client) handle(ctx context.Context, endpoint string, handler endpointHandler, m ...telebot.MiddlewareFunc) {
	logTriggeredHandler := createTriggeredHandlerLoggerMiddleware(c.logger, endpoint)
	metricsMiddleware := createEndpointMetricsMiddleware(endpoint)
	tracingMiddleware := createEndpointTracingMiddleware(endpoint)
	wrap := func(inner func(context.Context, telebotReducedContext) error) telebot.HandlerFunc {
		innerWithTracing := tracingMiddleware(inner)
		return logTriggeredHandler(metricsMiddleware(func(teleCtx telebot.Context) error {
			return innerWithTracing(ctx, teleCtx)
		}))
	}
	c.bot.Handle(endpoint, wrap(handler), m...)
}

type telebotReducedContext interface {
	Args() []string
	Send(what interface{}, opts ...interface{}) error
	Update() telebot.Update
	Message() *telebot.Message
	Sender() *telebot.User
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
		return teleCtx.Send(fmt.Sprintf("Failed to parse amount: %v", err))
	}

	day, err := time.Parse(dateLayout, date)
	if err != nil {
		return teleCtx.Send(fmt.Sprintf("Failed to parse date: %v", err))
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
	if err := exp.Validate(); err != nil {
		switch {
		case errors.Is(err, models.ErrExpenseAmountTooBig):
			return teleCtx.Send(expenseAmountIsTooBigMsg)
		case errors.Is(err, models.ErrExpenseAmountIsNotPositive):
			return teleCtx.Send(expenseAmountIsNotPositiveMsg)
		default:
			return errors.Wrapf(err, "unknown expense validation error")
		}
	}
	userID := models.UserID(teleMsg.Sender.ID)
	if _, err := c.expUC.AddExpense(ctx, userID, exp); err != nil {
		switch {
		case errors.Is(err, expense.ErrExpensesMonthlyLimitExcess):
			return teleCtx.Send(expensesAmountExceededMsg)
		default:
			return errors.Wrapf(err, "failed to create expense for userID=%d", userID)
		}
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
		return teleCtx.Send(fmt.Sprintf("Failed to parse since date: %v", err))
	}
	till, err := time.Parse(dateLayout, tillStr)
	if err != nil {
		return teleCtx.Send(fmt.Sprintf("Failed to parse till date: %v", err))
	}
	userID := models.UserID(teleCtx.Message().Sender.ID)
	report, err := c.expUC.GetExpensesSummaryByCategorySince(ctx, userID, since, till)
	if err != nil {
		return errors.Wrapf(err, "failed to create expenses report for userID=%d", userID)
	}
	if len(report) == 0 {
		return teleCtx.Send(noExpensesFoundMsg)
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
		return teleCtx.Send(fmt.Sprintf("Failed to parse since date: %v", err))
	}
	till, err := time.Parse(dateLayout, tillStr)
	if err != nil {
		return teleCtx.Send(fmt.Sprintf("Failed to parse till date: %v", err))
	}
	userID := models.UserID(teleCtx.Message().Sender.ID)
	expenses, err := c.expUC.GetExpensesAscendSinceTill(ctx, userID, since, till, maxExpensesList)
	if err != nil {
		return errors.Wrapf(err, "failed to create expenses report for userID=%d", userID)
	}
	if len(expenses) == 0 {
		return teleCtx.Send(noExpensesFoundMsg)
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
		return teleCtx.Send(startAlreadyWeKnowMsg)
	}
	if _, err := c.userUC.CreateUser(ctx, u); err != nil {
		return errors.Wrapf(err, "failed to create user with ID=%d if not exist", userID)
	}
	return teleCtx.Send(startNowWeKnowMsg)
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

func (c *Client) handleLimitCmd(ctx context.Context, teleCtx telebotReducedContext) error {
	args := teleCtx.Args()
	userID := models.UserID(teleCtx.Message().Sender.ID)
	if len(args) == 0 {
		limit, err := c.userUC.GetUserMonthlyLimit(ctx, userID)
		if err != nil {
			return errors.Wrapf(err, "failed to get monthly limit for userID=%d", userID)
		}
		limitArg := noneUserMonthlyLimitValue
		if limit != nil {
			limitArg = fmt.Sprintf("%v", *limit)
		}
		return teleCtx.Send(fmt.Sprintf("Your monthly limit is %q in %q", limitArg, c.baseCurr))
	}
	var limit *decimal.Decimal
	if limitArg := args[0]; limitArg != noneUserMonthlyLimitValue {
		limitValue, err := decimal.NewFromString(limitArg)
		if err != nil {
			return teleCtx.Send(fmt.Sprintf("Failed to parse monthly limit: %v", err))
		}
		limit = &limitValue
	}
	if err := models.ValidateUserMonthlyLimit(limit); err != nil {
		switch {
		case errors.Is(err, models.ErrUserMonthlyLimitTooBig):
			return teleCtx.Send(monthlyLimitIsTooBigMsg)
		case errors.Is(err, models.ErrUserMonthlyLimitIsNegative):
			return teleCtx.Send(monthlyLimitIsNegativeMsg)
		default:
			return errors.Wrapf(err, "unknown user monthly limit validation error")
		}
	}
	if err := c.userUC.SetUserMonthlyLimit(ctx, userID, limit); err != nil {
		return errors.Wrapf(err, "failed to set monthly limit for userID=%d", userID)
	}
	return teleCtx.Send("Monthly limit successfully set")
}

func (c *Client) Start(ctx context.Context) {
	c.initHandlers(ctx)
	c.bot.Start()
}

func (c *Client) Stop() {
	c.bot.Stop()
}
