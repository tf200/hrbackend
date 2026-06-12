package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestOrganizationServiceListAllLocationsInitialLoadCacheHit(t *testing.T) {
	ctx := context.Background()
	cachedPage := &domain.OrganizationLocationPage{
		Items:      []domain.OrganizationLocation{{ID: uuid.New(), Name: "Cached Location"}},
		TotalCount: 1,
	}
	repo := &fakeOrganizationRepository{}
	cache := &fakeCache{getPage: cachedPage, getHit: true}
	service := &OrganizationService{repository: repo, cache: cache, cacheTTL: time.Minute}

	page, err := service.ListAllLocations(ctx, domain.ListAllLocationsParams{
		Limit:  defaultLocationListCacheLimit,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListAllLocations returned error: %v", err)
	}
	if page.TotalCount != cachedPage.TotalCount || page.Items[0].Name != cachedPage.Items[0].Name {
		t.Fatalf("expected cached page, got %+v", page)
	}
	if repo.listAllLocationsCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.listAllLocationsCalls)
	}
	if cache.getKey != initialLocationListCacheKey {
		t.Fatalf("expected cache key %q, got %q", initialLocationListCacheKey, cache.getKey)
	}
}

func TestOrganizationServiceListAllLocationsInitialLoadCacheMissStoresPage(t *testing.T) {
	ctx := context.Background()
	repoPage := &domain.OrganizationLocationPage{
		Items:      []domain.OrganizationLocation{{ID: uuid.New(), Name: "DB Location"}},
		TotalCount: 1,
	}
	repo := &fakeOrganizationRepository{listAllLocationsPage: repoPage}
	cache := &fakeCache{}
	service := &OrganizationService{repository: repo, cache: cache, cacheTTL: 2 * time.Minute}

	page, err := service.ListAllLocations(ctx, domain.ListAllLocationsParams{
		Limit:  defaultLocationListCacheLimit,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListAllLocations returned error: %v", err)
	}
	if page != repoPage {
		t.Fatalf("expected repository page")
	}
	if repo.listAllLocationsCalls != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.listAllLocationsCalls)
	}
	if cache.setKey != initialLocationListCacheKey {
		t.Fatalf("expected set key %q, got %q", initialLocationListCacheKey, cache.setKey)
	}
	if cache.setTTL != 2*time.Minute {
		t.Fatalf("expected ttl %s, got %s", 2*time.Minute, cache.setTTL)
	}
}

func TestOrganizationServiceListAllLocationsBypassesCacheForNonInitialLoads(t *testing.T) {
	tests := []struct {
		name   string
		params domain.ListAllLocationsParams
	}{
		{
			name: "search",
			params: domain.ListAllLocationsParams{
				Limit:  defaultLocationListCacheLimit,
				Offset: 0,
				Search: "amsterdam",
			},
		},
		{
			name: "second page",
			params: domain.ListAllLocationsParams{
				Limit:  defaultLocationListCacheLimit,
				Offset: defaultLocationListCacheLimit,
			},
		},
		{
			name: "custom limit",
			params: domain.ListAllLocationsParams{
				Limit:  20,
				Offset: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrganizationRepository{listAllLocationsPage: &domain.OrganizationLocationPage{}}
			cache := &fakeCache{getHit: true, getPage: &domain.OrganizationLocationPage{TotalCount: 99}}
			service := &OrganizationService{repository: repo, cache: cache, cacheTTL: time.Minute}

			_, err := service.ListAllLocations(context.Background(), tt.params)
			if err != nil {
				t.Fatalf("ListAllLocations returned error: %v", err)
			}
			if cache.getCalls != 0 || cache.setCalls != 0 {
				t.Fatalf("expected cache bypass, got get=%d set=%d", cache.getCalls, cache.setCalls)
			}
			if repo.listAllLocationsCalls != 1 {
				t.Fatalf("expected repository call, got %d", repo.listAllLocationsCalls)
			}
		})
	}
}

