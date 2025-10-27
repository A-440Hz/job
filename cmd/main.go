package main

/*
   The cmd directory handles entry points for the application.

   package main means this file will be compiled into a runnable executable
   to interface with other frameworks.
*/

import (
	"flag"
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
	"time"

	"gorm.io/gorm"
)

func main() {
	// parse flags to allow selecting production mode at startup
	production := flag.Bool("production", false, "use production (Railway) DB")
	flag.Parse()

	var dBase *gorm.DB
	var err error
	log.Println("WE ARE IN MAIN")

	if *production {
		// In production we expect DATABASE_URL (Railway) to be present.
		// The database service (managed by the platform) may not be ready
		// immediately when the container starts. Retry with backoff until
		// we can open a connection or hit a timeout.
		log.Println("WE ARE IN PRODUCTION")
		err := handler.SetEnvForProduction()
		if err != nil {
			panic(err)
		}
		const maxAttempts = 30
		const baseDelay = 2 // seconds
		var attempt int
		for attempt = 1; attempt <= maxAttempts; attempt++ {
			dBase, err = db.InitGormRailwayDB()
			if err == nil {
				break
			}
			log.Printf("DB connect attempt %d/%d failed: %v", attempt, maxAttempts, err)
			// If we've exhausted attempts, break and panic below
			if attempt == maxAttempts {
				break
			}
			// exponential backoff with jitter
			wait := time.Duration(baseDelay*(1<<uint(attempt-1))) * time.Second
			if wait > 30*time.Second {
				wait = 30 * time.Second
			}
			time.Sleep(wait)
		}
		if err != nil {
			// final failure after retries
			panic(err)
		}
	} else {
		// Local / test mode: populate local testing envs and use local DB
		log.Println("WE ARE IN LOCAL TESTING")
		db.SetEnvForTesting()
		dBase, err = db.InitGormLocalDB()
		if err != nil {
			panic(err)
		}
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
