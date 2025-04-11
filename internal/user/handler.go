package user

import (
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

// if uuid is encrypted it definitely shouldn't be the same method as user credentials
func setUserCookie(w http.ResponseWriter, uuid string) error {
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
