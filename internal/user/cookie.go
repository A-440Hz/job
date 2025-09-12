package user

import (
	"errors"
	"net/http"

	"gorm.io/gorm"
)

const (
	sessionCookieName   = "sessionCookie"
	sessionCookieExpiry = 3600 * 24 * 365 // 1 year, but it refreshes every happy lookup
)

// GetUserIDFromCookie attempts to return the User ID from Session ID, clearing the cookie upon failure.
// it lookups the current day and updates session and cookie expiry every successful call.
// it clears the cookie on lookup failure
func (s *Service) GetUserIDFromCookie(r *http.Request, w http.ResponseWriter) (string, error) {
	// https://www.alexedwards.net/blog/working-with-cookies-in-go
	cookie, err := r.Cookie(sessionCookieName)
	// TODO: this is mutative behavior and probably wrong
	if err != nil {
		s.ClearSessionCookie(w)
		return "", err
	}
	return s.getUserIDFromSession(cookie.Value, w)
}

// GetSessionIDFromCookie just returns the session id from the cookie and does nothing otherwise
func (s *Service) GetSessionIDFromCookie(r *http.Request, w http.ResponseWriter) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s *Service) getUserIDFromSession(sID string, w http.ResponseWriter) (string, error) {
	sn, err := s.repo.lookupSession(sID)
	if err == gorm.ErrRecordNotFound {
		s.ClearSessionCookie(w)
		return "", errors.New("session expired -- please try logging in again")
	} else if err != nil {
		s.ClearSessionCookie(w)
		return "", err
	}
	// update repo session expiry
	s.UpdateSessionExpiry(sn)
	// update cookie expiry
	s.SetSessionCookie(w, sn.ID)
	return sn.UserID, nil
}

func (s *Service) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		MaxAge:   sessionCookieExpiry,
		Path:     "/",
		Secure:   true,
		HttpOnly: true, // prevents client-side JS from accessing the cookie
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}

// ClearSessionCookie clears the session cookie
func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}
