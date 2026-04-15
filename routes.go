package main

import (
	"net/http"
	"scheduler/appointments"
	"scheduler/handlers"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func newRouter(pool *pgxpool.Pool) http.Handler {
	aptDAO := appointments.NewDAO(pool)

	r := httprouter.New()
	r.GET("/appointments", handlers.ListAppointments(aptDAO))
	r.POST("/appointments", handlers.CreateAppointments(aptDAO))
	r.GET("/appointments/schedules", handlers.ListScheduledAppointments(aptDAO))
	return r
}
