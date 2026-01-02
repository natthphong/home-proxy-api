package email

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"net/smtp"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	HTML    string `json:"html"`
}

const (
	smtpHost = "smtp.gmail.com"
	smtpPort = "587"
)

func NewEmailHandler(from, user, pass string) fiber.Handler {
	auth := smtp.PlainAuth("", user, pass, smtpHost)
	return func(c *fiber.Ctx) error {
		var req EmailRequest
		if err := c.BodyParser(&req); err != nil {
			return api.BadRequest(c, "invalid JSON body")
		}
		if req.To == "" || req.Subject == "" || (req.Message == "" && req.HTML == "") {
			return api.BadRequest(c, "missing to, subject, or message/html")
		}
		header := make(map[string]string)
		header["From"] = from
		header["To"] = req.To
		header["Subject"] = req.Subject
		header["MIME-Version"] = "1.0"
		header["Content-Type"] = "text/html; charset=UTF-8"

		msg := ""
		for k, v := range header {
			msg += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		msg += "\r\n"
		if req.HTML != "" {
			msg += req.HTML
		} else {
			msg += fmt.Sprintf("<p>%s</p>", req.Message)
		}
		if err := smtp.SendMail(
			smtpHost+":"+smtpPort,
			auth,
			from,
			[]string{req.To},
			[]byte(msg),
		); err != nil {
			return api.InternalError(c, fmt.Sprintf("sending failed: %v", err))
		}
		return api.Ok(c, fiber.Map{
			"message": "Email sent successfully",
		})
	}
}
