package handlers

import (
	"avito_intern/internal/enteties"
	"avito_intern/internal/service"
	"avito_intern/mocks"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHander_SetIsActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	mockService := mocks.NewMockUserService(ctrl)
	userHandler := NewUserHandler(logger, mockService)

	app := fiber.New()
	app.Post("/users/setIsActive", userHandler.SetIsActive)

	tests := []struct {
		Name         string
		RequestBody  string
		User         enteties.User
		ExpectedCode int
		ExpectedBody string
		MockSetup    func(ms *mocks.MockUserService)
	}{
		{
			Name:        "Error_invalid_input_format",
			RequestBody: "invalid input",
			User: enteties.User{
				UserID:   "u1",
				UserName: "name",
				TeamName: "team",
				IsActive: true,
			},
			ExpectedCode: 400,
			ExpectedBody: `{
			"code":  "INVALID_INPUT",
			"message": "invalid input format"
			}`,
			MockSetup: nil,
		},
		{
			Name: "Error_not_enough_fileds",
			RequestBody: `{
				"user_id": "u1"
			}`,
			User: enteties.User{
				UserID:   "u1",
				UserName: "name",
				TeamName: "team",
				IsActive: true,
			},
			ExpectedCode: 400,
			ExpectedBody: `{
			"code":  "INVALID_INPUT",
			"message": "invalid input"
			}`,
			MockSetup: nil,
		},
		{
			Name: "Error_user_not_found",
			RequestBody: `{
				"user_id": "u2",
				"is_active": false
			}`,
			User: enteties.User{
				UserID:   "u1",
				UserName: "name",
				TeamName: "team",
				IsActive: true,
			},
			ExpectedCode: 404,
			ExpectedBody: `{
			"code":  "NOT_FOUND",
			"message": "user not found"
			}`,
			MockSetup: func(ms *mocks.MockUserService) {

				ms.EXPECT().SetIsActive(gomock.Any(), "u2", false).Return(nil, service.ErrorUserNotFound)
			},
		},
		{
			Name: "succes_changed",
			RequestBody: `{
				"user_id": "u1",
				"is_active": false
			}`,
			User: enteties.User{
				UserID:   "u1",
				UserName: "name",
				TeamName: "team",
				IsActive: true,
			},
			ExpectedCode: 200,
			ExpectedBody: `{
			"user_id":  "u1",
			"username": "name",
			"team_name": "team",
			"is_active": false
			}`,
			MockSetup: func(ms *mocks.MockUserService) {

				ms.EXPECT().SetIsActive(gomock.Any(), "u1", false).Return(&enteties.User{
					UserID:   "u1",
					UserName: "name",
					TeamName: "team",
					IsActive: false,
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			req := httptest.NewRequest("POST", "/users/setIsActive", strings.NewReader(tt.RequestBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.MockSetup != nil {
				tt.MockSetup(mockService)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assert.Equal(t, tt.ExpectedCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.JSONEq(t, tt.ExpectedBody, string(body))
		})
	}

}

func TestHandler_GetReview(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	mockService := mocks.NewMockUserService(ctrl)
	userHandler := NewUserHandler(logger, mockService)

	app := fiber.New()
	app.Get("/users/getReview", userHandler.GetReview)

	tests := []struct {
		Name               string
		RequestID          string
		Assigend_reviewers map[string][]string
		ExpectedCode       int
		ExpectedBody       string
		MockSetup          func(ms *mocks.MockUserService)
	}{
		{
			Name:               "invalid_user_id",
			RequestID:          "",
			Assigend_reviewers: map[string][]string{},
			ExpectedCode:       400,
			ExpectedBody: `{
			"code":  "INVALID_INPUT",
			"message": "invalid input"
			}`,
			MockSetup: nil,
		},
		{
			Name:      "Error_user_not_found",
			RequestID: "u1",
			Assigend_reviewers: map[string][]string{
				"u2": {"pr2"},
			},
			ExpectedCode: 404,
			ExpectedBody: `{
			"code":  "NOT_FOUND",
			"message": "user not found"
			}`,
			MockSetup: func(ms *mocks.MockUserService) {
				ms.EXPECT().GetReviews(gomock.Any(), "u1").Return(nil, service.ErrorUserNotFound)
			},
		},
		{
			Name:      "success_request_one_pr",
			RequestID: "u1",
			Assigend_reviewers: map[string][]string{
				"u1": {"pr2"},
			},
			ExpectedCode: 200,
			ExpectedBody: `{
			"user_id": "u1",
			"pull_requests": [
				{
					"pull_request_id": "pr1",
					"pull_request_name": "name",
					"author_id": "author",
					"status": "OPEN"
				}
			]
			}`,
			MockSetup: func(ms *mocks.MockUserService) {
				ms.EXPECT().GetReviews(gomock.Any(), "u1").Return(&enteties.UserReviews{
					UserID: "u1",
					PullRequests: []enteties.PullRequestShort{
						{
							PullRequestID:  "pr1",
							PulRequestName: "name",
							AuthorID:       "author",
							Status:         "OPEN",
						},
					},
				}, nil)
			},
		},
		{
			Name:      "success_request_any_pr",
			RequestID: "u1",
			Assigend_reviewers: map[string][]string{
				"u1": {"p1", "pr2"},
			},
			ExpectedCode: 200,
			ExpectedBody: `{
			"user_id": "u1",
			"pull_requests": [
				{
					"pull_request_id": "pr1",
					"pull_request_name": "name",
					"author_id": "author",
					"status": "OPEN"
				},
				{
					"pull_request_id": "pr2",
					"pull_request_name": "name2",
					"author_id": "author2",
					"status": "OPEN"
				}
			]
			}`,
			MockSetup: func(ms *mocks.MockUserService) {
				ms.EXPECT().GetReviews(gomock.Any(), "u1").Return(&enteties.UserReviews{
					UserID: "u1",
					PullRequests: []enteties.PullRequestShort{
						{
							PullRequestID:  "pr1",
							PulRequestName: "name",
							AuthorID:       "author",
							Status:         "OPEN",
						},
						{
							PullRequestID:  "pr2",
							PulRequestName: "name2",
							AuthorID:       "author2",
							Status:         "OPEN",
						},
					},
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			req := httptest.NewRequest("GET", fmt.Sprintf("/users/getReview?user_id=%s", tt.RequestID), nil)
			req.Header.Set("Content-Type", "application/json")

			if tt.MockSetup != nil {
				tt.MockSetup(mockService)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assert.Equal(t, tt.ExpectedCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.JSONEq(t, tt.ExpectedBody, string(body))
		})
	}
}
