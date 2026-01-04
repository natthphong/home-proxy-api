package voice

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/config"

	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai/gpt"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
)

type Handler struct {
	ttsFunc gpt.SendTextAndGetAudioWithStyleFunc
	sttFunc gpt.SendAudioAndGetTextFunc
	s3      *s3.Client
	s3cfg   config.AwsS3Config
}
type TTSRequest struct {
	Prompt string  `json:"prompt,required"`
	Style  string  `json:"style"`
	Speed  float32 `json:"speed"`
}

func New(ttsFunc gpt.SendTextAndGetAudioWithStyleFunc, sttFunc gpt.SendAudioAndGetTextFunc, s3c *s3.Client, s3cfg config.AwsS3Config) *Handler {
	return &Handler{ttsFunc: ttsFunc, sttFunc: sttFunc, s3: s3c, s3cfg: s3cfg}
}

func (h *Handler) TTS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logz.NewLogger()
		ctx := c.Context()
		var (
			req TTSRequest
		)

		err := c.BodyParser(&req)
		if err != nil {
			return api.BadRequest(c, "Invalidate Body Request")
		}
		if req.Prompt == "" {
			return api.BadRequest(c, "Prompt is required")
		}
		resp, err := h.ttsFunc(ctx, logger, req.Prompt, req.Style, req.Speed)
		if err != nil {
			return api.InternalError(c, "Failed to send TTS")
		}
		key := fmt.Sprintf("%s%s", h.s3cfg.PrefixTTSVoice, uuid.New().String())

		ct := "audio/mpeg"
		_, err = h.s3.PutObject(c.Context(), &s3.PutObjectInput{
			Bucket:             &h.s3cfg.Bucket,
			Key:                &key,
			Body:               bytes.NewReader(resp),
			ContentType:        &ct,
			ContentDisposition: ptr("inline"),
		})
		if err != nil {
			return api.InternalError(c, fmt.Sprintf("upload failed: %v", err))
		}

		return api.Ok(c, fiber.Map{
			"key": key,
			"url": fmt.Sprintf("%s/file?key=%s", h.s3cfg.PublicBase, key),
		})
	}
}

func (h *Handler) SST() fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logz.NewLogger()
		ctx := c.Context()
		if !h.s3cfg.Enable {
			return api.BadRequest(c, "S3 is disabled")
		}

		fh, err := c.FormFile("file")
		if err != nil {
			return api.BadRequest(c, "missing file")
		}

		mediaType := c.FormValue("mediaType")
		if mediaType == "" {
			mediaType = mime.TypeByExtension(strings.ToLower(filepath.Ext(fh.Filename)))
			if mediaType == "" {
				mediaType = "application/octet-stream"
			}
		}
		if !isAllowedMedia(mediaType) {
			return api.BadRequest(c, "unsupported mediaType")
		}
		src, err := fh.Open()
		if err != nil {
			return api.InternalError(c, err.Error())
		}
		defer src.Close()
		audioBytes, err := io.ReadAll(src)
		if err != nil {
			return api.InternalError(c, err.Error())
		}

		resp, err := h.sttFunc(ctx, logger, audioBytes, fh.Filename, mediaType)
		if err != nil {
			return err
		}

		return api.Ok(c, fiber.Map{
			"text": resp,
		})
	}
}
func isAllowedMedia(mt string) bool {
	return strings.HasPrefix(mt, "audio/") ||
		mt == "application/octet-stream"
}

func ptr(s string) *string { return &s }
