package api

import (
	"fmt"
	"log"
	"time"

	"github.com/Mshivam2409/hls-streamer/internal/db"
	"github.com/Mshivam2409/hls-streamer/internal/hls"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

type Question struct {
	Qid       string `json:"qid"`
	Uid       string `json:"uid"`
	Handshake string `json:"handshake"`
}

func Health(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func GetToken(c *fiber.Ctx) error {
	q := new(Question)

	if err := c.BodyParser(q); err != nil {
		return err
	}

	if q.Handshake != "HI" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Handshake"})
	}
	rid := uuid.NewV4().String()

	token, _ := GenerateToken(q.Qid, q.Uid, rid, 5)

	wavPath, err := db.WriteWAV(q.Qid)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	vttPath, err := db.WriteVTT(q.Qid)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err = hls.TranscodeHLS(wavPath, vttPath, rid); err != nil {
		log.Println(err)
		return err
	}

	db.GoStreamer.TTLCache.SetWithTTL(rid, fmt.Sprintf("%s/%s", viper.GetString("cache.static"), rid), time.Minute*time.Duration(2))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"token": token, "rid": rid})
}
