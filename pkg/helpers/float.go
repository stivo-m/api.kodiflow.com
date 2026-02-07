package helpers

import (
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

func NumericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}

	f, err := n.Float64Value()
	if err != nil {

		return 0
	}

	return f.Float64
}

func Float64ToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric

	s := strconv.FormatFloat(f, 'f', -1, 64)
	if err := n.Scan(s); err != nil {
		slog.Error("failed to parse float to pg type", "error", err)
		return pgtype.Numeric{}
	}

	return n
}
