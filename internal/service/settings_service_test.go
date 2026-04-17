package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"hrbackend/internal/domain"
)

func TestSettingsServiceGetOrganizationProfile(t *testing.T) {
	expected := &domain.OrganizationProfile{
		Name:            "Acme Care",
		DefaultTimezone: "Europe/Amsterdam",
		CreatedAt:       time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC),
	}
	repo := &fakeSettingsRepository{profile: expected}
	service := &SettingsService{repository: repo}

	profile, err := service.GetOrganizationProfile(context.Background())
	if err != nil {
		t.Fatalf("GetOrganizationProfile returned error: %v", err)
	}

	if profile != expected {
		t.Fatalf("expected same profile pointer to be returned")
	}
}

func TestSettingsServiceGetOrganizationProfileReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("boom")
	repo := &fakeSettingsRepository{err: expectedErr}
	service := &SettingsService{repository: repo}

	_, err := service.GetOrganizationProfile(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSettingsServiceUpdateOrganizationProfile(t *testing.T) {
	expected := &domain.OrganizationProfile{
		Name:            "Updated Care",
		DefaultTimezone: "Europe/Amsterdam",
		CreatedAt:       time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC),
	}
	name := "Updated Care"
	repo := &fakeSettingsRepository{profile: expected}
	service := &SettingsService{repository: repo}

	profile, err := service.UpdateOrganizationProfile(context.Background(), domain.UpdateOrganizationProfileParams{
		Name: &name,
	})
	if err != nil {
		t.Fatalf("UpdateOrganizationProfile returned error: %v", err)
	}

	if profile != expected {
		t.Fatalf("expected same profile pointer to be returned")
	}
	if repo.updatedParams == nil || repo.updatedParams.Name != &name {
		t.Fatalf("expected update params to be forwarded to repository")
	}
}

type fakeSettingsRepository struct {
	profile       *domain.OrganizationProfile
	err           error
	updatedParams *domain.UpdateOrganizationProfileParams
}

func (f *fakeSettingsRepository) GetOrganizationProfile(
	_ context.Context,
) (*domain.OrganizationProfile, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.profile, nil
}

func (f *fakeSettingsRepository) UpdateOrganizationProfile(
	_ context.Context,
	params domain.UpdateOrganizationProfileParams,
) (*domain.OrganizationProfile, error) {
	f.updatedParams = &params
	if f.err != nil {
		return nil, f.err
	}

	return f.profile, nil
}
