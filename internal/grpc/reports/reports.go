package reports

import (
	"context"
	"math/big"
	"net"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/proto/api"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/proto/types"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type MessageSender interface {
	SendMessage(chatID int64, message string) error
}

type Service struct {
	api.UnimplementedReportsServiceServer
	msgSender MessageSender
	logger    *zap.Logger
}

func NewService(msgSender MessageSender, logger *zap.Logger) (*Service, error) {
	return &Service{msgSender: msgSender, logger: logger}, nil
}

func (s *Service) SendReport(ctx context.Context, r *api.SendReportRequest) (*api.SendReportResponse, error) {
	// TODO: consider adding metrics
	switch report := r.GetReport().GetValue().(type) {
	case *types.Report_ByCategories_:
		value := report.ByCategories.GetValue()
		summaryReport := make(expense.SummaryReport, len(value))
		for category, sum := range value {
			cat := models.ExpenseCategory(category)
			summaryReport[cat] = decimal.NewFromBigInt(new(big.Int).SetBytes(sum.Mantissa), sum.Exponent)
		}
		msg, err := summaryReport.Text()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create report by categories text representation")
		}
		if err := s.msgSender.SendMessage(r.ChatId, msg); err != nil {
			return nil, errors.Wrapf(err, "failed to send message to chaiID=%d for userID=%v", r.ChatId, r.UserId)
		}
		s.logger.Info("Report successfully sent to chat", zap.Int64("chatID", r.ChatId), zap.Int64p("userID", r.UserId))
		return &api.SendReportResponse{}, nil
	case nil:
		return nil, status.Errorf(codes.InvalidArgument, "<nil> report value")
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported type of report (%T)", report)
	}
}

func (s *Service) Serve(ctx context.Context, l net.Listener) error {
	server := grpc.NewServer()
	api.RegisterReportsServiceServer(server, s)
	reflection.Register(server)
	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down gRPC reports server...")
		server.GracefulStop()
	}()
	s.logger.Info("Starting gRPC reports server", zap.String("address", l.Addr().String()))
	return server.Serve(l)
}
