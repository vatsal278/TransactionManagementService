package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	respModel "github.com/PereRohit/util/model"
	"github.com/PereRohit/util/testutil"
	"github.com/golang/mock/gomock"

	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
	"github.com/vatsal278/TransactionManagementService/pkg/mock"
)

func Test_transactionManagementService_Ping(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name     string
		setup    func() (TransactionManagementServiceHandler, http.ResponseWriter, *http.Request)
		validate func(w http.ResponseWriter)
	}{
		{
			name: "Success",
			setup: func() (TransactionManagementServiceHandler, http.ResponseWriter, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)

				req := &model.PingRequest{
					Data: "hello-world",
				}

				mockLogic.EXPECT().Ping(req).Return(&respModel.Response{
					Status:  http.StatusOK,
					Message: "Ok",
					Data:    "pong",
				}).Times(1)

				rec := &transactionManagementService{
					logic: mockLogic,
				}

				reqB, err := json.Marshal(req)
				if err != nil {
					t.Errorf(err.Error())
					t.Fail()
				}
				r := httptest.NewRequest(http.MethodGet, "https://ping", bytes.NewReader(reqB))
				w := httptest.NewRecorder()

				return rec, w, r
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

				diff = testutil.Diff(resp, respModel.Response{
					Status:  http.StatusOK,
					Message: "Ok",
					Data:    "pong",
				})
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
			},
		},
		{
			name: "Failure::request not as expected",
			setup: func() (TransactionManagementServiceHandler, http.ResponseWriter, *http.Request) {
				req := "hello world"

				rec := &transactionManagementService{}

				r := httptest.NewRequest(http.MethodGet, "https://ping", bytes.NewReader([]byte(req)))
				w := httptest.NewRecorder()

				return rec, w, r
			},
			validate: func(w http.ResponseWriter) {
				wIn := w.(*httptest.ResponseRecorder)

				diff := testutil.Diff(wIn.Code, http.StatusBadRequest)
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

				// ignore specific message
				resp.Message = ""

				diff = testutil.Diff(resp, respModel.Response{
					Status:  http.StatusBadRequest,
					Message: "",
					Data:    nil,
				})
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver, w, r := tt.setup()

			receiver.Ping(w, r)

			tt.validate(w)
		})
	}
}

func Test_transactionManagementService_HealthCheck(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name        string
		setup       func() TransactionManagementServiceHandler
		wantSvcName string
		wantMsg     string
		wantStat    bool
	}{
		{
			name: "Success",
			setup: func() TransactionManagementServiceHandler {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)

				mockLogic.EXPECT().HealthCheck().
					Return(true).Times(1)

				rec := &transactionManagementService{
					logic: mockLogic,
				}

				return rec
			},
			wantSvcName: TransactionManagementServiceName,
			wantMsg:     "",
			wantStat:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.setup()

			svcName, msg, stat := receiver.HealthCheck()

			diff := testutil.Diff(svcName, tt.wantSvcName)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}

			diff = testutil.Diff(msg, tt.wantMsg)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}

			diff = testutil.Diff(stat, tt.wantStat)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}

func TestNewTransactionManagementService(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name     string
		setup    func() datasource.DataSource
		wantStat bool
	}{
		{
			name: "Success",
			setup: func() datasource.DataSource {
				mockDs := mock.NewMockDataSource(mockCtrl)

				mockDs.EXPECT().HealthCheck().Times(1).
					Return(false)

				return mockDs
			},
			wantStat: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementService(tt.setup())

			_, _, stat := rec.HealthCheck()

			diff := testutil.Diff(stat, tt.wantStat)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}
