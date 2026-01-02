package line

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/line_message"
	"strings"
)

type NotificationRequest struct {
	UserId       string  `json:"userId"`
	Message      string  `json:"message"`
	Url          string  `json:"url"`
	MessageType  string  `json:"messageType"`
	ChannelToken *string `json:"channelToken,omitempty"`
}

func NewLineNotificationHandler(
	notificationFunc line_message.SendMessageFunc,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logz.NewLogger()
		var req NotificationRequest
		err := c.BodyParser(&req)
		if err != nil {
			logger.Error(err.Error())
			return api.BadRequest(c, "Request Invalid")
		}
		var messages []messaging_api.MessageInterface
		if strings.ToLower(req.MessageType) == "text" {
			temp := messaging_api.TextMessage{
				Text: req.Message,
			}
			messages = append(messages, &temp)
		} else if strings.ToLower(req.MessageType) == "audio" {
			temp := messaging_api.AudioMessage{
				OriginalContentUrl: req.Url,
				Duration:           100,
			}
			messages = append(messages, &temp)
		} else {
			return api.BadRequest(c, "Message Type not supported")
		}
		if len(messages) == 0 {
			return api.BadRequest(c, "No Message For Send Notification")
		}
		_, err = notificationFunc(messages, req.UserId, uuid.NewString(), false, req.ChannelToken)
		if err != nil {
			logger.Error(err.Error())
			return api.InternalError(c, "Cannot send notification")
		}
		return api.Ok(c, nil)
	}
}
