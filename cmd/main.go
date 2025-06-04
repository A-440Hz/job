package main

/*
   The cmd directory handles entry points for the application.

   package main means this file will be compiled into a runnable executable
   to interface with other frameworks.
*/

import (
	"job/internal/collection"
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
	db, err := db.InitGormLocalDB()
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&user.User{}, tracker.JobAppTracker{}, tracker.JobAppItem{}, collection.Collectable{})

	collectionService := collection.NewService(collection.NewRepository(db))
	userService := user.NewService(user.NewRepository(db), collectionService)
	trackerService := tracker.NewService(tracker.NewRepository(db), scheduler.NewScheduler(), collectionService)
	h := &handler.Handler{
		UserService:       userService,
		TrackerService:    trackerService,
		CollectionService: collectionService,
	}
	h.TrackerService.Start()

	http.HandleFunc("/careers", h.ServeMainPage)
	http.HandleFunc("/careers/test", h.SelectEverything)

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
