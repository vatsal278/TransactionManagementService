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
	srv := httptest.NewServer(router)
	tests := []testCase{
		{
			name: "Success",
			args: func() args {
				mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectClose()
				mock2.ExpectExec(regexp.QuoteMeta("create table if not exists ( transaction_id VARCHAR(255) NOT NULL PRIMARY KEY, account_number INT NOT NULL, user_id VARCHAR(255) NOT NULL, amount DECIMAL(18,2) NOT NULL DEFAULT 0.00, transfer_to VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, status VARCHAR(255) NOT NULL, type VARCHAR(255) NOT NULL, comment VARCHAR(255) );")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				//pdfmock.EXPECT().Register(gomock.Any()).Return("123", nil)
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
						HtmlTemplateFile: "./../../docs/scratch.html",
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
						HtmlTemplateFile: "./../../docs/scratch.html",
					},
					ServiceRouteVersion: "v2",
					SvrCfg:              config.ServerConfig{},
					PdfSvc:              PdfSvc{},
					UtilService:         UtilSvc{PdfSvc: PdfSvc{UuId: "123"}},
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
		//{
		//	name: "fail",
		//	args: func() (*sql.DB, args) {
		//		router := mux.NewRouter()
		//		router.HandleFunc("/register/publisher", func(w http.ResponseWriter, r *http.Request) {
		//			response.ToJson(w, http.StatusOK, "Success", map[string]interface{}{"id": "123"})
		//		})
		//		router.HandleFunc("/register/subscriber", func(w http.ResponseWriter, r *http.Request) {
		//			response.ToJson(w, http.StatusOK, "Success", nil)
		//		})
		//		srv := httptest.NewServer(router)
		//		db, mock, err := sqlmock.NewWithDSN(":@tcp(:)/?charset=utf8mb4&parseTime=True")
		//		if err != nil {
		//			t.Fail()
		//		}
		//		mock.ExpectPrepare("CREATE SCHEMA IF NOT EXISTS newTemp ;").ExpectExec().WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
		//		return db, args{
		//			cfg: Config{
		//				ServiceRouteVersion: "v2",
		//				ServerConfig:        config.ServerConfig{},
		//				DataBase: DbCfg{
		//					Driver: "sqlmock",
		//					DbName: "newTemp",
		//				},
		//				MessageQueue: MsgQueueCfg{SvcUrl: srv.URL, ActivatedAccountChannel: "account.activation.channel", NewAccountChannel: "new.account.channel", Key: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb1FJQkFBS0NBUUVBejdiaXVpck01TEJjN2pzeTFwS0dpNkZsNm0zWXYrZnFzYzhHMFNIQlFRQlUyYmJHCmI4MHhpM1h0bW1zVURkWGhyNXhaMnA0c0Q4T0J2QjV0eWpZZmJyY1oyc2gzcVp1Rm1EZExpbGJjSUVKdUdaM2UKTWRrZWpRRFdTKzJvMmFyQXI5dFBqTGVKTXk4THhUVVlKNmw4NnFTVVl0aDJNMWtXcUsrc2hWK3hlNnI4NjR0VApGeENvWEp1NU1ILys4Y3ZLeTlIMHh5OHR1V0JKVGF0V2lJd3pqaGlEU1NEclZiMVA2UDJYL2NZOGo5WmJQZW9hCm5aV3VkNlhDYnZJN1Faa1lSMFpreTEyb1grZWxTa2Vpa1pkQWxQbmh6OTNnTUw5cUhYWmIvdC9YWXZOSXRjalMKVzA4cFBXNGUvN2NsMkJVMHBaN2g1dWVEenBjSFhNbUd1WmlPT3dJREFRQUJBb0lCQVFDTHhsRFo0QlZTeXM4dQpUTTNRRUhmVG5EOWR1cCtCdkFsMXI0K3h5Vm9uYUphd2pzc0h6dmZKRmdsV3dUbVVlZG5OOTVPTGhxYTEwT1VMCmR4cUFXVjFiZm9FNmRXMzR4enZtQzBlZEJ3aEgrUXZuMXhEL1VGQzdwOVdNOEplUUtkUlNRbTFNanZFWGJWQXAKVzZvdWZtSWQ3N1Flcy9VT1pxUFZ6YWwxY3NpWEkyanhEd3F4ZkQ1Mk5CUUZqaXpLeXhJUHRxU2xNTGhjZG8rYQpwZzg1RTBDVTNPSGRXS0dack0vbjJ1UGhiRnRNWTNDOTVwTjh6eERHUnpQNlByME9xL1MrTHlid1FydUpNUU1CCmZ2enVPUGNHSzUrcVhlSlBPSzB1ZnFhbUhEQ2tHanJHZGRHZmZ0M0lMbzRCV0xrM1I4TXBVYUdoR0hTeng3QVoKNmtCOHhkUTVBb0dCQVBVNTlmOWNTcHh3ZXd4clY4Nm13OWRMTmNZMEsvVXhsME9EQlFBMlVleGJETFo4VnZocwpCcXRzZjdQZFRpdU53WEc0ZnE0bU5SVnJ1K2NyYWhlREI3ZTBqdk9LOGdpQ0QxdlhsZzN5WDZhbDZsUTQ1OXJYCkpHVmExc2hkRjBOdkFYTXR0N2FzQW1QK1kzRXlDSnpQK2NZOFZGaExJRHdlMFJzYzdJQ29Ma1Z2QW9HQkFOalgKQjlGVS8xYy8yUk9UaDdZYldjdHh2dDBtaDMrcStZaU9jQnZsQUMzbXNxc0NNTSt1ZGZ3U2FNQlpoSXhxUWRYUgpTQTF0VFgyY1hKZ0VpZ1RGNnJBNGdudmxOZW5QTGJvNlJydmZoNkdsejdIZXZ4cGY1ZWtYNHRTM3FEcWFWSDBGClZvY25jSzc3L21DWmM0VVVHeDJDTHYvbFNtZzRzaUtCQTltZFFoWDFBbjlyU1BCV3lBbmNaMWx1RlloVTRLRE4Ka0JuMm5OeWVhUlBFZFkyNmlnbE5Yb2d4VGpTK2VvUndld2RqcVc2Sm4zc0NSYlVtZTVDOXptUm12cGVyc2FldQp0MC9UUFBhbXdqLzE3bHUzdmxJYWxudnVYUGNTeHcwbFNwaXRFQTBkYzNNdThORnZHZEh4N1ZtVUxFK1lTMlQ3ClZXbVJOMHpqQUpoN1JDdzBIV0FoQW9HQUNHQ21hS3dFQVhieUNCT1hGcTRQMWhCYTgyaGRxODBMUHY5aHpYSVgKZzY1NkVLbFJBWFVZRWRrVU92bzZhTUppTU1TWktBdWxCc2xYdW5mU2JVVElRRzZ1ZStMckpsRmV6dWNaZklDeQpXTWh6TWNnTlVoT0thbXNGMUhvVUFjK2NuQWZzdytQK01vU0IyM0dTU1AzeDNqMzlXdDJjOWxIYWNBTFVCMEJRCklWRUNnWUJETHhjM1o3Y1o3VU5PNndRdy8xTWRGZXEwRGJ6dlY0Z1hlVkpoUHZWeU1RR2hiZjVESTFXMngwWGgKUjBHeUVyZTIrVGVXZEswRUZCS3FwTVRlSFNZNmM2QWY1UUZmOTJsRW1PSlVHOVhXQ3FBS3pqMHFUY0l1bldsWgpsaXpWdjJIM0hDbmlTWXMzbG9kSFFsSzcyZTBzcC9Kd25QN2hjMHlsNmh1cHdDeVQ3UT09Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0t"},
		//				Cookie:       CookieStruct{},
		//			},
		//		}
		//	},
		//	want: func(arg args, db *sql.DB) *SvcConfig {
		//		s := sdk.NewMsgBrokerSvc(arg.cfg.MessageQueue.SvcUrl)
		//		privateKey, _ := crypt.PEMStrAsPrivKey(arg.cfg.MessageQueue.Key)
		//		required := &SvcConfig{
		//			JwtSvc: JWTSvc{JwtSvc: jwtSvc.JWTAuthService("")},
		//			Cfg: &Config{
		//				ServiceRouteVersion: "v2",
		//				ServerConfig:        config.ServerConfig{},
		//				DataBase: DbCfg{
		//					Driver: "sqlmock",
		//					DbName: "newTemp",
		//				},
		//				MessageQueue: MsgQueueCfg{SvcUrl: arg.cfg.MessageQueue.SvcUrl, NewAccountChannel: "new.account.channel", ActivatedAccountChannel: "account.activation.channel", Key: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb1FJQkFBS0NBUUVBejdiaXVpck01TEJjN2pzeTFwS0dpNkZsNm0zWXYrZnFzYzhHMFNIQlFRQlUyYmJHCmI4MHhpM1h0bW1zVURkWGhyNXhaMnA0c0Q4T0J2QjV0eWpZZmJyY1oyc2gzcVp1Rm1EZExpbGJjSUVKdUdaM2UKTWRrZWpRRFdTKzJvMmFyQXI5dFBqTGVKTXk4THhUVVlKNmw4NnFTVVl0aDJNMWtXcUsrc2hWK3hlNnI4NjR0VApGeENvWEp1NU1ILys4Y3ZLeTlIMHh5OHR1V0JKVGF0V2lJd3pqaGlEU1NEclZiMVA2UDJYL2NZOGo5WmJQZW9hCm5aV3VkNlhDYnZJN1Faa1lSMFpreTEyb1grZWxTa2Vpa1pkQWxQbmh6OTNnTUw5cUhYWmIvdC9YWXZOSXRjalMKVzA4cFBXNGUvN2NsMkJVMHBaN2g1dWVEenBjSFhNbUd1WmlPT3dJREFRQUJBb0lCQVFDTHhsRFo0QlZTeXM4dQpUTTNRRUhmVG5EOWR1cCtCdkFsMXI0K3h5Vm9uYUphd2pzc0h6dmZKRmdsV3dUbVVlZG5OOTVPTGhxYTEwT1VMCmR4cUFXVjFiZm9FNmRXMzR4enZtQzBlZEJ3aEgrUXZuMXhEL1VGQzdwOVdNOEplUUtkUlNRbTFNanZFWGJWQXAKVzZvdWZtSWQ3N1Flcy9VT1pxUFZ6YWwxY3NpWEkyanhEd3F4ZkQ1Mk5CUUZqaXpLeXhJUHRxU2xNTGhjZG8rYQpwZzg1RTBDVTNPSGRXS0dack0vbjJ1UGhiRnRNWTNDOTVwTjh6eERHUnpQNlByME9xL1MrTHlid1FydUpNUU1CCmZ2enVPUGNHSzUrcVhlSlBPSzB1ZnFhbUhEQ2tHanJHZGRHZmZ0M0lMbzRCV0xrM1I4TXBVYUdoR0hTeng3QVoKNmtCOHhkUTVBb0dCQVBVNTlmOWNTcHh3ZXd4clY4Nm13OWRMTmNZMEsvVXhsME9EQlFBMlVleGJETFo4VnZocwpCcXRzZjdQZFRpdU53WEc0ZnE0bU5SVnJ1K2NyYWhlREI3ZTBqdk9LOGdpQ0QxdlhsZzN5WDZhbDZsUTQ1OXJYCkpHVmExc2hkRjBOdkFYTXR0N2FzQW1QK1kzRXlDSnpQK2NZOFZGaExJRHdlMFJzYzdJQ29Ma1Z2QW9HQkFOalgKQjlGVS8xYy8yUk9UaDdZYldjdHh2dDBtaDMrcStZaU9jQnZsQUMzbXNxc0NNTSt1ZGZ3U2FNQlpoSXhxUWRYUgpTQTF0VFgyY1hKZ0VpZ1RGNnJBNGdudmxOZW5QTGJvNlJydmZoNkdsejdIZXZ4cGY1ZWtYNHRTM3FEcWFWSDBGClZvY25jSzc3L21DWmM0VVVHeDJDTHYvbFNtZzRzaUtCQTltZFFoWDFBbjlyU1BCV3lBbmNaMWx1RlloVTRLRE4Ka0JuMm5OeWVhUlBFZFkyNmlnbE5Yb2d4VGpTK2VvUndld2RqcVc2Sm4zc0NSYlVtZTVDOXptUm12cGVyc2FldQp0MC9UUFBhbXdqLzE3bHUzdmxJYWxudnVYUGNTeHcwbFNwaXRFQTBkYzNNdThORnZHZEh4N1ZtVUxFK1lTMlQ3ClZXbVJOMHpqQUpoN1JDdzBIV0FoQW9HQUNHQ21hS3dFQVhieUNCT1hGcTRQMWhCYTgyaGRxODBMUHY5aHpYSVgKZzY1NkVLbFJBWFVZRWRrVU92bzZhTUppTU1TWktBdWxCc2xYdW5mU2JVVElRRzZ1ZStMckpsRmV6dWNaZklDeQpXTWh6TWNnTlVoT0thbXNGMUhvVUFjK2NuQWZzdytQK01vU0IyM0dTU1AzeDNqMzlXdDJjOWxIYWNBTFVCMEJRCklWRUNnWUJETHhjM1o3Y1o3VU5PNndRdy8xTWRGZXEwRGJ6dlY0Z1hlVkpoUHZWeU1RR2hiZjVESTFXMngwWGgKUjBHeUVyZTIrVGVXZEswRUZCS3FwTVRlSFNZNmM2QWY1UUZmOTJsRW1PSlVHOVhXQ3FBS3pqMHFUY0l1bldsWgpsaXpWdjJIM0hDbmlTWXMzbG9kSFFsSzcyZTBzcC9Kd25QN2hjMHlsNmh1cHdDeVQ3UT09Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0t"},
		//			},
		//			DbSvc:               DbSvc{Db: db},
		//			ServiceRouteVersion: "v2",
		//			SvrCfg:              config.ServerConfig{},
		//			MsgBrokerSvc:        MsgQueue{MsgBroker: s, PubId: "123", Channel: "account.activation.channel", PrivateKey: *privateKey},
		//		}
		//		return required
		//	},
		//},
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
			got.UtilService.PdfSvc.PdfService = nil
			got.Cacher.Cacher = nil
			diff := testutil.Diff(got, tt.want(s))
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})

	}
}
