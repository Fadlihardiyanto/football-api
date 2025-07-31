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

type PlayersUseCase interface {
	FindAll(ctx context.Context) ([]model.PlayerResponse, error)
	FindByID(ctx context.Context, request *model.PlayerRequestFindByID) (*model.PlayerResponse, error)
	Create(ctx context.Context, request *model.PlayerRequestCreate) (*model.PlayerResponse, error)
	Update(ctx context.Context, request *model.PlayerRequestUpdate) (*model.PlayerResponse, error)
	SoftDelete(ctx context.Context, request *model.PlayerRequestSoftDelete) (*model.PlayerResponse, error)
}

type playersUseCaseImpl struct {
	PlayersRepo  repository.PlayersRepository
	TeamsRepo    repository.TeamsRepository
	LogsProducer *messaging.LogProducer
	DB           *gorm.DB
	Log          *logrus.Logger
}

func NewPlayersUseCase(playersRepo repository.PlayersRepository, teamsRepo repository.TeamsRepository, logsProducer *messaging.LogProducer, db *gorm.DB, log *logrus.Logger) PlayersUseCase {
	return &playersUseCaseImpl{
		PlayersRepo:  playersRepo,
		TeamsRepo:    teamsRepo,
		LogsProducer: logsProducer,
		DB:           db,
		Log:          log,
	}
}

