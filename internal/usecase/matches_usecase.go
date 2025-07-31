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

type MatchesUseCase interface {
	FindAll(ctx context.Context) ([]model.MatchResponse, error)
	FindByID(ctx context.Context, request *model.MatchRequestFindByID) (*model.MatchResponse, error)
	Create(ctx context.Context, request *model.MatchRequestCreate) (*model.MatchResponse, error)
	Update(ctx context.Context, request *model.MatchRequestUpdate) (*model.MatchResponse, error)
	SoftDelete(ctx context.Context, request *model.MatchRequestSoftDelete) (*model.MatchResponse, error)
	GetMatchReport(ctx context.Context, request *model.MatchRequestFindByID) (*model.MatchReportResponse, error)
	FinishMatch(ctx context.Context, request *model.MatchRequestFinish) (*model.MatchResponse, error)
}

type matchesUseCaseImpl struct {
	MatchesRepo  repository.MatchesRepository
	LogsProducer *messaging.LogProducer
	DB           *gorm.DB
	Log          *logrus.Logger
}

func NewMatchesUseCase(matchesRepo repository.MatchesRepository, logsProducer *messaging.LogProducer, db *gorm.DB, log *logrus.Logger) MatchesUseCase {
	return &matchesUseCaseImpl{
		MatchesRepo:  matchesRepo,
		LogsProducer: logsProducer,
		DB:           db,
		Log:          log,
	}
}

