package service

import (
	"context"
	"errors"
	"fmt"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/repository"
)

// AlertServiceImpl implements the AlertService interface
type AlertServiceImpl struct {
	alertrepo repository.AlertRepository
	ruleRepo repository.RuleRepository
	// add more repos as needed
}

// NewAlertService creates a new AlertServiceImpl
func NewAlertService(alertRepo repository.AlertRepository, ruleRepo repository.RuleRepository) *AlertServiceImpl {
	return &AlertServiceImpl{
		alertRepo: alertRepo,
		ruleRepo: ruleRepo,
	}
}

func (s *AlertServiceImpl) ListAlerts(ctx context.Context, query dto.AlertQuery) ([]domain.Alert, *dto.MetaInfo, error) {
	alerts, total, err := s.alertRepo.FindAlerts(ctx, query)
	if err != nil {
		return nil, nil, WrapError(err)
	}

	meta := dto.CalculatePagination(query.Page, query.PageSize, total)
	return alerts, meta, nil
}

func (s *AlertServiceImpl) GetAlert(ctx context.Context, id uint) (*domain.Alert, error) {
	alert, err := s.alertRepo.GetAlertByID(ctx, id)
	if err != nil {
		return nil, WrapError(err)
	}
	return alert, nil
}

func (s *AlertServiceImpl) CreateAlert(ctx context.Context, input *dto.CreateAlertrequest) (*domain.Alert, error) {
	// convert DTO to a domain model
	alert := input.ToDomain()

	// validate the alert
	if !alert.IsValid() {
		return nil, ErrBadRequest
	}

	// verify that the rule exists
	_, err := s.rulerepo.GetRuleByID(ctx, alert.RuleID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: rule with ID %d not found", ErrBadRequest, alert.RuleID)
		}
		return nil, WrapError(err)
	}

	// TODO: Verify that the security event exists (will need a security event repository)

	// save the alert
	err = s.alertRepo.CreateAlert(ctx, alert)
	if err != nil {
		return nil, WrapError(err)
	}

	return alert, nil
}

func (s *AlertServiceImpl) UpdateAlert(ctx context.Context, id uint, input *dto.UpdateAlertRequest) (*domain.Alert, error) {
	// get the existing alert
	alert, err := s.alertRepo.GetAlertByID(ctx, id)
	if err != nil {
		return nil, WrapError(err)
	}

	//check if status transition is valid
	if input.Status != nil && !alert.CanTransitionToStatus(*input.Status) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s",
			ErrBadRequest, alert.Status, *inpit.Status)
		}

		// apply updates
		input.ApplytoAlert(alert)

		// validate updated alert
		if !alert.IsValid() {
			return nil, ErrBadRequest
		}

		// save the updated alert
		err = s.alertRepo.UpdateAlert(ctx, alert)
		if err != nil {
			return nil, WrapError(err)
		}

		return alert, nil
}

func (s *AlertServiceImpl) DeleteAlert(ctx context.Context, id uint) error {
	// check if alert exists
	_, err := s.alertRepo.GetAlertByID(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	// delete the alert
	err = s.alertRepo.DeleteAlert(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// AssignAlert assigns an alert to a user
func (s *AlertServiceImpl) AssignAlert(ctx context.Context, id uint, userID uint) (*domain.Alert, error) {
	// get the existing alert
	alert, err := s.alertRepo.GetAlertByID(ctx, id)
	if err != nil {
		return nil, WrapError(err)
	}

	// check if the alert can be assigned to the user
	if !alert.IsAssignableToUser(userID) {
		return nil, fmt.Errorf("%w: alert cannot be assigned to user %d",
			ErrBadRequest, userID)
	}

	//update the assigned user
	alert.AssignedTo = &userID

	// if alert is not in progress, update status
	if alert.Status == domain.AlertStatusOpen {
		inProgressStatus := domain.AlertStatusInProgress
		alert.Status = inProgressStatus
	}

	// save the updated alert
	err = s.alertRepo.UpdateAlert(ctx, alert)
	if err != nil {
		return nil, WrapError(err)
	}

	return alert, nil
}



