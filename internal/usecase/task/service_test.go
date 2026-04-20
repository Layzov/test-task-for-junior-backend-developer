package task

import (
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestShouldGenerateOnDate(t *testing.T) {
	t.Parallel()

	testDate := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		template taskdomain.TaskTemplate
		want     bool
	}{
		{
			name: "daily interval match",
			template: taskdomain.TaskTemplate{
				IsActive: true,
				StartsOn: time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC),
				Recurrence: taskdomain.Recurrence{
					Type:         taskdomain.RecurrenceDailyInterval,
					IntervalDays: 2,
				},
			},
			want: true,
		},
		{
			name: "monthly day match",
			template: taskdomain.TaskTemplate{
				IsActive: true,
				StartsOn: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Recurrence: taskdomain.Recurrence{
					Type:      taskdomain.RecurrenceMonthlyDays,
					MonthDays: []int{5, 20},
				},
			},
			want: true,
		},
		{
			name: "specific date miss",
			template: taskdomain.TaskTemplate{
				IsActive: true,
				StartsOn: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Recurrence: taskdomain.Recurrence{
					Type:          taskdomain.RecurrenceSpecificDates,
					SpecificDates: []time.Time{time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)},
				},
			},
			want: false,
		},
		{
			name: "odd-even match",
			template: taskdomain.TaskTemplate{
				IsActive: true,
				StartsOn: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Recurrence: taskdomain.Recurrence{
					Type:        taskdomain.RecurrenceOddEven,
					OddEvenType: taskdomain.OddEvenEven,
				},
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := shouldGenerateOnDate(tc.template, testDate)
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValidateRecurrenceInput(t *testing.T) {
	t.Parallel()

	if _, err := validateRecurrenceInput(RecurrenceInput{
		Type:         taskdomain.RecurrenceDailyInterval,
		IntervalDays: 0,
	}); err == nil {
		t.Fatal("expected error for invalid daily interval")
	}

	recurrence, err := validateRecurrenceInput(RecurrenceInput{
		Type:      taskdomain.RecurrenceMonthlyDays,
		MonthDays: []int{15, 1, 15},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recurrence.MonthDays) != 2 || recurrence.MonthDays[0] != 1 || recurrence.MonthDays[1] != 15 {
		t.Fatalf("unexpected normalized month_days: %#v", recurrence.MonthDays)
	}
}
