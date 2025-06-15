package resolvers

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// Resolver contiene las dependencias necesarias para los resolvers
type Resolver struct {
	memberService    input.MemberService
	familyService    input.FamilyService
	paymentService   input.PaymentService
	cashFlowService  input.CashFlowService
	authService      input.AuthService      // Añadimos authService
	userService      input.UserService      // Añadimos userService
	loginRateLimiter *auth.LoginRateLimiter // Añadimos rate limiter para login
}

// NewResolver crea una nueva instancia del Resolver principal para GraphQL
// con todas las dependencias necesarias para los resolvers anidados.
func NewResolver(
	memberService input.MemberService,
	familyService input.FamilyService,
	paymentService input.PaymentService,
	cashFlowService input.CashFlowService,
	authService input.AuthService, // Añadimos el parámetro
	userService input.UserService, // Añadimos el servicio de usuarios
	loginRateLimiter *auth.LoginRateLimiter, // Añadimos el rate limiter
) *Resolver {
	return &Resolver{
		memberService:    memberService,
		familyService:    familyService,
		paymentService:   paymentService,
		cashFlowService:  cashFlowService,
		authService:      authService,      // Asignamos el servicio
		userService:      userService,      // Asignamos el servicio de usuarios
		loginRateLimiter: loginRateLimiter, // Asignamos el rate limiter
	}
}

// Member mutation methods

// CreateMember creates a new member
func (r *Resolver) CreateMember(ctx context.Context, input model.CreateMemberInput) (*models.Member, error) {
	// Validate input
	memberResolver := &memberResolver{r}
	if err := memberResolver.validateCreateInput(&input); err != nil {
		return nil, err
	}

	// Map input to member model
	member, err := memberResolver.mapCreateInputToMember(&input)
	if err != nil {
		return nil, err
	}

	// Handle the mutation
	return memberResolver.handleMemberMutation(ctx, member)
}

// UpdateMember updates an existing member
func (r *Resolver) UpdateMember(ctx context.Context, input model.UpdateMemberInput) (*models.Member, error) {
	// Validate input
	memberResolver := &memberResolver{r}
	if err := memberResolver.validateUpdateInput(&input); err != nil {
		return nil, err
	}

	// Parse member ID
	memberID, err := parseID(input.MiembroID)
	if err != nil {
		return nil, err
	}

	// Get existing member
	existing, err := r.memberService.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.NewNotFoundError("Member")
	}

	// Map input to member model
	member := memberResolver.mapUpdateInputToMember(memberID, &input, existing)

	// Handle the mutation
	return memberResolver.handleMemberMutation(ctx, member)
}

// Auth methods (delegate to auth_resolver.go)

// Login handles user authentication
func (r *Resolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error) {
	authResolver := &authResolver{resolver: r}
	return authResolver.Login(ctx, input)
}

// Logout handles user logout
func (r *Resolver) Logout(ctx context.Context) (*model.MutationResponse, error) {
	authResolver := &authResolver{resolver: r}
	return authResolver.Logout(ctx)
}

// RefreshToken handles token refresh
func (r *Resolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (*model.TokenResponse, error) {
	authResolver := &authResolver{resolver: r}
	return authResolver.RefreshToken(ctx, input)
}

// User methods (delegate to user_resolver.go)

// GetUser retrieves a user by ID
func (r *Resolver) GetUser(ctx context.Context, id string) (*models.User, error) {
	return (&userResolver{r}).GetUser(ctx, id)
}

// ListUsers retrieves a list of users
func (r *Resolver) ListUsers(ctx context.Context, page *int, pageSize *int) ([]*models.User, error) {
	return (&userResolver{r}).ListUsers(ctx, page, pageSize)
}

// GetCurrentUser retrieves the current authenticated user
func (r *Resolver) GetCurrentUser(ctx context.Context) (*models.User, error) {
	return (&userResolver{r}).GetCurrentUser(ctx)
}

// CreateUser creates a new user
func (r *Resolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*models.User, error) {
	return (&userResolver{r}).CreateUser(ctx, input)
}

// UpdateUser updates an existing user
func (r *Resolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*models.User, error) {
	return (&userResolver{r}).UpdateUser(ctx, input)
}

// DeleteUser deletes a user
func (r *Resolver) DeleteUser(ctx context.Context, id string) (*model.MutationResponse, error) {
	return (&userResolver{r}).DeleteUser(ctx, id)
}

// ChangePassword changes the current user's password
func (r *Resolver) ChangePassword(ctx context.Context, input model.ChangePasswordInput) (*model.MutationResponse, error) {
	return (&userResolver{r}).ChangePassword(ctx, input)
}

// ResetUserPassword resets a user's password
func (r *Resolver) ResetUserPassword(ctx context.Context, userID string, newPassword string) (*model.MutationResponse, error) {
	return (&userResolver{r}).ResetUserPassword(ctx, userID, newPassword)
}
