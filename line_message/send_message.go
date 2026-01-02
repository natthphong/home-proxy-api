package line_message

import (
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"time"
)

type SendMessageFunc func(messages []messaging_api.MessageInterface, userId, retryToken string, notificationDisabled bool, channelToken *string) (*messaging_api.PushMessageResponse, error)

func SendMessage(bots []*messaging_api.MessagingApiAPI) SendMessageFunc {
	return func(messages []messaging_api.MessageInterface, userId, retryToken string, notificationDisabled bool, channelToken *string) (*messaging_api.PushMessageResponse, error) {
		index := time.Now().Day() % len(bots)
		bot := bots[index]
		if channelToken != nil && *channelToken != "" {
			newBot, err := messaging_api.NewMessagingApiAPI(*channelToken)
			if err == nil {
				bot = newBot
			}
		}
		req := &messaging_api.PushMessageRequest{
			To:                   userId,
			Messages:             messages,
			NotificationDisabled: notificationDisabled,
		}
		return bot.PushMessage(req, retryToken)
	}
}

func SendMessageChoseBot(bots []*messaging_api.MessagingApiAPI, index int) SendMessageFunc {
	return func(messages []messaging_api.MessageInterface, userId, retryToken string, notificationDisabled bool, channelToken *string) (*messaging_api.PushMessageResponse, error) {
		bot := bots[index]
		req := &messaging_api.PushMessageRequest{
			To:                   userId,
			Messages:             messages,
			NotificationDisabled: notificationDisabled,
		}
		return bot.PushMessage(req, retryToken)
	}
}

type SendMessageAllUserFunc func(messages []messaging_api.MessageInterface, retryToken string, notification bool) (*map[string]interface{}, error)

func SendMessageAllUser(bot *messaging_api.MessagingApiAPI) SendMessageAllUserFunc {
	return func(messages []messaging_api.MessageInterface, retryToken string, notification bool) (*map[string]interface{}, error) {
		req := &messaging_api.BroadcastRequest{
			Messages:             messages,
			NotificationDisabled: true,
		}
		return bot.Broadcast(req, retryToken)
	}
}
