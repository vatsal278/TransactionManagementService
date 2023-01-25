package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	respModel "github.com/PereRohit/util/model"
	"github.com/PereRohit/util/testutil"
	"github.com/golang/mock/gomock"

	"github.com/vatsal278/TransactionManagementService/pkg/mock"
)

func TestAddHealthChecker(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	defer func() {
		c = common{}
	}()

	hc := &mock.MockHealthChecker{}
	type args struct {
		h HealthChecker
	}
	tests := []struct {
		name     string
		args     args
		validate func()
	}{
		{
			name: "Success",
			args: args{
				h: hc,
			},
			validate: func() {
				diff := testutil.Diff(len(c.services), 1)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
				diff = testutil.Diff(c.services[0], HealthChecker(hc))
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddHealthChecker(tt.args.h)
		})
	}
}

func TestNewCommonSvc(t *testing.T) {
	tests := []struct {
		name string
		want Commoner
	}{
		{
			name: "Success",
			want: &c,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCommonSvc()
			diff := testutil.Diff(got, tt.want)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}

func Test_common_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() http.ResponseWriter
		validate func(http.ResponseWriter)
	}{
		{
			name: "Success",
			setup: func() http.ResponseWriter {
				return httptest.NewRecorder()
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.setup()
			c.MethodNotAllowed(w, nil)
			tt.validate(w)
		})
	}
}

func Test_common_RouteNotFound(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() http.ResponseWriter
		validate func(http.ResponseWriter)
	}{
		{
			name: "Success",
			setup: func() http.ResponseWriter {
				return httptest.NewRecorder()
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.setup()
			c.RouteNotFound(w, nil)
			tt.validate(w)
		})
	}
}

func Test_common_HealthCheck(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	defer func() {
		c = common{}
	}()

	tests := []struct {
		name     string
		setup    func() http.ResponseWriter
		validate func(http.ResponseWriter)
	}{
		{
			name: "Success::1 ok and 1 not ok",
			setup: func() http.ResponseWriter {
				// prepare mock
				okSvc := mock.NewMockHealthChecker(mockCtrl)
				notOkSvc := mock.NewMockHealthChecker(mockCtrl)

				// set mock expectation and mocked return
				okSvc.EXPECT().HealthCheck().
					Return("service-1", "ok", true).
					Times(1)
				notOkSvc.EXPECT().HealthCheck().
					Return("service-2", "down", false).
					Times(1)

				// inject mocks
				AddHealthChecker(okSvc)
				AddHealthChecker(notOkSvc)

				return httptest.NewRecorder()
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
				hCdata := map[string]svcHealthStat{}

				err = json.Unmarshal(respDataB, &hCdata)
				diff = testutil.Diff(err, nil)
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
				resp.Data = hCdata
				t.Logf("%v", resp.Data)

				diff = testutil.Diff(resp, respModel.Response{
					Status:  http.StatusOK,
					Message: http.StatusText(http.StatusOK),
					Data: map[string]svcHealthStat{
						"service-1": {
							Status:  http.StatusText(http.StatusOK),
							Message: "ok",
						},
						"service-2": {
							Status:  "Not " + http.StatusText(http.StatusOK),
							Message: "down",
						},
					},
				})
				if diff != "" {
					t.Error(testutil.Callers(), diff)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.setup()
			c.HealthCheck(w, nil)
			tt.validate(w)
		})
	}
}
