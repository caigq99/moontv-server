package router

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/internal/cache"
	"github.com/moontv/server/internal/handler"
	"github.com/moontv/server/internal/middleware"
	"github.com/moontv/server/web"
)

func Setup(jwtSecret string, apikeySecret []byte) *gin.Engine {
	r := gin.Default()

	searchCache := cache.NewSearchCache(10 * time.Minute)

	authH := &handler.AuthHandler{JWTSecret: jwtSecret}
	apikeyH := &handler.APIKeyHandler{Secret: apikeySecret}
	searchH := &handler.SearchHandler{Cache: searchCache}
	sourceH := &handler.SourceHandler{}
	adminH := &handler.AdminHandler{}

	api := r.Group("/api")

	// Public
	api.GET("/ping", handler.Ping)
	api.POST("/auth/login", authH.Login)
	api.POST("/auth/register", authH.Register)

	// JWT-protected (admin panel operations)
	jwtGroup := api.Group("")
	jwtGroup.Use(middleware.JWTAuth(jwtSecret))
	{
		jwtGroup.POST("/user/apikey", apikeyH.Generate)
		jwtGroup.DELETE("/user/apikey", apikeyH.Revoke)
	}

	// API-key-protected (external API calls)
	apiKeyGroup := api.Group("")
	apiKeyGroup.Use(middleware.APIKeyAuth(apikeySecret), middleware.RequestLog())
	{
		apiKeyGroup.GET("/search", searchH.Search)
		apiKeyGroup.GET("/search/sse", searchH.SearchSSE)
		apiKeyGroup.GET("/detail", searchH.Detail)
		apiKeyGroup.GET("/suggest", searchH.Suggest)

		apiKeyGroup.GET("/sources", sourceH.List)
		apiKeyGroup.POST("/sources", sourceH.Create)
		apiKeyGroup.PUT("/sources/sort", sourceH.Sort)
		apiKeyGroup.PUT("/sources/:key", sourceH.Update)
		apiKeyGroup.DELETE("/sources/:key", sourceH.Delete)
	}

	// Admin (JWT + admin role)
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth(jwtSecret), middleware.AdminOnly())
	{
		adminGroup.GET("/stats", adminH.Stats)
		adminGroup.GET("/users", adminH.ListUsers)
		adminGroup.PUT("/users/:id/ban", adminH.BanUser)
		adminGroup.DELETE("/users/:id", adminH.DeleteUser)

		adminGroup.POST("/invites", adminH.GenerateInvites)
		adminGroup.GET("/invites", adminH.ListInvites)
		adminGroup.DELETE("/invites/:code", adminH.DeleteInvite)

		adminGroup.GET("/sources", adminH.ListGlobalSources)
		adminGroup.POST("/sources", adminH.CreateGlobalSource)
		adminGroup.PUT("/sources/sort", adminH.SortGlobalSources)
		adminGroup.PUT("/sources/:key", adminH.UpdateGlobalSource)
		adminGroup.DELETE("/sources/:key", adminH.DeleteGlobalSource)
	}

	// Serve admin panel static files
	staticFS, _ := fs.Sub(web.Static, "static")
	r.StaticFS("/admin", http.FS(staticFS))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin")
	})

	return r
}
