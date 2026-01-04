package file

import (
	"bufio"
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
)

type Handler struct {
	s3    *s3.Client
	s3cfg config.AwsS3Config
}

func New(s3c *s3.Client, s3cfg config.AwsS3Config) *Handler {
	return &Handler{s3: s3c, s3cfg: s3cfg}
}

func (h *Handler) Upload() fiber.Handler {
	return func(c *fiber.Ctx) error {
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

		ext := strings.ToLower(filepath.Ext(fh.Filename))
		key := uuid.NewString() + ext

		_, err = h.s3.PutObject(c.Context(), &s3.PutObjectInput{
			Bucket:             &h.s3cfg.Bucket,
			Key:                &key,
			Body:               src,
			ContentType:        &mediaType,
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

func (h *Handler) Get() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Query("key")
		if key == "" {
			return api.BadRequest(c, "key is required")
		}

		out, err := h.s3.GetObject(c.Context(), &s3.GetObjectInput{
			Bucket: &h.s3cfg.Bucket,
			Key:    &key,
		})
		if err != nil {
			return api.InternalError(c, fmt.Sprintf("get failed: %v", err))
		}

		// headers
		if out.ContentType != nil && *out.ContentType != "" {
			c.Set("Content-Type", *out.ContentType)
		} else {
			c.Set("Content-Type", "application/octet-stream")
		}
		c.Set("Content-Disposition", "inline")

		if out.ContentLength != nil && *out.ContentLength != 0 {
			c.Set("Content-Length", fmt.Sprintf("%d", out.ContentLength))
		}

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			_, _ = io.Copy(w, out.Body)
			_ = out.Body.Close()
		})

		return nil
	}
}

func isAllowedMedia(mt string) bool {
	return strings.HasPrefix(mt, "image/") ||
		strings.HasPrefix(mt, "video/") ||
		strings.HasPrefix(mt, "audio/") ||
		mt == "application/octet-stream"
}

func ptr(s string) *string { return &s }
