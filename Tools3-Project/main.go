package main

import (
	"Tools3-Project/config"
	controllers "Tools3-Project/controller"
	"Tools3-Project/routes"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.ConnectDB()
	controllers.InitUserCollection(db)
	controllers.InitEventCollection(db)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//sessions
	store := cookie.NewStore([]byte("super-secret-key"))
	r.Use(sessions.Sessions("mysession", store))
	/*
		r.POST("/register", controllers.Register)
		r.POST("/login", controllers.Login)
		r.GET("/logout", controllers.Logout)

	*/

	//auth routes
	routes.AuthRoutes(r)
	//event routes
	routes.EventRoutes(r)

	fmt.Println("✅ Server running on http://localhost:8080")
	r.Run(":8080")
}
