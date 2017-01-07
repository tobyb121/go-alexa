package alexa

import (
	"github.com/gin-gonic/gin"
	"github.com/tobyb121/go-alexa/alexa/entities"
)

type Application struct {
	ApplicationID  string
	VerifyRequests bool
	OnLaunch       func(*gin.Context, entities.Request, entities.LaunchRequest)
	OnIntent       func(*gin.Context, entities.Request, entities.IntentRequest)
	OnSessionEnded func(*gin.Context, entities.Request, entities.SessionEndedRequest)
}

func NewApplication(ApplicationID string) *Application {
	return &Application{
		ApplicationID:  ApplicationID,
		VerifyRequests: true,
	}
}
