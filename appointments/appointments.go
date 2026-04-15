package appointments

// This is the domain, business logic layer to handle calculations.  Allowing reuse of business logic.
import (
	"context"
	"errors"
	"time"
)

// Using `AppointmentDL` for now but an appointments definition in the domain layer would protect from DB schema changes
type Appointments struct{}

func overlapsAny(bookings []AppointmentDL, start, end time.Time) bool {
	for _, b := range bookings {
		if start.Before(b.End) && end.After(b.Start) {
			return true
		}
	}
	return false
}

// AvailableAppointments - I haven't revised / validated this logic well
func AvailableAppointments(ctx context.Context, dao AppointmentDAO, trainerID int64, startsAt, endsAt time.Time) ([]AppointmentDL, error) {
	if !endsAt.After(startsAt) {
		return nil, nil
	}
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, err
	}
	booked, err := dao.GetAppointments(ctx, trainerID)
	if err != nil {
		return nil, err
	}
	var out []AppointmentDL
	startDay := startsAt.In(loc)
	endDay := endsAt.In(loc)
	firstDate := time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, loc)
	lastDate := time.Date(endDay.Year(), endDay.Month(), endDay.Day(), 0, 0, 0, 0, loc)
	for d := firstDate; !d.After(lastDate); d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		for _, minuteOfDay := range []int{
			8 * 60, 8*60 + 30, 9 * 60, 9*60 + 30, 10 * 60, 10*60 + 30,
			11 * 60, 11*60 + 30, 12 * 60, 12*60 + 30, 13 * 60, 13*60 + 30,
			14 * 60, 14*60 + 30, 15 * 60, 15*60 + 30, 16 * 60, 16*60 + 30,
		} {
			h, m := minuteOfDay/60, minuteOfDay%60
			slotStart := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, loc)
			slotEnd := slotStart.Add(30 * time.Minute)
			// Overlap with query window [startsAt, endsAt)
			if !slotStart.Before(endsAt) || !slotEnd.After(startsAt) {
				continue
			}
			if overlapsAny(booked, slotStart, slotEnd) {
				continue
			}
			out = append(out, AppointmentDL{
				TrainerID: trainerID,
				Start:     slotStart.UTC(),
				End:       slotEnd.UTC(),
			})
		}
	}
	return out, nil
}

// CreateAppointment - Again relying on AI for this logic due to time
func CreateAppointment(ctx context.Context, dao AppointmentDAO, userID, trainerID int64, startsAt, endsAt time.Time) (AppointmentDL, error) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return AppointmentDL{}, err
	}
	local := startsAt.In(loc)
	wd := local.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return AppointmentDL{}, errors.New("business hours are M-F 8am-5pm Pacific Time, appointment must match that")
	}
	m := local.Minute()
	if m != 0 && m != 30 || local.Second() != 0 || local.Nanosecond() != 0 {
		return AppointmentDL{}, errors.New("all appointments are 30 minutes long, and should be scheduled at :00, :30")
	}
	mins := local.Hour()*60 + local.Minute()
	if mins < 8*60 || mins > 16*60+30 {
		return AppointmentDL{}, errors.New("business hours are M-F 8am-5pm Pacific Time, appointment must match that")
	}
	end := startsAt.Add(30 * time.Minute)
	startUTC := startsAt.UTC()
	endUTC := end.UTC()
	// There's a gap here from get apts and checking overlap where you can double book, SQL constraint would fix that.
	booked, err := dao.GetAppointments(ctx, trainerID)
	if err != nil {
		return AppointmentDL{}, err
	}
	if overlapsAny(booked, startUTC, endUTC) {
		return AppointmentDL{}, errors.New("that time slot is already booked")
	}
	return dao.CreateAppointment(ctx, AppointmentDL{UserID: userID, TrainerID: trainerID, Start: startUTC, End: endUTC})
}
