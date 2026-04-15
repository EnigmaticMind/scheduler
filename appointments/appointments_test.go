package appointments

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubAppointmentDAO struct {
	getAppointmentsFn   func(ctx context.Context, trainerID int) ([]AppointmentDL, error)
	createAppointmentFn func(ctx context.Context, apt AppointmentDL) (AppointmentDL, error)
	getCalls            int
	createCalls         int
}

func (s *stubAppointmentDAO) GetAppointments(ctx context.Context, trainerID int) ([]AppointmentDL, error) {
	s.getCalls++
	if s.getAppointmentsFn == nil {
		return nil, nil
	}
	return s.getAppointmentsFn(ctx, trainerID)
}

func (s *stubAppointmentDAO) CreateAppointment(ctx context.Context, apt AppointmentDL) (AppointmentDL, error) {
	s.createCalls++
	if s.createAppointmentFn == nil {
		return apt, nil
	}
	return s.createAppointmentFn(ctx, apt)
}

func mustParseRFC3339(t *testing.T, value string) time.Time {
	t.Helper()
	out, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("failed to parse time %q: %v", value, err)
	}
	return out
}

func TestCreateAppointment(t *testing.T) {
	t.Run("creates valid weekday appointment and forces 30 minute end", func(t *testing.T) {
		start := mustParseRFC3339(t, "2026-04-13T09:00:00-07:00")
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int) ([]AppointmentDL, error) {
				return nil, nil
			},
			createAppointmentFn: func(_ context.Context, apt AppointmentDL) (AppointmentDL, error) {
				return AppointmentDL{
					ID:        10,
					UserID:    apt.UserID,
					TrainerID: apt.TrainerID,
					Start:     apt.Start,
					End:       apt.End,
				}, nil
			},
		}

		created, err := CreateAppointment(context.Background(), dao, 2, 1, start, start.Add(30*time.Minute))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dao.getCalls != 1 {
			t.Fatalf("expected one GetAppointments call, got %d", dao.getCalls)
		}
		if dao.createCalls != 1 {
			t.Fatalf("expected one CreateAppointment call, got %d", dao.createCalls)
		}

		if created.Start != start.UTC() {
			t.Fatalf("expected start %v, got %v", start.UTC(), created.Start)
		}
		wantEnd := start.UTC().Add(30 * time.Minute)
		if created.End != wantEnd {
			t.Fatalf("expected forced end %v, got %v", wantEnd, created.End)
		}
	})

	testCases := []struct {
		name   string
		start  string
		booked []AppointmentDL
	}{
		{
			name:  "rejects weekend appointment",
			start: "2026-04-12T09:00:00-07:00",
		},
		{
			name:  "rejects non half hour boundary",
			start: "2026-04-13T09:15:00-07:00",
		},
		{
			name:  "rejects start before business hours",
			start: "2026-04-13T07:30:00-07:00",
		},
		{
			name:  "rejects start after last slot",
			start: "2026-04-13T17:00:00-07:00",
		},
		{
			name:  "rejects overlapping appointment",
			start: "2026-04-13T09:00:00-07:00",
			booked: []AppointmentDL{
				{
					TrainerID: 1,
					Start:     mustParseRFC3339(t, "2026-04-13T16:00:00Z"),
					End:       mustParseRFC3339(t, "2026-04-13T16:30:00Z"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := mustParseRFC3339(t, tc.start)
			dao := &stubAppointmentDAO{
				getAppointmentsFn: func(context.Context, int) ([]AppointmentDL, error) {
					return tc.booked, nil
				},
			}

			_, err := CreateAppointment(context.Background(), dao, 2, 1, start, start.Add(30*time.Minute))
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if dao.createCalls != 0 {
				t.Fatalf("expected create not to be called, got %d", dao.createCalls)
			}
		})
	}

	t.Run("returns DAO GetAppointments error", func(t *testing.T) {
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int) ([]AppointmentDL, error) {
				return nil, errors.New("read failed")
			},
		}
		start := mustParseRFC3339(t, "2026-04-13T09:00:00-07:00")

		_, err := CreateAppointment(context.Background(), dao, 2, 1, start, start.Add(30*time.Minute))
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("returns DAO CreateAppointment error", func(t *testing.T) {
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int) ([]AppointmentDL, error) {
				return nil, nil
			},
			createAppointmentFn: func(context.Context, AppointmentDL) (AppointmentDL, error) {
				return AppointmentDL{}, errors.New("write failed")
			},
		}
		start := mustParseRFC3339(t, "2026-04-13T09:00:00-07:00")

		_, err := CreateAppointment(context.Background(), dao, 2, 1, start, start.Add(30*time.Minute))
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

func TestAvailableAppointments(t *testing.T) {
	t.Run("returns nil when endsAt is not after startsAt", func(t *testing.T) {
		start := mustParseRFC3339(t, "2026-04-13T09:00:00-07:00")
		dao := &stubAppointmentDAO{}

		got, err := AvailableAppointments(context.Background(), dao, 1, start, start)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil result, got %v", got)
		}
	})

	t.Run("returns weekday slots only during business hours", func(t *testing.T) {
		start := mustParseRFC3339(t, "2026-04-10T00:00:00-07:00") // Friday
		end := mustParseRFC3339(t, "2026-04-13T23:59:59-07:00")   // Monday
		dao := &stubAppointmentDAO{}

		got, err := AvailableAppointments(context.Background(), dao, 1, start, end)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(got) == 0 {
			t.Fatal("expected available slots, got none")
		}

		loc, _ := time.LoadLocation("America/Los_Angeles")
		for _, slot := range got {
			localStart := slot.Start.In(loc)
			if localStart.Weekday() == time.Saturday || localStart.Weekday() == time.Sunday {
				t.Fatalf("unexpected weekend slot: %v", localStart)
			}
			mins := localStart.Hour()*60 + localStart.Minute()
			if mins < 8*60 || mins > 16*60+30 {
				t.Fatalf("unexpected out-of-hours slot: %v", localStart)
			}
		}
	})

	t.Run("excludes overlapping booked slots", func(t *testing.T) {
		start := mustParseRFC3339(t, "2026-04-13T08:00:00-07:00")
		end := mustParseRFC3339(t, "2026-04-13T10:00:00-07:00")
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int) ([]AppointmentDL, error) {
				return []AppointmentDL{
					{
						TrainerID: 1,
						Start:     mustParseRFC3339(t, "2026-04-13T15:30:00Z"),
						End:       mustParseRFC3339(t, "2026-04-13T16:00:00Z"),
					},
				}, nil
			},
		}

		got, err := AvailableAppointments(context.Background(), dao, 1, start, end)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("expected 3 slots after excluding overlap, got %d", len(got))
		}
		for _, slot := range got {
			if slot.Start.Equal(mustParseRFC3339(t, "2026-04-13T15:30:00Z")) {
				t.Fatalf("booked slot should not be available: %v", slot.Start)
			}
		}
	})

	t.Run("respects query window boundaries", func(t *testing.T) {
		start := mustParseRFC3339(t, "2026-04-13T08:30:00-07:00")
		end := mustParseRFC3339(t, "2026-04-13T09:00:00-07:00")
		dao := &stubAppointmentDAO{}

		got, err := AvailableAppointments(context.Background(), dao, 1, start, end)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected exactly one slot, got %d", len(got))
		}
		if got[0].Start != mustParseRFC3339(t, "2026-04-13T15:30:00Z") {
			t.Fatalf("unexpected slot start: %v", got[0].Start)
		}
	})

	t.Run("propagates DAO errors", func(t *testing.T) {
		start := mustParseRFC3339(t, "2026-04-13T08:00:00-07:00")
		end := mustParseRFC3339(t, "2026-04-13T09:00:00-07:00")
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int) ([]AppointmentDL, error) {
				return nil, errors.New("read failed")
			},
		}

		_, err := AvailableAppointments(context.Background(), dao, 1, start, end)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}
