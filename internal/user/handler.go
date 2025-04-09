package user

import (
	"job/internal/db"
	"net/http"
)

const (
	userCookie                    = "userCookie"
	defaultUnregisteredUserExpiry = 3600 * 24 * 7
)

// todo: probably incorporate encryption to get/set cookies
// https://www.alexedwards.net/blog/working-with-cookies-in-go
func getUserCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie(userCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func setUserCookie(w http.ResponseWriter, r *http.Request) error {
	// probably separate this logic to better handle unregistered users
	uuid := db.NewPublicID(db.UserIdPrefix)
	cookie := http.Cookie{
		Name:     userCookie,
		Value:    uuid,
		MaxAge:   defaultUnregisteredUserExpiry,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	return nil
}
