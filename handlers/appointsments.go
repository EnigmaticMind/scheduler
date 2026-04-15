package handlers

// This layer exist to handle request + responses
import (
	"encoding/json"
	"net/http"
	"scheduler/appointsments"
	"scheduler/helpers/jsonwriter"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

type potentialAppointmentRes struct {
	TrainerID int64     `json:"trainer_id"`
	Start     time.Time `json:"started_at"`
	End       time.Time `json:"ended_at"`
}

type scheduledAppointmentRes struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TrainerID int64     `json:"trainer_id"`
	Start     time.Time `json:"started_at"`
	End       time.Time `json:"ended_at"`
}

type createAppointmentReq struct {
	TrainerID int64  `json:"trainer_id"`
	UserID    int64  `json:"user_id"`
	EndsAt    string `json:"ends_at"`
	StartsAt  string `json:"starts_at"`
}

// Map data object to response types so they remain uncoupled
// Should decouple `AppointmentDL` from the handlers, creating a domain specific definition of appointsments
func listAppointsmentsToRes(apts []appointsments.AppointmentDL) []potentialAppointmentRes {
	out := make([]potentialAppointmentRes, 0, len(apts))
	for _, a := range apts {
		out = append(out, potentialAppointmentRes{
			TrainerID: a.TrainerID,
			Start:     a.Start,
			End:       a.End,
		})
	}
	return out
}

func scheduledAppointmentsToRes(apts []appointsments.AppointmentDL) []scheduledAppointmentRes {
	out := make([]scheduledAppointmentRes, 0, len(apts))
	for _, a := range apts {
		out = append(out, scheduledAppointmentRes{
			ID:        a.ID,
			UserID:    a.UserID,
			TrainerID: a.TrainerID,
			Start:     a.Start,
			End:       a.End,
		})
	}
	return out
}

// ListAppointments
func ListAppointments(aptDAO appointsments.AppointmentDAO) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		q := r.URL.Query()
		trainerID, err := strconv.ParseInt(q.Get("trainer_id"), 10, 64)
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusBadRequest, "bad trainer_id", err)
			return
		}
		startsAt, err := time.Parse(time.RFC3339, q.Get("starts_at"))
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusBadRequest, "bad starts_at (use RFC3339)", err)
			return
		}
		endsAt, err := time.Parse(time.RFC3339, q.Get("ends_at"))
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusBadRequest, "bad ends_at (use RFC3339)", err)
			return
		}

		apts, err := appointsments.AvailableAppointments(r.Context(), aptDAO, trainerID, startsAt, endsAt)
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusInternalServerError, "failed to list appointments", err)
			return
		}

		jsonwriter.WriteJSONArray(w, http.StatusOK, listAppointsmentsToRes(apts))
	}
}

// CreateAppointments
func CreateAppointments(aptDAO appointsments.AppointmentDAO) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var body createAppointmentReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusBadRequest, "invalid JSON", err)
			return
		}
		startsAt, err := time.Parse(time.RFC3339, body.StartsAt)
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusBadRequest, "bad starts_at (use RFC3339)", err)
			return
		}
		endsAt, err := time.Parse(time.RFC3339, body.EndsAt)
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusBadRequest, "bad ends_at (use RFC3339)", err)
			return
		}

		created, err := appointsments.CreateAppointment(r.Context(), aptDAO, body.UserID, body.TrainerID, startsAt, endsAt)
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusInternalServerError, err.Error(), err)
			return
		}

		jsonwriter.WriteJSONArray(w, http.StatusCreated, scheduledAppointmentsToRes([]appointsments.AppointmentDL{created}))
	}
}

// ListScheduledAppointments
func ListScheduledAppointments(aptDAO appointsments.AppointmentDAO) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		q := r.URL.Query()
		trainerID, err := strconv.ParseInt(q.Get("trainer_id"), 10, 64)
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusBadRequest, "bad trainer_id", err)
			return
		}

		apts, err := aptDAO.GetAppointments(r.Context(), trainerID)
		if err != nil {
			jsonwriter.WriteJSONErr(w, http.StatusInternalServerError, "failed to list appointments", err)
			return
		}

		jsonwriter.WriteJSONArray(w, http.StatusOK, scheduledAppointmentsToRes(apts))
	}
}
