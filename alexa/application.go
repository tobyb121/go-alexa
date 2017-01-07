package alexa

import (
	"github.com/gin-gonic/gin"
	"github.com/tobyb121/go-alexa/alexa/entities"
)

type Application struct {
	ApplicationID  string
	VerifyRequests bool
	OnLaunch       func(*gin.Context, *entities.Request, *entities.LaunchRequest) (*entities.Response, error)
	OnIntent       func(*gin.Context, *entities.Request, *entities.IntentRequest) (*entities.Response, error)
	OnSessionEnded func(*gin.Context, *entities.Request, *entities.SessionEndedRequest) (*entities.Response, error)
}

func NewApplication(ApplicationID string) *Application {
	return &Application{
		ApplicationID:  ApplicationID,
		VerifyRequests: true,
		OnLaunch:       nil,
		OnIntent:       nil,
		OnSessionEnded: nil,
	}
}
