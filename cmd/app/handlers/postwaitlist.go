package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/form"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
)

type PostWaitlistHandler struct {
	waitlistStore store.WaitlistStore
}

type PostWaitlistHandlerParams struct {
	WaitlistStore store.WaitlistStore
}

func NewPostWaitlistHandler(params PostWaitlistHandlerParams) *PostWaitlistHandler {
	return &PostWaitlistHandler{
		waitlistStore: params.WaitlistStore,
	}
}

func (h *PostWaitlistHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var f templates.JoinWaitlistFormData
	fieldErrors, err := form.Decode(&f, r)
	if fieldErrors.NotNil() || err != nil {
		templates.WaitlistForm(f, fieldErrors, err, "").Render(ctx, w)
		return
	}

	if err := h.waitlistStore.Add(ctx, f.Email); err != nil {
		if err != store.ErrDuplicateWaitlistEntry {
			// be very forgiving here... can recover from errors by looking at logs
			logger.FromContext(ctx).Error("error adding waitlist entry", "email", f.Email, "error", err)
		}
	}

	templates.WaitlistForm(f, fieldErrors, err, "you're on the list! 👍").Render(ctx, w)
}
