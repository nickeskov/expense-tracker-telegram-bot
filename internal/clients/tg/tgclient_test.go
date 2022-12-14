package tg

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	clMock "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/mocks/clients"
	expMock "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/mocks/expense"
	userMock "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/mocks/user"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user"
	"gopkg.in/telebot.v3"
)

func newClient(ctx context.Context, t *testing.T, expUC expense.UseCase, userUC user.UseCase) *Client {
	cl, err := NewWithOptions("stub", "stub", []models.CurrencyCode{"stub"}, expUC, userUC, Options{offline: true})
	require.NoError(t, err)
	go cl.Start(ctx)
	t.Cleanup(cl.Stop)
	return cl
}

func Test_handleExpenseCmd(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	var (
		expUCMock   = expMock.NewMockUseCase(ctrl)
		userUCMock  = userMock.NewMockUseCase(ctrl)
		teleCtxMock = clMock.NewMocktelebotReducedContext(ctrl)
	)
	var (
		userID      = 11
		messageID   = 22
		comment     = "     sdf f fds"
		category    = "test"
		amount      = decimal.NewFromFloat(111.1)
		day         = time.Date(2022, time.October, 3, 0, 0, 0, 0, time.UTC)
		expectedExp = models.Expense{
			ID:       models.ExpenseID(messageID),
			Category: models.ExpenseCategory(category),
			Amount:   amount,
			Date:     day,
			Comment:  comment,
		}
	)
	args := []string{category, fmt.Sprintf("%v", amount), day.Format(dateLayout)}
	args = append(args, strings.Split(comment, " ")...)

	argCall := teleCtxMock.EXPECT().Args().Times(1).Return(args)
	msgCall := teleCtxMock.EXPECT().Message().Times(1).Return(&telebot.Message{
		ID:     messageID,
		Sender: &telebot.User{ID: int64(userID)},
	}).After(argCall)
	addExpCall := expUCMock.EXPECT().AddExpense(ctx, models.UserID(userID), expectedExp).MaxTimes(1).Return(expectedExp, nil).After(msgCall)
	teleCtxMock.EXPECT().Send("Expense successfully created").Times(1).After(addExpCall) // send call

	cl := newClient(ctx, t, expUCMock, userUCMock)
	err := cl.handleExpenseCmd(ctx, teleCtxMock)
	require.NoError(t, err)
}

func Test_handleExpensesReportCmd(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	var (
		expUCMock   = expMock.NewMockUseCase(ctrl)
		userUCMock  = userMock.NewMockUseCase(ctrl)
		teleCtxMock = clMock.NewMocktelebotReducedContext(ctrl)
	)
	var (
		userID    = 11
		messageID = 22
		since     = time.Date(2022, time.October, 3, 0, 0, 0, 0, time.UTC)
		till      = time.Date(2022, time.October, 4, 0, 0, 0, 0, time.UTC)
		report    = expense.SummaryReport{
			"cat1": decimal.NewFromFloat(111.1),
			"cat2": decimal.NewFromFloat(222.2),
			"aaa":  decimal.NewFromFloat(333.3),
		}
		reportMsg, err = report.Text()
	)
	require.NoError(t, err)

	argCall := teleCtxMock.EXPECT().Args().Times(1).Return([]string{since.Format(dateLayout), till.Format(dateLayout)})
	msgCall := teleCtxMock.EXPECT().Message().Times(1).Return(&telebot.Message{
		ID:     messageID,
		Sender: &telebot.User{ID: int64(userID)},
	}).After(argCall)
	reportCall := expUCMock.EXPECT().GetExpensesSummaryByCategorySince(ctx, models.UserID(userID), since, till).Times(1).
		Return(report, nil).After(msgCall)
	teleCtxMock.EXPECT().Send(reportMsg).Times(1).Return(nil).After(reportCall)

	cl := newClient(ctx, t, expUCMock, userUCMock)
	err = cl.handleExpensesReportCmd(ctx, teleCtxMock)
	require.NoError(t, err)
}

func Test_handleExpensesListCmd(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	var (
		expUCMock  = expMock.NewMockUseCase(ctrl)
		userUCMock = userMock.NewMockUseCase(ctrl)

		teleCtxMock = clMock.NewMocktelebotReducedContext(ctrl)
	)
	var (
		userID      = 11
		messageID   = 22
		since       = time.Date(2022, time.October, 3, 0, 0, 0, 0, time.UTC)
		till        = time.Date(2022, time.October, 4, 0, 0, 0, 0, time.UTC)
		comment     = "     sdf f fds"
		category    = "test"
		amount      = decimal.NewFromFloat(111.1)
		day         = time.Date(2022, time.October, 3, 0, 0, 0, 0, time.UTC)
		expectedExp = models.Expense{
			ID:       models.ExpenseID(messageID),
			Category: models.ExpenseCategory(category),
			Amount:   amount,
			Date:     day,
			Comment:  comment,
		}
	)

	argCall := teleCtxMock.EXPECT().Args().Times(1).Return([]string{since.Format(dateLayout), till.Format(dateLayout)})
	msgCall := teleCtxMock.EXPECT().Message().Times(1).Return(&telebot.Message{
		ID:     messageID,
		Sender: &telebot.User{ID: int64(userID)},
	}).After(argCall)
	reportCall := expUCMock.EXPECT().GetExpensesAscendSinceTill(ctx, models.UserID(userID), since, till, maxExpensesList).Times(1).
		Return([]models.Expense{expectedExp, expectedExp}, nil).After(msgCall)
	teleCtxMock.EXPECT().Send(printExpense(expectedExp)).Times(2).Return(nil).After(reportCall)

	cl := newClient(ctx, t, expUCMock, userUCMock)
	err := cl.handleExpensesListCmd(ctx, teleCtxMock)
	require.NoError(t, err)
}
