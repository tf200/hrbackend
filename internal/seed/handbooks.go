package seed

import (
	"context"
	"fmt"
	"strings"

	"hrbackend/internal/domain"
	"hrbackend/internal/repository"
	dbrepo "hrbackend/internal/repository/db"
	"hrbackend/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type HandbookStepSeed struct {
	SortOrder  int32
	Kind       string
	Title      string
	Body       *string
	Content    []byte
	IsRequired *bool
}

type HandbookTemplateSeed struct {
	Alias              string
	DepartmentAlias    string
	ActorEmployeeAlias *string
	Title              string
	Description        *string
	Steps              []HandbookStepSeed
}

type HandbooksSeeder struct {
	Templates []HandbookTemplateSeed
}

func (s HandbooksSeeder) Name() string {
	return "handbooks"
}

func (s HandbooksSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Templates) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed handbooks: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed handbooks: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	handbookRepo := repository.NewHandbookRepository(store)
	handbookService := service.NewHandbookService(handbookRepo, nil)

	for _, item := range s.Templates {
		if strings.TrimSpace(item.Alias) == "" {
			return fmt.Errorf("seed handbooks: alias is required")
		}
		if strings.TrimSpace(item.DepartmentAlias) == "" {
			return fmt.Errorf("seed handbooks[%s]: department alias is required", item.Alias)
		}
		if strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("seed handbooks[%s]: title is required", item.Alias)
		}
		if len(item.Steps) == 0 {
			return fmt.Errorf("seed handbooks[%s]: at least one step is required", item.Alias)
		}

		departmentID, ok := env.State.DepartmentID(item.DepartmentAlias)
		if !ok {
			return fmt.Errorf(
				"seed handbooks[%s]: missing department alias %q in seed state",
				item.Alias,
				item.DepartmentAlias,
			)
		}

		actorEmployeeID, err := resolveOptionalHandbookActor(env, item)
		if err != nil {
			return fmt.Errorf("seed handbooks[%s]: %w", item.Alias, err)
		}

		existingTemplates, err := handbookService.ListTemplatesByDepartment(ctx, departmentID)
		if err != nil {
			return fmt.Errorf("list templates by department: %w", err)
		}

		if published := findPublishedTemplate(existingTemplates, item.Title); published != nil {
			env.State.PutHandbook(item.Alias, published.ID)
			continue
		}

		template, err := findDraftTemplate(existingTemplates, item.Title)
		if err != nil {
			return fmt.Errorf("inspect existing draft templates: %w", err)
		}
		if template == nil {
			template, err = handbookService.CreateTemplateForDepartment(ctx, actorEmployeeID, domain.CreateTemplateForDepartmentParams{
				DepartmentID: departmentID,
				Title:        item.Title,
				Description:  item.Description,
			})
			if err != nil {
				return fmt.Errorf("create draft template: %w", err)
			}
		}

		currentSteps, err := handbookService.ListStepsByTemplate(ctx, template.ID)
		if err != nil {
			return fmt.Errorf("list template steps: %w", err)
		}
		if len(currentSteps) == 0 {
			for _, step := range item.Steps {
				if _, err := handbookService.CreateStep(ctx, domain.CreateStepParams{
					TemplateID: template.ID,
					SortOrder:  step.SortOrder,
					Kind:       step.Kind,
					Title:      step.Title,
					Body:       step.Body,
					Content:    step.Content,
					IsRequired: step.IsRequired,
				}); err != nil {
					return fmt.Errorf("create handbook step %d: %w", step.SortOrder, err)
				}
			}
		}

		published, err := handbookService.PublishTemplate(ctx, actorEmployeeID, domain.PublishTemplateParams{
			TemplateID: template.ID,
		})
		if err != nil {
			return fmt.Errorf("publish template: %w", err)
		}

		env.State.PutHandbook(item.Alias, published.ID)
	}

	return nil
}

func resolveOptionalHandbookActor(env Env, item HandbookTemplateSeed) (uuid.UUID, error) {
	if item.ActorEmployeeAlias == nil || strings.TrimSpace(*item.ActorEmployeeAlias) == "" {
		return uuid.Nil, nil
	}

	employeeID, ok := env.State.EmployeeID(strings.TrimSpace(*item.ActorEmployeeAlias))
	if !ok {
		return uuid.Nil, fmt.Errorf("missing actor employee alias %q in seed state", strings.TrimSpace(*item.ActorEmployeeAlias))
	}
	return employeeID, nil
}

func findPublishedTemplate(items []domain.HandbookTemplate, title string) *domain.HandbookTemplate {
	normalizedTitle := strings.TrimSpace(title)
	for _, item := range items {
		if item.Status == "published" && strings.TrimSpace(item.Title) == normalizedTitle {
			template := item
			return &template
		}
	}
	return nil
}

func findDraftTemplate(items []domain.HandbookTemplate, title string) (*domain.HandbookTemplate, error) {
	normalizedTitle := strings.TrimSpace(title)
	var draftCount int
	for _, item := range items {
		if item.Status != "draft" {
			continue
		}
		draftCount++
		if strings.TrimSpace(item.Title) != normalizedTitle {
			return nil, fmt.Errorf("department already has a different draft template")
		}
		template := item
		return &template, nil
	}
	if draftCount == 0 {
		return nil, nil
	}
	return nil, nil
}
