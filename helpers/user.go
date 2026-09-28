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
        setPostLoginRedirect(c)
        return models.User{}, errors.New("Session not found");
    }

    db := ConnectDB()
    var session models.Session
    if err := db.Raw("SELECT * FROM Sessions WHERE session_id = UNHEX(?) AND expires > ?", sessionId, time.Now()).Scan(&session).Error; err != nil {
        setPostLoginRedirect(c)
        return models.User{}, err
    }
    if session.Id == "" {
        setPostLoginRedirect(c)
        return models.User{}, errors.New("Session not found");
    }

    user, err := models.BlobToUser(session.Data)
    if err != nil {
        setPostLoginRedirect(c)
        return models.User{}, err
    }
    return user, nil
}

// setPostLoginRedirect remembers where to come back to after signing in. It is
// only written when there is no usable session: writing it on every request put
// a Set-Cookie on every response, which stops anything in front of the app from
// caching.
func setPostLoginRedirect(c *fiber.Ctx) {
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
}
