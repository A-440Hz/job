package handler

import (
	"fmt"
	"job/internal/db"
	"net/http"
)

// todo: probably incorporate encryption to get/set cookies
// https://www.alexedwards.net/blog/working-with-cookies-in-go
func getUserCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie("userCookie")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func setUserCookie() error {
	uuid := db.NewID(db.UserIdPrefix)
	fmt.Println(uuid)
	return nil
}
