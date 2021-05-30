package app

import (
	"context"
	"fmt"
	"log"
	"mlock/onprem/hab"
	"mlock/onprem/sqs"
	"mlock/shared"
	"mlock/shared/protos/messaging"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type App struct {
	sqsClient *sqs.Client
}

func NewApp(sqsClient *sqs.Client) (*App, error) {
	return &App{
		sqsClient: sqsClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	shortDelay := 1 * time.Second
	longDelay := 1 * time.Minute

	ticker := time.NewTicker(shortDelay)

	for {
		select {
		case <-ctx.Done():
			log.Printf("context says it's time to quit, ending run loop")
			return nil
		case <-ticker.C:
			handledMessage, err := a.processMessage(ctx)
			if err != nil {
				return fmt.Errorf("error on process message: %s", err.Error())
			}

			nextDelay := longDelay
			if handledMessage {
				nextDelay = shortDelay
			}

			log.Printf("sleeping for %.0f seconds", nextDelay.Seconds())
			ticker.Reset(nextDelay)
		}
	}
}

func (a *App) processMessage(ctx context.Context) (bool, error) {
	messages, err := a.sqsClient.GetMessages(ctx)
	if err != nil {
		return false, fmt.Errorf("error getting messages: %s", err.Error())
	}

	for _, m := range messages {
		// For now, we want to clear every message regardless of how we handled it (or failed to do so).
		defer func(message types.Message) {
			a.sqsClient.AcknowledgeMessage(ctx, message)
		}(m)

		mType, ok := m.MessageAttributes[shared.SQSProtoMessageTypeKey]
		if !ok {
			// TODO: signal back error to the cloud
			return true, fmt.Errorf("no message type")
		}

		switch *mType.StringValue {
		case string((&messaging.HabCommand{}).ProtoReflect().Descriptor().FullName()):
			message, err := shared.DecodeMessageHabCommand(*m.Body)
			if err != nil {
				return true, fmt.Errorf("error decoding messages: %s", err.Error())
			}
			log.Printf("message description: %s", message.Description)

			resp, err := hab.ProcessCommand(ctx, message)
			if err != nil {
				return true, fmt.Errorf("error processing command: %s", err.Error())
			}
			log.Printf("response: %s", resp.Description)

			if err := a.sqsClient.SendMessage(ctx, resp); err != nil {
				return true, fmt.Errorf("error sending response: %s", err.Error())
			}

		default:
			// TODO: signal back error to the cloud
			return true, fmt.Errorf("unhandled message type: %s", *mType.StringValue)
		}
	}

	return len(messages) > 0, nil
}
