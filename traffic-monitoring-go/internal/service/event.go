package service

import (
	"context"
	"fmt"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/repository"
)


// the following type implements the SecurityEventService interface
type SecurityEventServiceImpl struct {
	eventRepo	repository.SecurityEventRepository
	alertRepo	repository.AlertRepository
	rulerepo	repository.RuleRepository
}

func NewSecurityEventService(
	eventRepo	repository.SecurityEventRepository,
	alertRepo	repository.AlertRepository,
	ruleRepo	repository.RuleRepository,
) *SecurityEventServiceImpl {
	return &SecurityEventServiceImpl{
		eventRepo: 	eventRepo,
		alertRepo: 	alertRepo,
		ruleRepo: 	ruleRepo,
	}
}


func (s *SecurityEventServiceImpl) ListSecurityEvents(ctx context.Context, query dto.SecurityEventQuery) ([]domain.SecurityEvent, *dto.MetaInfo, error) {
	events, total, err := s.eventRepo.FindSecurityEvents(ctx, query)
	if err != nil {
		return nil, nil, WrapError(err)
	}

	meta := dto.CalculatePagination(query.Page, query.PageSize, total)
	return events, meta, nil
}


func (s *SecurityEventServiceImpl) GetSecurityEvent(ctx context.Context, id uint) (*domain.SecurityEvent, error) {
	event, err := s.eventRepo.GetSecurityEventByID(ctx, id)
	if err != nil {
		return nil, WrapError(err)
	}
	return event, nil
}

func (s *SecurityEventServiceImpl) CreateSecurityEvent(ctx context.Context, input *dto.CreateSecurityEvent(ctx context.Context, input *dto.CreateSecurityEventRequest) (*domain.SecurityEvent, error) {
	// Convert DTO to a domain model
    event := input.ToDomain()

    // Validate the event
    if !event.IsValid() {
        return nil, ErrBadRequest
    }

    // Save the event
    err := s.eventRepo.CreateSecurityEvent(ctx, event)
    if err != nil {
        return nil, WrapError(err)
    }

    // Process rules and create alerts if needed
    if err := s.processRules(ctx, event); err != nil {
        // Log error but don't fail the event creation
        fmt.Printf("Error processing rules for event %d: %v\n", event.ID, err)
    }

    return event, nil
}

// BatchCreateSecurityEvents creates multiple security events
func (s *SecurityEventServiceImpl) BatchCreateSecurityEvents(ctx context.Context, inputs []*dto.CreateSecurityEventRequest) ([]*domain.SecurityEvent, error) {
    if len(inputs) == 0 {
        return []*domain.SecurityEvent{}, nil
    }

    // Convert DTOs to domain models
    events := make([]*domain.SecurityEvent, len(inputs))
    for i, input := range inputs {
        events[i] = input.ToDomain()

        // Validate each event
        if !events[i].IsValid() {
            return nil, fmt.Errorf("%w: invalid event at index %d", ErrBadRequest, i)
        }
    }

    // Save the events
    err := s.eventRepo.BatchCreateSecurityEvents(ctx, events)
    if err != nil {
        return nil, WrapError(err)
    }

    // Process rules and create alerts for each event
    for _, event := range events {
        if err := s.processRules(ctx, event); err != nil {
            // Log error but don't fail the batch creation
            fmt.Printf("Error processing rules for event %d: %v\n", event.ID, err)
        }
    }

    return events, nil
}

// DeleteSecurityEvent removes a security event by ID
func (s *SecurityEventServiceImpl) DeleteSecurityEvent(ctx context.Context, id uint) error {
    // Check if event exists
    _, err := s.eventRepo.GetSecurityEventByID(ctx, id)
    if err != nil {
        return WrapError(err)
    }

    // Delete the event
    err = s.eventRepo.DeleteSecurityEvent(ctx, id)
    if err != nil {
        return WrapError(err)
    }

    return nil
}

// processRules processes rule evaluation for a security event
func (s *SecurityEventServiceImpl) processRules(ctx context.Context, event *domain.SecurityEvent) error {
    // For now, simple default alert creation for critical events
    if event.ShouldTriggerAlert() {
        alert := &domain.Alert{
            RuleID:          1, // Default rule ID - in a real implementation, find matching rules
            SecurityEventID: event.ID,
            Timestamp:       event.Timestamp,
            Severity:        event.Severity,
            Status:          domain.AlertStatusOpen,
        }

        if err := s.alertRepo.CreateAlert(ctx, alert); err != nil {
            return fmt.Errorf("create alert for event: %w", err)
        }
    }

    // TODO: Implement rule engine with proper rule evaluation
    return nil
}
