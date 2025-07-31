package http

import (
	"net/http"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/Fadlihardiyanto/football-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PlayersController struct {
	PlayersUseCase usecase.PlayersUseCase
	Log            *logrus.Logger
}

func NewPlayersController(playersUseCase usecase.PlayersUseCase, log *logrus.Logger) *PlayersController {
	return &PlayersController{
		PlayersUseCase: playersUseCase,
		Log:            log,
	}
}

func (c *PlayersController) FindAll(ctx *gin.Context) {
	players, err := c.PlayersUseCase.FindAll(ctx)
	if err != nil {
		c.Log.Errorf("Failed to find all players: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(players, "Players found"))
}

func (c *PlayersController) FindByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Player ID is required"),
		))
		return
	}

	player, err := c.PlayersUseCase.FindByID(ctx, &model.PlayerRequestFindByID{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to find player by ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}
	if player == nil {
		ctx.JSON(http.StatusNotFound, common.NewStandardErrorResponse(
			common.ErrNotFound("Player not found").WithDetail("id", id),
		))
		return
	}
	ctx.JSON(http.StatusOK, model.NewSuccessResponse(player, "Player found"))
}

func (c *PlayersController) Create(ctx *gin.Context) {
	var req model.PlayerRequestCreate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	res, err := c.PlayersUseCase.Create(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to create player: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusCreated, model.NewSuccessResponse(res, "Player created successfully"))
}

func (c *PlayersController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Player ID is required"),
		))
		return
	}

	var req model.PlayerRequestUpdate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	req.ID = id

	c.Log.Infof("Updating player with ID %s", id)

	res, err := c.PlayersUseCase.Update(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to update player with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Player updated successfully"))
}

func (c *PlayersController) SoftDelete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Player ID is required"),
		))
		return
	}

	res, err := c.PlayersUseCase.SoftDelete(ctx, &model.PlayerRequestSoftDelete{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to soft delete player with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	c.Log.Infof("Player with ID %s soft deleted successfully", id)

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Player soft deleted successfully"))
}
