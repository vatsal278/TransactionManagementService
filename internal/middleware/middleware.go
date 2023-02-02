package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/PereRohit/util/log"
	"github.com/PereRohit/util/response"
	"github.com/dgrijalva/jwt-go"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	svcCfg "github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/authentication"
	"github.com/vatsal278/TransactionManagementService/pkg/session"
	"github.com/vatsal278/go-redis-cache"
	"net/http"
	"strings"
)

type TransactionMgmtMiddleware struct {
	cfg    *svcCfg.Config
	jwt    authentication.JWTService
	cacher redis.Cacher
}

type respWriterWithStatus struct {
	status   int
	response string
	http.ResponseWriter
}

func (w *respWriterWithStatus) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *respWriterWithStatus) Write(d []byte) (int, error) {
	w.response = string(d)
	return w.ResponseWriter.Write(d)
}

func NewTransactionMgmtMiddleware(cfg *svcCfg.SvcConfig) *TransactionMgmtMiddleware {
	return &TransactionMgmtMiddleware{
		cfg:    cfg.Cfg,
		jwt:    cfg.JwtSvc.JwtSvc,
		cacher: cfg.Cacher.Cacher,
	}
}
func (u TransactionMgmtMiddleware) ExtractUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("token")
		if err != nil {
			log.Error(err)
			response.ToJson(w, http.StatusUnauthorized, codes.GetErr(codes.ErrUnauthorized), nil)
			return
		}
		if cookie.Value == "" {
			log.Error(err)
			response.ToJson(w, http.StatusUnauthorized, codes.GetErr(codes.ErrUnauthorized), nil)
			return
		}
		token, err := u.jwt.ValidateToken(cookie.Value)
		if err != nil {
			log.Error(err)
			if strings.Contains(err.Error(), "Token is expired") {
				response.ToJson(w, http.StatusUnauthorized, codes.GetErr(codes.ErrTokenExpired), nil)
				return
			}
			response.ToJson(w, http.StatusUnauthorized, codes.GetErr(codes.ErrMatchingToken), nil)
			return
		}
		if !token.Valid {
			response.ToJson(w, http.StatusUnauthorized, codes.GetErr(codes.ErrUnauthorized), nil)
			return
		}
		mapClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.ToJson(w, http.StatusInternalServerError, codes.GetErr(codes.ErrAssertClaims), nil)
			return
		}
		userId, ok := mapClaims["user_id"]
		if !ok {
			response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
			return
		}
		userIdStr, ok := userId.(string)
		if !ok {
			response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
			return
		}
		sessionStruct := model.SessionStruct{UserId: userIdStr, Cookie: cookie.Value}
		ctx := session.SetSession(r.Context(), sessionStruct)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (t TransactionMgmtMiddleware) Cacher(requireAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var key string
			var cacheResponse model.CacheResponse
			key = fmt.Sprint(r.URL.String())
			if requireAuth != false {
				sessionStruct := session.GetSession(r.Context())
				session, ok := sessionStruct.(model.SessionStruct)
				if !ok {
					response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
					return
				}
				key = fmt.Sprint(key + "/auth/" + session.UserId)
			}
			Cacher := t.cacher
			by, err := Cacher.Get(key)
			if err == nil {
				err = json.Unmarshal(by, &cacheResponse)
				if err != nil {
					log.Error(err)
					return
				}
				w.Header().Set("Content-Type", cacheResponse.ContentType)
				w.Write([]byte(cacheResponse.Response))
				w.WriteHeader(cacheResponse.Status)
				return
			}
			hijackedWriter := &respWriterWithStatus{-1, "", w}
			next.ServeHTTP(hijackedWriter, r)
			if hijackedWriter.status < 200 || hijackedWriter.status >= 300 {
				return
			}
			cacheResponse = model.CacheResponse{
				Status:      hijackedWriter.status,
				Response:    hijackedWriter.response,
				ContentType: w.Header().Get("Content-Type"),
			}
			byt, err := json.Marshal(cacheResponse)
			if err != nil {
				log.Error(err)
				return
			}
			log.Info(string(byt))
			err = Cacher.Set(key, byt, t.cfg.Cache.Time)
			if err != nil {
				log.Error(err)
				return
			}
		})
	}
}
