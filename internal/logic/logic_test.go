package logic

import (
	"encoding/json"
	"errors"
	respModel "github.com/PereRohit/util/model"
	"github.com/PereRohit/util/testutil"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
	"github.com/vatsal278/TransactionManagementService/pkg/mock"
)

type Reader string

func (Reader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func Test_transactionManagementServiceLogic_HealthCheck(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name  string
		setup func() datasource.DataSourceI
		want  bool
	}{
		{
			name: "Success",
			setup: func() datasource.DataSourceI {
				mockDs := mock.NewMockDataSourceI(mockCtrl)

				mockDs.EXPECT().HealthCheck().Times(1).
					Return(true)

				return mockDs
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementServiceLogic(tt.setup(), "")

			got := rec.HealthCheck()

			diff := testutil.Diff(got, tt.want)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}

func TestTransactionManagementServiceLogic_GetTransactions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name   string
		userId string
		setup  func() (datasource.DataSourceI, string)
		want   func(*respModel.Response)
	}{
		{
			name:   "Success :: Get Transaction",
			userId: "123",
			setup: func() (datasource.DataSourceI, string) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"user_id": "123"}, 5, 0).Times(1).Return(trans, 1, nil)
				return mockDs, ""
			},
			want: func(resp *respModel.Response) {
				var users = model.Transaction{AccountNumber: 1}
				temp := respModel.Response{
					Status:  http.StatusOK,
					Message: "SUCCESS",
					Data:    users,
				}
				res, ok := resp.Data.(model.PaginatedResponse)
				if !ok {
					t.Log("fail")
					t.Fail()
				}
				if !reflect.DeepEqual(&resp.Status, &temp.Status) {
					t.Errorf("Want: %v, Got: %v", &temp.Status, &resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", &temp.Message, &resp.Message)
				}
				if !reflect.DeepEqual(res.Response[0].AccountNumber, 1) {
					t.Errorf("Want: %v, Got: %v", 1, res.Response[0].AccountNumber)
				}
				if !reflect.DeepEqual(res.Pagination, model.Paginate{
					CurrentPage: 1,
					NextPage:    -1,
					TotalPage:   1,
				}) {
					t.Errorf("Want: %v, Got: %v", model.Paginate{
						CurrentPage: 1,
						NextPage:    -1,
						TotalPage:   1,
					}, res.Pagination)
				}
			},
		},
		{
			name:   "Success :: Get Transaction:: count_offset>limit",
			userId: "123",
			setup: func() (datasource.DataSourceI, string) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"user_id": "123"}, 5, 0).Times(1).Return(trans, 100, nil)
				return mockDs, ""
			},
			want: func(resp *respModel.Response) {
				var users = model.Transaction{AccountNumber: 1}
				temp := respModel.Response{
					Status:  http.StatusOK,
					Message: "SUCCESS",
					Data:    users,
				}
				res, ok := resp.Data.(model.PaginatedResponse)
				if !ok {
					t.Log("fail")
					t.Fail()
				}
				if !reflect.DeepEqual(&resp.Status, &temp.Status) {
					t.Errorf("Want: %v, Got: %v", &temp.Status, &resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", &temp.Message, &resp.Message)
				}
				if !reflect.DeepEqual(res.Response[0].AccountNumber, 1) {
					t.Errorf("Want: %v, Got: %v", 1, res.Response[0].AccountNumber)
				}
				if !reflect.DeepEqual(res.Pagination, model.Paginate{
					CurrentPage: 1,
					NextPage:    2,
					TotalPage:   20,
				}) {
					t.Errorf("Want: %v, Got: %v", model.Paginate{
						CurrentPage: 1,
						NextPage:    -1,
						TotalPage:   1,
					}, res.Pagination)
				}
			},
		},
		{
			name:   "Failure :: Get Transaction :: db err",
			userId: "123",
			setup: func() (datasource.DataSourceI, string) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Get(map[string]interface{}{"user_id": "123"}, 5, 0).Times(1).Return(nil, 0, errors.New("error"))
				return mockDs, ""
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrGetTransaction),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", &temp, resp)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementServiceLogic(tt.setup())

			got := rec.GetTransactions(tt.userId, 5, 1)

			tt.want(got)
		})
	}
}

type TestServer struct {
	srv *httptest.Server
	t   *testing.T
	wg  *sync.WaitGroup
}

func testClient(c *TestServer) {
	router := mux.NewRouter()
	router.HandleFunc("/microbank/v1/account/update/transaction", func(w http.ResponseWriter, r *http.Request) {
		defer c.wg.Done()
		defer c.t.Log("Hit")
		expectedReqBody, _ := json.Marshal(model.UpdateTransaction{
			AccountNumber:   0,
			Amount:          1000,
			TransactionType: "debit",
		})
		reqBody, _ := ioutil.ReadAll(r.Body)
		if string(reqBody) != string(expectedReqBody) {
			c.t.Errorf("Expected request body '%s', got '%s'", expectedReqBody, reqBody)
			c.t.Fail()
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPut)
	c.srv = httptest.NewServer(router)
}

func TestTransactionManagementServiceLogic_NewTransaction(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tStruct := &TestServer{
		t:  t,
		wg: &sync.WaitGroup{},
	}
	tests := []struct {
		name        string
		credentials model.NewTransaction
		setup       func() (datasource.DataSourceI, string)
		want        func(*respModel.Response)
	}{
		{
			name: "Success::transaction status != approved",
			credentials: model.NewTransaction{
				UserId: "123",
			},
			setup: func() (datasource.DataSourceI, string) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Insert(gomock.Any()).Times(1).DoAndReturn(func(tr model.Transaction) error {
					tr.TransactionId = ""
					tr.CreatedAt = time.Time{}
					diff := testutil.Diff(tr, model.Transaction{
						UserId:        "123",
						AccountNumber: 0,
						Amount:        0,
					})
					if diff != "" {
						t.Error(testutil.Callers(), diff)
					}
					return nil
				})
				return mockDs, ""
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusCreated,
					Message: "SUCCESS",
					Data:    nil,
				}
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", temp, resp)
				}
			},
		},
		{
			name: "Success::transaction status = approved",
			credentials: model.NewTransaction{
				UserId: "123",
				Type:   "debit",
				Status: "approved",
				Amount: 1000,
			},
			setup: func() (datasource.DataSourceI, string) {
				tStruct.wg.Add(1)
				testClient(tStruct)
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Insert(gomock.Any()).Times(1).Return(nil)

				return mockDs, tStruct.srv.URL
			},
			want: func(resp *respModel.Response) {
				tStruct.wg.Wait()
				tStruct.srv.Close()
				temp := respModel.Response{
					Status:  http.StatusCreated,
					Message: "SUCCESS",
					Data:    nil,
				}
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", temp, resp)
				}
			},
		},
		{
			name: "Failure::Get from db err",
			credentials: model.NewTransaction{
				UserId: "123",
			},
			setup: func() (datasource.DataSourceI, string) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Insert(gomock.Any()).Return(errors.New("error"))
				return mockDs, ""
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrNewTransaction),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", &temp, resp)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementServiceLogic(tt.setup())

			got := rec.NewTransaction(tt.credentials)

			tt.want(got)
		})
	}
}
