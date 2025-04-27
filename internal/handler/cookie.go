package handler

import (
	"net/http"
)

const (
	userCookieName   = "userCookie"
	userCookieExpiry = 3600 * 24 * 30 // 30 days
)

// todo: probably incorporate encryption to get/set cookies
// https://www.alexedwards.net/blog/working-with-cookies-in-go
func getUserIdFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(userCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// if uuid is encrypted it definitely shouldn't be the same method as user credentials
func setUserCookie(w http.ResponseWriter, uuid string) error {
	cookie := http.Cookie{
		Name:     userCookieName,
		Value:    uuid,
		MaxAge:   userCookieExpiry,
		Secure:   true,
		HttpOnly: true, // prevents client-side JS from accessing the cookie
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	return nil
}

// clearUserCookie clears the user cookie from the response
func clearUserCookie(w http.ResponseWriter) error {
	cookie := http.Cookie{
		Name:     userCookieName,
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	return nil
}
