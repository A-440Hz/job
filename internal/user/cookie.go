package user

import (
	"net/http"
)

const (
	userCookieName   = "userCookie"
	userCookieExpiry = 3600 * 24 * 30 // 30 days
)

// GetUserIdFromCookie attempts to return the User ID from Session ID, clearing the cookie upon failure.
// it lookups the current day and updates session expiry every call.
func (s *Service) GetUserIdFromCookie(r *http.Request, w http.ResponseWriter) (string, error) {
	// https://www.alexedwards.net/blog/working-with-cookies-in-go
	cookie, err := r.Cookie(userCookieName)
	if err != nil {
		s.ClearUserCookie(w, "")
		return "", err
	}
	return s.getUserIDFromSession(cookie.Value, w)
}

func (s *Service) getUserIDFromSession(sID string, w http.ResponseWriter) (string, error) {
	sn, err := s.repo.lookupSession(sID)
	if err != nil {
		s.ClearUserCookie(w, sID)
		return "", err
	}
	// update repo session expiry
	s.UpdateSessionExpiry(sn)
	return sn.UserID, nil
}

func (s *Service) SetUserCookie(w http.ResponseWriter, sessionID string) {
	cookie := http.Cookie{
		Name:  userCookieName,
		Value: sessionID,
		// MaxAge:   userCookieExpiry,
		Secure:   true,
		HttpOnly: true, // prevents client-side JS from accessing the cookie
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}

// ClearUserCookie clears the user cookie and optionally, session, from the response
// TODO: make it make sense
func (s *Service) ClearUserCookie(w http.ResponseWriter, prevSession string) {
	cookie := http.Cookie{
		Name:     userCookieName,
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	s.repo.deleteSession(prevSession)
}
