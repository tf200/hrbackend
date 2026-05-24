package service

import (
	"context"
	"fmt"
	"strings"

	"hrbackend/internal/domain"
	"hrbackend/pkg/password"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EmployeeService struct {
	repo      domain.EmployeeRepository
	taskQueue domain.TaskQueue
	logger    domain.Logger
}

func NewEmployeeService(
	repo domain.EmployeeRepository,
	taskQueue domain.TaskQueue,
	logger domain.Logger,
) domain.EmployeeService {
	return &EmployeeService{repo: repo, taskQueue: taskQueue, logger: logger}
}

func (s *EmployeeService) GetEmployeeByID(
	ctx context.Context,
	id uuid.UUID,
	currentUserID uuid.UUID,
) (*domain.EmployeeDetail, error) {
	emp, err := s.repo.GetEmployeeByID(ctx, id)
	if err != nil {
		s.logError(ctx, "GetEmployeeByID", err, zap.String("employee_id", id.String()))
		return nil, err
	}

	attachments, err := s.repo.ListEmployeeAttachments(ctx, id)
	if err != nil {
		s.logError(ctx, "GetEmployeeByID", err, zap.String("employee_id", id.String()))
		return nil, err
	}
	emp.Attachments = attachments

	qualifications, err := s.repo.ListQualifications(ctx, id)
	if err != nil {
		s.logError(ctx, "GetEmployeeByID", err, zap.String("employee_id", id.String()))
		return nil, err
	}
	emp.Qualifications = qualifications

	authorizations, err := s.repo.ListEmployeeAuthorizations(ctx, id)
	if err != nil {
		s.logError(ctx, "GetEmployeeByID", err, zap.String("employee_id", id.String()))
		return nil, err
	}
	emp.Authorizations = authorizations

	return emp, nil
}

func (s *EmployeeService) GetEmployeeProfile(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.EmployeeProfile, error) {
	profile, err := s.repo.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		s.logError(ctx, "GetEmployeeProfile", err, zap.String("user_id", userID.String()))
		return nil, err
	}
	profile.PortalAccess = computePortalAccess(profile.Permissions)
	return profile, nil
}

// computePortalAccess derives portal routing from effective permissions.
func computePortalAccess(permissions []domain.Permission) string {
	hasAdmin := false
	hasEmployee := false
	for _, p := range permissions {
		switch p.Name {
		case domain.PortalPermissionAdmin:
			hasAdmin = true
		case domain.PortalPermissionEmployee:
			hasEmployee = true
		}
	}

	switch {
	case hasAdmin && hasEmployee:
		return domain.PortalAccessBoth
	case hasAdmin:
		return domain.PortalAccessAdmin
	case hasEmployee:
		return domain.PortalAccessEmployee
	default:
		// Safe fallback — employee portal is the least-privilege default.
		return domain.PortalAccessEmployee
	}
}

func (s *EmployeeService) ListEmployees(
	ctx context.Context,
	params domain.ListEmployeesParams,
) (*domain.EmployeePage, error) {
	page, err := s.repo.ListEmployees(ctx, params)
	if err != nil {
		s.logError(ctx, "ListEmployees", err)
		return nil, err
	}
	return page, nil
}

