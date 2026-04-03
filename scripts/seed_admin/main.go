package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"hrbackend/pkg/password"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type seedConfig struct {
	DBSource      string
	AdminEmail    string
	AdminPassword string
	Profile       adminProfileDefaults
}

type adminProfileDefaults struct {
	FirstName           string
	LastName            string
	BSN                 string
	Street              string
	HouseNumber         string
	HouseNumberAddition string
	PostalCode          string
	City                string
	Position            string
	PrivateEmail        string
	WorkEmail           string
	PrivatePhone        string
	WorkPhone           string
	HomeTelephone       string
	Gender              string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := loadConfigFromEnv()
	if err != nil {
		exitErr(err)
	}

	pool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		exitErr(fmt.Errorf("connect db: %w", err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		exitErr(fmt.Errorf("ping db: %w", err))
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		exitErr(fmt.Errorf("begin tx: %w", err))
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	adminRoleID, err := ensureAdminRole(ctx, tx)
	if err != nil {
		exitErr(err)
	}

	grantedCount, err := syncAdminRolePermissions(ctx, tx, adminRoleID)
	if err != nil {
		exitErr(err)
	}

	userID, createdUser, err := ensureAdminUser(ctx, tx, cfg.AdminEmail, cfg.AdminPassword)
	if err != nil {
		exitErr(err)
	}

	createdProfile, err := ensureEmployeeProfile(ctx, tx, userID, cfg)
	if err != nil {
		exitErr(err)
	}

	if err := ensureUserRole(ctx, tx, userID, adminRoleID); err != nil {
		exitErr(err)
	}

	if err := tx.Commit(ctx); err != nil {
		exitErr(fmt.Errorf("commit tx: %w", err))
	}

	fmt.Println("Admin seed completed.")
	fmt.Printf("Admin email: %s\n", cfg.AdminEmail)
	fmt.Printf("User created: %t\n", createdUser)
	fmt.Printf("Profile created: %t\n", createdProfile)
	fmt.Printf("Admin permission links ensured: %d\n", grantedCount)
}

func ensureAdminRole(ctx context.Context, tx pgx.Tx) (uuid.UUID, error) {
	if _, err := tx.Exec(ctx, `
		INSERT INTO roles (name, description)
		VALUES ('admin', 'System administrator with full access')
		ON CONFLICT (name) DO NOTHING
	`); err != nil {
		return uuid.Nil, fmt.Errorf("ensure admin role: %w", err)
	}

	var roleID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'admin'`).
		Scan(&roleID); err != nil {
		return uuid.Nil, fmt.Errorf("read admin role id: %w", err)
	}
	return roleID, nil
}

func syncAdminRolePermissions(ctx context.Context, tx pgx.Tx, roleID uuid.UUID) (int64, error) {
	var permissionCount int64
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM permissions`).
		Scan(&permissionCount); err != nil {
		return 0, fmt.Errorf("count permissions: %w", err)
	}
	if permissionCount == 0 {
		return 0, errors.New("no permissions found; run migrations before seeding admin")
	}

	tag, err := tx.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT $1, p.id
		FROM permissions p
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`, roleID)
	if err != nil {
		return 0, fmt.Errorf("sync admin role permissions: %w", err)
	}

	return tag.RowsAffected(), nil
}

func ensureAdminUser(
	ctx context.Context,
	tx pgx.Tx,
	email, plainPassword string,
) (uuid.UUID, bool, error) {
	var userID uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM custom_user WHERE email = $1`, email).Scan(&userID)
	if err == nil {
		return userID, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, fmt.Errorf("lookup admin user: %w", err)
	}

	hashedPassword, err := password.HashPassword(plainPassword)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("hash admin password: %w", err)
	}

	if err := tx.QueryRow(ctx, `
		INSERT INTO custom_user (password, email, is_active, profile_picture)
		VALUES ($1, $2, true, NULL)
		RETURNING id
	`, hashedPassword, email).Scan(&userID); err != nil {
		return uuid.Nil, false, fmt.Errorf("create admin user: %w", err)
	}

	return userID, true, nil
}

