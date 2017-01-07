package alexa

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"encoding/json"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tobyb121/go-alexa/alexa/entities"
)

type Server struct {
	Application    Application
	VerifyRequests bool
}

func peekBody(request *http.Request) []byte {
	bodyData, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil
	}

	// Recreate the body for future reads
	request.Body = ioutil.NopCloser(bytes.NewReader(bodyData))
	return bodyData
}

func (s *Server) HandleRequest(ctx *gin.Context) {
	var req entities.Request
	bodyData := peekBody(ctx.Request)
	if json.Unmarshal(bodyData, &req) == nil {
		if err := s.VerifyRequest(ctx, &req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest", "message": "Request signature error"})
			return
		}
		var baseRequest entities.BaseRequest
		if err := json.Unmarshal(req.Request, &baseRequest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest", "message": "Invalid request format"})
			return
		}
		switch strings.ToLower(baseRequest.Type) {
		case "launchrequest":
			var launchRequest entities.LaunchRequest
			if err := json.Unmarshal(req.Request, &launchRequest); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest", "message": "Invalid request format"})
				return
			}
			if s.Application.OnLaunch != nil {
				if response, err := s.Application.OnLaunch(ctx, &req, &launchRequest); err != nil {
					ctx.JSON(http.StatusOK, response)
					return
				}
				ctx.JSON(http.StatusInternalServerError, gin.H{"status": "ServerError", "message": "Internal Server Error"})
				return

			}
		case "intentrequest":
			var intentRequest entities.IntentRequest
			if err := json.Unmarshal(req.Request, &intentRequest); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest", "message": "Invalid request format"})
				return
			}
			if s.Application.OnIntent != nil {
				if response, err := s.Application.OnIntent(ctx, &req, &intentRequest); err != nil {
					ctx.JSON(http.StatusOK, response)
					return
				}
				ctx.JSON(http.StatusInternalServerError, gin.H{"status": "ServerError", "message": "Internal Server Error"})
				return

			}
		case "sessionendedrequest":
			var sessionEndedRequest entities.SessionEndedRequest
			if err := json.Unmarshal(req.Request, &sessionEndedRequest); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest", "message": "Invalid request format"})
				return
			}
			if s.Application.OnSessionEnded != nil {
				if response, err := s.Application.OnSessionEnded(ctx, &req, &sessionEndedRequest); err != nil {
					ctx.JSON(http.StatusOK, response)
					return
				}
				ctx.JSON(http.StatusInternalServerError, gin.H{"status": "ServerError", "message": "Internal Server Error"})
				return

			}
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"status": "BadRequest", "message": "Request signature error"})
}
