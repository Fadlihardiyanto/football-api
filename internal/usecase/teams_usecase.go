package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	messaging "github.com/Fadlihardiyanto/football-api/internal/gateway"
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/Fadlihardiyanto/football-api/internal/model/converter"
	"github.com/Fadlihardiyanto/football-api/internal/repository"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TeamsUseCase interface {
	FindAll(ctx context.Context) ([]model.TeamResponse, error)
	FindByID(ctx context.Context, request *model.TeamRequestFindByID) (*model.TeamResponse, error)
	Create(ctx context.Context, request *model.TeamRequestCreate) (*model.TeamResponse, error)
	Update(ctx context.Context, request *model.TeamRequestUpdate) (*model.TeamResponse, error)
	SoftDelete(ctx context.Context, request *model.TeamRequestSoftDelete) (*model.TeamResponse, error)
	UploadLogo(ctx context.Context, request *model.TeamRequestUploadLogo) (*model.TeamResponse, error)
}

type teamsUseCaseImpl struct {
	TeamsRepo    repository.TeamsRepository
	LogsProducer *messaging.LogProducer
	DB           *gorm.DB
	Log          *logrus.Logger
}

func NewTeamsUseCase(teamsRepo repository.TeamsRepository, logsProducer *messaging.LogProducer, db *gorm.DB, log *logrus.Logger) TeamsUseCase {
	return &teamsUseCaseImpl{
		TeamsRepo:    teamsRepo,
		LogsProducer: logsProducer,
		DB:           db,
		Log:          log,
	}
}

