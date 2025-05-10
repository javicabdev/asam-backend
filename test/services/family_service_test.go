package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

func TestCreateFamily(t *testing.T) {
	tests := []struct {
		name      string
		family    *models.Family
		setupRepo func(*test.MockFamilyRepository, *test.MockMemberRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:   "successful creation",
			family: test.CreateValidFamily(),
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				fr.On("Create", mock.Anything, mock.AnythingOfType("*models.Family")).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "validation failed - empty numero socio",
			family: &models.Family{
				NumeroSocio: "",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				// No se llama al repositorio porque la validación falla antes
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsValidationError(err), "debería ser un error de validación")
			},
		},
		{
			name: "member origin not found",
			family: &models.Family{
				NumeroSocio:              "B0002",
				MiembroOrigenID:          test.UintPtr(999),
				EsposoNombre:             "Juan",
				EsposoApellidos:          "Pérez",
				EsposaNombre:             "Maria",
				EsposaApellidos:          "Lopez",
				EsposoDocumentoIdentidad: "12345678A",
				EsposaDocumentoIdentidad: "87654321B",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				mr.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("member"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFoundError(err), "debería ser un error de no encontrado")
			},
		},
		{
			name:   "repository error",
			family: test.CreateValidFamily(),
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				fr.On("Create", mock.Anything, mock.AnythingOfType("*models.Family")).Return(errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			familyRepo := new(test.MockFamilyRepository)
			memberRepo := new(test.MockMemberRepository)
			tt.setupRepo(familyRepo, memberRepo)

			service := services.NewFamilyService(familyRepo, memberRepo)
			err := service.Create(context.Background(), tt.family)

			tt.checkErr(t, err)
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
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "successful update",
			family: &models.Family{
				ID:                       1,
				NumeroSocio:              "B0001",
				EsposoNombre:             "Juan Updated",
				EsposoApellidos:          "Pérez",
				EsposaNombre:             "Maria",
				EsposaApellidos:          "Lopez",
				EsposoDocumentoIdentidad: "12345678A",
				EsposaDocumentoIdentidad: "87654321B",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				// Se debe devolver una familia completa con documentos para que pase la validación
				validFamily := &models.Family{
					ID:                       1,
					NumeroSocio:              "B0001",
					EsposoNombre:             "Juan",
					EsposoApellidos:          "Pérez",
					EsposaNombre:             "Maria",
					EsposaApellidos:          "Lopez",
					EsposoDocumentoIdentidad: "12345678A",
					EsposaDocumentoIdentidad: "87654321B",
				}
				fr.On("GetByID", mock.Anything, uint(1)).Return(validFamily, nil)
				fr.On("Update", mock.Anything, mock.AnythingOfType("*models.Family")).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "family not found",
			family: &models.Family{
				ID:          999,
				NumeroSocio: "A0999",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				fr.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("family"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFoundError(err), "debería ser un error de no encontrado")
			},
		},
		{
			name: "repository error",
			family: &models.Family{
				ID:                       1,
				NumeroSocio:              "B0001",
				EsposoNombre:             "Juan",
				EsposoApellidos:          "Pérez",
				EsposaNombre:             "Maria",
				EsposaApellidos:          "Lopez",
				EsposoDocumentoIdentidad: "12345678A",
				EsposaDocumentoIdentidad: "87654321B",
			},
			setupRepo: func(fr *test.MockFamilyRepository, mr *test.MockMemberRepository) {
				// Se debe devolver una familia completa con documentos para que pase la validación
				validFamily := &models.Family{
					ID:                       1,
					NumeroSocio:              "B0001",
					EsposoNombre:             "Juan",
					EsposoApellidos:          "Pérez",
					EsposaNombre:             "Maria",
					EsposaApellidos:          "Lopez",
					EsposoDocumentoIdentidad: "12345678A",
					EsposaDocumentoIdentidad: "87654321B",
				}
				fr.On("GetByID", mock.Anything, uint(1)).Return(validFamily, nil)
				fr.On("Update", mock.Anything, mock.AnythingOfType("*models.Family")).Return(errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			familyRepo := new(test.MockFamilyRepository)
			memberRepo := new(test.MockMemberRepository)
			tt.setupRepo(familyRepo, memberRepo)

			service := services.NewFamilyService(familyRepo, memberRepo)
			err := service.Update(context.Background(), tt.family)

			tt.checkErr(t, err)
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
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:     "successful delete",
			familyID: 1,
			setupRepo: func(fr *test.MockFamilyRepository) {
				// Retornamos una familia válida completa
				validFamily := &models.Family{
					ID:                       1,
					NumeroSocio:              "B0001",
					EsposoNombre:             "Juan",
					EsposoApellidos:          "Pérez",
					EsposaNombre:             "Maria",
					EsposaApellidos:          "Lopez",
					EsposoDocumentoIdentidad: "12345678A",
					EsposaDocumentoIdentidad: "87654321B",
				}
				fr.On("GetByID", mock.Anything, uint(1)).Return(validFamily, nil)
				fr.On("Delete", mock.Anything, uint(1)).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "family not found",
			familyID: 999,
			setupRepo: func(fr *test.MockFamilyRepository) {
				fr.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("family"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFoundError(err), "debería ser un error de no encontrado")
			},
		},
		{
			name:     "repository error",
			familyID: 1,
			setupRepo: func(fr *test.MockFamilyRepository) {
				// Retornamos una familia válida completa
				validFamily := &models.Family{
					ID:                       1,
					NumeroSocio:              "B0001",
					EsposoNombre:             "Juan",
					EsposoApellidos:          "Pérez",
					EsposaNombre:             "Maria",
					EsposaApellidos:          "Lopez",
					EsposoDocumentoIdentidad: "12345678A",
					EsposaDocumentoIdentidad: "87654321B",
				}
				fr.On("GetByID", mock.Anything, uint(1)).Return(validFamily, nil)
				fr.On("Delete", mock.Anything, uint(1)).Return(errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			familyRepo := new(test.MockFamilyRepository)
			tt.setupRepo(familyRepo)

			service := services.NewFamilyService(familyRepo, nil)
			err := service.Delete(context.Background(), tt.familyID)

			tt.checkErr(t, err)
			familyRepo.AssertExpectations(t)
		})
	}
}
