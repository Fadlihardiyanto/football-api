package http

import (
	"net/http"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/Fadlihardiyanto/football-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MatchesController struct {
	MatchesUseCase usecase.MatchesUseCase
	Log            *logrus.Logger
}

func NewMatchesController(matchesUseCase usecase.MatchesUseCase, log *logrus.Logger) *MatchesController {
	return &MatchesController{
		MatchesUseCase: matchesUseCase,
		Log:            log,
	}
}

func (c *MatchesController) FindAll(ctx *gin.Context) {
	matches, err := c.MatchesUseCase.FindAll(ctx)
	if err != nil {
		c.Log.Errorf("Failed to find all matches: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(matches, "Matches found"))
}

func (c *MatchesController) FindByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Match ID is required"),
		))
		return
	}

	match, err := c.MatchesUseCase.FindByID(ctx, &model.MatchRequestFindByID{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to find match by ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}
	if match == nil {
		ctx.JSON(http.StatusNotFound, common.NewStandardErrorResponse(
			common.ErrNotFound("Match not found").WithDetail("id", id),
		))
		return
	}
	ctx.JSON(http.StatusOK, model.NewSuccessResponse(match, "Match found"))
}

func (c *MatchesController) Create(ctx *gin.Context) {
	var req model.MatchRequestCreate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Log.Errorf("Invalid JSON format: %v", err)
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"+err.Error()),
		))
		return
	}

	res, err := c.MatchesUseCase.Create(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to create match: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusCreated, model.NewSuccessResponse(res, "Match created successfully"))
}

func (c *MatchesController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Match ID is required"),
		))
		return
	}

	var req model.MatchRequestUpdate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	req.ID = id

	c.Log.Infof("Updating match with ID %s", id)

	res, err := c.MatchesUseCase.Update(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to update match with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Match updated successfully"))
}

func (c *MatchesController) SoftDelete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Match ID is required"),
		))
		return
	}

	res, err := c.MatchesUseCase.SoftDelete(ctx, &model.MatchRequestSoftDelete{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to soft delete match with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	c.Log.Infof("Match with ID %s soft deleted successfully", id)

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Match soft deleted successfully"))
}

func (c *MatchesController) GetMatchReport(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Match ID is required"),
		))
		return
	}

	res, err := c.MatchesUseCase.GetMatchReport(ctx, &model.MatchRequestFindByID{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to get match report for ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Match report retrieved successfully"))
}

func (c *MatchesController) FinishMatch(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Match ID is required"),
		))
		return
	}

	var req model.MatchRequestFinish
	req.ID = id

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Log.Errorf("Invalid JSON format: %v", err)
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"+err.Error()),
		))
		return
	}

	res, err := c.MatchesUseCase.FinishMatch(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to finish match with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Match finished successfully"))
}
