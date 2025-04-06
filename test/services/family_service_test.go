package services_test

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestCreateFamily(t *testing.T) {
	tests := []struct {
		name      string
		family    *models.Family
		setupRepo func(*test.MockFamilyRepository, *test.MockMemberRepository)
		wantErr   bool
	}{
		{
			name:   "successful creation",
			family: test.CreateValidFamily(),
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				fr.On("Create", mock.Anything, mock.AnythingOfType("*models.Family")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "validation failed - empty numero socio",
			family: &models.Family{
				NumeroSocio: "",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				// No expectations needed as validation should fail before repo call
			},
			wantErr: true,
		},
		{
			name: "member origin not found",
			family: &models.Family{
				NumeroSocio:     "B0002",
				MiembroOrigenID: test.UintPtr(999),
				EsposoNombre:    "Juan",
				EsposoApellidos: "Pérez",
				EsposaNombre:    "Maria",
				EsposaApellidos: "Lopez",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				mr.On("GetByID", mock.Anything, uint(999)).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup repos
			familyRepo := new(test.MockFamilyRepository)
			memberRepo := new(test.MockMemberRepository)
			tt.setupRepo(familyRepo, memberRepo)

			// Create service
			service := services.NewFamilyService(familyRepo, memberRepo)

			// Execute test
			err := service.Create(context.Background(), tt.family)

			// Assert results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			familyRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateFamily(t *testing.T) {
	tests := []struct {
		name      string
		family    *models.Family
		setupRepo func(*test.MockFamilyRepository, *test.MockMemberRepository)
		wantErr   bool
	}{
		{
			name: "successful update",
			family: &models.Family{
				ID:              1,
				NumeroSocio:     "B0001",
				EsposoNombre:    "Juan Updated",
				EsposoApellidos: "Pérez",
				EsposaNombre:    "Maria",
				EsposaApellidos: "Lopez",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				fr.On("GetByID", mock.Anything, uint(1)).Return(&models.Family{ID: 1}, nil)
				fr.On("Update", mock.Anything, mock.AnythingOfType("*models.Family")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "family not found",
			family: &models.Family{
				ID:          999,
				NumeroSocio: "A0999",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				fr.On("GetByID", mock.Anything, uint(999)).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup repos
			familyRepo := new(test.MockFamilyRepository)
			memberRepo := new(test.MockMemberRepository)
			tt.setupRepo(familyRepo, memberRepo)

			// Create service
			service := services.NewFamilyService(familyRepo, memberRepo)

			// Execute test
			err := service.Update(context.Background(), tt.family)

			// Assert results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			familyRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteFamily(t *testing.T) {
	tests := []struct {
		name      string
		familyID  uint
		setupRepo func(*test.MockFamilyRepository)
		wantErr   bool
	}{
		{
			name:     "successful delete",
			familyID: 1,
			setupRepo: func(fr *test.MockFamilyRepository) {
				fr.On("GetByID", mock.Anything, uint(1)).Return(&models.Family{ID: 1}, nil)
				fr.On("Delete", mock.Anything, uint(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "family not found",
			familyID: 999,
			setupRepo: func(fr *test.MockFamilyRepository) {
				fr.On("GetByID", mock.Anything, uint(999)).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup repo
			familyRepo := new(test.MockFamilyRepository)
			tt.setupRepo(familyRepo)

			// Create service
			service := services.NewFamilyService(familyRepo, nil)

			// Execute test
			err := service.Delete(context.Background(), tt.familyID)

			// Assert results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			familyRepo.AssertExpectations(t)
		})
	}
}