func ensureEmployeeProfile(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	cfg seedConfig,
) (bool, error) {
	var profileID uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM employee_profile WHERE user_id = $1`, userID).
		Scan(&profileID)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("lookup admin employee profile: %w", err)
	}

	profile := cfg.Profile
	if profile.PrivateEmail == "" {
		profile.PrivateEmail = cfg.AdminEmail
	}
	if profile.WorkEmail == "" {
		profile.WorkEmail = cfg.AdminEmail
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO employee_profile (
			user_id,
			first_name,
			last_name,
			bsn,
			street,
			house_number,
			house_number_addition,
			postal_code,
			city,
			position,
			private_email_address,
			work_email_address,
			private_phone_number,
			work_phone_number,
			home_telephone_number,
			gender
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9, NULLIF($10, ''),
			NULLIF($11, ''), NULLIF($12, ''), NULLIF($13, ''), NULLIF($14, ''), NULLIF($15, ''), $16::gender_enum
		)
	`, userID, profile.FirstName, profile.LastName, profile.BSN, profile.Street, profile.HouseNumber,
		profile.HouseNumberAddition, profile.PostalCode, profile.City, profile.Position,
		profile.PrivateEmail, profile.WorkEmail, profile.PrivatePhone, profile.WorkPhone,
		profile.HomeTelephone, profile.Gender); err != nil {
		return false, fmt.Errorf("create admin employee profile: %w", err)
	}

	return true, nil
}

func ensureUserRole(ctx context.Context, tx pgx.Tx, userID, roleID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET role_id = EXCLUDED.role_id
	`, userID, roleID); err != nil {
		return fmt.Errorf("assign admin role to user: %w", err)
	}
	return nil
}

func loadConfigFromEnv() (seedConfig, error) {
	dbSource := strings.TrimSpace(os.Getenv("MIGRATION_DB_SOURCE"))
	adminEmail := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	adminPassword := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))

	if dbSource == "" {
		return seedConfig{}, errors.New("MIGRATION_DB_SOURCE is required")
	}
	if adminEmail == "" {
		return seedConfig{}, errors.New("ADMIN_EMAIL is required")
	}
	if adminPassword == "" {
		return seedConfig{}, errors.New("ADMIN_PASSWORD is required")
	}

	profile := adminProfileDefaults{
		FirstName:           envOrDefault("ADMIN_FIRST_NAME", "System"),
		LastName:            envOrDefault("ADMIN_LAST_NAME", "Admin"),
		BSN:                 envOrDefault("ADMIN_BSN", "000000000"),
		Street:              envOrDefault("ADMIN_STREET", "Adminstraat"),
		HouseNumber:         envOrDefault("ADMIN_HOUSE_NUMBER", "1"),
		HouseNumberAddition: envOrDefault("ADMIN_HOUSE_NUMBER_ADDITION", ""),
		PostalCode:          envOrDefault("ADMIN_POSTAL_CODE", "1000AA"),
		City:                envOrDefault("ADMIN_CITY", "Amsterdam"),
		Position:            envOrDefault("ADMIN_POSITION", "Administrator"),
		PrivateEmail:        strings.TrimSpace(os.Getenv("ADMIN_PRIVATE_EMAIL")),
		WorkEmail:           strings.TrimSpace(os.Getenv("ADMIN_WORK_EMAIL")),
		PrivatePhone:        strings.TrimSpace(os.Getenv("ADMIN_PRIVATE_PHONE")),
		WorkPhone:           strings.TrimSpace(os.Getenv("ADMIN_WORK_PHONE")),
		HomeTelephone:       strings.TrimSpace(os.Getenv("ADMIN_HOME_TELEPHONE")),
		Gender:              normalizeGender(envOrDefault("ADMIN_GENDER", "unknown")),
	}

	return seedConfig{
		DBSource:      dbSource,
		AdminEmail:    adminEmail,
		AdminPassword: adminPassword,
		Profile:       profile,
	}, nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func normalizeGender(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "male", "female", "other", "unknown":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "unknown"
	}
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "seed-admin: %v\n", err)
	os.Exit(1)
}
