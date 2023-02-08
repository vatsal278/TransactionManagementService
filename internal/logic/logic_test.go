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
				var paginatedResponse = model.PaginatedResponse{Response: []model.Transaction{{AccountNumber: 1, UserId: "123"}}, Pagination: model.Paginate{
					CurrentPage: 1,
					NextPage:    -1,
					TotalPage:   1,
				}}
				res, ok := resp.Data.(model.PaginatedResponse)
				if !ok {
					t.Log("fail")
					t.Fail()
				}
				if !reflect.DeepEqual(&res, &paginatedResponse) {
					t.Errorf("Want: %v, Got: %v", &paginatedResponse, &res)
					return
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
				var paginatedResponse = model.PaginatedResponse{Response: []model.Transaction{{AccountNumber: 1, UserId: "123"}}, Pagination: model.Paginate{
					CurrentPage: 1,
					NextPage:    2,
					TotalPage:   20,
				}}
				res, ok := resp.Data.(model.PaginatedResponse)
				if !ok {
					t.Log("fail")
					t.Fail()
				}
				if !reflect.DeepEqual(&res, &paginatedResponse) {
					t.Errorf("Want: %v, Got: %v", &paginatedResponse, &res)
					return
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
	hit bool
}

func testClient(hit *bool) func(*TestServer) {
	return func(c *TestServer) {
		router := mux.NewRouter()
		router.HandleFunc("/microbank/v1/account/update/transaction", func(w http.ResponseWriter, r *http.Request) {
			defer c.wg.Done()
			defer c.t.Log("Hit")
			*hit = true
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
		router.HandleFunc("/fail/microbank/v1/account/update/transaction", func(w http.ResponseWriter, r *http.Request) {
			defer c.wg.Done()
			defer c.t.Log("Hit")
			w.WriteHeader(http.StatusInternalServerError)
		}).Methods(http.MethodPut)
		c.srv = httptest.NewServer(router)
	}
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
				x := testClient(&tStruct.hit)
				x(tStruct)
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Insert(gomock.Any()).Times(1).Return(nil)

				return mockDs, config.ExternalSvc{AccSvcUrl: tStruct.srv.URL}
			},
			want: func(resp *respModel.Response) {
				tStruct.wg.Wait()
				tStruct.srv.Close()
				if !tStruct.hit {
					t.Errorf("Want: %v, Got: %v", true, tStruct.hit)
				}
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
			name: "Failure::client do failure",
			credentials: model.NewTransaction{
				UserId: "123",
				Type:   "debit",
				Status: "approved",
				Amount: 1000,
			},
			setup: func() (datasource.DataSourceI, string) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				mockDs.EXPECT().Insert(gomock.Any()).Times(1).Return(nil)

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
		name          string
		transactionId string
		setup         func() (datasource.DataSourceI, config.ExternalSvc)
		want          func(*respModel.Response)
	}{
		{
			name:          "Success :: DownloadPdf",
			transactionId: "123",
			setup: func() (datasource.DataSourceI, config.ExternalSvc) {
				mockDs := mock.NewMockDataSourceI(mockCtrl)
				var transactions []model.Transaction
				transactions = append(transactions, model.Transaction{UserId: "123", AccountNumber: 1})
				mockDs.EXPECT().Get(map[string]interface{}{"transaction_id": "123"}, 0, 0).Times(1).Return(transactions, 1, nil)
				router := mux.NewRouter()
				router.HandleFunc("/microbank/v1/user", func(w http.ResponseWriter, r *http.Request) {
					val, err := r.Cookie("token")
					if err != nil {
						t.Errorf("Want: %v, Got: %v", "", err)
						return
					}
					if val.Value != "123" {
						t.Errorf("Want: %v, Got: %v", "", val.Value)
						return
					}
					response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"Name": "abc"})
				})
				mockPdf := pdfMock.NewMockHtmlToPdfSvcI(mockCtrl)

				mockPdf.EXPECT().GeneratePdf(map[string]interface{}{
					"Name":                      "abc",
					"TransferFromAccountNumber": transactions[0].AccountNumber,
					"TransferToAccountNumber":   transactions[0].TransferTo,
					"TransactionId":             transactions[0].TransactionId,
					"Amount":                    transactions[0].Amount,
					"Date":                      transactions[0].CreatedAt,
					"Status":                    transactions[0].Status,
					"Type":                      transactions[0].Type,
					"Comment":                   transactions[0].Comment,
				}, "11-22-33-44").Return([]byte("PDF"), nil)
				srv := httptest.NewServer(router)
				return mockDs, config.ExternalSvc{UserSvc: srv.URL, PdfSvc: config.PdfSvc{UuId: "11-22-33-44", PdfService: mockPdf}}
			},
			want: func(resp *respModel.Response) {
				temp := respModel.Response{
					Status:  http.StatusOK,
					Message: "SUCCESS",
					Data:    []byte("PDF"),
				}
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", &temp, resp)
				}
			},
		},
		{
			name:          "Failure :: DownloadPdf :: error from db",
			transactionId: "123",
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
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", &temp, resp)
				}
			},
		},
		{
			name:          "Failure :: DownloadPdf :: no transaction found in db",
			transactionId: "123",
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
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", temp, &resp)
				}
			},
		},
		{
			name:          "Failure :: DownloadPdf :: error making request to user svc",
			transactionId: "123",
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
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", temp, &resp)
				}
			},
		},
		{
			name:          "Failure :: DownloadPdf ::not ok status code",
			transactionId: "123",
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
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", temp, &resp)
				}
			},
		},
		{
			name:          "Failure :: DownloadPdf :: error unmarshall response",
			transactionId: "123",
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
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", temp, &resp)
				}
			},
		},
		{
			name:          "Failure :: DownloadPdf :: error assert response data",
			transactionId: "123",
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
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", temp, &resp)
				}
			},
		},
		{
			name:          "Failure :: DownloadPdf :: error generate pdf",
			transactionId: "123",
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
				if !reflect.DeepEqual(resp, &temp) {
					t.Errorf("Want: %v, Got: %v", temp, &resp)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementServiceLogic(tt.setup())

			got := rec.DownloadTransaction(tt.transactionId, "123")

			tt.want(got)
		})
	}
}
