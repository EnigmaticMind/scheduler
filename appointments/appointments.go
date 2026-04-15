package appointments

// This is the domain, business logic layer to handle calculations.  Allowing reuse of business logic.
import (
	"context"
	"slices"
	"time"
)

// Data object here represent how they exist in our go code, decoupling our data layer
// Allowing reuse of the domain layer, with changes to the data layer being limited

// ScheduledAppointment - This could be normalized with BookingAppointments and allow null if preferred
type ScheduledAppointment struct {
	ID        int       `json:"-" db:"-"`
	UserID    int       `json:"-" db:"-"`
	TrainerID int       `json:"-" db:"-"`
	End       time.Time `json:"-" db:"-"`
	Start     time.Time `json:"-" db:"-"`
}

// BookingAppointments - Appointments which are not yet scheduled
type BookingAppointments struct {
	TrainerID int       `json:"-" db:"-"`
	End       time.Time `json:"-" db:"-"`
	Start     time.Time `json:"-" db:"-"`
}

// This should probably accept GetScheduledAppointsments output instead,
// This will prevent downstream impact a bit
func overlapsAny(bookings []AppointmentDL, start, end time.Time) bool {
	// Not optimized due to time
	// Due to assumption of sorting you can bail earlier or otherwise optimize this
	for _, b := range bookings {
		if start.Before(b.End) && end.After(b.Start) {
			return true
		}
	}
	return false
}

// Round up to the next 30-minute boundary (:00 or :30).
// If t is already exactly on a boundary, it returns t unchanged.
func nextHalfHourSlot(t time.Time) time.Time {
	const slot = 30 * time.Minute
	// Work in Unix seconds to avoid float math.
	sec := t.Unix()
	slotSec := int64(slot / time.Second)
	rounded := ((sec + slotSec - 1) / slotSec) * slotSec
	return time.Unix(rounded, 0).In(t.Location())
}

// AvailableAppointments
func AvailableAppointments(ctx context.Context, dao AppointmentDAO, trainerID int, startsAt, endsAt time.Time) ([]BookingAppointments, error) {
	if endsAt.Before(startsAt) {
		return nil, ErrInvalidAptEndBeforeStart
	}

	// There's a `delay` in getting appointments and checking them for overlap
	// Accepting it for now
	booked, err := dao.GetAppointments(ctx, trainerID)
	if err != nil {
		return nil, err
	}

	var results []BookingAppointments
	for s := nextHalfHourSlot(startsAt); !s.After(endsAt); s = s.Add(time.Minute * 30) {
		e := s.Add(time.Minute * 30)
		withinBusiness, err := withinBusinessHours(s, e)
		if err != nil {
			// I regret adding loc into this function
			// Ignoring for now due to time
			return nil, err
		}
		if !withinBusiness {
			// more ideal, this should set s to next available time to limit loops
			// keeping this for now due to time
			continue
		}

		if overlapsAny(booked, s, e) {
			continue
		}

		// Make sure end is before our requested end time
		if e.After(endsAt) {
			continue
		}

		results = append(results, BookingAppointments{
			TrainerID: trainerID,
			// Convert back to UTC, HOW we should display our times should be up to the front end
			// I should probably enforce input to be UTC as well
			Start: s.UTC(),
			End:   e.UTC(),
		})
	}

	return results, nil
}

func withinBusinessHours(start, end time.Time) (bool, error) {
	// I should probably enforce everythig in UTC and do the business time compare in UTC values as well
	// Due to time keeping this mistake
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return false, err
	}

	// Convert to pacific time
	startLocal, endLocal := start.In(loc), end.In(loc)

	// Check for weekday occurances
	weekends := []time.Weekday{time.Saturday, time.Sunday}
	if slices.Contains(weekends, startLocal.Weekday()) ||
		slices.Contains(weekends, endLocal.Weekday()) {
		return false, nil
	}

	// Check for timeframe allowance
	openMins := 8 * 60
	closeMins := 17 * 60
	startMins := startLocal.Hour()*60 + startLocal.Minute()
	endMins := endLocal.Hour()*60 + endLocal.Minute()
	if startMins < openMins || endMins > closeMins {
		return false, nil
	}

	return true, nil
}

// CreateAppointment
func CreateAppointment(ctx context.Context, dao AppointmentDAO, userID, trainerID int, startsAt, endsAt time.Time) (ScheduledAppointment, error) {
	var results ScheduledAppointment

	if endsAt.Sub(startsAt).Minutes() != float64(30) {
		return results, ErrInvalidAptDur
	}

	halfHour := []int{0, 30}
	if !slices.Contains(halfHour, endsAt.Minute()) ||
		!slices.Contains(halfHour, startsAt.Minute()) {
		return results, ErrInvalidAptStartEnd
	}

	withinBH, err := withinBusinessHours(startsAt, endsAt)
	if err != nil {
		return results, err
	}

	if !withinBH {
		return results, ErrInvalidAptBusinessHours
	}

	// Do this last because of race conditions, technically it can still exist and a sql constraint should be added
	booked, err := dao.GetAppointments(ctx, trainerID)
	if err != nil {
		return results, err
	}
	if overlapsAny(booked, startsAt, endsAt) {
		return results, ErrInvalidAptOverlap
	}

	aptDL, err := dao.CreateAppointment(ctx, AppointmentDL{UserID: userID, TrainerID: trainerID, Start: startsAt.UTC(), End: endsAt.UTC()})
	if err != nil {
		return results, err
	}

	return ScheduledAppointment{
		ID:        aptDL.ID,
		UserID:    aptDL.UserID,
		TrainerID: aptDL.TrainerID,
		End:       aptDL.End,
		Start:     aptDL.Start,
	}, nil
}

// GetScheduledAppointsments
func GetScheduledAppointsments(ctx context.Context, dao AppointmentDAO, trainerID int) ([]ScheduledAppointment, error) {
	aptsDL, err := dao.GetAppointments(ctx, trainerID)

	if err != nil {
		return []ScheduledAppointment{}, err
	}

	results := make([]ScheduledAppointment, 0, len(aptsDL))
	for _, apt := range aptsDL {
		results = append(results, ScheduledAppointment{
			ID:        apt.ID,
			UserID:    apt.UserID,
			TrainerID: apt.TrainerID,
			Start:     apt.Start,
			End:       apt.End,
		})
	}

	return results, nil
}
