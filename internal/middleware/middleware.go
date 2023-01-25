package middleware

import (
	"github.com/PereRohit/util/log"
	"github.com/PereRohit/util/response"
	"github.com/dgrijalva/jwt-go"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	svcCfg "github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/repo/authentication"
	"github.com/vatsal278/TransactionManagementService/pkg/session"
	"io"
	"net/http"
	"strings"
)

type SessionStruct struct {
	UserId interface{}
	cookie string
}
type TransactionMgmtMiddleware struct {
	cfg *svcCfg.Config
	jwt authentication.JWTService
	msg func(io.ReadCloser) (string, error)
}

func NewTransactionMgmtMiddleware(cfg *svcCfg.SvcConfig) *TransactionMgmtMiddleware {
	return &TransactionMgmtMiddleware{
		cfg: cfg.Cfg,
		jwt: cfg.JwtSvc.JwtSvc,
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
			response.ToJson(w, http.StatusInternalServerError, codes.GetErr(codes.ErrAssertUserid), nil)
			return
		}
		sessionStruct := SessionStruct{UserId: userId, cookie: cookie.Value}
		ctx := session.SetSession(r.Context(), sessionStruct)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
