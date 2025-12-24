package user

import (
	"errors"
	my_db "job/internal/db"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	sessionCookieName      = "sessionCookie"
	sessionCookieExpiry    = 3600 * 24 * 365     // 1 year, but it refreshes every happy lookup
	cookieRefreshThreshold = 30 * 24 * time.Hour // only refresh cookie when within 30 days of expiry
)

var cookieDomain string // set from allowOrigin in production

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

// SetCookieDomain sets the domain for session cookies, extracted from the allowed origin URL
// This should be called during initialization with the allowOrigin value
func SetCookieDomain(allowOrigin string) {
	// Extract domain from URL (e.g., "https://app.railway.app" -> ".railway.app")
	// For localhost, leave empty to default to exact host match
	if allowOrigin == "" || allowOrigin == "localhost:5173" {
		cookieDomain = ""
		return
	}
	ao, _ := strings.CutPrefix("https://", allowOrigin)
	cookieDomain = ao
}

func (s *Service) getUserIDFromSession(sID string, w http.ResponseWriter) (string, error) {
	var sn *Session
	var err error

	// Retry the session lookup to handle transient connection errors
	retryErr := my_db.WithRetry(func() error {
		sn, err = s.repo.lookupSession(sID)
		return err
	})
	if retryErr == gorm.ErrRecordNotFound {
		s.ClearSessionCookie(w)
		return "", errors.New("session expired -- please try logging in again")
	} else if retryErr != nil {
		s.ClearSessionCookie(w)
		return "", retryErr
	}

	// Only refresh cookie when close to expiring
	if time.Until(sn.ExpiresAt) < cookieRefreshThreshold {
		s.UpdateSessionExpiry(sn)
		s.SetSessionCookie(w, sn.ID)
	}

	return sn.UserID, nil
}

func (s *Service) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		MaxAge:   sessionCookieExpiry,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,                  // prevents client-side JS from accessing the cookie
		SameSite: http.SameSiteNoneMode, // switch to this to try to get site to work on Edge browser
	}

	// Set Domain if configured (for cross-origin cookie support)
	if cookieDomain != "" {
		cookie.Domain = cookieDomain
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
		SameSite: http.SameSiteNoneMode,
	}

	// Set Domain if configured (must match the domain used when setting the cookie)
	if cookieDomain != "" {
		cookie.Domain = cookieDomain
	}

	http.SetCookie(w, &cookie)
}
