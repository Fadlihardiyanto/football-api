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

type GoalsUseCase interface {
	FindAll(ctx context.Context) ([]model.GoalResponse, error)
	FindByID(ctx context.Context, request *model.GoalRequestFindByID) (*model.GoalResponse, error)
	Create(ctx context.Context, request *model.GoalRequestCreate) (*model.GoalResponse, error)
	Update(ctx context.Context, request *model.GoalRequestUpdate) (*model.GoalResponse, error)
	SoftDelete(ctx context.Context, request *model.GoalRequestSoftDelete) (*model.GoalResponse, error)
}

type goalsUseCaseImpl struct {
	GoalsRepo    repository.GoalsRepository
	MatchesRepo  repository.MatchesRepository
	PlayersRepo  repository.PlayersRepository
	LogsProducer *messaging.LogProducer
	DB           *gorm.DB
	Log          *logrus.Logger
}

func NewGoalsUseCase(goalsRepo repository.GoalsRepository, matchesRepo repository.MatchesRepository, playersRepo repository.PlayersRepository, logsProducer *messaging.LogProducer, db *gorm.DB, log *logrus.Logger) GoalsUseCase {
	return &goalsUseCaseImpl{
		GoalsRepo:    goalsRepo,
		MatchesRepo:  matchesRepo,
		PlayersRepo:  playersRepo,
		LogsProducer: logsProducer,
		DB:           db,
		Log:          log,
	}
}

