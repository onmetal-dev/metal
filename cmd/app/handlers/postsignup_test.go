package handlers

import (
	"bytes"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onmetal-dev/metal/lib/store"
	storemock "github.com/onmetal-dev/metal/lib/store/mock"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRegisterUserHandler(t *testing.T) {

	// assume user is invited. TODO: test uninvited registers
	inviteStore := &storemock.InviteStoreMock{}
	inviteStore.On("Get", "test@example.com").Return(&store.InvitedUser{Email: "test@example.com"}, nil)

	testCases := []struct {
		name               string
		email              string
		password           string
		createUserError    error
		expectedStatusCode int
		expectedBody       []byte
	}{
		{
			name:               "success",
			email:              "test@example.com",
			password:           "password",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "fail - error creating user",
			email:              "test@example.com",
			password:           "password",
			createUserError:    gorm.ErrDuplicatedKey,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			userStore := &storemock.UserStoreMock{}

			userStore.On("CreateUser", tc.email, tc.password).Return(tc.createUserError)

			handler := NewPostSignUpHandler(userStore, inviteStore, &storemock.TeamStoreMock{})
			body := bytes.NewBufferString("email=" + tc.email + "&password=" + tc.password)
			req, _ := http.NewRequest("POST", "/", body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(tc.expectedStatusCode, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, tc.expectedStatusCode)
			userStore.AssertExpectations(t)
		})
	}
}
