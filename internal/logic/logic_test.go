package logic

import (
	"encoding/json"
	"errors"
	respModel "github.com/PereRohit/util/model"
	"github.com/PereRohit/util/response"
	"github.com/PereRohit/util/testutil"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
	"github.com/vatsal278/TransactionManagementService/pkg/mock"
	pdfMock "github.com/vatsal278/html-pdf-service/pkg/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
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
			rec := NewTransactionManagementServiceLogic(tt.setup(), config.ExternalSvc{})

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
		setup  func() (datasource.DataSourceI, config.ExternalSvc)
		want   func(*respModel.Response)
	}{
		{
			name:   "Success :: Get Transaction",
			userId: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"user_id": "123"}, 5, 0).Times(1).Return(trans, 1, nil)
				return mockDs, config.ExternalSvc{}
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
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"user_id": "123"}, 5, 0).Times(1).Return(trans, 100, nil)
				return mockDs, config.ExternalSvc{}
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
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Get(map[string]interface{}{"user_id": "123"}, 5, 0).Times(1).Return(nil, 0, errors.New("error"))
				return mockDs, config.ExternalSvc{}
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
		setup       func() (datasource.DataSourceI, config.ExternalSvc)
		want        func(*respModel.Response)
	}{
		{
			name: "Success::transaction status != approved",
			credentials: model.NewTransaction{
				UserId: "123",
			},
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
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
				return mockDs, config.ExternalSvc{}
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
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				tStruct.wg.Add(1)
				testClient(tStruct)
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Insert(gomock.Any()).Times(1).Return(nil)

				return mockDs, config.ExternalSvc{UserSvc: tStruct.srv.URL}
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
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Insert(gomock.Any()).Return(errors.New("error"))
				return mockDs, config.ExternalSvc{}
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

func TestTransactionManagementServiceLogic_DownloadTransactions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name        string
		credentials string
		setup       func() (datasource.DataSourceI, config.ExternalSvc)
		want        func(*respModel.Response)
	}{
		{
			name:        "Success :: DownloadPdf",
			credentials: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(trans, 1, nil)
				router := mux.NewRouter()
				router.HandleFunc("/microbank/v1/user", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"Name": "abc"})
				})
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)
				mockPdf.EXPECT().GeneratePdf(gomock.Any(), gomock.Any()).Return([]byte("PDF"), nil)
				srv := httptest.NewServer(router)
				return mockDs, config.ExternalSvc{UserSvc: srv.URL, PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusOK,
					Message: "SUCCESS",
					Data:    []byte("PDF"),
				}
				if !reflect.DeepEqual(resp.Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", temp.Message, resp.Message)
				}
				if !reflect.DeepEqual(&resp.Data, &temp.Data) {
					t.Errorf("Want: %v, Got: %v", temp.Data, resp.Data)
				}
			},
		},
		{
			name:        "Failure :: DownloadPdf :: error from db",
			credentials: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(trans, 1, errors.New("error db"))
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)
				return mockDs, config.ExternalSvc{UserSvc: "", PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrGetTransaction),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp.Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", temp.Message, resp.Message)
				}
				if !reflect.DeepEqual(&resp.Data, &temp.Data) {
					t.Errorf("Want: %v, Got: %v", temp.Data, resp.Data)
				}
			},
		},
		{
			name:        "Failure :: DownloadPdf :: no transaction found in db",
			credentials: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(trans, 1, nil)
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)
				return mockDs, config.ExternalSvc{UserSvc: "", PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusBadRequest,
					Message: codes.GetErr(codes.ErrGetTransaction),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp.Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", temp.Message, resp.Message)
				}
				if !reflect.DeepEqual(&resp.Data, &temp.Data) {
					t.Errorf("Want: %v, Got: %v", temp.Data, resp.Data)
				}
			},
		},
		{
			name:        "Failure :: DownloadPdf :: error making request to user svc",
			credentials: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(trans, 1, nil)
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)
				return mockDs, config.ExternalSvc{UserSvc: "", PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrFetchinDataUserSvc),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp.Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", temp.Message, resp.Message)
				}
				if !reflect.DeepEqual(&resp.Data, &temp.Data) {
					t.Errorf("Want: %v, Got: %v", temp.Data, resp.Data)
				}
			},
		},
		{
			name:        "Failure :: DownloadPdf ::not ok status code",
			credentials: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(trans, 1, nil)
				router := mux.NewRouter()
				router.HandleFunc("/microbank/v1/user", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusBadRequest, "Success", map[string]interface{}{"Name": "abc"})
				})
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)
				srv := httptest.NewServer(router)
				return mockDs, config.ExternalSvc{UserSvc: srv.URL, PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrFetchinDataUserSvc),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp.Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", temp.Message, resp.Message)
				}
				if !reflect.DeepEqual(&resp.Data, &temp.Data) {
					t.Errorf("Want: %v, Got: %v", temp.Data, resp.Data)
				}
			},
		},
		{
			name:        "Failure :: DownloadPdf :: error unmarshall response",
			credentials: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(trans, 1, nil)
				router := mux.NewRouter()
				router.HandleFunc("/microbank/v1/user", func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(Reader(""))
				})
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)
				srv := httptest.NewServer(router)
				return mockDs, config.ExternalSvc{UserSvc: srv.URL, PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrUnmarshall),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp.Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", temp.Message, resp.Message)
				}
				if !reflect.DeepEqual(&resp.Data, &temp.Data) {
					t.Errorf("Want: %v, Got: %v", temp.Data, resp.Data)
				}
			},
		},
		{
			name:        "Failure :: DownloadPdf :: error assert response data",
			credentials: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(trans, 1, nil)
				router := mux.NewRouter()
				router.HandleFunc("/microbank/v1/user", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", Reader(""))
				})
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)
				srv := httptest.NewServer(router)
				return mockDs, config.ExternalSvc{UserSvc: srv.URL, PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrAssertResp),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp.Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", temp.Message, resp.Message)
				}
				if !reflect.DeepEqual(&resp.Data, &temp.Data) {
					t.Errorf("Want: %v, Got: %v", temp.Data, resp.Data)
				}
			},
		},
		{
			name:        "Failure :: DownloadPdf :: error generate pdf",
			credentials: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var trans []model.Transaction
				trans = append(trans, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(trans, 1, nil)
				router := mux.NewRouter()
				router.HandleFunc("/microbank/v1/user", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"Name": "abc"})
				})
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)
				mockPdf.EXPECT().GeneratePdf(gomock.Any(), gomock.Any()).Return([]byte("PDF"), errors.New("pdf generate error"))
				srv := httptest.NewServer(router)
				return mockDs, config.ExternalSvc{UserSvc: srv.URL, PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrPdf),
					Data:    nil,
				}
				if !reflect.DeepEqual(resp.Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, resp.Status)
				}
				if !reflect.DeepEqual(&resp.Message, &temp.Message) {
					t.Errorf("Want: %v, Got: %v", temp.Message, resp.Message)
				}
				if !reflect.DeepEqual(&resp.Data, &temp.Data) {
					t.Errorf("Want: %v, Got: %v", temp.Data, resp.Data)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementServiceLogic(tt.setup())

			got := rec.DownloadTransaction(tt.credentials, "123")

			tt.want(got)
		})
	}
}
