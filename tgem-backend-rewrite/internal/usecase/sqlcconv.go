package usecase

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// Boundary helpers between the model layer (uint, string, time.Time,
// decimal.Decimal, bool) and the sqlc-generated nullable types inherited
// from the phase-5 baseline schema. Every column GORM did not explicitly
// mark NOT NULL is nullable in the dump, so most non-id columns surface
// as pgtype.* in db.* structs.
//
// These helpers encode the GORM coercion rule: NULL becomes the zero
// value on read, and zero values become non-NULL writes. They will be
// removed once a phase-7 schema-tightening migration adds NOT NULL
// constraints to the columns that GORM treated as non-nullable.

func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func pgInt8(v uint) pgtype.Int8 {
	return pgtype.Int8{Int64: int64(v), Valid: true}
}

func uintFromPgInt8(v pgtype.Int8) uint {
	if !v.Valid {
		return 0
	}
	return uint(v.Int64)
}

func pgBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timeFromPgTimestamptz(v pgtype.Timestamptz) time.Time {
	if !v.Valid {
		return time.Time{}
	}
	return v.Time
}

func pgNumericFromDecimal(d decimal.Decimal) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   d.Coefficient(),
		Exp:   d.Exponent(),
		Valid: true,
	}
}

func decimalFromPgNumeric(v pgtype.Numeric) decimal.Decimal {
	if !v.Valid || v.NaN || v.Int == nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(v.Int, v.Exp)
}

// pgNumericFromFloat64 routes a Go float64 into a Postgres numeric column
// without going through decimal.Decimal. Used for columns where the model
// itself is float64 (operations.planned_amount_for_project, etc.).
func pgNumericFromFloat64(v float64) pgtype.Numeric {
	return pgNumericFromDecimal(decimal.NewFromFloat(v))
}

// float64FromPgNumeric reads a Postgres numeric column into a Go float64.
// Mirrors the inline conversion from operation_usecase's toModelOperation,
// promoted here for object-family aggregates where multiple `length`-style
// fields are float64 in the model but numeric in the schema.
func float64FromPgNumeric(v pgtype.Numeric) float64 {
	if !v.Valid {
		return 0
	}
	f, err := v.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

// decimalFromString parses a Postgres numeric stored as text (a GORM idiom
// for some columns like auction_participant_prices.unit_price). Returns
// decimal.Zero on empty input or parse error to mirror the GORM scanner's
// permissive behavior.
func decimalFromString(s string) decimal.Decimal {
	if s == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}