func (s *EmployeeService) CreateEmployee(
	ctx context.Context,
	params domain.CreateEmployeeParams,
) (*domain.EmployeeDetail, error) {
	if params.Contract != nil {
		if _, err := parseContractHoursType(params.Contract.ContractHoursType, params.Contract.HoursPerWeek, params.Contract.MinHoursPerWeek, params.Contract.MaxHoursPerWeek); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrContractChangeInvalid, err)
		}
		if params.Contract.ContractEndDate != nil && params.Contract.ContractEndDate.Before(params.Contract.StartDate) {
			return nil, fmt.Errorf("%w: contract_end_date cannot be before start_date", domain.ErrContractChangeInvalid)
		}
	}
	if params.SalaryAssignment != nil {
		if params.SalaryAssignment.EffectiveFrom == nil {
			if params.Contract == nil {
				return nil, fmt.Errorf("%w: salary effective_from is required without contract", domain.ErrContractChangeInvalid)
			}
			params.SalaryAssignment.EffectiveFrom = &params.Contract.StartDate
		}
		if params.SalaryAssignment.EffectiveTo != nil && !params.SalaryAssignment.EffectiveTo.After(*params.SalaryAssignment.EffectiveFrom) {
			return nil, fmt.Errorf("%w: salary effective_to must be after effective_from", domain.ErrContractChangeInvalid)
		}
	}
	hashedPassword, err := password.HashPassword(params.UserPassword)
	if err != nil {
		s.logError(ctx, "CreateEmployee", err)
		return nil, domain.ErrPasswordHashFailed
	}
	params.UserPassword = hashedPassword

	var emp *domain.EmployeeDetail
	err = s.repo.WithTx(ctx, func(tx domain.EmployeeTxRepository) error {
		userID, err := tx.CreateUser(ctx, params.UserEmail, params.UserPassword)
		if err != nil {
			return err
		}

		empID, err := tx.CreateEmployeeProfile(ctx, userID, params)
		if err != nil {
			return err
		}

		err = tx.AssignRoleToUser(ctx, userID, params.RoleID)
		if err != nil {
			return err
		}

		var contractID *uuid.UUID
		if params.Contract != nil {
			cid, err := tx.AddEmployeeContractDetails(ctx, empID, *params.Contract)
			if err != nil {
				return err
			}
			contractID = &cid
		}

		if params.SalaryAssignment != nil {
			_, err = tx.CreateEmployeeSalaryAssignment(ctx, empID, contractID, *params.SalaryAssignment)
			if err != nil {
				return err
			}
		}

		if len(params.AttachmentIDs) > 0 {
			err = tx.LinkEmployeeAttachments(ctx, empID, params.AttachmentIDs, "document")
			if err != nil {
				return err
			}
			err = tx.UpdateAttachmentsUsed(ctx, params.AttachmentIDs, true)
			if err != nil {
				return err
			}
		}

		if len(params.Qualifications) > 0 {
			err = tx.AddEmployeeQualificationsBatch(ctx, empID, params.Qualifications)
			if err != nil {
				return err
			}
		}

		if len(params.Authorizations) > 0 {
			err = tx.AddEmployeeAuthorizationsBatch(ctx, empID, params.Authorizations)
			if err != nil {
				return err
			}
		}

		emp, err = tx.GetEmployeeByID(ctx, empID)
		return err
	})
	if err != nil {
		s.logError(ctx, "CreateEmployee", err,
			zap.String("first_name", params.FirstName),
			zap.String("last_name", params.LastName),
		)
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) UpdateEmployee(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateEmployeeParams,
) (*domain.EmployeeDetail, error) {
	emp, err := s.repo.UpdateEmployee(ctx, id, params)
	if err != nil {
		s.logError(ctx, "UpdateEmployee", err, zap.String("employee_id", id.String()))
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) GetEmployeeCounts(ctx context.Context) (*domain.EmployeeCounts, error) {
	counts, err := s.repo.GetEmployeeCounts(ctx)
	if err != nil {
		s.logError(ctx, "GetEmployeeCounts", err)
		return nil, err
	}
	return counts, nil
}

func (s *EmployeeService) SearchEmployeesByNameOrEmail(
	ctx context.Context,
	search *string,
) ([]domain.EmployeeSearchResult, error) {
	results, err := s.repo.SearchEmployeesByNameOrEmail(ctx, search)
	if err != nil {
		s.logError(ctx, "SearchEmployeesByNameOrEmail", err)
		return nil, err
	}
	return results, nil
}

func (s *EmployeeService) ListEducation(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.Education, error) {
	items, err := s.repo.ListEducation(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "ListEducation", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return items, nil
}

func (s *EmployeeService) AddEducation(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateEducationParams,
) (*domain.Education, error) {
	edu, err := s.repo.AddEducation(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddEducation", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return edu, nil
}

func (s *EmployeeService) UpdateEducation(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateEducationParams,
) (*domain.Education, error) {
	edu, err := s.repo.UpdateEducation(ctx, id, params)
	if err != nil {
		s.logError(ctx, "UpdateEducation", err, zap.String("education_id", id.String()))
		return nil, err
	}
	return edu, nil
}

func (s *EmployeeService) DeleteEducation(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Education, error) {
	edu, err := s.repo.DeleteEducation(ctx, id)
	if err != nil {
		s.logError(ctx, "DeleteEducation", err, zap.String("education_id", id.String()))
		return nil, err
	}
	return edu, nil
}

func (s *EmployeeService) ListExperience(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.Experience, error) {
	items, err := s.repo.ListExperience(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "ListExperience", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return items, nil
}

func (s *EmployeeService) AddExperience(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateExperienceParams,
) (*domain.Experience, error) {
	exp, err := s.repo.AddExperience(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddExperience", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return exp, nil
}

func (s *EmployeeService) UpdateExperience(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateExperienceParams,
) (*domain.Experience, error) {
	exp, err := s.repo.UpdateExperience(ctx, id, params)
	if err != nil {
		s.logError(ctx, "UpdateExperience", err, zap.String("experience_id", id.String()))
		return nil, err
	}
	return exp, nil
}

func (s *EmployeeService) DeleteExperience(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Experience, error) {
	exp, err := s.repo.DeleteExperience(ctx, id)
	if err != nil {
		s.logError(ctx, "DeleteExperience", err, zap.String("experience_id", id.String()))
		return nil, err
	}
	return exp, nil
}

func (s *EmployeeService) ListQualifications(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.Qualification, error) {
	items, err := s.repo.ListQualifications(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "ListQualifications", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return items, nil
}

func (s *EmployeeService) AddQualification(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateQualificationParams,
) (*domain.Qualification, error) {
	qual, err := s.repo.AddQualification(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddQualification", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return qual, nil
}

func (s *EmployeeService) UpdateQualification(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateQualificationParams,
) (*domain.Qualification, error) {
	qual, err := s.repo.UpdateQualification(ctx, id, params)
	if err != nil {
		s.logError(ctx, "UpdateQualification", err, zap.String("qualification_id", id.String()))
		return nil, err
	}
	return qual, nil
}

func (s *EmployeeService) DeleteQualification(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Qualification, error) {
	qual, err := s.repo.DeleteQualification(ctx, id)
	if err != nil {
		s.logError(ctx, "DeleteQualification", err, zap.String("qualification_id", id.String()))
		return nil, err
	}
	return qual, nil
}

func (s *EmployeeService) ListEmployeeAuthorizations(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.EmployeeAuthorization, error) {
	items, err := s.repo.ListEmployeeAuthorizations(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "ListEmployeeAuthorizations", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return items, nil
}

func (s *EmployeeService) AddEmployeeAuthorization(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateEmployeeAuthorizationParams,
) (*domain.EmployeeAuthorization, error) {
	authRecord, err := s.repo.AddEmployeeAuthorization(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddEmployeeAuthorization", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return authRecord, nil
}

func (s *EmployeeService) UpdateEmployeeAuthorization(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateEmployeeAuthorizationParams,
) (*domain.EmployeeAuthorization, error) {
	authRecord, err := s.repo.UpdateEmployeeAuthorization(ctx, id, params)
	if err != nil {
		s.logError(ctx, "UpdateEmployeeAuthorization", err, zap.String("authorization_id", id.String()))
		return nil, err
	}
	return authRecord, nil
}

func (s *EmployeeService) DeleteEmployeeAuthorization(
	ctx context.Context,
	id uuid.UUID,
) (*domain.EmployeeAuthorization, error) {
	authRecord, err := s.repo.DeleteEmployeeAuthorization(ctx, id)
	if err != nil {
		s.logError(ctx, "DeleteEmployeeAuthorization", err, zap.String("authorization_id", id.String()))
		return nil, err
	}
	return authRecord, nil
}

func (s *EmployeeService) ResetPassword(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.ResetPasswordParams,
) (*domain.ResetPasswordResult, error) {
	emp, err := s.repo.GetEmployeeByID(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "ResetPassword", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}

	if params.SendEmail {
		if s.taskQueue == nil {
			s.logError(ctx, "ResetPassword", domain.ErrEmailDeliveryFailed,
				zap.String("employee_id", employeeID.String()))
			return nil, domain.ErrEmailDeliveryFailed
		}
		if emp.WorkEmailAddress == nil || strings.TrimSpace(*emp.WorkEmailAddress) == "" {
			s.logError(ctx, "ResetPassword", domain.ErrEmailDeliveryFailed,
				zap.String("employee_id", employeeID.String()))
			return nil, domain.ErrEmailDeliveryFailed
		}
	}

	var plainPassword string

	if params.Generated {
		gen, err := password.GenerateRandomPassword(16)
		if err != nil {
			s.logError(ctx, "ResetPassword", err, zap.String("employee_id", employeeID.String()))
			return nil, domain.ErrPasswordHashFailed
		}
		plainPassword = gen
	} else {
		if params.Password == nil || strings.TrimSpace(*params.Password) == "" {
			return nil, domain.ErrInvalidPasswordResetRequest
		}
		plainPassword = *params.Password
	}

	hashedPassword, err := password.HashPassword(plainPassword)
	if err != nil {
		s.logError(ctx, "ResetPassword", err, zap.String("employee_id", employeeID.String()))
		return nil, domain.ErrPasswordHashFailed
	}

	if err := s.repo.UpdatePassword(ctx, emp.UserID, hashedPassword); err != nil {
		s.logError(ctx, "ResetPassword", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}

	if params.SendEmail {
		payload := domain.EmailDeliveryTaskPayload{
			To:           *emp.WorkEmailAddress,
			Name:         emp.FirstName + " " + emp.LastName,
			UserEmail:    *emp.WorkEmailAddress,
			UserPassword: plainPassword,
		}
		if err := s.taskQueue.EnqueueEmailDelivery(ctx, payload, nil); err != nil {
			s.logError(ctx, "ResetPassword", err, zap.String("employee_id", employeeID.String()))
			return nil, domain.ErrEmailDeliveryFailed
		}
	}

	return &domain.ResetPasswordResult{TemporaryPassword: plainPassword}, nil
}

func (s *EmployeeService) logError(
	ctx context.Context,
	method string,
	err error,
	fields ...zap.Field,
) {
	if s.logger != nil {
		s.logger.LogError(ctx, "EmployeeService."+method, err.Error(), err, fields...)
	}
}

func isValidIrregularHoursProfile(value string) bool {
	switch value {
	case domain.IrregularHoursProfileNone,
		domain.IrregularHoursProfileRoster,
		domain.IrregularHoursProfileNonRoster:
		return true
	default:
		return false
	}
}

func parseContractHoursType(ct string, hoursPerWeek, minHours, maxHours *float64) (string, error) {
	switch ct {
	case "fixed":
		if hoursPerWeek == nil || *hoursPerWeek <= 0 {
			return "", fmt.Errorf("fixed contract requires hours_per_week > 0")
		}
	case "min_max":
		if minHours == nil || *minHours <= 0 {
			return "", fmt.Errorf("min_max contract requires min_hours_per_week > 0")
		}
		if maxHours == nil || *maxHours <= 0 {
			return "", fmt.Errorf("min_max contract requires max_hours_per_week > 0")
		}
		if *minHours > *maxHours {
			return "", fmt.Errorf("min_hours_per_week cannot exceed max_hours_per_week")
		}
	case "zero_hours":
	default:
		return "", fmt.Errorf("invalid contract_hours_type: %s", ct)
	}
	return ct, nil
}
