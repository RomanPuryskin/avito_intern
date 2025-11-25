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

func TestHanlder_CreateTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	mockService := mocks.NewMockTeamService(ctrl)
	teamHandler := NewTeamHandler(logger, mockService)

	app := fiber.New()
	app.Post("/team/add", teamHandler.CreateTeam)

	tests := []struct {
		Name             string
		RequestBody      string
		ExistedTeams     []string
		ExistedUsersID   []string
		ExistedUserNames []string
		ExpectedCode     int
		ExpectedBody     string
		MockSetup        func(ms *mocks.MockTeamService)
	}{
		{
			Name:             "error_invalid_input_format",
			RequestBody:      "invaid_input_format",
			ExistedTeams:     []string{},
			ExistedUsersID:   []string{},
			ExistedUserNames: []string{},
			ExpectedCode:     400,
			ExpectedBody: `{
			"code":  "INVALID_INPUT",
			"message": "invalid input format"
			}`,
			MockSetup: nil,
		},
		{
			Name: "error_invalid_input",
			RequestBody: `{
			"team_name": "name",
			"members": [
				{
					"user_id": "id",
					"is_active": true
				}
			]
			}`,
			ExistedTeams:     []string{},
			ExistedUsersID:   []string{},
			ExistedUserNames: []string{},
			ExpectedCode:     400,
			ExpectedBody: `{
			"code":  "INVALID_INPUT",
			"message": "invalid input"
			}`,
			MockSetup: nil,
		},
		{
			Name: "error_team_already_exist",
			RequestBody: `{
			"team_name": "name",
			"members": [
				{
					"user_id": "id",
					"username": "name",
					"is_active": true
				}
			]
			}`,
			ExistedTeams:     []string{"name"},
			ExistedUsersID:   []string{},
			ExistedUserNames: []string{},
			ExpectedCode:     400,
			ExpectedBody: `{
			"code":  "TEAM_EXISTS",
			"message": "team already exists"
			}`,
			MockSetup: func(ms *mocks.MockTeamService) {

				ms.EXPECT().CreateTeam(gomock.Any(), gomock.Any()).Return(nil, service.ErrorTeamExists)
			},
		},
		{
			Name: "error_user_already_exist",
			RequestBody: `{
			"team_name": "name",
			"members": [
				{
					"user_id": "id",
					"username": "name",
					"is_active": true
				}
			]
			}`,
			ExistedTeams:     []string{"name1"},
			ExistedUsersID:   []string{"id"},
			ExistedUserNames: []string{},
			ExpectedCode:     400,
			ExpectedBody: `{
			"code":  "USER_EXISTS",
			"message": "user with user_id already exists"
			}`,
			MockSetup: func(ms *mocks.MockTeamService) {

				ms.EXPECT().CreateTeam(gomock.Any(), gomock.Any()).Return(nil, service.ErrorUserAlreadyExists)
			},
		},
		{
			Name: "error_user_already_exist_by_username",
			RequestBody: `{
			"team_name": "name1",
			"members": [
				{
					"user_id": "id1",
					"username": "name1",
					"is_active": true
				}
			]
			}`,
			ExistedTeams:     []string{"name2"},
			ExistedUsersID:   []string{"id2"},
			ExistedUserNames: []string{"name1"},
			ExpectedCode:     400,
			ExpectedBody: `{
			"code":  "USER_EXISTS",
			"message": "user with username already exists"
			}`,
			MockSetup: func(ms *mocks.MockTeamService) {

				ms.EXPECT().CreateTeam(gomock.Any(), gomock.Any()).Return(nil, service.ErrorUserAlreadyExistsByUserName)
			},
		},
		{
			Name: "success_create",
			RequestBody: `{
			"team_name": "name1",
			"members": [
				{
					"user_id": "id1",
					"username": "name1",
					"is_active": true
				},
				{
					"user_id": "id2",
					"username": "name2",
					"is_active": true
				}
			]
			}`,
			ExistedTeams:     []string{"name3"},
			ExistedUsersID:   []string{"id3", "id4"},
			ExistedUserNames: []string{"name3", "name4"},
			ExpectedCode:     201,
			ExpectedBody: `{
			"team_name": "name1",
			"members": [
				{
					"user_id": "id1",
					"username": "name1",
					"is_active": true
				},
				{
					"user_id": "id2",
					"username": "name2",
					"is_active": true
				}
			]
			}`,
			MockSetup: func(ms *mocks.MockTeamService) {

				ms.EXPECT().CreateTeam(gomock.Any(), gomock.Any()).Return(&enteties.Team{
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
							IsActive: true,
						},
					},
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			req := httptest.NewRequest("POST", "/team/add", strings.NewReader(tt.RequestBody))
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

func TestHanlder_GetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	mockService := mocks.NewMockTeamService(ctrl)
	teamHandler := NewTeamHandler(logger, mockService)

	app := fiber.New()
	app.Get("/team/get", teamHandler.GetTeam)

	tests := []struct {
		Name         string
		RequestName  string
		Teams        []enteties.Team
		ExpectedCode int
		ExpectedBody string
		MockSetup    func(ms *mocks.MockTeamService)
	}{
		{
			Name:         "error_invalid_input",
			RequestName:  "",
			Teams:        nil,
			ExpectedCode: 400,
			ExpectedBody: `{
			"code":  "INVALID_INPUT",
			"message": "invalid input"
			}`,
			MockSetup: nil,
		},
		{
			Name:        "success_found",
			RequestName: "name1",
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
			ExpectedCode: 200,
			ExpectedBody: `{
			"team_name": "name1",
			"members": [
				{
					"user_id": "id1",
					"username": "name1",
					"is_active": true
				}
			]
			}`,
			MockSetup: func(ms *mocks.MockTeamService) {

				ms.EXPECT().GetTeam(gomock.Any(), "name1").Return(&enteties.Team{
					TeamName: "name1",
					Members: []enteties.TeamMember{
						{
							UserID:   "id1",
							UserName: "name1",
							IsActive: true,
						},
					},
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			req := httptest.NewRequest("GET", fmt.Sprintf("/team/get?team_name=%s", tt.RequestName), nil)
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
