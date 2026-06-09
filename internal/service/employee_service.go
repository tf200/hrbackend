package service

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		if err := validateContractHours(params.Contract.ContractType, params.Contract.HoursPerWeek); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrContractChangeInvalid, err)
		}
		if params.Contract.ContractEndDate != nil &&
			params.Contract.ContractEndDate.Before(params.Contract.StartDate) {
			return nil, fmt.Errorf(
				"%w: contract_end_date cannot be before start_date",
				domain.ErrContractChangeInvalid,
			)
		}
	}
	if params.SalaryAssignment != nil {
		if params.SalaryAssignment.EffectiveFrom == nil {
			if params.Contract == nil {
				return nil, fmt.Errorf(
					"%w: salary effective_from is required without contract",
					domain.ErrContractChangeInvalid,
				)
			}
			params.SalaryAssignment.EffectiveFrom = &params.Contract.StartDate
		}
		if params.SalaryAssignment.EffectiveTo != nil &&
			!params.SalaryAssignment.EffectiveTo.After(*params.SalaryAssignment.EffectiveFrom) {
			return nil, fmt.Errorf(
				"%w: salary effective_to must be after effective_from",
				domain.ErrContractChangeInvalid,
			)
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
			_, err = tx.CreateEmployeeSalaryAssignment(
				ctx,
				empID,
				contractID,
				*params.SalaryAssignment,
			)
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
	if params.SalaryAssignment != nil {
		if params.SalaryAssignment.EffectiveFrom == nil {
			return nil, fmt.Errorf(
				"%w: salary effective_from is required",
				domain.ErrContractChangeInvalid,
			)
		}
		if params.SalaryAssignment.EffectiveTo != nil &&
			!params.SalaryAssignment.EffectiveTo.After(*params.SalaryAssignment.EffectiveFrom) {
			return nil, fmt.Errorf(
				"%w: salary effective_to must be after effective_from",
				domain.ErrContractChangeInvalid,
			)
		}
	}

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

func (s *EmployeeService) AddQualifications(
	ctx context.Context,
	employeeID uuid.UUID,
	params []domain.CreateQualificationParams,
) (int, error) {
	if _, err := s.repo.GetEmployeeByID(ctx, employeeID); err != nil {
		return 0, err
	}

	count, err := s.repo.AddQualifications(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddQualifications", err, zap.String("employee_id", employeeID.String()))
		return 0, err
	}
	return count, nil
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
		s.logError(
			ctx,
			"ListEmployeeAuthorizations",
			err,
			zap.String("employee_id", employeeID.String()),
		)
		return nil, err
	}
	return items, nil
}

func (s *EmployeeService) AddEmployeeAuthorizations(
	ctx context.Context,
	employeeID uuid.UUID,
	params []domain.CreateEmployeeAuthorizationParams,
) (int, error) {
	if _, err := s.repo.GetEmployeeByID(ctx, employeeID); err != nil {
		return 0, err
	}

	count, err := s.repo.AddEmployeeAuthorizations(ctx, employeeID, params)
	if err != nil {
		s.logError(
			ctx,
			"AddEmployeeAuthorizations",
			err,
			zap.String("employee_id", employeeID.String()),
		)
		return 0, err
	}
	return count, nil
}

func (s *EmployeeService) UpdateEmployeeAuthorization(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateEmployeeAuthorizationParams,
) (*domain.EmployeeAuthorization, error) {
	authRecord, err := s.repo.UpdateEmployeeAuthorization(ctx, id, params)
	if err != nil {
		s.logError(
			ctx,
			"UpdateEmployeeAuthorization",
			err,
			zap.String("authorization_id", id.String()),
		)
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
		s.logError(
			ctx,
			"DeleteEmployeeAuthorization",
			err,
			zap.String("authorization_id", id.String()),
		)
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

func (s *EmployeeService) CreateContractAmendment(
	ctx context.Context,
	employeeID uuid.UUID,
	contractID uuid.UUID,
	params domain.CreateContractAmendmentParams,
) (*domain.EmployeeDetail, error) {
	baseContract, err := s.repo.GetEmployeeContractByID(ctx, contractID)
	if err != nil {
		s.logError(ctx, "CreateContractAmendment", err,
			zap.String("contract_id", contractID.String()),
			zap.String("employee_id", employeeID.String()),
		)
		return nil, fmt.Errorf("%w: base contract not found", domain.ErrContractChangeInvalid)
	}

	if baseContract.EmployeeID != employeeID {
		return nil, fmt.Errorf(
			"%w: contract does not belong to employee",
			domain.ErrContractChangeInvalid,
		)
	}

	if !params.StartDate.After(baseContract.StartDate) {
		return nil, fmt.Errorf(
			"%w: amendment start_date must be after base contract start_date",
			domain.ErrContractChangeInvalid,
		)
	}

	if params.ContractEndDate != nil && !params.ContractEndDate.After(params.StartDate) {
		return nil, fmt.Errorf(
			"%w: amendment contract_end_date must be after start_date",
			domain.ErrContractChangeInvalid,
		)
	}
	if err := validateContractHours(params.ContractType, params.HoursPerWeek); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrContractChangeInvalid, err)
	}

	var emp *domain.EmployeeDetail
	err = s.repo.WithTx(ctx, func(tx domain.EmployeeTxRepository) error {
		oldEndDate := params.StartDate.AddDate(0, 0, -1)
		if err := tx.EndEmployeeContractSegment(ctx, contractID, oldEndDate, nil); err != nil {
			return err
		}

		newContractID, err := tx.AddEmployeeContractAmendment(ctx, employeeID, contractID, params)
		if err != nil {
			return err
		}

		if err := s.createSalaryForContractSegment(
			ctx,
			tx,
			employeeID,
			&contractID,
			newContractID,
			params.StartDate,
			params.SalaryAssignment,
		); err != nil {
			return err
		}

		emp, err = tx.GetEmployeeByID(ctx, employeeID)
		return err
	})
	if err != nil {
		s.logError(ctx, "CreateContractAmendment", err,
			zap.String("contract_id", contractID.String()),
			zap.String("employee_id", employeeID.String()),
		)
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) CreateNewContract(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateNewContractParams,
) (*domain.EmployeeDetail, error) {
	if err := validateContractHours(params.ContractType, params.HoursPerWeek); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrContractChangeInvalid, err)
	}
	if params.ContractEndDate != nil && !params.ContractEndDate.After(params.StartDate) {
		return nil, fmt.Errorf(
			"%w: contract_end_date must be after start_date",
			domain.ErrContractChangeInvalid,
		)
	}

	var emp *domain.EmployeeDetail
	err := s.repo.WithTx(ctx, func(tx domain.EmployeeTxRepository) error {
		overlapping, err := tx.GetEmployeeContractAtDate(ctx, employeeID, params.StartDate)
		if err != nil {
			return err
		}

		var previousContractID *uuid.UUID
		if overlapping != nil {
			oldEndDate := params.StartDate.AddDate(0, 0, -1)
			if err := tx.EndEmployeeContractSegment(ctx, overlapping.ID, oldEndDate, nil); err != nil {
				return err
			}
			previousContractID = &overlapping.ID
		}

		newContractID, err := tx.AddNewContract(ctx, employeeID, previousContractID, params)
		if err != nil {
			return err
		}

		if err := s.createSalaryForContractSegment(
			ctx,
			tx,
			employeeID,
			previousContractID,
			newContractID,
			params.StartDate,
			params.SalaryAssignment,
		); err != nil {
			return err
		}

		emp, err = tx.GetEmployeeByID(ctx, employeeID)
		return err
	})
	if err != nil {
		s.logError(ctx, "CreateNewContract", err,
			zap.String("employee_id", employeeID.String()),
		)
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) createSalaryForContractSegment(
	ctx context.Context,
	tx domain.EmployeeTxRepository,
	employeeID uuid.UUID,
	previousContractID *uuid.UUID,
	newContractID uuid.UUID,
	startDate time.Time,
	requested *domain.CreateEmployeeSalaryAssignmentParams,
) error {
	var previousSalary *domain.EmployeeSalaryAssignmentInfo
	if previousContractID != nil {
		salary, err := tx.GetActiveEmployeeSalaryAssignment(
			ctx,
			employeeID,
			previousContractID,
			startDate,
		)
		if err != nil {
			return err
		}
		previousSalary = salary
	}

	salaryScaleStepID := uuid.Nil
	if requested != nil {
		salaryScaleStepID = requested.SalaryScaleStepID
	} else if previousSalary != nil {
		salaryScaleStepID = previousSalary.SalaryScaleStepID
	} else {
		return fmt.Errorf(
			"%w: salary assignment is required when no active previous salary exists",
			domain.ErrContractChangeInvalid,
		)
	}

	if previousSalary != nil {
		if err := tx.EndEmployeeSalaryAssignment(ctx, previousSalary.ID, startDate); err != nil {
			return err
		}
	}

	params := domain.CreateEmployeeSalaryAssignmentParams{
		SalaryScaleStepID: salaryScaleStepID,
		EffectiveFrom:     &startDate,
	}
	_, err := tx.CreateEmployeeSalaryAssignment(ctx, employeeID, &newContractID, params)
	return err
}

func (s *EmployeeService) ListEmployeeContracts(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.EmployeeContractDetail, error) {
	return s.repo.ListEmployeeContracts(ctx, employeeID)
}

func (s *EmployeeService) UpdateEmployeeContract(
	ctx context.Context,
	employeeID uuid.UUID,
	contractID uuid.UUID,
	params domain.UpdateEmployeeContractParams,
) (*domain.EmployeeContractDetail, error) {
	if err := validateOptionalContractHours(params.ContractType, params.HoursPerWeek); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrContractChangeInvalid, err)
	}

	contract, err := s.repo.UpdateEmployeeContract(ctx, employeeID, contractID, params)
	if err != nil {
		s.logError(ctx, "UpdateEmployeeContract", err,
			zap.String("contract_id", contractID.String()),
			zap.String("employee_id", employeeID.String()),
		)
		return nil, err
	}

	return contract, nil
}