func (t *teamsUseCaseImpl) FindAll(ctx context.Context) ([]model.TeamResponse, error) {
	tx := t.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	teams, err := t.TeamsRepo.FindAllWithRelations(tx, "Players")
	if err != nil {
		t.Log.Errorf("Failed to find all teams: %v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to find teams")
	}

	if err := tx.Commit().Error; err != nil {
		t.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	var responses []model.TeamResponse
	for _, team := range teams {
		responses = append(responses, *converter.ToTeamResponse(&team))
	}

	return responses, nil
}

func (t *teamsUseCaseImpl) FindByID(ctx context.Context, request *model.TeamRequestFindByID) (*model.TeamResponse, error) {
	tx := t.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validasi input
	if err := common.ValidateStruct(request); err != nil {
		t.Log.Warnf("Invalid request body: %+v", err)
		return nil, err
	}

	team, err := t.TeamsRepo.FindByIDWithRelations(tx, request.ID, "Players")
	if err != nil {
		t.Log.Errorf("Failed to find team by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Team not found").WithDetail("id", request.ID)
	}

	if err := tx.Commit().Error; err != nil {
		t.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	return converter.ToTeamResponse(team), nil
}

func (t *teamsUseCaseImpl) Create(ctx context.Context, request *model.TeamRequestCreate) (*model.TeamResponse, error) {
	tx := t.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		t.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	team := &entity.Team{
		ID:                  uuid.New().String(),
		Name:                request.Name,
		Logo:                request.Logo,
		FoundedYear:         request.FoundedYear,
		HeadquartersAddress: request.HeadquartersAddress,
		HeadquartersCity:    request.HeadquartersCity,
	}

	// check if team already exists by name
	exists, err := t.TeamsRepo.CheckTeamExistsByName(tx, team.Name)
	if err != nil {
		tx.Rollback()
		t.Log.Errorf("Failed to check team existence by name %s: %v", team.Name, err)
		// send log event
		logEvent := &model.LogEvent{
			Level:   "error",
			Message: fmt.Sprintf("Failed to check team existence by name %s: %v", team.Name, err),
			Service: "teams",
			Time:    time.Now().Format(time.RFC3339),
		}
		t.LogsProducer.Send(logEvent)
		return nil, common.ErrInternalServer("Failed to check team existence")
	}
	if exists {
		tx.Rollback()
		t.Log.Warnf("Team with name %s already exists", team.Name)
		return nil, common.ErrConflict("Team with this name already exists").WithDetail("name", team.Name)
	}

	// check validation year founded
	currentYear := time.Now().Year()
	if team.FoundedYear > currentYear {
		tx.Rollback()
		t.Log.Warnf("Invalid founded year %d for team %s", team.FoundedYear, team.Name)
		return nil, common.ErrInvalidInput("Invalid founded year").WithDetail("year", fmt.Sprintf("%d", team.FoundedYear))
	}

	if err := t.TeamsRepo.Create(tx, team); err != nil {
		tx.Rollback()
		t.Log.Errorf("Failed to create team: %v", err)
		// send log event
		logEvent := &model.LogEvent{
			Level:   "error",
			Message: fmt.Sprintf("Failed to create team %s with error: %v", team.Name, err),
			Service: "teams",
			Time:    time.Now().Format(time.RFC3339),
		}
		t.LogsProducer.Send(logEvent)
		return nil, common.ErrInternalServer("Failed to create team")
	}

	if err := tx.Commit().Error; err != nil {
		t.Log.Errorf("Failed to commit transaction: %v", err)
		// send log event
		logEvent := &model.LogEvent{
			Level:   "error",
			Message: fmt.Sprintf("Failed to commit transaction for team %s: %v", team.Name, err),
			Service: "teams",
			Time:    time.Now().Format(time.RFC3339),
		}
		t.LogsProducer.Send(logEvent)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: "Team created successfully",
		Service: "teams",
		Time:    time.Now().Format(time.RFC3339),
	}

	t.Log.Infof("Sending log event: %+v", logEvent)
	if err := t.LogsProducer.Send(logEvent); err != nil {
		t.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToTeamResponse(team), nil
}

func (t *teamsUseCaseImpl) Update(ctx context.Context, request *model.TeamRequestUpdate) (*model.TeamResponse, error) {
	tx := t.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		t.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	team, err := t.TeamsRepo.FindByID(tx, request.ID)
	if err != nil {
		t.Log.Errorf("Failed to find team by ID %s: %v", request.ID, err)
		tx.Rollback()
		// send log event
		logEvent := &model.LogEvent{
			Level:   "error",
			Message: fmt.Sprintf("Failed to find team by ID %s: %v", request.ID, err),
			Service: "teams",
			Time:    time.Now().Format(time.RFC3339),
		}
		t.LogsProducer.Send(logEvent)
		return nil, common.ErrNotFound("Team not found").WithDetail("id", request.ID)
	}

	if request.Name != "" && request.Name != team.Name {
		exist, err := t.TeamsRepo.CheckTeamExistsByName(tx, request.Name)
		if err != nil {
			tx.Rollback()
			t.Log.Errorf("Failed to check team name existence: %v", err)
			return nil, common.ErrInternalServer("Failed to check team name")
		}
		if exist {
			tx.Rollback()
			return nil, common.ErrConflict("Team name already exists")
		}
		team.Name = request.Name
	}

	if request.FoundedYear != 0 && request.FoundedYear != team.FoundedYear {
		currentYear := time.Now().Year()
		if request.FoundedYear > currentYear {
			tx.Rollback()
			return nil, common.ErrInvalidInput("Invalid founded year").WithDetail("year", fmt.Sprintf("%d", request.FoundedYear))
		}
		team.FoundedYear = request.FoundedYear
	}

	if request.Name != "" {
		team.Name = request.Name
	}
	if request.Logo != "" {
		team.Logo = request.Logo
	}
	if request.FoundedYear != 0 {
		team.FoundedYear = request.FoundedYear
	}
	if request.HeadquartersAddress != "" {
		team.HeadquartersAddress = request.HeadquartersAddress
	}
	if request.HeadquartersCity != "" {
		team.HeadquartersCity = request.HeadquartersCity
	}

	if err := t.TeamsRepo.Update(tx, team); err != nil {
		tx.Rollback()
		t.Log.Errorf("Failed to update team: %v", err)
		return nil, common.ErrInternalServer("Failed to update team")
	}

	if err := tx.Commit().Error; err != nil {
		t.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	return converter.ToTeamResponse(team), nil
}

func (t *teamsUseCaseImpl) SoftDelete(ctx context.Context, request *model.TeamRequestSoftDelete) (*model.TeamResponse, error) {
	tx := t.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		t.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	team, err := t.TeamsRepo.FindByID(tx, request.ID)
	if err != nil {
		t.Log.Errorf("Failed to find team by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Team not found").WithDetail("id", request.ID)
	}

	if err := t.TeamsRepo.SoftDelete(tx, team.ID); err != nil {
		tx.Rollback()
		t.Log.Errorf("Failed to soft delete team: %v", err)
		return nil, common.ErrInternalServer("Failed to soft delete team")
	}

	if err := tx.Commit().Error; err != nil {
		t.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Team with ID %s soft deleted successfully", request.ID),
		Service: "teams",
		Time:    time.Now().Format(time.RFC3339),
	}
	t.Log.Infof("Sending log event: %+v", logEvent)

	return converter.ToTeamResponse(team), nil
}

func (t *teamsUseCaseImpl) UploadLogo(ctx context.Context, request *model.TeamRequestUploadLogo) (*model.TeamResponse, error) {
	tx := t.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		t.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	team, err := t.TeamsRepo.FindByID(tx, request.ID)
	if err != nil {
		t.Log.Errorf("Failed to find team by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Team not found").WithDetail("id", request.ID)
	}

	team.Logo = request.Logo

	if err := t.TeamsRepo.Update(tx, team); err != nil {
		tx.Rollback()
		t.Log.Errorf("Failed to update team logo: %v", err)
		return nil, common.ErrInternalServer("Failed to update team logo")
	}

	if err := tx.Commit().Error; err != nil {
		t.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Logo for team with ID %s uploaded successfully", request.ID),
		Service: "teams",
		Time:    time.Now().Format(time.RFC3339),
	}
	t.Log.Infof("Sending log event: %+v", logEvent)
	if err := t.LogsProducer.Send(logEvent); err != nil {
		t.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToTeamResponse(team), nil
}