func (m *matchesUseCaseImpl) FindAll(ctx context.Context) ([]model.MatchResponse, error) {
	tx := m.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	matches, err := m.MatchesRepo.FindAllWithRelations(tx, "HomeTeam", "AwayTeam")
	if err != nil {
		m.Log.Errorf("Failed to find all matches: %v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to find matches")
	}

	if err := tx.Commit().Error; err != nil {
		m.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	var responses []model.MatchResponse
	for _, match := range matches {
		responses = append(responses, *converter.ToMatchResponse(&match))
	}

	return responses, nil
}

func (m *matchesUseCaseImpl) FindByID(ctx context.Context, request *model.MatchRequestFindByID) (*model.MatchResponse, error) {
	tx := m.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		m.Log.Warnf("Invalid request body: %+v", err)
		return nil, err
	}

	match, err := m.MatchesRepo.FindByIDWithRelations(tx, request.ID, "HomeTeam", "AwayTeam")
	if err != nil {
		m.Log.Errorf("Failed to find match by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Match not found").WithDetail("id", request.ID)
	}

	if err := tx.Commit().Error; err != nil {
		m.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	return converter.ToMatchResponse(match), nil
}

func (m *matchesUseCaseImpl) Create(ctx context.Context, request *model.MatchRequestCreate) (*model.MatchResponse, error) {
	tx := m.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		m.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	match := &entity.Match{
		ID:         uuid.New().String(),
		MatchDate:  common.ConvertStringToDate(request.MatchDate),
		MatchTime:  request.MatchTime,
		HomeTeamID: request.HomeTeamID,
		AwayTeamID: request.AwayTeamID,
		HomeScore:  request.HomeScore,
		AwayScore:  request.AwayScore,
		Status:     request.Status,
	}

	// check if home team and away team are the same
	if match.HomeTeamID == match.AwayTeamID {
		tx.Rollback()
		m.Log.Warnf("Home team and away team cannot be the same: %s", match.HomeTeamID)
		return nil, common.ErrInvalidInput("Home team and away team cannot be the same").WithDetail("home_team_id", match.HomeTeamID)
	}

	// check if match date is in the past
	if match.MatchDate.Before(time.Now()) {
		tx.Rollback()
		m.Log.Warnf("Match date cannot be in the past: %s", match.MatchDate)
		return nil, common.ErrInvalidInput("Match date cannot be in the past").WithDetail("match_date", fmt.Sprintf("%s %s", match.MatchDate.Format("2006-01-02"), match.MatchTime))
	}

	m.Log.Infof("Creating match: %+v", match)

	if err := m.MatchesRepo.Create(tx, match); err != nil {
		tx.Rollback()
		m.Log.Errorf("Failed to create match: %v", err)
		return nil, common.ErrInternalServer("Failed to create match")
	}

	if err := tx.Commit().Error; err != nil {
		m.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	m.Log.Infof("Match created successfully: %+v", match)
	event := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Match with ID %s created successfully", match.ID),
		Service: "matches",
		Time:    time.Now().Format(time.RFC3339),
	}

	if err := m.LogsProducer.Send(event); err != nil {
		m.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToMatchResponse(match), nil
}

func (m *matchesUseCaseImpl) Update(ctx context.Context, request *model.MatchRequestUpdate) (*model.MatchResponse, error) {
	tx := m.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		m.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	match, err := m.MatchesRepo.FindByID(tx, request.ID)
	if err != nil {
		m.Log.Errorf("Failed to find match by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Match not found").WithDetail("id", request.ID)
	}

	// check if home team and away team are the same
	if match.HomeTeamID == match.AwayTeamID {
		tx.Rollback()
		m.Log.Warnf("Home team and away team cannot be the same: %s", match.HomeTeamID)
		return nil, common.ErrInvalidInput("Home team and away team cannot be the same").WithDetail("home_team_id", match.HomeTeamID)
	}

	if request.HomeTeamID != "" {
		match.HomeTeamID = request.HomeTeamID
	}
	if request.AwayTeamID != "" {
		match.AwayTeamID = request.AwayTeamID
	}
	if request.MatchDate != "" {
		match.MatchDate = common.ConvertStringToDate(request.MatchDate)
	}
	if request.MatchTime != "" {
		match.MatchTime = request.MatchTime
	}
	if request.HomeScore != nil {
		match.HomeScore = request.HomeScore
	}
	if request.AwayScore != nil {
		match.AwayScore = request.AwayScore
	}
	if request.Status != "" {
		match.Status = request.Status
	}

	if err := m.MatchesRepo.Update(tx, match); err != nil {
		tx.Rollback()
		m.Log.Errorf("Failed to update match: %v", err)
		return nil, common.ErrInternalServer("Failed to update match")
	}

	if err := tx.Commit().Error; err != nil {
		m.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	m.Log.Infof("Match updated successfully: %+v", match)
	event := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Match with ID %s updated successfully", match.ID),
		Service: "matches",
		Time:    time.Now().Format(time.RFC3339),
	}
	if err := m.LogsProducer.Send(event); err != nil {
		m.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToMatchResponse(match), nil
}

func (m *matchesUseCaseImpl) SoftDelete(ctx context.Context, request *model.MatchRequestSoftDelete) (*model.MatchResponse, error) {
	tx := m.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		m.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	match, err := m.MatchesRepo.FindByID(tx, request.ID)
	if err != nil {
		m.Log.Errorf("Failed to find match by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Match not found").WithDetail("id", request.ID)
	}

	if err := m.MatchesRepo.SoftDelete(tx, match.ID); err != nil {
		tx.Rollback()
		m.Log.Errorf("Failed to soft delete match: %v", err)
		return nil, common.ErrInternalServer("Failed to soft delete match")
	}

	if err := tx.Commit().Error; err != nil {
		m.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}
	m.Log.Infof("Match with ID %s soft deleted successfully", match.ID)
	event := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("Match with ID %s soft deleted successfully", match.ID),
		Service: "matches",
		Time:    time.Now().Format(time.RFC3339),
	}
	if err := m.LogsProducer.Send(event); err != nil {
		m.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to send log event")
	}

	return converter.ToMatchResponse(match), nil
}

func (m *matchesUseCaseImpl) GetMatchReport(ctx context.Context, request *model.MatchRequestFindByID) (*model.MatchReportResponse, error) {
	tx := m.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		m.Log.Warnf("Invalid request body: %+v", err)
		return nil, err
	}

	match, err := m.MatchesRepo.FindByIDWithRelations(tx, request.ID, "HomeTeam", "AwayTeam")
	if err != nil {
		tx.Rollback()
		return nil, common.ErrNotFound("Match not found").WithDetail("id", request.ID)
	}

	goals, err := m.MatchesRepo.FindGoalsByMatchIDWithPlayer(tx, match.ID)
	if err != nil {
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to get goals for match")
	}

	pastMatches, err := m.MatchesRepo.FindAllBeforeDate(tx, match.MatchDate)
	if err != nil {
		m.Log.Errorf("Failed to get past matches: %v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to get past matches")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	return converter.ToMatchReportResponse(match, goals, pastMatches), nil
}

func (m *matchesUseCaseImpl) FinishMatch(ctx context.Context, request *model.MatchRequestFinish) (*model.MatchResponse, error) {
	tx := m.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := common.ValidateStruct(request); err != nil {
		m.Log.Warnf("Invalid request body: %+v", err)
		return nil, err
	}
	match, err := m.MatchesRepo.FindByID(tx, request.ID)
	if err != nil {
		m.Log.Errorf("Failed to find match by ID %s: %v", request.ID, err)
		tx.Rollback()
		return nil, common.ErrNotFound("Match not found").WithDetail("id", request.ID)
	}

	if match.Status != "Scheduled" {
		tx.Rollback()
		m.Log.Warnf("Match with ID %s is not scheduled", request.ID)
		return nil, common.ErrInvalidInput("Match is not scheduled").WithDetail("id", request.ID)
	}
	match.Status = "Finished"

	if err := m.MatchesRepo.Update(tx, match); err != nil {
		tx.Rollback()
		m.Log.Errorf("Failed to update match: %v", err)
		return nil, common.ErrInternalServer("Failed to update match")
	}
	if err := tx.Commit().Error; err != nil {
		m.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}
	return converter.ToMatchResponse(match), nil
}
