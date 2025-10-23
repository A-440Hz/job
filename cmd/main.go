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
	dBase, err := db.InitGormLocalDB()
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// use AutoMigrate instead of CleanDB to perserve models
	// db.CleanDB(*dBase, user.User{}, user.Session{}, collection.UserInventory{}, tracker.JobAppTracker{}, tracker.JobAppItem{}, collection.Collectable{})
	dBase.AutoMigrate(user.User{}, user.Session{}, collection.Collectable{}, collection.UserCollectable{}, collection.UserInventory{}, tracker.JobAppTracker{}, tracker.JobAppItem{})

	collectionService := collection.NewService(collection.NewRepository(dBase))
	userService := user.NewService(user.NewRepository(dBase), collectionService)
	trackerService := tracker.NewService(tracker.NewRepository(dBase), scheduler.NewScheduler(), collectionService)
	h := &handler.Handler{
		UserService:       userService,
		TrackerService:    trackerService,
		CollectionService: collectionService,
	}
	h.TrackerService.Start()
	h.UserService.StartCleanupCron()

	http.HandleFunc("/careers", h.ServeMainPage)
	http.HandleFunc("/careers/test", h.SelectEverything)
	http.HandleFunc("/careers/profile", h.ServeUserMainPage)
	http.HandleFunc("/careers/profile/open", h.HandleAwardCollectableRequest)
	http.HandleFunc("/careers/login", h.HandleLoginRequest)
	http.HandleFunc("/careers/logout", h.HandleLogoutRequest)
	http.HandleFunc("/careers/register", h.RegisterBaseUser)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server running at ", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down")
	h.TrackerService.Stop()
}
