package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	respModel "github.com/PereRohit/util/model"
	"github.com/gorilla/mux"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/pkg/session"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/PereRohit/util/testutil"
	"github.com/golang/mock/gomock"

	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
	"github.com/vatsal278/TransactionManagementService/pkg/mock"
)

type Reader string

func (Reader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
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
		setup    func() datasource.DataSourceI
		wantStat bool
	}{
		{
			name: "Success",
			setup: func() datasource.DataSourceI {
				mockDs := mock.NewMockDataSourceI(mockCtrl)

				mockDs.EXPECT().HealthCheck().Times(1).
					Return(false)

				return mockDs
			},
			wantStat: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementService(tt.setup(), config.UtilSvc{})

			_, _, stat := rec.HealthCheck()

			diff := testutil.Diff(stat, tt.wantStat)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}

func TestTransactionManagementService_NewTransaction(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name  string
		model model.Transaction
		setup func() (*transactionManagementService, *http.Request)
		want  func(recorder httptest.ResponseRecorder)
	}{
		{
			name: "Success",
			model: model.Transaction{
				AccountNumber: 1,
				Amount:        1000,
				TransferTo:    2,
				Status:        "approved",
				Type:          "debit",
				Comment:       "shopping",
			},
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				mockLogic.EXPECT().NewTransaction(model.NewTransaction{
					UserId:        "1234",
					AccountNumber: 1,
					Amount:        1000,
					TransferTo:    2,
					Status:        "approved",
					Type:          "debit",
					Comment:       "shopping",
				}).Times(1).Return(&respModel.Response{
					Status:  http.StatusCreated,
					Message: codes.GetErr(codes.Success),
					Data:    nil,
				})
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				by, err := json.Marshal(model.Transaction{
					AccountNumber: 1,
					Amount:        1000,
					TransferTo:    2,
					Status:        "approved",
					Type:          "debit",
					Comment:       "shopping",
				})
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				r := httptest.NewRequest("POST", "/transactions/new", bytes.NewBuffer(by))
				ctx := session.SetSession(r.Context(), model.SessionStruct{UserId: "1234"})
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusCreated,
					Message: codes.GetErr(codes.Success),
					Data:    nil,
				}
				if !reflect.DeepEqual(rec.Code, http.StatusCreated) {
					t.Errorf("Want: %v, Got: %v", http.StatusCreated, rec.Code)
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}

			},
		},
		{
			name: "Failure :: NewTransaction:: Failure assert user_id",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("POST", "/new_account", Reader(""))
				ctx := session.SetSession(r.Context(), "")
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusBadRequest,
					Message: codes.GetErr(codes.ErrAssertUserid),
					Data:    nil,
				}
				if !reflect.DeepEqual(rec.Code, http.StatusBadRequest) {
					t.Errorf("Want: %v, Got: %v", http.StatusBadRequest, rec.Code)
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}
			},
		},
		{
			name: "Failure :: NewTransaction:: Read all failure",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("POST", "/new_account", Reader(""))
				ctx := session.SetSession(r.Context(), model.SessionStruct{UserId: "1234"})
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: "request body read : test error",
					Data:    nil,
				}
				if !reflect.DeepEqual(rec.Code, http.StatusInternalServerError) {
					t.Errorf("Want: %v, Got: %v", http.StatusInternalServerError, rec.Code)
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}
			},
		},
		{
			name: "Failure :: NewTransaction:: json unmarshall failure",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("POST", "/new_account", bytes.NewBuffer([]byte("")))
				ctx := session.SetSession(r.Context(), model.SessionStruct{UserId: "1234"})
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusBadRequest,
					Message: "put data into data: unexpected end of JSON input",
					Data:    nil,
				}
				if !reflect.DeepEqual(rec.Code, http.StatusBadRequest) {
					t.Errorf("Want: %v, Got: %v", http.StatusBadRequest, rec.Code)
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			x, r := tt.setup()
			x.NewTransaction(w, r)
			tt.want(*w)
		})
	}
}

