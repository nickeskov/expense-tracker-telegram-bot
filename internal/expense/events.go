package expense

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/proto/events"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/models"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventGenerateSummaryReportByCategories struct {
	ChatID int64
	UserID models.UserID
	Since  time.Time
	Till   time.Time
}

func (e *EventGenerateSummaryReportByCategories) MarshalBinary() (data []byte, err error) {
	event := &events.Event{Value: &events.Event_GenerateReport_{GenerateReport: &events.Event_GenerateReport{
		ChatId: e.ChatID,
		UserId: int64(e.UserID),
		Request: &events.Event_GenerateReport_ByCategories_{ByCategories: &events.Event_GenerateReport_ByCategories{
			Since: timestamppb.New(e.Since),
			Till:  timestamppb.New(e.Till),
		}},
	}}}
	return proto.Marshal(event)
}

func (e *EventGenerateSummaryReportByCategories) UnmarshalBinary(data []byte) error {
	event := &events.Event{}
	if err := proto.Unmarshal(data, event); err != nil {
		return err
	}
	genReportEvent, ok := event.GetValue().(*events.Event_GenerateReport_)
	if !ok {
		return errors.Errorf("unexpected protobuf event type (%T)", event)
	}
	genReportReq := genReportEvent.GenerateReport.GetRequest()
	genByCategoriesRequest, ok := genReportReq.(*events.Event_GenerateReport_ByCategories_)
	if !ok {
		return errors.Errorf("unexpected protobuf generate report by categories request type (%T)", genReportReq)
	}
	*e = EventGenerateSummaryReportByCategories{
		ChatID: genReportEvent.GenerateReport.GetChatId(),
		UserID: models.UserID(genReportEvent.GenerateReport.GetUserId()),
		Since:  genByCategoriesRequest.ByCategories.GetSince().AsTime(),
		Till:   genByCategoriesRequest.ByCategories.GetTill().AsTime(),
	}
	return nil
}
