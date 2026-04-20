package task

import "time"

type RecurrenceType string

const (
	RecurrenceDailyInterval RecurrenceType = "daily_interval"
	RecurrenceMonthlyDays   RecurrenceType = "monthly_days"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceOddEven       RecurrenceType = "odd_even"
)

type OddEvenType string

const (
	OddEvenOdd  OddEvenType = "odd"
	OddEvenEven OddEvenType = "even"
)

func (t RecurrenceType) Valid() bool {
	switch t {
	case RecurrenceDailyInterval, RecurrenceMonthlyDays, RecurrenceSpecificDates, RecurrenceOddEven:
		return true
	default:
		return false
	}
}

func (t OddEvenType) Valid() bool {
	switch t {
	case OddEvenOdd, OddEvenEven:
		return true
	default:
		return false
	}
}

type Recurrence struct {
	Type          RecurrenceType `json:"type"`
	IntervalDays  int            `json:"interval_days,omitempty"`
	MonthDays     []int          `json:"month_days,omitempty"`
	SpecificDates []time.Time    `json:"specific_dates,omitempty"`
	OddEvenType   OddEvenType    `json:"odd_even_type,omitempty"`
}

type TaskTemplate struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Recurrence  Recurrence `json:"recurrence"`
	StartsOn    time.Time  `json:"starts_on"`
	EndsOn      *time.Time `json:"ends_on,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
