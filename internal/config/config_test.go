package config

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PereRohit/util/config"
	"github.com/PereRohit/util/response"
	"github.com/PereRohit/util/testutil"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	jwtSvc "github.com/vatsal278/TransactionManagementService/internal/repo/authentication"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

func TestInitSvcConfig(t *testing.T) {
	type args struct {
		cfg Config
	}
	type testCase struct {
		name string
		args func() args
		want func(args) *SvcConfig
	}
	_, mock, err := sqlmock.NewWithDSN(":@tcp(:)/?charset=utf8mb4&parseTime=True")
	if err != nil {
		t.Log(err)
		t.Fail()

	}
	_, mock2, err := sqlmock.NewWithDSN(":@tcp(:)/newTemp?charset=utf8mb4&parseTime=True")
	if err != nil {
		t.Log()
		t.Fail()
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	router := mux.NewRouter()
	router.HandleFunc("/v1/register", func(w http.ResponseWriter, r *http.Request) {
		response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"id": "123"})
	})
	router.HandleFunc("/v1/registerfail", func(w http.ResponseWriter, r *http.Request) {
		response.ToJson(w, http.StatusInternalServerError, "Success", nil)
	})
	srv := httptest.NewServer(router)
	tests := []testCase{
		{
			name: "Success",
			args: func() args {
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( transaction_id VARCHAR(255) NOT NULL PRIMARY KEY, account_number INT NOT NULL, user_id VARCHAR(255) NOT NULL, amount DECIMAL(18,2) NOT NULL DEFAULT 0.00, transfer_to VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, status VARCHAR(255) NOT NULL, type VARCHAR(255) NOT NULL, comment VARCHAR(255) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				return args{
					cfg: Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:            CacheCfg{Duration: "1m"},
						Cookie:           CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl:    srv.URL,
						HtmlTemplateFile: "./../../docs/transaction-template.html",
					},
				}
			},
			want: func(arg args) *SvcConfig {
				required := &SvcConfig{
					JwtSvc: JWTSvc{JwtSvc: jwtSvc.JWTAuthService("")},
					Cfg: &Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:            CacheCfg{Duration: "1m", Time: time.Minute},
						Cookie:           CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl:    srv.URL,
						HtmlTemplateFile: "./../../docs/transaction-template.html",
						TemplateUuid:     "123",
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
					PdfSvc:              PdfSvc{},
					ExternalService:     ExternalSvc{PdfSvc: PdfSvc{UuId: "123"}},
				}
				return required
			},
		},
		{
			name: "Failure::Register file failure",
			args: func() args {
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( transaction_id VARCHAR(255) NOT NULL PRIMARY KEY, account_number INT NOT NULL, user_id VARCHAR(255) NOT NULL, amount DECIMAL(18,2) NOT NULL DEFAULT 0.00, transfer_to VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, status VARCHAR(255) NOT NULL, type VARCHAR(255) NOT NULL, comment VARCHAR(255) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				return args{
					cfg: Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:            CacheCfg{Duration: "1m"},
						Cookie:           CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl:    srv.URL + "fail",
						HtmlTemplateFile: "./../../docs/transaction-template.html",
					},
				}
			},
			want: func(arg args) *SvcConfig {
				required := &SvcConfig{
					JwtSvc: JWTSvc{JwtSvc: jwtSvc.JWTAuthService("")},
					Cfg: &Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:            CacheCfg{Duration: "1m", Time: time.Minute},
						Cookie:           CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl:    srv.URL,
						HtmlTemplateFile: "./../../docs/transaction-template.html",
						TemplateUuid:     "123",
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
					PdfSvc:              PdfSvc{},
					ExternalService:     ExternalSvc{PdfSvc: PdfSvc{UuId: "123"}},
				}
				return required
			},
		},
		{
			name: "Failure::error file not found",
			args: func() args {
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( transaction_id VARCHAR(255) NOT NULL PRIMARY KEY, account_number INT NOT NULL, user_id VARCHAR(255) NOT NULL, amount DECIMAL(18,2) NOT NULL DEFAULT 0.00, transfer_to VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, status VARCHAR(255) NOT NULL, type VARCHAR(255) NOT NULL, comment VARCHAR(255) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				return args{
					cfg: Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:            CacheCfg{Duration: "1m"},
						Cookie:           CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl:    srv.URL,
						HtmlTemplateFile: "",
					},
				}
			},
			want: func(arg args) *SvcConfig {
				required := &SvcConfig{
					JwtSvc: JWTSvc{JwtSvc: jwtSvc.JWTAuthService("")},
					Cfg: &Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:            CacheCfg{Duration: "1m", Time: time.Minute},
						Cookie:           CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl:    srv.URL,
						HtmlTemplateFile: "./../../docs/transaction-template.html",
						TemplateUuid:     "123",
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
					PdfSvc:              PdfSvc{},
					ExternalService:     ExternalSvc{PdfSvc: PdfSvc{UuId: "123"}},
				}
				return required
			},
		},
		{
			name: "Failure:: Parse time error",
			args: func() args {
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( transaction_id VARCHAR(255) NOT NULL PRIMARY KEY, account_number INT NOT NULL, user_id VARCHAR(255) NOT NULL, amount DECIMAL(18,2) NOT NULL DEFAULT 0.00, transfer_to VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, status VARCHAR(255) NOT NULL, type VARCHAR(255) NOT NULL, comment VARCHAR(255) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				return args{
					cfg: Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:            CacheCfg{Duration: "5"},
						Cookie:           CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl:    srv.URL,
						HtmlTemplateFile: "./../../docs/transaction-template.html",
					},
				}
			},
			want: func(arg args) *SvcConfig {
				required := &SvcConfig{
					JwtSvc: JWTSvc{JwtSvc: jwtSvc.JWTAuthService("")},
					Cfg: &Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:            CacheCfg{Duration: "1m", Time: time.Minute},
						Cookie:           CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl:    srv.URL,
						HtmlTemplateFile: "./../../docs/transaction-template.html",
						TemplateUuid:     "123",
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
					PdfSvc:              PdfSvc{},
					ExternalService:     ExternalSvc{PdfSvc: PdfSvc{UuId: "123"}},
				}
				return required
			},
		},
		{
			name: "Failure::Err Register Sub",
			args: func() args {
				router := mux.NewRouter()
				router.HandleFunc("/register/publisher", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"id": "123"})
				})
				srv := httptest.NewServer(router)
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( user_id varchar(225) not null unique, account_number int AUTO_INCREMENT, income dec(18,2) DEFAULT 0.00, spends dec(18,2) DEFAULT 0.00, created_on timestamp not null DEFAULT CURRENT_TIMESTAMP, updated_on timestamp not null DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, active_services json, inactive_services json, primary key (account_number), index(user_id) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))

				return args{
					cfg: Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:         CacheCfg{Duration: "1m"},
						Cookie:        CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl: srv.URL,
					},
				}
			},
			want: func(arg args) *SvcConfig {
				required := &SvcConfig{
					Cfg: &Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:  CacheCfg{Duration: "1m", Time: time.Minute},
						Cookie: CookieStruct{ExpiryStr: "5m"},
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
				}
				return required
			},
		},
		{
			name: "Failure::Err::Register Publisher",
			args: func() args {
				router := mux.NewRouter()
				router.HandleFunc("/register/subscriber", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", nil)
				})
				srv := httptest.NewServer(router)
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( user_id varchar(225) not null unique, account_number int AUTO_INCREMENT, income dec(18,2) DEFAULT 0.00, spends dec(18,2) DEFAULT 0.00, created_on timestamp not null DEFAULT CURRENT_TIMESTAMP, updated_on timestamp not null DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, active_services json, inactive_services json, primary key (account_number), index(user_id) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				return args{
					cfg: Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:         CacheCfg{Duration: "1m"},
						Cookie:        CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl: srv.URL,
					},
				}
			},
			want: func(arg args) *SvcConfig {
				required := &SvcConfig{
					Cfg: &Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:  CacheCfg{Duration: "1m", Time: time.Minute},
						Cookie: CookieStruct{ExpiryStr: "5m"},
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
				}
				return required
			},
		},
		{
			name: "Failure::Incorrect Private Key as Pem String",
			args: func() args {
				router := mux.NewRouter()
				router.HandleFunc("/register/publisher", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"id": "123"})
				})
				router.HandleFunc("/register/subscriber", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", nil)
				})
				srv := httptest.NewServer(router)
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( user_id varchar(225) not null unique, account_number int AUTO_INCREMENT, income dec(18,2) DEFAULT 0.00, spends dec(18,2) DEFAULT 0.00, created_on timestamp not null DEFAULT CURRENT_TIMESTAMP, updated_on timestamp not null DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, active_services json, inactive_services json, primary key (account_number), index(user_id) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))

				return args{
					cfg: Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:         CacheCfg{Duration: "1m"},
						Cookie:        CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl: srv.URL,
					},
				}
			},
			want: func(arg args) *SvcConfig {
				required := &SvcConfig{
					Cfg: &Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:  CacheCfg{Duration: "1m", Time: time.Minute},
						Cookie: CookieStruct{ExpiryStr: "5m"},
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
				}
				return required
			},
		},
		{
			name: "Failure:: Incorrect Cache Duration",
			args: func() args {
				router := mux.NewRouter()
				router.HandleFunc("/register/publisher", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"id": "123"})
				})
				router.HandleFunc("/register/subscriber", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", nil)
				})
				srv := httptest.NewServer(router)
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( user_id varchar(225) not null unique, account_number int AUTO_INCREMENT, income dec(18,2) DEFAULT 0.00, spends dec(18,2) DEFAULT 0.00, created_on timestamp not null DEFAULT CURRENT_TIMESTAMP, updated_on timestamp not null DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, active_services json, inactive_services json, primary key (account_number), index(user_id) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				return args{
					cfg: Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:         CacheCfg{Duration: "abc"},
						Cookie:        CookieStruct{ExpiryStr: "5m"},
						PdfServiceUrl: srv.URL,
					},
				}
			},
			want: func(arg args) *SvcConfig {
				required := &SvcConfig{
					Cfg: &Config{
						ServiceRouteVersion: "v2",
						ServerConfig:        config.ServerConfig{},
						DataBase: DbCfg{
							Driver: "sqlmock",
							DbName: "newTemp",
						},
						Cache:  CacheCfg{Duration: "1m", Time: time.Minute},
						Cookie: CookieStruct{ExpiryStr: "5m"},
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
				}
				return required
			},
		},
		{
			name: "Failure::DB Open 1",
			args: func() args {
				return args{cfg: Config{DataBase: DbCfg{Driver: "", DbName: "newTemp"}}}
			},
		},
		{
			name: "Failure::DB Prepare",
			args: func() args {

				router := mux.NewRouter()
				router.HandleFunc("/register/publisher", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"id": "123"})
				})
				router.HandleFunc("/register/subscriber", func(w http.ResponseWriter, r *http.Request) {
					response.ToJson(w, http.StatusOK, "Success", nil)
				})
				_ = httptest.NewServer(router)
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").WillReturnError(errors.New("error "))
				return args{cfg: Config{DataBase: DbCfg{Driver: "sqlmock", DbName: "newTemp"}}}
			},
		},
		{
			name: "Failure:DB Exec",
			args: func() args {
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(errors.New("error")).WillReturnResult(sqlmock.NewResult(1, 1))
				return args{cfg: Config{DataBase: DbCfg{Driver: "sqlmock", DbName: "newTemp"}}}
			},
		},
		{
			name: "Failure:: Exec err 2",
			args: func() args {
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( user_id varchar(225) not null unique, account_number int AUTO_INCREMENT, income dec(18,2) DEFAULT 0.00, spends dec(18,2) DEFAULT 0.00, created_on timestamp not null DEFAULT CURRENT_TIMESTAMP, updated_on timestamp not null DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, active_services json, inactive_services json, primary key (account_number), index(user_id) );")).WillReturnError(errors.New("error exec")).WillReturnResult(sqlmock.NewResult(1, 1))
				return args{cfg: Config{DataBase: DbCfg{Driver: "sqlmock", DbName: "newTemp"}}}
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				a := recover()
				if a != nil {
					t.Log("RECOVER"+tt.name, a)
				}
			}()
			s := tt.args()
			got := InitSvcConfig(s.cfg)
			got.DbSvc.Db = nil
			got.ExternalService.PdfSvc.PdfService = nil
			got.Cacher.Cacher = nil
			diff := testutil.Diff(got, tt.want(s))
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})

	}
}