func TestTransactionManagementService_GetTransactions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name  string
		model model.GetTransaction
		setup func() (*transactionManagementService, *http.Request)
		want  func(recorder httptest.ResponseRecorder)
	}{
		{
			name: "Success::GetTransaction",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				mockLogic.EXPECT().GetTransactions("1234", 2, 2).Times(1).Return(&respModel.Response{
					Status:  http.StatusOK,
					Message: codes.GetErr(codes.Success),
					Data: model.PaginatedResponse{Response: []model.Transaction{{Amount: 1000, AccountNumber: 1}}, Pagination: model.Paginate{
						CurrentPage: 1,
						NextPage:    -1,
						TotalPage:   1,
					}},
				})
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("GET", "/transactions", nil)
				req := r.URL.Query()
				req.Add("limit", "2")
				req.Add("page", "2")
				r.URL.RawQuery = req.Encode()
				ctx := session.SetSession(r.Context(), model.SessionStruct{UserId: "1234"})
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				tempResp := &respModel.Response{
					Status:  http.StatusOK,
					Message: codes.GetErr(codes.Success),
					Data: model.PaginatedResponse{Response: []model.Transaction{{Amount: 1000, AccountNumber: 1}}, Pagination: model.Paginate{
						CurrentPage: 1,
						NextPage:    -1,
						TotalPage:   1,
					}},
				}
				marshal, err := json.Marshal(&tempResp)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				if !reflect.DeepEqual(rec.Code, http.StatusOK) {
					t.Errorf("Want: %v, Got: %v", http.StatusOK, rec.Code)
				}
				if strings.Compare(string(marshal), string(b)) != -1 {
					t.Errorf("Want: %v, Got: %v", string(marshal), string(b))
				}
			},
		},
		{
			name: "Success::GetTransaction:: default limit and page",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				mockLogic.EXPECT().GetTransactions("1234", 5, 1).Times(1).Return(&respModel.Response{
					Status:  http.StatusOK,
					Message: codes.GetErr(codes.Success),
					Data: model.PaginatedResponse{Response: []model.Transaction{{Amount: 1000, AccountNumber: 1}}, Pagination: model.Paginate{
						CurrentPage: 1,
						NextPage:    -1,
						TotalPage:   1,
					}},
				})
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("GET", "/transactions", nil)
				r.URL.Query().Set("limit", "10")
				r.URL.Query().Set("page", "1")
				ctx := session.SetSession(r.Context(), model.SessionStruct{UserId: "1234"})
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				tempResp := &respModel.Response{
					Status:  http.StatusOK,
					Message: codes.GetErr(codes.Success),
					Data: model.PaginatedResponse{Response: []model.Transaction{{Amount: 1000, AccountNumber: 1}}, Pagination: model.Paginate{
						CurrentPage: 1,
						NextPage:    -1,
						TotalPage:   1,
					}},
				}
				marshal, err := json.Marshal(&tempResp)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				if !reflect.DeepEqual(rec.Code, http.StatusOK) {
					t.Errorf("Want: %v, Got: %v", http.StatusOK, rec.Code)
				}
				if strings.Compare(string(marshal), string(b)) != -1 {
					t.Errorf("Want: %v, Got: %v", string(marshal), string(b))
				}
			},
		},
		{
			name: "Failure::GetTransaction:: logic-internal server error",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("PUT", "/activate", nil)
				ctx := session.SetSession(r.Context(), "1234")
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusBadRequest,
					Message: codes.GetErr(codes.ErrAssertUserid),
					Data:    nil,
				}
				if !reflect.DeepEqual(rec.Code, http.StatusBadRequest) {
					t.Errorf("Want: %v, Got: %v", http.StatusBadRequest, rec.Code)
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}
			},
		},
		{
			name: "Failure::GetTransaction:: err asserting to string",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("PUT", "/activate", nil)
				ctx := session.SetSession(r.Context(), 1.11)
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					t.Log(err)
					t.Fail()
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusBadRequest,
					Message: codes.GetErr(codes.ErrAssertUserid),
					Data:    nil,
				}
				if !reflect.DeepEqual(rec.Code, http.StatusBadRequest) {
					t.Errorf("Want: %v, Got: %v", http.StatusBadRequest, rec.Code)
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			x, r := tt.setup()
			x.GetTransactions(w, r)
			tt.want(*w)
		})
	}
}
func TestTransactionManagementService_DownloadTransaction(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name  string
		model model.GetTransaction
		setup func() (*transactionManagementService, *http.Request)
		want  func(recorder httptest.ResponseRecorder)
	}{
		{
			name: "Success",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				mockLogic.EXPECT().DownloadTransaction(gomock.Any(), "456").Times(1).Return(&respModel.Response{
					Status:  http.StatusOK,
					Message: codes.GetErr(codes.Success),
					Data:    []byte("PDF"),
				})
				svc := &transactionManagementService{
					logic: mockLogic,
				}

				r := httptest.NewRequest("GET", "/transactions/download/:123", nil)

				ctx := session.SetSession(r.Context(), model.SessionStruct{UserId: "1234", Cookie: "456"})
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					return
				}
				if strings.Compare("PDF", string(b)) != 0 {
					t.Errorf("Want: %v, Got: %v", "PDF", string(b))
				}
			},
		},
		{
			name: "Failure:: DownloadTransaction :: err assert userid",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("GET", "/transactions/download/:123", nil)
				ctx := session.SetSession(r.Context(), "")
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					return
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusBadRequest,
					Message: codes.GetErr(codes.ErrAssertUserid),
					Data:    nil,
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}
			},
		},
		{
			name: "Failure:: DownloadTransaction :: not ok status code",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				mockLogic.EXPECT().DownloadTransaction(gomock.Any(), gomock.Any()).Return(&respModel.Response{
					Status:  http.StatusBadRequest,
					Message: "",
					Data:    nil,
				})
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("GET", "/transactions/download/:123", nil)
				ctx := session.SetSession(r.Context(), model.SessionStruct{UserId: "1234", Cookie: "4321"})
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					return
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusBadRequest,
					Message: "",
					Data:    nil,
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}
			},
		},
		{
			name: "Failure:: DownloadTransaction :: err asserting pdf data",
			setup: func() (*transactionManagementService, *http.Request) {
				mockLogic := mock.NewMockTransactionManagementServiceLogicIer(mockCtrl)
				mockLogic.EXPECT().DownloadTransaction("123", "4321").Return(&respModel.Response{
					Status:  http.StatusOK,
					Message: "Success",
					Data:    123,
				})
				svc := &transactionManagementService{
					logic: mockLogic,
				}
				r := httptest.NewRequest("GET", "/transactions/download/123", nil)
				r = mux.SetURLVars(r, map[string]string{"transaction_id": "123"})
				ctx := session.SetSession(r.Context(), model.SessionStruct{UserId: "1234", Cookie: "4321"})
				return svc, r.WithContext(ctx)
			},
			want: func(rec httptest.ResponseRecorder) {
				b, err := ioutil.ReadAll(rec.Body)
				if err != nil {
					return
				}
				var response respModel.Response
				err = json.Unmarshal(b, &response)
				tempResp := &respModel.Response{
					Status:  http.StatusBadRequest,
					Message: "error assert pdf []byte",
					Data:    nil,
				}
				if !reflect.DeepEqual(&response, tempResp) {
					t.Errorf("Want: %v, Got: %v", tempResp, &response)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			x, r := tt.setup()
			x.DownloadTransaction(w, r)
			tt.want(*w)
		})
	}
}
