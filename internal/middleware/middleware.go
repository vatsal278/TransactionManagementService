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

// TransactionMgmtMiddleware is a middleware struct that includes a configuration object, a JWT service,
// and a Redis cacher. It is responsible for handling authentication and caching for requests.
type TransactionMgmtMiddleware struct {
	cfg    *svcCfg.Config
	jwt    authentication.JWTService
	cacher redis.Cacher
}

// respWriterWithStatus is a wrapper for http.ResponseWriter that includes the status code and response
// from a given request.
type respWriterWithStatus struct {
	status   int
	response string
	http.ResponseWriter
}

// WriteHeader overrides the default WriteHeader method to set the status code of the response.
func (w *respWriterWithStatus) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Write overrides the default Write method to set the response of the request.
func (w *respWriterWithStatus) Write(d []byte) (int, error) {
	w.response = string(d)
	return w.ResponseWriter.Write(d)
}

// NewTransactionMgmtMiddleware is a constructor function that returns a new instance of the TransactionMgmtMiddleware struct.
func NewTransactionMgmtMiddleware(cfg *svcCfg.SvcConfig) *TransactionMgmtMiddleware {
	return &TransactionMgmtMiddleware{
		cfg:    cfg.Cfg,
		jwt:    cfg.JwtSvc.JwtSvc,
		cacher: cfg.Cacher.Cacher,
	}
}

// ExtractUser is a middleware function that extracts user information from a JWT cookie and sets it in the request context.
func (u TransactionMgmtMiddleware) ExtractUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("token")
		if err != nil {
			log.Error(err)
			response.ToJson(w, http.StatusUnauthorized, codes.GetErr(codes.ErrUnauthorized), nil)
			return
		}
		if cookie.Value == "" {
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

// Cacher returns a middleware function that can be used to cache HTTP responses using the provided cache implementation.
// The middleware checks the cache for an existing response for the current request URL and, if found, writes it to the response writer and returns without invoking the next handler.
// Otherwise, the middleware calls the next handler to generate a response and caches the response for future requests.
// The middleware also optionally adds the user ID from the session to the cache key if requireAuth is true.
func (t TransactionMgmtMiddleware) Cacher(requireAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var key string
			var cacheResponse model.CacheResponse
			key = fmt.Sprint(r.URL.String())

			// If authentication is required, append the user ID to the cache key
			if requireAuth != false {
				sessionStruct := session.GetSession(r.Context())
				session, ok := sessionStruct.(model.SessionStruct)
				if !ok {
					response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
					return
				}
				key = fmt.Sprint(key + "/auth/" + session.UserId)
			}

			// Check the cache for an existing response
			Cacher := t.cacher
			by, err := Cacher.Get(key)
			if err == nil {
				// If a cached response is found, write it to the response writer and return
				err = json.Unmarshal(by, &cacheResponse)
				if err != nil {
					log.Error(err)
					response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrUnmarshall), nil)
					return
				}
				w.Header().Set("Content-Type", cacheResponse.ContentType)
				w.Write([]byte(cacheResponse.Response))
				w.WriteHeader(cacheResponse.Status)
				return
			}

			// If no cached response is found, call the next handler to generate a response
			hijackedWriter := &respWriterWithStatus{-1, "", w}
			next.ServeHTTP(hijackedWriter, r)

			// If the response status is not in the 2xx range, do not cache the response
			if hijackedWriter.status < 200 || hijackedWriter.status >= 300 {
				return
			}

			// Otherwise, cache the response for future requests
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
			err = Cacher.Set(key, byt, t.cfg.Cache.Time)
			if err != nil {
				log.Error(err)
				return
			}
		})
	}
}
