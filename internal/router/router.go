package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/example/currency-converter-api/internal/auth"
	"github.com/example/currency-converter-api/internal/config"
	"github.com/example/currency-converter-api/internal/models"
	"github.com/example/currency-converter-api/internal/rates"
	"github.com/shopspring/decimal"
)

func Register(g *gin.Engine, db *gorm.DB, rateSvc *rates.Service, cfg config.Config) {

	g.GET("/check", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	authGroup := g.Group("/auth")
	{
		authGroup.POST("/register", func(c *gin.Context) {
			var req struct{ Email, Password string }
			if err := c.ShouldBindJSON(&req); err != nil || !strings.Contains(req.Email, "@") || len(req.Password) < 6 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_email_or_password"})
				return
			}
			h, _ := auth.HashPassword(req.Password)
			u := models.User{Email: strings.ToLower(strings.TrimSpace(req.Email)), Password: h}
			if err := db.Create(&u).Error; err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": "email_taken"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email})
		})
		authGroup.POST("/login", func(c *gin.Context) {
			var req struct{ Email, Password string }
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
				return
			}
			var u models.User
			if err := db.Where("email = ?", strings.ToLower(req.Email)).First(&u).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
				return
			}
			if !auth.CheckPassword(u.Password, req.Password) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
				return
			}

			var prev models.Session
			if u.CurrentSessionID != nil {
				db.First(&prev, *u.CurrentSessionID)
				if prev.ID != 0 && prev.RevokedAt == nil {
					now := time.Now()
					db.Model(&prev).Update("revoked_at", &now)
				}
			}
			sess := models.Session{UserID: u.ID}
			if err := db.Create(&sess).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot_create_session"})
				return
			}
			db.Model(&u).Update("current_session_id", sess.ID)
			tok, _ := auth.MakeToken(cfg.JWTSecret, u.ID, sess.ID, 24*time.Hour)
			c.JSON(http.StatusOK, gin.H{"access_token": tok, "token_type": "Bearer"})
		})
	}

	// auth middleware
	mw := auth.AuthMiddleware(cfg.JWTSecret, func(userID, sessionID uint) bool {
		var u models.User
		if err := db.Select("id, current_session_id").First(&u, userID).Error; err != nil {
			return false
		}
		return u.CurrentSessionID != nil && *u.CurrentSessionID == sessionID
	})
	api := g.Group("/api", mw)

	// Rates
	api.GET("/rates", func(c *gin.Context) {
		base := c.Query("base")
		if base == "" {
			base = "USD"
		}
		base = strings.ToUpper(base)
		rm, fetchedAt, err := rateSvc.Latest(base)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rates_unavailable"})
			return
		}
		// optional symbols filter
		sym := c.Query("symbols")
		out := map[string]string{}
		if sym != "" {
			set := map[string]struct{}{}
			for _, s := range strings.Split(sym, ",") {
				set[strings.ToUpper(strings.TrimSpace(s))] = struct{}{}
			}
			for k, v := range rm {
				if _, ok := set[k]; ok {
					out[k] = v.String()
				}
			}
		} else {
			for k, v := range rm {
				out[k] = v.String()
			}
		}
		c.JSON(http.StatusOK, gin.H{"base": base, "fetchedAt": fetchedAt, "rates": out})
	})

	api.POST("/convert", func(c *gin.Context) {
		var req struct {
			Amount         string `json:"amount"`
			From, To, Base string
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}
		if req.Base == "" {
			req.Base = "USD"
		}
		amt, err := decimal.NewFromString(req.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_amount"})
			return
		}
		res, err := rateSvc.Convert(amt, req.From, req.To, req.Base)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"amount": req.Amount, "from": strings.ToUpper(req.From), "to": strings.ToUpper(req.To), "base": strings.ToUpper(req.Base), "converted": res.StringFixedBank(6)})
	})

	// Admin: manual refresh
	api.POST("/admin/refresh-rates", func(c *gin.Context) {
		var req struct{ Base string }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}
		if req.Base == "" {
			req.Base = "USD"
		}
		if err := rateSvc.RefreshAll(req.Base); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
