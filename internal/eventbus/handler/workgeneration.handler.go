package handler

import (
	"context"
	"encoding/json"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/collection"
	labels_constant "personal_schedule_service/internal/constant/labels"
	workgeneration_constant "personal_schedule_service/internal/constant/work"
	event_models "personal_schedule_service/internal/eventbus/models"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/internal/grpc/validation"
	"personal_schedule_service/internal/repos"
	"strings"
	"time"

	"github.com/thanvuc/go-core-lib/eventbus"
	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"github.com/wagslane/go-rabbitmq"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.uber.org/zap"
)

type WorkGenerationHandler struct {
	logger            log.Logger
	workRepo          repos.WorkRepo
	labelRepo         repos.LabelRepo
	workValidator     validation.WorkValidator
	eventbusConnector *eventbus.RabbitMQConnector
	mongoConnector    *mongolib.MongoConnector
	publisher         eventbus.Publisher
}

func NewWorkGenerationHandler(
	workRepo repos.WorkRepo,
	workValidator validation.WorkValidator,
	labelRepo repos.LabelRepo,
) *WorkGenerationHandler {
	publisher := eventbus.NewPublisher(
		global.EventBusConnector,
		workgeneration_constant.NOTIFICATION_GENERATE_WORK_EXCHANGE,
		eventbus.ExchangeTypeDirect,
		nil,
		nil,
		false,
	)
	return &WorkGenerationHandler{
		logger:            global.Logger,
		workRepo:          workRepo,
		eventbusConnector: global.EventBusConnector,
		workValidator:     workValidator,
		labelRepo:         labelRepo,
		mongoConnector:    global.MongoDbConntector,
		publisher:         publisher,
	}
}

func (n *WorkGenerationHandler) ConsumeWorks(ctx context.Context, d rabbitmq.Delivery) rabbitmq.Action {
	workMessages, err := n.DecodeWorkMessage(d.Body)
	userId, ok := d.Headers["user_id"].(string)
	if !ok {
		n.logger.Error("missing user_id header", "")
		return rabbitmq.NackDiscard
	}

	messageId, ok := d.Headers["message_id"].(string)
	if !ok {
		n.logger.Error("missing message_id header", "")
		return rabbitmq.NackDiscard
	}

	if err != nil {
		n.logger.Error("Failed to decode work message", "")
		n.PublishErrorNotification(ctx, userId, messageId)
		return rabbitmq.NackDiscard
	}

	labelMap := make(map[string]collection.Label)
	labels, err := n.labelRepo.GetLabels(ctx)
	if err != nil {
		n.logger.Error("Failed to get labels", "")
		n.PublishErrorNotification(ctx, userId, messageId)
		return rabbitmq.NackDiscard
	}

	for _, label := range labels {
		if _, exists := labelMap[label.Key]; !exists {
			labelMap[label.Key] = label
		}
	}

	err = n.workValidator.ValidateWorkMessages(ctx, labelMap, workMessages)
	if err != nil {
		n.logger.Error("Work message validation failed", "")
		n.PublishErrorNotification(ctx, userId, messageId)
		return rabbitmq.NackDiscard
	}

	works := make([]*collection.Work, 0, len(workMessages))
	subTasks := make([]*collection.SubTask, 0)
	now := time.Now().UTC()
	draftId := labelMap[labels_constant.LabelDraft].ID

	for _, wm := range workMessages {
		startDate, err := utils.ParseLocalTimePtrToUTC(wm.StartDate, "2006-01-02 15:04")
		if err != nil {
			n.PublishErrorNotification(ctx, userId, messageId)
			return rabbitmq.NackDiscard
		}

		endDate, err := utils.ParseLocalTimeToUTC(wm.EndDate, "2006-01-02 15:04")
		if err != nil {
			n.PublishErrorNotification(ctx, userId, messageId)
			return rabbitmq.NackDiscard
		}

		work := collection.Work{
			ID:             bson.NewObjectID(),
			Name:           wm.Name,
			NameNormalized: strings.ToLower(strings.TrimSpace(wm.Name)),

			ShortDescriptions:   utils.ToStringPointer(wm.ShortDescriptions),
			DetailedDescription: utils.ToStringPointer(wm.DetailedDescription),

			StartDate: startDate,
			EndDate:   endDate,

			// map KEY → ID (giả sử bạn đã có mapping)
			DifficultyID: labelMap[wm.DifficultyKey].ID,
			PriorityID:   labelMap[wm.PriorityKey].ID,
			CategoryID:   labelMap[wm.CategoryKey].ID,

			// default values
			StatusID:       labelMap[labels_constant.LabelPending].ID,
			TypeID:         labelMap[labels_constant.LabelInDay].ID,
			UserID:         userId,
			CreatedAt:      now,
			LastModifiedAt: now,

			// Draft work - pending
			DraftID: &draftId,
		}

		for _, stm := range wm.SubTasks {
			subTask := collection.SubTask{
				ID:             bson.NewObjectID(),
				Name:           stm,
				IsCompleted:    false,
				WorkID:         work.ID,
				CreatedAt:      now,
				LastModifiedAt: now,
			}

			subTasks = append(subTasks, &subTask)
		}

		works = append(works, &work)
	}

	wc := writeconcern.Majority()
	txnOptions := options.Transaction().SetWriteConcern(wc)
	session, err := n.mongoConnector.Client.StartSession()
	if err != nil {
		n.logger.Error("Failed to start mongo session", "")
		n.PublishErrorNotification(ctx, userId, messageId)
		return rabbitmq.NackDiscard
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		workDocs := make([]interface{}, len(works))
		for i, work := range works {
			workDocs[i] = work
		}

		err := n.workRepo.BulkInsertWorks(ctx, workDocs)
		if err != nil {
			n.logger.Error("Failed to bulk insert works", "", zap.Error(err))
			return nil, err
		}

		subTaskDocs := make([]interface{}, len(subTasks))
		for i, subTask := range subTasks {
			subTaskDocs[i] = subTask
		}
		err = n.workRepo.BulkInsertSubTasks(ctx, subTaskDocs)
		if err != nil {
			n.logger.Error("Failed to bulk insert sub tasks", "", zap.Error(err))
			return nil, err
		}
		return nil, nil
	}, txnOptions)

	if err != nil {
		n.logger.Error("Transaction to insert works and subtasks failed", "")
		n.PublishErrorNotification(ctx, userId, messageId)
		return rabbitmq.NackDiscard
	}

	err = n.PublishSuccessNotification(ctx, userId, messageId)

	if err != nil {
		n.logger.Error("Failed to publish work generated notification", "")
		n.PublishErrorNotification(ctx, userId, messageId)
		return rabbitmq.NackDiscard
	}

	return rabbitmq.Ack
}

