package handlers

import (
	"avito_intern/internal/enteties"
	"avito_intern/internal/service"
	"avito_intern/mocks"
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

func TestHander_CreatePR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	mockService := mocks.NewMockPRService(ctrl)
	prHandler := NewPRHandler(logger, mockService)

	app := fiber.New()
	app.Post("/pullRequest/create", prHandler.CreatePR)

	tests := []struct {
		Name              string
		RequestBody       string
		ExistingPRs       []string
		Teams             []enteties.Team
		ExpectedReviewers []string
		ExpectedCode      int
		ExpectedBody      string
		MockSetup         func(ms *mocks.MockPRService)
	}{
		{
			Name:              "error_invalid_input_format",
			RequestBody:       "invaid_input_format",
			ExistingPRs:       nil,
			Teams:             nil,
			ExpectedReviewers: nil,
			ExpectedCode:      400,
			ExpectedBody: `{
			"code":  "INVALID_INPUT",
			"message": "invalid input format"
			}`,
			MockSetup: nil,
		},
		{
			Name: "error_invalid_input",
			RequestBody: `{
			"pull_request_id": "id",
			"pull_request_name": "name"
			}`,
			ExistingPRs:       nil,
			Teams:             nil,
			ExpectedReviewers: nil,
			ExpectedCode:      400,
			ExpectedBody: `{
			"code":  "INVALID_INPUT",
			"message": "invalid input"
			}`,
			MockSetup: nil,
		},
		{
			Name: "error_author_not_exists",
			RequestBody: `{
			"pull_request_id": "id",
			"pull_request_name": "name",
			"author_id": "id2"
			}`,
			ExistingPRs: nil,
			Teams: []enteties.Team{
				{
					TeamName: "name1",
					Members: []enteties.TeamMember{
						{
							UserID:   "id1",
							UserName: "name1",
							IsActive: true,
						},
					},
				},
			},
			ExpectedReviewers: nil,
			ExpectedCode:      404,
			ExpectedBody: `{
			"code":  "NOT_FOUND",
			"message": "user not found"
			}`,
			MockSetup: func(ms *mocks.MockPRService) {
				ms.EXPECT().CreatePR(gomock.Any(), gomock.Any()).Return(nil, service.ErrorUserNotFound)
			},
		},
		{
			Name: "error_pr_already_exists",
			RequestBody: `{
			"pull_request_id": "id",
			"pull_request_name": "name",
			"author_id": "id2"
			}`,
			ExistingPRs: nil,
			Teams: []enteties.Team{
				{
					TeamName: "name1",
					Members: []enteties.TeamMember{
						{
							UserID:   "id1",
							UserName: "name1",
							IsActive: true,
						},
					},
				},
			},
			ExpectedReviewers: []string{"id"},
			ExpectedCode:      409,
			ExpectedBody: `{
			"code":  "PR_EXISTS",
			"message": "pr already exists"
			}`,
			MockSetup: func(ms *mocks.MockPRService) {
				ms.EXPECT().CreatePR(gomock.Any(), gomock.Any()).Return(nil, service.ErrorPRAlreadyExists)
			},
		},
		{
			Name: "success_assigned_0_reviewers",
			RequestBody: `{
			"pull_request_id": "id",
			"pull_request_name": "name",
			"author_id": "id1"
			}`,
			ExistingPRs: nil,
			Teams: []enteties.Team{
				{
					TeamName: "name1",
					Members: []enteties.TeamMember{
						{
							UserID:   "id1",
							UserName: "name1",
							IsActive: true,
						},
						{
							UserID:   "id2",
							UserName: "name2",
							IsActive: false,
						},
					},
				},
			},
			ExpectedReviewers: nil,
			ExpectedCode:      201,
			ExpectedBody: `{
			"pull_request_id":  "id",
			"pull_request_name": "name",
			"author_id": "id1",
			"status": "OPEN",
			"assigned_reviewers": [
			
			]
			}`,
			MockSetup: func(ms *mocks.MockPRService) {
				ms.EXPECT().CreatePR(gomock.Any(), gomock.Any()).Return(&enteties.PullRequest{
					PullRequestID:     "id",
					PulRequestName:    "name",
					AuthorID:          "id1",
					Status:            "OPEN",
					AssignedReviewers: []string{},
				}, nil)
			},
		},
		{
			Name: "success_assigned_1_reviewer",
			RequestBody: `{
			"pull_request_id": "id",
			"pull_request_name": "name",
			"author_id": "id1"
			}`,
			ExistingPRs: nil,
			Teams: []enteties.Team{
				{
					TeamName: "name1",
					Members: []enteties.TeamMember{
						{
							UserID:   "id1",
							UserName: "name1",
							IsActive: true,
						},
						{
							UserID:   "id2",
							UserName: "name2",
							IsActive: false,
						},
						{
							UserID:   "id3",
							UserName: "name3",
							IsActive: true,
						},
					},
				},
			},
			ExpectedReviewers: nil,
			ExpectedCode:      201,
			ExpectedBody: `{
			"pull_request_id":  "id",
			"pull_request_name": "name",
			"author_id": "id1",
			"status": "OPEN",
			"assigned_reviewers": [
				"id3"
			]
			}`,
			MockSetup: func(ms *mocks.MockPRService) {
				ms.EXPECT().CreatePR(gomock.Any(), gomock.Any()).Return(&enteties.PullRequest{
					PullRequestID:     "id",
					PulRequestName:    "name",
					AuthorID:          "id1",
					Status:            "OPEN",
					AssignedReviewers: []string{"id3"},
				}, nil)
			},
		},
		{
			Name: "success_assigned_1_reviewer",
			RequestBody: `{
			"pull_request_id": "id",
			"pull_request_name": "name",
			"author_id": "id1"
			}`,
			ExistingPRs: nil,
			Teams: []enteties.Team{
				{
					TeamName: "name1",
					Members: []enteties.TeamMember{
						{
							UserID:   "id1",
							UserName: "name1",
							IsActive: true,
						},
						{
							UserID:   "id2",
							UserName: "name2",
							IsActive: false,
						},
						{
							UserID:   "id3",
							UserName: "name3",
							IsActive: true,
						},
						{
							UserID:   "id4",
							UserName: "name4",
							IsActive: true,
						},
					},
				},
			},
			ExpectedReviewers: nil,
			ExpectedCode:      201,
			ExpectedBody: `{
			"pull_request_id":  "id",
			"pull_request_name": "name",
			"author_id": "id1",
			"status": "OPEN",
			"assigned_reviewers": [
				"id3","id4"
			]
			}`,
			MockSetup: func(ms *mocks.MockPRService) {
				ms.EXPECT().CreatePR(gomock.Any(), gomock.Any()).Return(&enteties.PullRequest{
					PullRequestID:     "id",
					PulRequestName:    "name",
					AuthorID:          "id1",
					Status:            "OPEN",
					AssignedReviewers: []string{"id3", "id4"},
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			req := httptest.NewRequest("POST", "/pullRequest/create", strings.NewReader(tt.RequestBody))
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
