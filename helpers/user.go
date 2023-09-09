package helpers

import (
	"errors"
    "time"

	"app/models"

	"github.com/gofiber/fiber/v2"
)

func GetUserFromSession(c *fiber.Ctx) (models.User, error) {
    sessionId := c.Cookies("session_id", "")
    if sessionId == "" {
        return models.User{}, errors.New("Session not found");
    }
    db := ConnectDB()
    var session models.Session
    db.Raw("SELECT * FROM Sessions WHERE session_id = UNHEX(?) AND expires > ?", sessionId, time.Now()).Scan(&session)
    if session.Id == "" {
            redirectUrl := c.GetReqHeaders()["Hx-Current-Url"]
            if redirectUrl == "" {
                redirectUrl = c.Request().URI().String()
            }
            c.Cookie(&fiber.Cookie{
                Name: "post_login_redirect",
                Value: redirectUrl,
                Expires: time.Now().Add(time.Minute),
                Secure: true,
                HTTPOnly: true,
                SameSite: "Strict",
            })
        return models.User{}, errors.New("Session not found");
    }
    return models.BlobToUser(session.Data)
}
