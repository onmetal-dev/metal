package handlers

import (
	"bytes"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	hashmock "github.com/onmetal-dev/metal/cmd/app/hash/mock"
	"github.com/onmetal-dev/metal/lib/store"
	storemock "github.com/onmetal-dev/metal/lib/store/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestLogin(t *testing.T) {

	user := &store.User{Common: store.Common{Id: "user_asdf"}, Email: "test@example.com", Password: "password", TeamMemberships: []store.TeamMember{{TeamId: "team_test", Role: store.TeamRoleAdmin}}}
	teamStore := &storemock.TeamStoreMock{}
	teamStore.On("GetTeam", mock.Anything, "team_test").Return(&store.Team{
		Common:         store.Common{Id: "team_test"},
		Name:           "test",
		PaymentMethods: []store.PaymentMethod{{Common: store.Common{Id: "pm_test"}}},
	}, nil)

	testCases := []struct {
		name               string
		email              string
		password           string
		expectedStatusCode int
	}{
		{
			name:               "success",
			email:              user.Email,
			password:           user.Password,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "fail",
			email:              user.Email,
			password:           user.Password + "incorrect",
			expectedStatusCode: http.StatusUnauthorized,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			userStore := &storemock.UserStoreMock{}
			sessionStore := sessions.NewCookieStore([]byte("secret"))
			passwordHash := &hashmock.PasswordHashMock{}
			passwordHash.On("ComparePasswordAndHash", tc.password, user.Password).Return(tc.password == user.Password, nil)
			if tc.email == user.Email {
				userStore.On("GetUser", tc.email).Return(user, nil)
			} else {
				userStore.On("GetUser", tc.email).Return(nil, gorm.ErrRecordNotFound)
			}

			handler := NewPostLoginHandler(userStore, teamStore, passwordHash, sessionStore, "session")
			body := bytes.NewBufferString("email=" + tc.email + "&password=" + tc.password)
			req, _ := http.NewRequest("POST", "/", body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			assert.Equal(tc.expectedStatusCode, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, tc.expectedStatusCode)
			userStore.AssertExpectations(t)
			passwordHash.AssertExpectations(t)
		})
	}
}
