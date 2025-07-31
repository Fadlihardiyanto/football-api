package http

import (
	"net/http"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/Fadlihardiyanto/football-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type GoalsController struct {
	GoalsUseCase usecase.GoalsUseCase
	Log          *logrus.Logger
}

func NewGoalsController(goalsUseCase usecase.GoalsUseCase, log *logrus.Logger) *GoalsController {
	return &GoalsController{
		GoalsUseCase: goalsUseCase,
		Log:          log,
	}
}

func (c *GoalsController) FindAll(ctx *gin.Context) {
	goals, err := c.GoalsUseCase.FindAll(ctx)
	if err != nil {
		c.Log.Errorf("Failed to find all goals: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(goals, "Goals found"))
}

func (c *GoalsController) FindByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Goal ID is required"),
		))
		return
	}

	goal, err := c.GoalsUseCase.FindByID(ctx, &model.GoalRequestFindByID{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to find goal by ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}
	if goal == nil {
		ctx.JSON(http.StatusNotFound, common.NewStandardErrorResponse(
			common.ErrNotFound("Goal not found").WithDetail("id", id),
		))
		return
	}
	ctx.JSON(http.StatusOK, model.NewSuccessResponse(goal, "Goal found"))
}

func (c *GoalsController) Create(ctx *gin.Context) {
	var req model.GoalRequestCreate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	res, err := c.GoalsUseCase.Create(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to create goal: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusCreated, model.NewSuccessResponse(res, "Goal created successfully"))
}

func (c *GoalsController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Goal ID is required"),
		))
		return
	}

	var req model.GoalRequestUpdate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	req.ID = id

	c.Log.Infof("Updating goal with ID %s", id)

	res, err := c.GoalsUseCase.Update(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to update goal with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Goal updated successfully"))
}

func (c *GoalsController) SoftDelete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Goal ID is required"),
		))
		return
	}

	res, err := c.GoalsUseCase.SoftDelete(ctx, &model.GoalRequestSoftDelete{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to soft delete goal with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	c.Log.Infof("Goal with ID %s soft deleted successfully", id)

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Goal soft deleted successfully"))
}
