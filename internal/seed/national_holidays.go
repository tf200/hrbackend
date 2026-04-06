package seed

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type NationalHolidaySeed struct {
	CountryCode string
	HolidayDate time.Time
	Name        string
	IsNational  bool
}

type NationalHolidaysSeeder struct {
	Holidays []NationalHolidaySeed
}

func (s NationalHolidaysSeeder) Name() string {
	return "national_holidays"
}

func (s NationalHolidaysSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Holidays) == 0 {
		return nil
	}

	for _, item := range s.Holidays {
		if strings.TrimSpace(item.CountryCode) == "" {
			return fmt.Errorf("seed national_holidays: country code is required")
		}
		if item.HolidayDate.IsZero() {
			return fmt.Errorf("seed national_holidays[%s]: holiday date is required", item.Name)
		}
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("seed national_holidays: holiday name is required")
		}

		if _, err := env.DB.Exec(ctx, `
			INSERT INTO national_holidays (
				country_code,
				holiday_date,
				name,
				is_national
			) VALUES ($1, $2, $3, $4)
			ON CONFLICT (country_code, holiday_date) DO UPDATE
			SET
				name = EXCLUDED.name,
				is_national = EXCLUDED.is_national,
				updated_at = CURRENT_TIMESTAMP
		`, strings.TrimSpace(item.CountryCode), dateOnlyUTC(item.HolidayDate), item.Name, item.IsNational); err != nil {
			return fmt.Errorf("seed national_holidays[%s %s]: %w",
				item.CountryCode,
				dateOnlyUTC(item.HolidayDate).Format("2006-01-02"),
				err,
			)
		}
	}

	return nil
}

func dateOnlyUTC(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
