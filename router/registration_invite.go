package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"

	"github.com/gin-gonic/gin"
)

func registerRegistrationInviteRoutes(apiRouter *gin.RouterGroup) {
	registrationInviteRoute := apiRouter.Group("/registration-invites")
	registrationInviteRoute.Use(middleware.RootAuth())
	{
		registrationInviteRoute.GET("", controller.ListRegistrationInvites)
		registrationInviteRoute.POST("", middleware.DisableCache(), controller.CreateRegistrationInvites)
		registrationInviteRoute.POST("/:id/revoke", controller.RevokeRegistrationInvite)
	}
}