func (s *EmployeeService) UpdateEmployeeContractSalary(
	ctx context.Context,
	employeeID uuid.UUID,
	contractID uuid.UUID,
	params domain.UpdateEmployeeContractSalaryParams,
) (*domain.EmployeeSalaryAssignmentDetail, error) {
	salary, err := s.repo.UpdateEmployeeContractSalary(ctx, employeeID, contractID, params)
	if err != nil {
		s.logError(ctx, "UpdateEmployeeContractSalary", err,
			zap.String("contract_id", contractID.String()),
			zap.String("employee_id", employeeID.String()),
		)
		return nil, err
	}

	return salary, nil
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

func validateContractHours(contractType string, hoursPerWeek *float64) error {
	switch contractType {
	case "permanent", "temporary":
		if hoursPerWeek == nil || *hoursPerWeek <= 0 {
			return fmt.Errorf("%s contract requires hours_per_week > 0", contractType)
		}
	case "on_call":
		if hoursPerWeek != nil {
			return fmt.Errorf("on_call contract cannot have hours_per_week")
		}
	default:
		return fmt.Errorf("invalid contract_type: %s", contractType)
	}
	return nil
}

func validateOptionalContractHours(contractType *string, hoursPerWeek *float64) error {
	if contractType == nil {
		if hoursPerWeek != nil && *hoursPerWeek <= 0 {
			return fmt.Errorf("hours_per_week must be > 0")
		}
		return nil
	}
	return validateContractHours(*contractType, hoursPerWeek)
}
