package service

import (
	"context"

	"hrbackend/internal/domain"
	"hrbackend/pkg/password"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EmployeeService struct {
	repo   domain.EmployeeRepository
	logger domain.Logger
}

func NewEmployeeService(repo domain.EmployeeRepository, logger domain.Logger) domain.EmployeeService {
	return &EmployeeService{repo: repo, logger: logger}
}

func (s *EmployeeService) GetEmployeeByID(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID) (*domain.EmployeeDetail, error) {
	emp, err := s.repo.GetEmployeeByID(ctx, id)
	if err != nil {
		s.logError(ctx, "GetEmployeeByID", err, zap.String("employee_id", id.String()))
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) GetEmployeeProfile(ctx context.Context, userID uuid.UUID) (*domain.EmployeeProfile, error) {
	profile, err := s.repo.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		s.logError(ctx, "GetEmployeeProfile", err, zap.String("user_id", userID.String()))
		return nil, err
	}
	return profile, nil
}

func (s *EmployeeService) ListEmployees(ctx context.Context, params domain.ListEmployeesParams) (*domain.EmployeePage, error) {
	page, err := s.repo.ListEmployees(ctx, params)
	if err != nil {
		s.logError(ctx, "ListEmployees", err)
		return nil, err
	}
	return page, nil
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, params domain.CreateEmployeeParams) (*domain.EmployeeDetail, error) {
	hashedPassword, err := password.HashPassword(params.UserPassword)
	if err != nil {
		s.logError(ctx, "CreateEmployee", err)
		return nil, domain.ErrPasswordHashFailed
	}
	params.UserPassword = hashedPassword

	emp, err := s.repo.CreateEmployee(ctx, params)
	if err != nil {
		s.logError(ctx, "CreateEmployee", err,
			zap.String("first_name", params.FirstName),
			zap.String("last_name", params.LastName),
		)
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, id uuid.UUID, params domain.UpdateEmployeeParams) (*domain.EmployeeDetail, error) {
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

func (s *EmployeeService) SearchEmployeesByNameOrEmail(ctx context.Context, search *string) ([]domain.EmployeeSearchResult, error) {
	results, err := s.repo.SearchEmployeesByNameOrEmail(ctx, search)
	if err != nil {
		s.logError(ctx, "SearchEmployeesByNameOrEmail", err)
		return nil, err
	}
	return results, nil
}

func (s *EmployeeService) GetContractDetails(ctx context.Context, employeeID uuid.UUID) (*domain.ContractDetails, error) {
	details, err := s.repo.GetContractDetails(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "GetContractDetails", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return details, nil
}

func (s *EmployeeService) AddContractDetails(ctx context.Context, employeeID uuid.UUID, params domain.AddContractDetailsParams) (*domain.EmployeeDetail, error) {
	emp, err := s.repo.AddContractDetails(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddContractDetails", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) UpdateIsSubcontractor(ctx context.Context, employeeID uuid.UUID, params domain.UpdateIsSubcontractorParams) (*domain.EmployeeDetail, error) {
	contractType := "loondienst"
	if params.IsSubcontractor {
		contractType = "ZZP"
	}
	emp, err := s.repo.UpdateIsSubcontractor(ctx, employeeID, contractType)
	if err != nil {
		s.logError(ctx, "UpdateIsSubcontractor", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) ListEducation(ctx context.Context, employeeID uuid.UUID) ([]domain.Education, error) {
	items, err := s.repo.ListEducation(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "ListEducation", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return items, nil
}

func (s *EmployeeService) AddEducation(ctx context.Context, employeeID uuid.UUID, params domain.CreateEducationParams) (*domain.Education, error) {
	edu, err := s.repo.AddEducation(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddEducation", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return edu, nil
}

func (s *EmployeeService) UpdateEducation(ctx context.Context, id uuid.UUID, params domain.UpdateEducationParams) (*domain.Education, error) {
	edu, err := s.repo.UpdateEducation(ctx, id, params)
	if err != nil {
		s.logError(ctx, "UpdateEducation", err, zap.String("education_id", id.String()))
		return nil, err
	}
	return edu, nil
}

func (s *EmployeeService) DeleteEducation(ctx context.Context, id uuid.UUID) (*domain.Education, error) {
	edu, err := s.repo.DeleteEducation(ctx, id)
	if err != nil {
		s.logError(ctx, "DeleteEducation", err, zap.String("education_id", id.String()))
		return nil, err
	}
	return edu, nil
}

func (s *EmployeeService) ListExperience(ctx context.Context, employeeID uuid.UUID) ([]domain.Experience, error) {
	items, err := s.repo.ListExperience(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "ListExperience", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return items, nil
}

func (s *EmployeeService) AddExperience(ctx context.Context, employeeID uuid.UUID, params domain.CreateExperienceParams) (*domain.Experience, error) {
	exp, err := s.repo.AddExperience(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddExperience", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return exp, nil
}

func (s *EmployeeService) UpdateExperience(ctx context.Context, id uuid.UUID, params domain.UpdateExperienceParams) (*domain.Experience, error) {
	exp, err := s.repo.UpdateExperience(ctx, id, params)
	if err != nil {
		s.logError(ctx, "UpdateExperience", err, zap.String("experience_id", id.String()))
		return nil, err
	}
	return exp, nil
}

func (s *EmployeeService) DeleteExperience(ctx context.Context, id uuid.UUID) (*domain.Experience, error) {
	exp, err := s.repo.DeleteExperience(ctx, id)
	if err != nil {
		s.logError(ctx, "DeleteExperience", err, zap.String("experience_id", id.String()))
		return nil, err
	}
	return exp, nil
}

func (s *EmployeeService) ListCertification(ctx context.Context, employeeID uuid.UUID) ([]domain.Certification, error) {
	items, err := s.repo.ListCertification(ctx, employeeID)
	if err != nil {
		s.logError(ctx, "ListCertification", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return items, nil
}

func (s *EmployeeService) AddCertification(ctx context.Context, employeeID uuid.UUID, params domain.CreateCertificationParams) (*domain.Certification, error) {
	cert, err := s.repo.AddCertification(ctx, employeeID, params)
	if err != nil {
		s.logError(ctx, "AddCertification", err, zap.String("employee_id", employeeID.String()))
		return nil, err
	}
	return cert, nil
}

func (s *EmployeeService) UpdateCertification(ctx context.Context, id uuid.UUID, params domain.UpdateCertificationParams) (*domain.Certification, error) {
	cert, err := s.repo.UpdateCertification(ctx, id, params)
	if err != nil {
		s.logError(ctx, "UpdateCertification", err, zap.String("certification_id", id.String()))
		return nil, err
	}
	return cert, nil
}

func (s *EmployeeService) DeleteCertification(ctx context.Context, id uuid.UUID) (*domain.Certification, error) {
	cert, err := s.repo.DeleteCertification(ctx, id)
	if err != nil {
		s.logError(ctx, "DeleteCertification", err, zap.String("certification_id", id.String()))
		return nil, err
	}
	return cert, nil
}

func (s *EmployeeService) logError(ctx context.Context, method string, err error, fields ...zap.Field) {
	if s.logger != nil {
		s.logger.LogError(ctx, "EmployeeService."+method, err.Error(), err, fields...)
	}
}
