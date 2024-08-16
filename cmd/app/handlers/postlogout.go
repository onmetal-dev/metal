package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type PostLogoutHandler struct {
	sessionStore sessions.Store
	sessionName  string
}

type PostLogoutHandlerParams struct {
	SessionStore sessions.Store
	SessionName  string
}

func NewPostLogoutHandler(params PostLogoutHandlerParams) *PostLogoutHandler {
	return &PostLogoutHandler{
		sessionStore: params.SessionStore,
		sessionName:  params.SessionName,
	}
}

func (h *PostLogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := h.sessionStore.Get(r, h.sessionName)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session.Options.MaxAge = -1
	if err := h.sessionStore.Save(r, w, session); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
