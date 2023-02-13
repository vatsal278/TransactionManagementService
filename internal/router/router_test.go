package router

import (
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	respModel "github.com/PereRohit/util/model"
	"github.com/PereRohit/util/testutil"
	"github.com/golang/mock/gomock"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/repo/authentication"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name     string
		setup    func() *config.SvcConfig
		validate func(http.ResponseWriter)
		give     *http.Request
	}{
		{
			name: "Success health check",
			setup: func() *config.SvcConfig {
				return &config.SvcConfig{
					JwtSvc: config.JWTSvc{JwtSvc: authentication.JWTAuthService("")},
					Cfg: &config.Config{
						ServiceRouteVersion: "v1",
						DataBase: config.DbCfg{
							Driver: "mysql",
						},
					},
					ServiceRouteVersion: "v1",
					DbSvc:               config.DbSvc{}}
			},
			validate: func(w http.ResponseWriter) {
				wIn := w.(*httptest.ResponseRecorder)

				diff := testutil.Diff(wIn.Code, http.StatusOK)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				diff = testutil.Diff(wIn.Header().Get("Content-Type"), "application/json")
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				resp := respModel.Response{}
				err := json.NewDecoder(wIn.Body).Decode(&resp)
				diff = testutil.Diff(err, nil)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
				respDataB, err := json.Marshal(resp.Data)
				diff = testutil.Diff(err, nil)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				type svcHealthStat struct {
					Status  string `json:"status"`
					Message string `json:"message,omitempty"`
				}
				hCdata := map[string]interface{}{}
				hCdata["transactionManagementService"] = svcHealthStat{Status: fmt.Sprint(http.StatusOK)}

				err = json.Unmarshal(respDataB, &hCdata)
				diff = testutil.Diff(err, nil)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				diff = testutil.Diff(resp, respModel.Response{
					Status:  http.StatusOK,
					Message: http.StatusText(http.StatusOK),
					Data:    hCdata,
				})
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
			},
			give: httptest.NewRequest(http.MethodGet, "/v1/health", nil),
		},
		{
			name: "No route found",
			setup: func() *config.SvcConfig {
				return &config.SvcConfig{
					JwtSvc: config.JWTSvc{JwtSvc: authentication.JWTAuthService("")},
					Cfg: &config.Config{
						ServiceRouteVersion: "v1",
						DataBase: config.DbCfg{
							Driver: "mysql",
						},
					},
					ServiceRouteVersion: "v1",
					DbSvc:               config.DbSvc{}}
			},
			validate: func(w http.ResponseWriter) {
				wIn := w.(*httptest.ResponseRecorder)

				diff := testutil.Diff(wIn.Code, http.StatusNotFound)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				diff = testutil.Diff(wIn.Header().Get("Content-Type"), "application/json")
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				resp := respModel.Response{}
				err := json.NewDecoder(wIn.Body).Decode(&resp)
				diff = testutil.Diff(err, nil)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				diff = testutil.Diff(resp, respModel.Response{
					Status:  http.StatusNotFound,
					Message: http.StatusText(http.StatusNotFound),
					Data:    nil,
				})
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
			},
			give: httptest.NewRequest(http.MethodGet, "/no-route", nil),
		},
		{
			name: "Method not allowed",
			setup: func() *config.SvcConfig {
				return &config.SvcConfig{
					JwtSvc: config.JWTSvc{JwtSvc: authentication.JWTAuthService("")},
					Cfg: &config.Config{
						ServiceRouteVersion: "v1",
						DataBase: config.DbCfg{
							Driver: "mysql",
						},
					},
					ServiceRouteVersion: "v1",
					DbSvc:               config.DbSvc{}}
			},
			validate: func(w http.ResponseWriter) {
				wIn := w.(*httptest.ResponseRecorder)

				diff := testutil.Diff(wIn.Code, http.StatusMethodNotAllowed)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				diff = testutil.Diff(wIn.Header().Get("Content-Type"), "application/json")
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				resp := respModel.Response{}
				err := json.NewDecoder(wIn.Body).Decode(&resp)
				diff = testutil.Diff(err, nil)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}

				diff = testutil.Diff(resp, respModel.Response{
					Status:  http.StatusMethodNotAllowed,
					Message: http.StatusText(http.StatusMethodNotAllowed),
					Data:    nil,
				})
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
			},
			give: httptest.NewRequest(http.MethodPut, "/v1/health", nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			sqlmock.MonitorPingsOption(true)
			c := tt.setup()
			c.DbSvc.Db = db
			r := Register(c)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.give)
			tt.validate(w)
		})
	}
}
