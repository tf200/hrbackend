package seed

import (
	"sync"

	"github.com/google/uuid"
)

type State struct {
	mu            sync.RWMutex
	organizations map[string]uuid.UUID
	locations     map[string]uuid.UUID
	departments   map[string]uuid.UUID
	employees     map[string]uuid.UUID
	handbooks     map[string]uuid.UUID
	schedules     map[string]uuid.UUID
	payPeriods    map[string]uuid.UUID
}

func NewState() *State {
	return &State{
		organizations: make(map[string]uuid.UUID),
		locations:     make(map[string]uuid.UUID),
		departments:   make(map[string]uuid.UUID),
		employees:     make(map[string]uuid.UUID),
		handbooks:     make(map[string]uuid.UUID),
		schedules:     make(map[string]uuid.UUID),
		payPeriods:    make(map[string]uuid.UUID),
	}
}

func (s *State) PutOrganization(alias string, id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.organizations[alias] = id
}

func (s *State) OrganizationID(alias string) (uuid.UUID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.organizations[alias]
	return id, ok
}

func (s *State) PutLocation(alias string, id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.locations[alias] = id
}

func (s *State) LocationID(alias string) (uuid.UUID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.locations[alias]
	return id, ok
}

func (s *State) PutDepartment(alias string, id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.departments[alias] = id
}

func (s *State) DepartmentID(alias string) (uuid.UUID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.departments[alias]
	return id, ok
}

func (s *State) PutEmployee(alias string, id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.employees[alias] = id
}

func (s *State) EmployeeID(alias string) (uuid.UUID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.employees[alias]
	return id, ok
}

func (s *State) PutHandbook(alias string, id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handbooks[alias] = id
}

func (s *State) HandbookID(alias string) (uuid.UUID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.handbooks[alias]
	return id, ok
}

func (s *State) PutSchedule(alias string, id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[alias] = id
}

func (s *State) ScheduleID(alias string) (uuid.UUID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.schedules[alias]
	return id, ok
}

func (s *State) PutPayPeriod(alias string, id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payPeriods[alias] = id
}

func (s *State) PayPeriodID(alias string) (uuid.UUID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.payPeriods[alias]
	return id, ok
}
