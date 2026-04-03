package conv

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TimeFromPgTimestamptz(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}

func PgTimestamptzFromTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

func StringFromPgTime(value pgtype.Time) string {
	if !value.Valid {
		return ""
	}

	totalSeconds := value.Microseconds / 1000000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func TimeFromPgDate(value pgtype.Date) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}

func PgDateFromTime(value time.Time) pgtype.Date {
	if value.IsZero() {
		return pgtype.Date{}
	}
	return pgtype.Date{
		Time:  value,
		Valid: true,
	}
}

func TimePtrFromPgDate(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func PgTimeFromString(value string) (pgtype.Time, error) {
	parsed, err := time.Parse("15:04:05", value)
	if err != nil {
		parsed, err = time.Parse("15:04", value)
		if err != nil {
			return pgtype.Time{}, err
		}
	}

	return pgtype.Time{
		Microseconds: int64(
			parsed.Hour(),
		)*3600*1000000 + int64(
			parsed.Minute(),
		)*60*1000000 + int64(
			parsed.Second(),
		)*1000000,
		Valid: true,
	}, nil
}