func TestOrganizationServiceListAllLocationsCacheErrorsFallBackToRepository(t *testing.T) {
	repoPage := &domain.OrganizationLocationPage{Items: []domain.OrganizationLocation{{ID: uuid.New()}}}
	repo := &fakeOrganizationRepository{listAllLocationsPage: repoPage}
	cache := &fakeCache{getErr: errors.New("redis unavailable")}
	service := &OrganizationService{repository: repo, cache: cache, cacheTTL: time.Minute}

	page, err := service.ListAllLocations(context.Background(), domain.ListAllLocationsParams{
		Limit:  defaultLocationListCacheLimit,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListAllLocations returned error: %v", err)
	}
	if page != repoPage {
		t.Fatalf("expected repository page")
	}
	if repo.listAllLocationsCalls != 1 {
		t.Fatalf("expected repository call, got %d", repo.listAllLocationsCalls)
	}
}

func TestOrganizationServiceInvalidatesLocationListCacheAfterLocationWrite(t *testing.T) {
	locationID := uuid.New()
	repo := &fakeOrganizationRepository{
		createLocation: &domain.OrganizationLocation{ID: locationID},
	}
	cache := &fakeCache{}
	service := &OrganizationService{repository: repo, cache: cache, cacheTTL: time.Minute}

	_, err := service.CreateOrganizationLocation(
		context.Background(),
		uuid.New(),
		domain.CreateOrganizationLocationParams{Name: "New Location"},
	)
	if err != nil {
		t.Fatalf("CreateOrganizationLocation returned error: %v", err)
	}
	if cache.deletePrefix != locationListCachePrefix {
		t.Fatalf("expected invalidation prefix %q, got %q", locationListCachePrefix, cache.deletePrefix)
	}
}

func TestOrganizationServiceInvalidatesLocationListCacheAfterShiftWrite(t *testing.T) {
	shiftID := uuid.New()
	repo := &fakeOrganizationRepository{
		createShift: &domain.OrganizationLocationShift{ID: shiftID},
	}
	cache := &fakeCache{}
	service := &OrganizationService{repository: repo, cache: cache, cacheTTL: time.Minute}

	_, err := service.CreateShift(context.Background(), domain.CreateShiftParams{
		LocationID: uuid.New(),
		ShiftName:  "Morning",
		StartTime:  "09:00",
		EndTime:    "17:00",
	})
	if err != nil {
		t.Fatalf("CreateShift returned error: %v", err)
	}
	if cache.deletePrefix != locationListCachePrefix {
		t.Fatalf("expected invalidation prefix %q, got %q", locationListCachePrefix, cache.deletePrefix)
	}
}

type fakeCache struct {
	getCalls int
	getKey   string
	getHit   bool
	getErr   error
	getPage  *domain.OrganizationLocationPage

	setCalls int
	setKey   string
	setTTL   time.Duration
	setErr   error

	deletePrefix string
	deleteErr    error
}

func (f *fakeCache) Get(_ context.Context, key string, dest any) (bool, error) {
	f.getCalls++
	f.getKey = key
	if f.getErr != nil {
		return false, f.getErr
	}
	if !f.getHit || f.getPage == nil {
		return false, nil
	}
	page, ok := dest.(*domain.OrganizationLocationPage)
	if !ok {
		return false, errors.New("unexpected cache destination")
	}
	*page = *f.getPage
	return true, nil
}

func (f *fakeCache) Set(_ context.Context, key string, _ any, ttl time.Duration) error {
	f.setCalls++
	f.setKey = key
	f.setTTL = ttl
	return f.setErr
}

func (f *fakeCache) Delete(context.Context, ...string) error { return nil }

func (f *fakeCache) DeleteByPrefix(_ context.Context, prefix string) error {
	f.deletePrefix = prefix
	return f.deleteErr
}

func (f *fakeCache) Close() error { return nil }

type fakeOrganizationRepository struct {
	listAllLocationsCalls int
	listAllLocationsPage  *domain.OrganizationLocationPage
	listAllLocationsErr   error

	createLocation *domain.OrganizationLocation
	createShift    *domain.OrganizationLocationShift
}

func (f *fakeOrganizationRepository) CreateOrganization(
	context.Context,
	domain.CreateOrganizationParams,
) (*domain.Organization, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) UpdateOrganization(
	context.Context,
	uuid.UUID,
	domain.UpdateOrganizationParams,
) (*domain.Organization, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) DeleteOrganization(context.Context, uuid.UUID) error { return nil }

func (f *fakeOrganizationRepository) CreateOrganizationLocation(
	context.Context,
	uuid.UUID,
	domain.CreateOrganizationLocationParams,
) (*domain.OrganizationLocation, error) {
	return f.createLocation, nil
}

func (f *fakeOrganizationRepository) UpdateLocation(
	context.Context,
	uuid.UUID,
	domain.UpdateOrganizationLocationParams,
) (*domain.OrganizationLocation, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) DeleteLocation(context.Context, uuid.UUID) error { return nil }

func (f *fakeOrganizationRepository) CreateShift(
	context.Context,
	domain.CreateShiftParams,
) (*domain.OrganizationLocationShift, error) {
	return f.createShift, nil
}

func (f *fakeOrganizationRepository) UpdateShift(
	context.Context,
	uuid.UUID,
	domain.UpdateShiftParams,
) (*domain.OrganizationLocationShift, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) DeleteShift(context.Context, uuid.UUID) error { return nil }

func (f *fakeOrganizationRepository) GetShiftsByLocationID(
	context.Context,
	uuid.UUID,
) ([]domain.OrganizationLocationShift, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) GetOrganizationCounts(
	context.Context,
	uuid.UUID,
) (*domain.OrganizationCounts, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) GetGlobalOrganizationCounts(
	context.Context,
) (*domain.GlobalOrganizationCounts, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) GetOrganizationByID(
	context.Context,
	uuid.UUID,
) (*domain.Organization, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) GetLocationByID(
	context.Context,
	uuid.UUID,
) (*domain.OrganizationLocation, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) ListOrganizations(
	context.Context,
	domain.ListOrganizationsParams,
) (*domain.OrganizationPage, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) ListOrganizationLocations(
	context.Context,
	domain.ListOrganizationLocationsParams,
) (*domain.OrganizationLocationPage, error) {
	return nil, nil
}

func (f *fakeOrganizationRepository) ListAllLocations(
	context.Context,
	domain.ListAllLocationsParams,
) (*domain.OrganizationLocationPage, error) {
	f.listAllLocationsCalls++
	return f.listAllLocationsPage, f.listAllLocationsErr
}

func (f *fakeOrganizationRepository) ListOrganizationalRoles(
	context.Context,
	domain.ListOrganizationalRolesParams,
) ([]domain.OrganizationalRole, error) {
	return nil, nil
}

var _ domain.Cache = (*fakeCache)(nil)
var _ domain.OrganizationRepository = (*fakeOrganizationRepository)(nil)
