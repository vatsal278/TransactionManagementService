package middleware

import (
	"encoding/json"
	"errors"
	"github.com/PereRohit/util/model"
	"github.com/PereRohit/util/response"
	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	model2 "github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/authentication"
	"github.com/vatsal278/TransactionManagementService/pkg/mock"
	"github.com/vatsal278/TransactionManagementService/pkg/session"
	redisMock "github.com/vatsal278/go-redis-cache/mocks"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

var x func()

var hit = false

func test(w http.ResponseWriter, r *http.Request) {
	hit = true
	c := r.Context()
	sessionStruct := session.GetSession(c)
	session, ok := sessionStruct.(model2.SessionStruct)
	if !ok {
		response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
		return
	}
	response.ToJson(w, http.StatusOK, "passed", session.UserId)

}

func TestTransactionMgmtMiddleware_ExtractUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		name      string
		setupFunc func() (*http.Request, authentication.JWTService)
		validator func(*httptest.ResponseRecorder)
	}{
		{
			name: "SUCCESS::ExtractUser",
			setupFunc: func() (*http.Request, authentication.JWTService) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)

				token := jwtGo.Token{
					Claims: jwtGo.MapClaims{"user_id": "123"},
					Valid:  true,
				}
				mockJwtSvc.EXPECT().ValidateToken("jwtToken").Return(&token, nil)
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: "jwtToken",
				})
				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != true {
					t.Errorf("Want: %v, Got: %v", true, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					return
				}
				expected := &model.Response{
					Status:  http.StatusOK,
					Message: "passed",
					Data:    "123",
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name: "Failure::ExtractUser:: empty token value",
			setupFunc: func() (*http.Request, authentication.JWTService) {
				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: "",
				})
				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					return
				}
				expected := &model.Response{
					Status:  http.StatusUnauthorized,
					Message: codes.GetErr(codes.ErrUnauthorized),
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name: "Failure::ExtractUser:: no cookie found",
			setupFunc: func() (*http.Request, authentication.JWTService) {
				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)
				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					t.Fail()
					return
				}
				expected := &model.Response{
					Status:  http.StatusUnauthorized,
					Message: codes.GetErr(codes.ErrUnauthorized),
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name: "Failure::ExtractUser:: compared literals not same",
			setupFunc: func() (*http.Request, authentication.JWTService) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: " jwtToken",
				})
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)
				err := errors.New(" err ")
				mockJwtSvc.EXPECT().ValidateToken(gomock.Any()).Return(nil, err)

				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					t.Fail()
					return
				}
				expected := &model.Response{
					Status:  http.StatusUnauthorized,
					Message: codes.GetErr(codes.ErrMatchingToken),
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name: "Failure::ExtractUser:: Token is expired ",
			setupFunc: func() (*http.Request, authentication.JWTService) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				//authentication := jwtSvc.JWTAuthService()
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: "123",
				})
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)
				err := errors.New(" Token is expired ")
				mockJwtSvc.EXPECT().ValidateToken(gomock.Any()).Return(nil, err)

				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					t.Fail()
					return
				}
				expected := &model.Response{
					Status:  http.StatusUnauthorized,
					Message: codes.GetErr(codes.ErrTokenExpired),
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name: "Failure::ExtractUser:: Token is invalid ",
			setupFunc: func() (*http.Request, authentication.JWTService) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: "123",
				})
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)
				token := jwtGo.Token{Valid: false}
				mockJwtSvc.EXPECT().ValidateToken(gomock.Any()).Return(&token, nil)

				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					t.Fail()
					return
				}
				expected := &model.Response{
					Status:  http.StatusUnauthorized,
					Message: codes.GetErr(codes.ErrUnauthorized),
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name: "Failure::ExtractUser:: mapClaims not ok",
			setupFunc: func() (*http.Request, authentication.JWTService) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				//authentication := jwtSvc.JWTAuthService()
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: "123",
				})
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)
				token := jwtGo.Token{
					Claims: nil,
					Valid:  true,
				}
				mockJwtSvc.EXPECT().ValidateToken(gomock.Any()).Return(&token, nil)

				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					t.Fail()
					return
				}
				expected := &model.Response{
					Status:  http.StatusInternalServerError,
					Message: codes.GetErr(codes.ErrAssertClaims),
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name: "Failure::ExtractUser:: user id not in claims",
			setupFunc: func() (*http.Request, authentication.JWTService) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: "123",
				})
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)
				token := jwtGo.Token{
					Claims: jwtGo.MapClaims{},
					Valid:  true,
				}
				mockJwtSvc.EXPECT().ValidateToken(gomock.Any()).Return(&token, nil)

				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					t.Fail()
					return
				}
				expected := &model.Response{
					Status:  http.StatusBadRequest,
					Message: codes.GetErr(codes.ErrAssertUserid),
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name: "Failure::ExtractUser:: user id not string",
			setupFunc: func() (*http.Request, authentication.JWTService) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: "123",
				})
				mockJwtSvc := mock.NewMockJWTService(mockCtrl)
				token := jwtGo.Token{
					Claims: jwtGo.MapClaims{"user_id": 1200},
					Valid:  true,
				}
				mockJwtSvc.EXPECT().ValidateToken(gomock.Any()).Return(&token, nil)

				return req, mockJwtSvc
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
				}
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					t.Fail()
					return
				}
				expected := &model.Response{
					Status:  http.StatusBadRequest,
					Message: codes.GetErr(codes.ErrAssertUserid),
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
	}

	// to execute the tests in the table
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// STEP 1: seting up all instances for the specific test case
			res := httptest.NewRecorder()
			req, jwt := tt.setupFunc()

			// STEP 2: call the test function
			middleware := NewTransactionMgmtMiddleware(&config.SvcConfig{
				JwtSvc: config.JWTSvc{
					JwtSvc: jwt,
				},
				Cfg: &config.Config{},
			})
			hit := false
			x := middleware.ExtractUser(http.HandlerFunc(test))
			x.ServeHTTP(res, req)

			tt.validator(res)

		})
	}
}
func TestTransactionMgmtMiddleware_Cacher(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		name           string
		config         config.Config
		extractMsgFunc func(closer io.ReadCloser) (string, error)
		setupFunc      func() (*http.Request, *redisMock.MockCacher)
		validator      func(*httptest.ResponseRecorder)
	}{
		{
			name:   "SUCCESS::Cacher::Cached Response",
			config: config.Config{},
			setupFunc: func() (*http.Request, *redisMock.MockCacher) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				ctx := session.SetSession(req.Context(), model2.SessionStruct{UserId: "123"})
				mockCacher := redisMock.NewMockCacher(mockCtrl)
				cacheResponse := model2.CacheResponse{Status: http.StatusOK, Response: "ok", ContentType: "application/json"}
				b, _ := json.Marshal(cacheResponse)
				mockCacher.EXPECT().Get("http://localhost:80/auth/123").Return(b, nil)
				return req.WithContext(ctx), mockCacher
			},
			extractMsgFunc: func(closer io.ReadCloser) (string, error) {
				return "", nil
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
					return
				}
				by, _ := ioutil.ReadAll(res.Body)
				if !reflect.DeepEqual([]byte("ok"), by) {
					t.Errorf("Want: %v, Got: %v", "ok", string(by))
				}
			},
		},
		{
			name:   "Failure::Cacher::Cached Response::Err assert id",
			config: config.Config{},
			setupFunc: func() (*http.Request, *redisMock.MockCacher) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				ctx := session.SetSession(req.Context(), 1)
				mockCacher := redisMock.NewMockCacher(mockCtrl)
				return req.WithContext(ctx), mockCacher
			},
			extractMsgFunc: func(closer io.ReadCloser) (string, error) {
				return "", nil
			},
			validator: func(res *httptest.ResponseRecorder) {
				by, _ := ioutil.ReadAll(res.Body)
				result := model.Response{}
				err := json.Unmarshal(by, &result)
				if err != nil {
					t.Log(err)
					return
				}
				expected := &model.Response{
					Status:  http.StatusBadRequest,
					Message: "1004: unable to assert userid",
					Data:    nil,
				}
				if !reflect.DeepEqual(&result, expected) {
					t.Errorf("Want: %v, Got: %v", expected, result)
				}
			},
		},
		{
			name:   "Failure::Cacher::Cached Response::unmarshal error",
			config: config.Config{},
			setupFunc: func() (*http.Request, *redisMock.MockCacher) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				ctx := session.SetSession(req.Context(), model2.SessionStruct{UserId: "123"})
				mockCacher := redisMock.NewMockCacher(mockCtrl)
				mockCacher.EXPECT().Get("http://localhost:80/auth/123").Return([]byte("123"), nil)
				return req.WithContext(ctx), mockCacher
			},
			extractMsgFunc: func(closer io.ReadCloser) (string, error) {
				return "", nil
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != false {
					t.Errorf("Want: %v, Got: %v", false, hit)
					return
				}
				by, _ := ioutil.ReadAll(res.Body)
				if !reflect.DeepEqual([]byte(""), by) {
					t.Errorf("Want: %v, Got: %v", "", string(by))
				}
			},
		},
		{
			name:   "SUCCESS::Cacher::Normal Response",
			config: config.Config{Cache: config.CacheCfg{Time: time.Minute}},
			setupFunc: func() (*http.Request, *redisMock.MockCacher) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				ctx := session.SetSession(req.Context(), model2.SessionStruct{UserId: "123"})
				mockCacher := redisMock.NewMockCacher(mockCtrl)
				//x := model.Response{Status: 200, Message: "passed", Data: "123"}
				//y, _ := json.Marshal(x)
				//cacheResponse := model2.CacheResponse{Status: http.StatusOK, Response: string(y), ContentType: "application/json"}
				mockCacher.EXPECT().Get("http://localhost:80/auth/123").Return(nil, errors.New("error"))
				mockCacher.EXPECT().Set("http://localhost:80/auth/123", []byte("{\"Status\":200,\"Response\":\"{\\\"status\\\":200,\\\"message\\\":\\\"passed\\\",\\\"data\\\":\\\"123\\\"}\\n\",\"ContentType\":\"application/json\"}"), time.Minute)
				return req.WithContext(ctx), mockCacher
			},
			extractMsgFunc: func(closer io.ReadCloser) (string, error) {
				return "", nil
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != true {
					t.Errorf("Want: %v, Got: %v", true, hit)
					return
				}
				var resp model.Response
				by, _ := ioutil.ReadAll(res.Body)
				json.Unmarshal(by, &resp)
				expectedResp := model.Response{Status: 200, Message: "passed", Data: "123"}
				if !reflect.DeepEqual(expectedResp, resp) {
					t.Errorf("Want: %v, Got: %v", expectedResp, resp)
				}
			},
		},
		{
			name:   "Failure::Cacher::Normal Response::Redis fail",
			config: config.Config{Cache: config.CacheCfg{Time: time.Minute}},
			setupFunc: func() (*http.Request, *redisMock.MockCacher) {

				req := httptest.NewRequest(http.MethodGet, "http://localhost:80", nil)
				ctx := session.SetSession(req.Context(), model2.SessionStruct{UserId: "123"})
				mockCacher := redisMock.NewMockCacher(mockCtrl)
				mockCacher.EXPECT().Get("http://localhost:80/auth/123").Return(nil, errors.New("error"))
				mockCacher.EXPECT().Set("http://localhost:80/auth/123", []byte("{\"Status\":200,\"Response\":\"{\\\"status\\\":200,\\\"message\\\":\\\"passed\\\",\\\"data\\\":\\\"123\\\"}\\n\",\"ContentType\":\"application/json\"}"), time.Minute).Return(errors.New("error"))
				return req.WithContext(ctx), mockCacher
			},
			extractMsgFunc: func(closer io.ReadCloser) (string, error) {
				return "", nil
			},
			validator: func(res *httptest.ResponseRecorder) {
				if hit != true {
					t.Errorf("Want: %v, Got: %v", true, hit)
					return
				}
				var resp model.Response
				by, _ := ioutil.ReadAll(res.Body)
				json.Unmarshal(by, &resp)
				expectedResp := model.Response{Status: 200, Message: "passed", Data: "123"}
				if !reflect.DeepEqual(expectedResp, resp) {
					t.Errorf("Want: %v, Got: %v", expectedResp, string(by))
				}
			},
		},
	}

	// to execute the tests in the table
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// STEP 1: seting up all instances for the specific test case
			res := httptest.NewRecorder()
			req, cacher := tt.setupFunc()
			cfg := config.SvcConfig{Cfg: &tt.config}
			middleware := TransactionMgmtMiddleware{
				cacher: cacher,
				cfg:    cfg.Cfg}
			hit = false
			x := middleware.Cacher(true)
			x(http.HandlerFunc(test)).ServeHTTP(res, req)

			tt.validator(res)

		})
	}
}