func (p *playersUseCaseImpl) FindAll(ctx context.Context) ([]model.PlayerResponse, error) {
	tx := p.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	players, err := p.PlayersRepo.FindAllWithRelations(tx, "Team")
	if err != nil {
		p.Log.Errorf("Failed to find all players: %v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to find players")
	}

	if err := tx.Commit().Error; err != nil {
		p.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	var responses []model.PlayerResponse
	for _, player := range players {
		responses = append(responses, *converter.ToPlayerResponse(&player))
	}

	return responses, nil
}

func (p *playersUseCaseImpl) FindByID(ctx context.Context, request *model.PlayerRequestFindByID) (*model.PlayerResponse, error) {
	tx := p.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		p.Log.Warnf("Invalid request body: %+v", err)
		return nil, err
	}

	player, err := p.PlayersRepo.FindByIDWithRelations(tx, request.ID, "Team")
	if err != nil {
		p.Log.Errorf("Failed to find player by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Player not found").WithDetail("id", request.ID)
	}

	if err := tx.Commit().Error; err != nil {
		p.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	return converter.ToPlayerResponse(player), nil
}

func (p *playersUseCaseImpl) Create(ctx context.Context, request *model.PlayerRequestCreate) (*model.PlayerResponse, error) {
	tx := p.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		p.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	player := &entity.Player{
		ID:           uuid.New().String(),
		Name:         request.Name,
		Position:     request.Position,
		TeamID:       request.TeamID,
		Height:       request.Height,
		Weight:       request.Weight,
		JerseyNumber: request.JerseyNumber,
	}

	// check if team exists
	if request.TeamID != "" {
		exists, err := p.TeamsRepo.CheckTeamExistsByTeamID(tx, request.TeamID)
		if err != nil {
			tx.Rollback()
			p.Log.Errorf("Failed to check team existence: %v", err)
			return nil, common.ErrInternalServer("Failed to check team existence")
		}
		if !exists {
			tx.Rollback()
			return nil, common.ErrNotFound("Team not found").WithDetail("id", request.TeamID)
		}
	}

	// check if jersey number already exists for the team
	exist, err := p.PlayersRepo.CheckNumberJerseyByNumberAndTeamID(tx, player.JerseyNumber, player.TeamID)
	if err != nil {
		tx.Rollback()
		p.Log.Errorf("Failed to check jersey number: %v", err)
		return nil, common.ErrInternalServer("Failed to check jersey number")
	}
	if exist {
		tx.Rollback()
		return nil, common.ErrConflict("Jersey number already taken by another player in the same team")
	}

	if err := p.PlayersRepo.Create(tx, player); err != nil {
		tx.Rollback()
		p.Log.Errorf("Failed to create player: %v", err)
		return nil, common.ErrInternalServer("Failed to create player")
	}

	if err := tx.Commit().Error; err != nil {
		p.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Player %s created successfully", player.Name),
		Service: "players",
		Time:    player.CreatedAt.Format(time.RFC3339),
	}
	p.Log.Infof("Sending log event: %+v", logEvent)
	if err := p.LogsProducer.Send(logEvent); err != nil {
		p.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToPlayerResponse(player), nil
}

func (p *playersUseCaseImpl) Update(ctx context.Context, request *model.PlayerRequestUpdate) (*model.PlayerResponse, error) {
	tx := p.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		p.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	player, err := p.PlayersRepo.FindByID(tx, request.ID)
	if err != nil {
		p.Log.Errorf("Failed to find player by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Player not found").WithDetail("id", request.ID)
	}

	// check if player changes team
	if request.TeamID != "" {
		exists, err := p.TeamsRepo.CheckTeamExistsByTeamID(tx, request.TeamID)
		if err != nil {
			tx.Rollback()
			p.Log.Errorf("Failed to check team existence: %v", err)
			return nil, common.ErrInternalServer("Failed to check team existence")
		}
		if !exists {
			tx.Rollback()
			return nil, common.ErrNotFound("Team not found").WithDetail("id", request.TeamID)
		}
	}

	// check if player changes jersey number or team
	if (request.TeamID != "" && request.TeamID != player.TeamID) ||
		(request.JerseyNumber != 0 && request.JerseyNumber != player.JerseyNumber) {

		exist, err := p.PlayersRepo.CheckNumberJerseyByNumberAndTeamIDExceptPlayerID(tx, request.JerseyNumber, request.TeamID, request.ID)
		if err != nil {
			tx.Rollback()
			p.Log.Errorf("Failed to check jersey number: %v", err)
			return nil, common.ErrInternalServer("Failed to check jersey number")
		}
		if exist {
			tx.Rollback()
			return nil, common.ErrConflict("Jersey number already taken by another player in the same team")
		}
	}

	if request.Position != "" {
		player.Position = request.Position
	}
	if request.TeamID != "" {
		player.TeamID = request.TeamID
	}
	if request.Height != 0 {
		player.Height = request.Height
	}
	if request.Weight != 0 {
		player.Weight = request.Weight
	}
	if request.JerseyNumber != 0 {
		player.JerseyNumber = request.JerseyNumber
	}

	if err := p.PlayersRepo.Update(tx, player); err != nil {
		tx.Rollback()
		p.Log.Errorf("Failed to update player: %v", err)
		return nil, common.ErrInternalServer("Failed to update player")
	}

	if err := tx.Commit().Error; err != nil {
		p.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}
	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Player %s updated successfully", player.Name),
		Service: "players",
		Time:    time.Now().Format(time.RFC3339),
	}
	p.Log.Infof("Sending log event: %+v", logEvent)
	if err := p.LogsProducer.Send(logEvent); err != nil {
		p.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToPlayerResponse(player), nil
}

func (p *playersUseCaseImpl) SoftDelete(ctx context.Context, request *model.PlayerRequestSoftDelete) (*model.PlayerResponse, error) {
	tx := p.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		p.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	player, err := p.PlayersRepo.FindByID(tx, request.ID)
	if err != nil {
		p.Log.Errorf("Failed to find player by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Player not found").WithDetail("id", request.ID)
	}

	if err := p.PlayersRepo.SoftDelete(tx, player.ID); err != nil {
		tx.Rollback()
		p.Log.Errorf("Failed to soft delete player: %v", err)
		return nil, common.ErrInternalServer("Failed to soft delete player")
	}

	if err := tx.Commit().Error; err != nil {
		p.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}
	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Player %s soft deleted successfully", player.Name),
		Service: "players",
		Time:    time.Now().Format(time.RFC3339),
	}
	p.Log.Infof("Sending log event: %+v", logEvent)
	if err := p.LogsProducer.Send(logEvent); err != nil {
		p.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToPlayerResponse(player), nil
}