func (g *goalsUseCaseImpl) FindAll(ctx context.Context) ([]model.GoalResponse, error) {
	tx := g.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	goals, err := g.GoalsRepo.FindAllWithRelations(tx, "Match", "Player", "Match.HomeTeam", "Match.AwayTeam")
	if err != nil {
		g.Log.Errorf("Failed to find all goals: %v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to find goals")
	}

	if err := tx.Commit().Error; err != nil {
		g.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	var responses []model.GoalResponse
	for _, goal := range goals {
		responses = append(responses, *converter.ToGoalResponse(&goal))
	}

	return responses, nil
}

func (g *goalsUseCaseImpl) FindByID(ctx context.Context, request *model.GoalRequestFindByID) (*model.GoalResponse, error) {
	tx := g.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		g.Log.Warnf("Invalid request body: %+v", err)
		return nil, err
	}

	goal, err := g.GoalsRepo.FindByIDWithRelations(tx, request.ID, "Match", "Player", "Match.HomeTeam", "Match.AwayTeam")
	if err != nil {
		g.Log.Errorf("Failed to find goal by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Goal not found").WithDetail("id", request.ID)
	}

	if err := tx.Commit().Error; err != nil {
		g.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	return converter.ToGoalResponse(goal), nil
}

func (g *goalsUseCaseImpl) Create(ctx context.Context, request *model.GoalRequestCreate) (*model.GoalResponse, error) {
	tx := g.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		g.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	goal := &entity.Goal{
		ID:       uuid.New().String(),
		MatchID:  request.MatchID,
		PlayerID: request.PlayerID,
		GoalTime: request.GoalTime,
	}

	// Validation for GoalTime
	if request.GoalTime < 0 || request.GoalTime > 120 {
		tx.Rollback()
		return nil, common.ErrInvalidInput("Goal time must be between 0 and 120 minutes").WithDetail("goal_time", fmt.Sprintf("%d", request.GoalTime))
	}

	// Check if Match exists
	match, err := g.MatchesRepo.FindByID(tx, request.MatchID)
	if err != nil {
		tx.Rollback()
		return nil, common.ErrNotFound("Match not found").WithDetail("id", request.MatchID)
	}

	// Check if Player exists
	player, err := g.PlayersRepo.FindByID(tx, request.PlayerID)
	if err != nil {
		tx.Rollback()
		return nil, common.ErrNotFound("Player not found").WithDetail("id", request.PlayerID)
	}

	// Check if Player belongs to either HomeTeam or AwayTeam
	if player.TeamID != match.HomeTeamID && player.TeamID != match.AwayTeamID {
		tx.Rollback()
		return nil, common.ErrInvalidInput("Player does not belong to either team in the match").WithDetail("player_id", request.PlayerID)
	}

	// Update score in the match
	if player.TeamID == match.HomeTeamID {
		if match.HomeScore == nil {
			match.HomeScore = new(int)
		}
		*match.HomeScore++
	} else if player.TeamID == match.AwayTeamID {
		if match.AwayScore == nil {
			match.AwayScore = new(int)
		}
		*match.AwayScore++
	}

	// Update match status if necessary

	if err := g.MatchesRepo.Update(tx, match); err != nil {
		tx.Rollback()
		g.Log.Errorf("Failed to update match score: %v", err)
		return nil, common.ErrInternalServer("Failed to update match score")
	}

	if err := g.GoalsRepo.Create(tx, goal); err != nil {
		tx.Rollback()
		g.Log.Errorf("Failed to create goal: %v", err)
		return nil, common.ErrInternalServer("Failed to create goal")
	}

	if err := tx.Commit().Error; err != nil {
		g.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Goal created successfully for match %s by player %s", request.MatchID, request.PlayerID),
		Service: "goals",
		Time:    goal.CreatedAt.Format(time.RFC3339),
	}
	g.Log.Infof("Sending log event: %+v", logEvent)
	if err := g.LogsProducer.Send(logEvent); err != nil {
		g.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToGoalResponse(goal), nil
}

func (g *goalsUseCaseImpl) Update(ctx context.Context, request *model.GoalRequestUpdate) (*model.GoalResponse, error) {
	tx := g.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		g.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	goal, err := g.GoalsRepo.FindByID(tx, request.ID)
	if err != nil {
		g.Log.Errorf("Failed to find goal by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Goal not found").WithDetail("id", request.ID)
	}

	if request.MatchID != "" {
		goal.MatchID = request.MatchID
	}
	if request.PlayerID != "" {
		goal.PlayerID = request.PlayerID
	}
	if request.GoalTime >= 0 {
		goal.GoalTime = request.GoalTime
	}

	if err := g.GoalsRepo.Update(tx, goal); err != nil {
		tx.Rollback()
		g.Log.Errorf("Failed to update goal: %v", err)
		return nil, common.ErrInternalServer("Failed to update goal")
	}

	if err := tx.Commit().Error; err != nil {
		g.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}
	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Goal with ID %s updated successfully", request.ID),
		Service: "goals",
		Time:    time.Now().Format(time.RFC3339),
	}
	g.Log.Infof("Sending log event: %+v", logEvent)
	if err := g.LogsProducer.Send(logEvent); err != nil {
		g.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToGoalResponse(goal), nil
}

func (g *goalsUseCaseImpl) SoftDelete(ctx context.Context, request *model.GoalRequestSoftDelete) (*model.GoalResponse, error) {
	tx := g.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		g.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	goal, err := g.GoalsRepo.FindByID(tx, request.ID)
	if err != nil {
		g.Log.Errorf("Failed to find goal by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Goal not found").WithDetail("id", request.ID)
	}

	if err := g.GoalsRepo.SoftDelete(tx, goal.ID); err != nil {
		tx.Rollback()
		g.Log.Errorf("Failed to soft delete goal: %v", err)
		return nil, common.ErrInternalServer("Failed to soft delete goal")
	}

	if err := tx.Commit().Error; err != nil {
		g.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}
	// Send log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Goal with ID %s soft deleted successfully", request.ID),
		Service: "goals",
		Time:    time.Now().Format(time.RFC3339),
	}
	g.Log.Infof("Sending log event: %+v", logEvent)
	if err := g.LogsProducer.Send(logEvent); err != nil {
		g.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToGoalResponse(goal), nil
}
