package shared

import (
	"encoding/base64"
	"fmt"
	"mlock/shared/protos/messaging"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func EncodeMessage(message protoreflect.ProtoMessage) (string, error) {
	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("error marshalling message: %s", err.Error())
	}

	return base64.StdEncoding.EncodeToString(messageBytes), nil
}

func DecodeMessage(encodedMessage string, out protoreflect.ProtoMessage) error {
	messageBytes, err := base64.StdEncoding.DecodeString(encodedMessage)
	if err != nil {
		return fmt.Errorf("error decoding message: %s", err.Error())
	}

	if err := proto.Unmarshal(messageBytes, out); err != nil {
		return fmt.Errorf("error unmarshalling message: %s", err.Error())
	}

	return nil
}

func DecodeMessageHabCommand(encodedMessage string) (*messaging.HabCommand, error) {
	messageBytes, err := base64.StdEncoding.DecodeString(encodedMessage)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %s", err.Error())
	}

	message := &messaging.HabCommand{}
	if err := proto.Unmarshal(messageBytes, message); err != nil {
		return nil, fmt.Errorf("error unmarshalling message: %s", err.Error())
	}

	return message, nil
}

func DecodeMessageOnPremHabCommandResponse(encodedMessage string) (*messaging.OnPremHabCommandResponse, error) {
	messageBytes, err := base64.StdEncoding.DecodeString(encodedMessage)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %s", err.Error())
	}

	message := &messaging.OnPremHabCommandResponse{}
	if err := proto.Unmarshal(messageBytes, message); err != nil {
		return nil, fmt.Errorf("error unmarshalling message: %s", err.Error())
	}

	return message, nil
}

func HabCommandListThings(description string) *messaging.HabCommand {
	return &messaging.HabCommand{
		Description: description,
		Request: &messaging.HTTPRequest{
			Method: "GET",
			Path:   "/rest/things",
		},
		CommandType: messaging.HabCommand_LIST_THINGS,
	}
}
