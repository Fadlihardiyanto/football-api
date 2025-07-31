package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/Fadlihardiyanto/football-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TeamsController struct {
	TeamsUseCase usecase.TeamsUseCase
	Log          *logrus.Logger
}

func NewTeamsController(teamsUseCase usecase.TeamsUseCase, log *logrus.Logger) *TeamsController {
	return &TeamsController{
		TeamsUseCase: teamsUseCase,
		Log:          log,
	}
}

func (c *TeamsController) FindAll(ctx *gin.Context) {
	teams, err := c.TeamsUseCase.FindAll(ctx)
	if err != nil {
		c.Log.Errorf("Failed to find all teams: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(teams, "Teams found"))
}

func (c *TeamsController) FindByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Team ID is required"),
		))
		return
	}

	team, err := c.TeamsUseCase.FindByID(ctx, &model.TeamRequestFindByID{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to find team by ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}
	if team == nil {
		ctx.JSON(http.StatusNotFound, common.NewStandardErrorResponse(
			common.ErrNotFound("Team not found").WithDetail("id", id),
		))
		return
	}
	ctx.JSON(http.StatusOK, model.NewSuccessResponse(team, "Team found"))
}

func (c *TeamsController) Create(ctx *gin.Context) {
	var req model.TeamRequestCreate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	res, err := c.TeamsUseCase.Create(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to create team: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusCreated, model.NewSuccessResponse(res, "Team created successfully"))
}

func (c *TeamsController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Team ID is required"),
		))
		return
	}

	var req model.TeamRequestUpdate

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	req.ID = id

	c.Log.Infof("Updating team with ID %s", id)

	res, err := c.TeamsUseCase.Update(ctx, &req)
	if err != nil {
		c.Log.Errorf("Failed to update team with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Team updated successfully"))
}

func (c *TeamsController) SoftDelete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Team ID is required"),
		))
		return
	}

	c.Log.Infof("Soft deleting team with ID %s", id)

	res, err := c.TeamsUseCase.SoftDelete(ctx, &model.TeamRequestSoftDelete{ID: id})
	if err != nil {
		c.Log.Errorf("Failed to soft delete team with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Team soft deleted successfully"))
}

func (c *TeamsController) UploadLogo(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Team ID is required"),
		))
		return
	}

	file, err := ctx.FormFile("logo")
	if err != nil {
		c.Log.Errorf("Failed to get logo file: %v", err)
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Logo file is required"),
		))
		return
	}

	// validate extension and size if needed
	ext := filepath.Ext(file.Filename)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		c.Log.Warnf("Invalid logo file extension %s for team with ID %s", ext, id)
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid logo file extension"),
		))
		return
	}

	// validate file size
	if file.Size > 2*1024*1024 {
		c.Log.Warnf("Logo file size exceeds limit for team with ID %s", id)
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Logo file size exceeds limit"),
		))
		return
	}

	// path to save the logo
	destinationDir := fmt.Sprintf("uploads/logos/%s", id)
	err = os.MkdirAll(destinationDir, os.ModePerm)
	if err != nil {
		c.Log.Errorf("Failed to create directory for logo upload for team with ID %s: %v", id, err)
		ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
			common.ErrInternalServer("Failed to create directory"),
		))
		return
	}

	// generate file name
	// using original file name or generate a unique one
	if file.Filename == "" {
		c.Log.Warnf("Logo file name is empty for team with ID %s", id)
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Logo file name is required"),
		))
		return
	}

	uniqueID := uuid.New().String()
	fileName := fmt.Sprintf("%s%s", uniqueID, ext)
	filePath := filepath.Join(destinationDir, fileName)

	// save the file
	if err := ctx.SaveUploadedFile(file, filePath); err != nil {
		c.Log.Errorf("Failed to save logo file for team with ID %s: %v", id, err)
		ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
			common.ErrInternalServer("Failed to save logo file"),
		))
		return
	}

	res, err := c.TeamsUseCase.UploadLogo(ctx, &model.TeamRequestUploadLogo{
		ID:   id,
		Logo: filePath,
	})
	if err != nil {
		c.Log.Errorf("Failed to upload logo for team with ID %s: %v", id, err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Logo uploaded successfully"))
}