func (n *WorkGenerationHandler) DecodeWorkMessage(body []byte) ([]event_models.WorkMessage, error) {
	var works []event_models.WorkMessage

	if err := json.Unmarshal(body, &works); err != nil {
		return nil, err
	}

	return works, nil
}

func (n *WorkGenerationHandler) PublishErrorNotification(ctx context.Context, userId string, messageId string) error {
	logger := global.Logger
	notification := event_models.Notification{
		Title:   "Tạo công việc với AI thất bại",
		Message: "Hệ thống gặp lỗi khi tạo công việc cho bạn. Vui lòng thử lại sau.",

		SenderID:    "system",
		ReceiverIDs: []string{userId},

		CorrelationID:   messageId,
		CorrelationType: 2,

		Link:     utils.ToStringPointer(workgeneration_constant.LINK),
		ImageURL: utils.ToStringPointer(workgeneration_constant.IMGAGE_URL),
	}

	body, err := json.Marshal(notification)
	if err != nil {
		n.logger.Error("Failed to marshal error notification", "")
	}

	err = n.publisher.Publish(
		ctx,
		"work_generation_handler_error_notification",
		[]string{workgeneration_constant.NOTIFICATION_GENERATE_WORK_ROUTING_KEY},
		body,
		nil,
	)
	if err != nil {
		logger.Error("Failed to publish error notification", "")
	}
	return err
}

func (n *WorkGenerationHandler) PublishSuccessNotification(ctx context.Context, userId string, messageId string) error {
	notification := event_models.Notification{
		Title:           "Tạo công việc với AI thành công",
		Message:         "Hệ thống đã tạo công việc cho bạn thành công. Vui lòng kiểm tra trong ứng dụng.",
		SenderID:        "system",
		ReceiverIDs:     []string{userId},
		CorrelationID:   messageId,
		CorrelationType: 2,
		Link:            utils.ToStringPointer(workgeneration_constant.LINK),
		ImageURL:        utils.ToStringPointer(workgeneration_constant.IMGAGE_URL),
	}

	body, err := json.Marshal(notification)
	if err != nil {
		n.logger.Error("Failed to marshal success notification", "")
	}

	err = n.publisher.Publish(
		ctx,
		"work_generation_handler_success_notification",
		[]string{workgeneration_constant.NOTIFICATION_GENERATE_WORK_ROUTING_KEY},
		body,
		nil,
	)

	return err
}
