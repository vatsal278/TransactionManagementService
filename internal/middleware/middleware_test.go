package middleware

import (
	"encoding/json"
	"errors"
	"github.com/PereRohit/util/log"
	"github.com/PereRohit/util/model"
	"github.com/PereRohit/util/response"
	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/repo/authentication"
	"github.com/vatsal278/TransactionManagementService/pkg/mock"
	"github.com/vatsal278/TransactionManagementService/pkg/session"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var x func()

var hit = false

func test(w http.ResponseWriter, r *http.Request) {
	hit = true
	c := r.Context()
	session := session.GetSession(c)
	sessionStruct, ok := session.(SessionStruct)
	if !ok {
		log.Error("error assert")
	}
	response.ToJson(w, http.StatusOK, "passed", sessionStruct.UserId)

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
					Status:  http.StatusInternalServerError,
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
			hit = false
			x := middleware.ExtractUser(http.HandlerFunc(test))
			x.ServeHTTP(res, req)

			tt.validator(res)

		})
	}
}
