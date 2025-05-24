package main

/*
   The cmd directory handles entry points for the application.

   package main means this file will be compiled into a runnable executable
   to interface with other frameworks.
*/

import (
	"job/internal/db"
	"job/internal/handler"
	"job/internal/scheduler"
	"job/internal/tracker"
	"job/internal/user"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	db.SetEnvForTesting()
	db, err := db.InitGormDB()
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&user.User{}, tracker.JobAppTracker{}, tracker.JobAppItem{})

	userService := user.NewService(user.NewRepository(db))
	trackerService := tracker.NewService(tracker.NewRepository(db), scheduler.NewScheduler())
	h := &handler.Handler{
		UserService:    userService,
		TrackerService: trackerService,
	}
	h.TrackerService.Start()

	http.HandleFunc("/careers", h.ServeJobAppTrackerMainPage)
	http.HandleFunc("/careers/edit", h.UpdateJobAppTrackerFields)
	http.HandleFunc("/careers/new-entry", h.CreateJobAppItem)
	http.HandleFunc("/careers/edit-entry", h.UpdateJobAppItemFields)
	http.HandleFunc("/careers/delete-entry", h.DeleteJobAppItem)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server running at http://localhost:8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down")
	h.TrackerService.Stop()
}
