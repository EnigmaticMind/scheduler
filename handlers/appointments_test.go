package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"scheduler/appointments"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
)

type stubAppointmentDAO struct {
	getAppointmentsFn   func(ctx context.Context, trainerID int64) ([]appointments.AppointmentDL, error)
	createAppointmentFn func(ctx context.Context, apt appointments.AppointmentDL) (appointments.AppointmentDL, error)
}

func (s *stubAppointmentDAO) GetAppointments(ctx context.Context, trainerID int64) ([]appointments.AppointmentDL, error) {
	if s.getAppointmentsFn == nil {
		return nil, nil
	}
	return s.getAppointmentsFn(ctx, trainerID)
}

func (s *stubAppointmentDAO) CreateAppointment(ctx context.Context, apt appointments.AppointmentDL) (appointments.AppointmentDL, error) {
	if s.createAppointmentFn == nil {
		return apt, nil
	}
	return s.createAppointmentFn(ctx, apt)
}

func decodeJSONBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	return out
}

func TestListAppointments(t *testing.T) {
	t.Run("returns 400 for invalid trainer_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/appointments?trainer_id=x&starts_at=2026-04-13T08:00:00-07:00&ends_at=2026-04-13T09:00:00-07:00", nil)
		rr := httptest.NewRecorder()

		ListAppointments(&stubAppointmentDAO{})(rr, req, httprouter.Params{})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
		}
		body := decodeJSONBody(t, rr)
		if _, ok := body["error"]; !ok {
			t.Fatalf("expected error envelope, got %v", body)
		}
	})

	t.Run("returns 400 for invalid starts_at", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/appointments?trainer_id=1&starts_at=bad&ends_at=2026-04-13T09:00:00-07:00", nil)
		rr := httptest.NewRecorder()

		ListAppointments(&stubAppointmentDAO{})(rr, req, httprouter.Params{})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns 400 for invalid ends_at", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/appointments?trainer_id=1&starts_at=2026-04-13T08:00:00-07:00&ends_at=bad", nil)
		rr := httptest.NewRecorder()

		ListAppointments(&stubAppointmentDAO{})(rr, req, httprouter.Params{})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns 200 with data envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/appointments?trainer_id=1&starts_at=2026-04-13T08:00:00-07:00&ends_at=2026-04-13T09:00:00-07:00", nil)
		rr := httptest.NewRecorder()
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int64) ([]appointments.AppointmentDL, error) {
				return nil, nil
			},
		}

		ListAppointments(dao)(rr, req, httprouter.Params{})
		if rr.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
		}
		body := decodeJSONBody(t, rr)
		if _, ok := body["data"]; !ok {
			t.Fatalf("expected data envelope, got %v", body)
		}
	})

	t.Run("returns 500 when DAO read fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/appointments?trainer_id=1&starts_at=2026-04-13T08:00:00-07:00&ends_at=2026-04-13T09:00:00-07:00", nil)
		rr := httptest.NewRecorder()
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int64) ([]appointments.AppointmentDL, error) {
				return nil, errors.New("dao failed")
			},
		}

		ListAppointments(dao)(rr, req, httprouter.Params{})
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
		}
		body := decodeJSONBody(t, rr)
		if _, ok := body["error"]; !ok {
			t.Fatalf("expected error envelope, got %v", body)
		}
	})
}

func TestCreateAppointments(t *testing.T) {
	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/appointments", strings.NewReader("{bad json"))
		rr := httptest.NewRecorder()

		CreateAppointments(&stubAppointmentDAO{})(rr, req, httprouter.Params{})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns 400 for invalid starts_at", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/appointments", strings.NewReader(`{"trainer_id":1,"user_id":2,"starts_at":"bad","ends_at":"2026-04-13T09:30:00-07:00"}`))
		rr := httptest.NewRecorder()

		CreateAppointments(&stubAppointmentDAO{})(rr, req, httprouter.Params{})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns 400 for invalid ends_at", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/appointments", strings.NewReader(`{"trainer_id":1,"user_id":2,"starts_at":"2026-04-13T09:00:00-07:00","ends_at":"bad"}`))
		rr := httptest.NewRecorder()

		CreateAppointments(&stubAppointmentDAO{})(rr, req, httprouter.Params{})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns 201 with created appointment envelope", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2026-04-13T09:00:00-07:00")
		req := httptest.NewRequest(http.MethodPost, "/appointments", strings.NewReader(`{"trainer_id":1,"user_id":2,"starts_at":"2026-04-13T09:00:00-07:00","ends_at":"2026-04-13T09:30:00-07:00"}`))
		rr := httptest.NewRecorder()
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int64) ([]appointments.AppointmentDL, error) {
				return nil, nil
			},
			createAppointmentFn: func(_ context.Context, apt appointments.AppointmentDL) (appointments.AppointmentDL, error) {
				return appointments.AppointmentDL{
					ID:        50,
					UserID:    apt.UserID,
					TrainerID: apt.TrainerID,
					Start:     apt.Start,
					End:       apt.End,
				}, nil
			},
		}

		CreateAppointments(dao)(rr, req, httprouter.Params{})
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, rr.Code)
		}
		body := decodeJSONBody(t, rr)
		data, ok := body["data"].([]any)
		if !ok || len(data) != 1 {
			t.Fatalf("expected single created appointment, got %v", body["data"])
		}
		row := data[0].(map[string]any)
		if row["trainer_id"] != float64(1) {
			t.Fatalf("unexpected trainer_id in response: %v", row["trainer_id"])
		}
		if row["started_at"] != start.UTC().Format(time.RFC3339) {
			t.Fatalf("unexpected started_at in response: %v", row["started_at"])
		}
	})

	t.Run("returns 500 on business-rule violation from domain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/appointments", strings.NewReader(`{"trainer_id":1,"user_id":2,"starts_at":"2026-04-12T09:00:00-07:00","ends_at":"2026-04-12T09:30:00-07:00"}`))
		rr := httptest.NewRecorder()

		CreateAppointments(&stubAppointmentDAO{})(rr, req, httprouter.Params{})
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
		}
		body := decodeJSONBody(t, rr)
		if _, ok := body["error"]; !ok {
			t.Fatalf("expected error envelope, got %v", body)
		}
	})
}

func TestListScheduledAppointments(t *testing.T) {
	t.Run("returns 400 for invalid trainer_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/appointments/schedules?trainer_id=bad", nil)
		rr := httptest.NewRecorder()

		ListScheduledAppointments(&stubAppointmentDAO{})(rr, req, httprouter.Params{})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns 200 with data envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/appointments/schedules?trainer_id=1", nil)
		rr := httptest.NewRecorder()
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int64) ([]appointments.AppointmentDL, error) {
				return []appointments.AppointmentDL{
					{ID: 1, UserID: 2, TrainerID: 1, Start: time.Now().UTC(), End: time.Now().UTC().Add(30 * time.Minute)},
				}, nil
			},
		}

		ListScheduledAppointments(dao)(rr, req, httprouter.Params{})
		if rr.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
		}
		body := decodeJSONBody(t, rr)
		if _, ok := body["data"]; !ok {
			t.Fatalf("expected data envelope, got %v", body)
		}
	})

	t.Run("returns 500 when DAO fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/appointments/schedules?trainer_id=1", nil)
		rr := httptest.NewRecorder()
		dao := &stubAppointmentDAO{
			getAppointmentsFn: func(context.Context, int64) ([]appointments.AppointmentDL, error) {
				return nil, errors.New("read failed")
			},
		}

		ListScheduledAppointments(dao)(rr, req, httprouter.Params{})
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
		}
	})
}
