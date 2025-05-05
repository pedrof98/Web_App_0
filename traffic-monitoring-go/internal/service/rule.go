package service

Import (
	"context"
	"errors"
	"fmt"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/repository"
)

// RuleServiceImpl implements the RuleService interface
type RuleServiceImpl struct {
	ruleRepo repository.RuleRepository
}

// NewRuleService creates a new RuleServiceImpl
func NewRuleService(ruleRepo repository.RuleRepository) *RuleServiceImpl {
	return &RuleServiceImpl{
		ruleRepo: ruleRepo,
	}
}

// ListRules retrieves rules based on query parameters
func (s *RuleServiceImpl) ListRules(ctx context.Context, query dto.RuleQuery) ([]domain.Rule, *dto.MetaInfo, error) {
	rules, total, err := s.ruleRepo.FindRules(ctx, query)
	if err != nil {
		return nil, nil, WrapError(err)
	}

	meta := dto.CalculatedPagination(query.Page, query.PageSize, total)
	return rules, meta, nil
}

// GetRule retrieves a single rule by ID
func (s *RuleServiceImpl) GetRule(ctx context.Context, id uint) (*domain.Rule, error) {
	rule, err := s.rulerepo.GetRuleByID(ctx, id)
	if err != nil {
		return nil, WrapError(err)
	}
	return rule, nil
}


// CreateRule creates a new rule
func (s *RuleServiceImpl) CreateRule(ctx context.Context, input *dto.CreateRuleRequest, userID uint) (*domain.Rule, error) {
	// convert DTO to a domain model
	rule := input.ToDomain(userID)

	// validate the rule
	if !rule.IsValid() {
		return nil, ErrBadRequest
	}

	// Save the rule
	err := s.ruleRepo.CreateRule(ctx, rule)
	if err != nil {
		return nil, WrapError(err)
	}

	return rule, nil
}


// UpdateRule updates an existing rule
func (s *RuleServiceImpl) UpdateRule(ctx context.Context, id uint, input *dto.UpdateRuleRequest) (*domain.Rule, error) {
	// get the existing rule
	rule, err := s.ruleRepo.GetRuleByID(ctx, id)
	if err != nil {
		return nil, WrapError(err)
	}

	// Apply updates
	input.ApplyToRule(rule)
	
	// validate the updated rule
	if !rule.IsValid() {
		return nil, ErrBadRequest
	}

	// Save the updated rule
	err = s.ruleRepo.UpdateRule(ctx, rule)
	if err != nil {
		return nil, WrapError(err)
	}

	return rule, nil
}


// DeleteRule removes a rule by ID
func (s *RuleServiceImpl) DeleteRule(ctx context.Context, id uint) error {
	// check if there are any alerts associated with this rule
	alertCount, err := s.ruleRepo.CountAlertsByRuleID(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	// get the rule to check if it can be deleted
	rule, err := s.ruleRepo.GetRuleByID(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	// check if the rule can be deleted
	if !rule.CanBeDeleted(alertCount) {
		return fmt.Errorf("%w: rule has %d associated alerts",
		ErrForbidden, alertCount)
	}

	// delete the rule
	err = s.ruleRepo.DeleteRule(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	return nil
}


