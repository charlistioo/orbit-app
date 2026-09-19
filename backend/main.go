package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"orbit/backend/internal/api"
	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// withCORS applies a permissive CORS policy to every route, and answers
// the browser's preflight OPTIONS request directly. This supports local
// web-based testing of the mobile client (see TASK-001's CORS incident)
// and is harmless for the native iOS/Android target, which isn't
// subject to browser CORS at all.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("%s environment variable is required", name)
	}
	return v
}

func main() {
	databaseURL := mustEnv("DATABASE_URL")
	googleClientID := mustEnv("GOOGLE_CLIENT_ID")
	sessionSecret := mustEnv("SESSION_SECRET")

	db, err := store.Open(databaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	users := store.NewUserStore(db)
	categories := store.NewCategoryStore(db)
	items := store.NewItemStore(db)
	plans := store.NewPlanStore(db)
	transactions := store.NewTransactionStore(db, plans)
	bank := store.NewConsumptionBankStore(db)
	mood := store.NewMoodStore(db)
	sessions := auth.NewSessionIssuer(sessionSecret)
	googleVerifier := auth.NewRealGoogleVerifier(googleClientID)

	authHandler := &api.AuthHandler{
		Verifier: googleVerifier,
		Users:    users,
		Sessions: sessions,
	}
	categoriesHandler := &api.CategoriesHandler{
		Categories: categories,
		Items:      items,
	}
	plansHandler := &api.PlansHandler{Plans: plans}
	transactionsHandler := &api.TransactionsHandler{
		Transactions: transactions,
		Categories:   categories,
		Items:        items,
		Bank:         bank,
	}
	bankHandler := &api.ConsumptionBankHandler{Bank: bank}
	moodHandler := &api.MoodHandler{Mood: mood}
	homeHandler := &api.HomeHandler{
		Plans:        plans,
		Transactions: transactions,
		Mood:         mood,
		Bank:         bank,
	}
	timeline := store.NewTimelineStore(db)
	timelineHandler := &api.TimelineHandler{Timeline: timeline}
	calendar := store.NewCalendarStore(db)
	calendarHandler := &api.CalendarHandler{Calendar: calendar}
	insights := store.NewInsightStore(db)
	insightHandler := &api.InsightHandler{Insights: insights}
	dailyReviews := store.NewDailyReviewStore(db)
	dailyReviewHandler := &api.DailyReviewHandler{Reviews: dailyReviews}
	profileHandler := &api.ProfileHandler{Users: users, Bank: bank}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", healthHandler)
	mux.HandleFunc("/api/v1/auth/google", authHandler.GoogleSignIn)
	mux.HandleFunc("/api/v1/me", auth.RequireAuth(sessions, authHandler.Me))
	mux.HandleFunc("GET /api/v1/categories", auth.RequireAuth(sessions, categoriesHandler.ListCategories))
	mux.HandleFunc("POST /api/v1/categories", auth.RequireAuth(sessions, categoriesHandler.CreateCategory))
	mux.HandleFunc("GET /api/v1/categories/{category_id}/items", auth.RequireAuth(sessions, categoriesHandler.ListItems))
	mux.HandleFunc("POST /api/v1/categories/{category_id}/items", auth.RequireAuth(sessions, categoriesHandler.CreateItem))
	mux.HandleFunc("POST /api/v1/plans", auth.RequireAuth(sessions, plansHandler.CreatePlan))
	mux.HandleFunc("GET /api/v1/plans/{date}", auth.RequireAuth(sessions, plansHandler.GetPlan))
	mux.HandleFunc("PATCH /api/v1/plans/{plan_id}/categories/{plan_category_id}", auth.RequireAuth(sessions, plansHandler.UpdatePlanCategory))
	mux.HandleFunc("POST /api/v1/transactions", auth.RequireAuth(sessions, transactionsHandler.CreateTransaction))
	mux.HandleFunc("GET /api/v1/transactions", auth.RequireAuth(sessions, timelineHandler.GetTransactions))
	mux.HandleFunc("GET /api/v1/consumption-bank", auth.RequireAuth(sessions, bankHandler.GetBank))
	mux.HandleFunc("POST /api/v1/consumption-bank/apply", auth.RequireAuth(sessions, bankHandler.ApplyBank))
	mux.HandleFunc("POST /api/v1/mood", auth.RequireAuth(sessions, moodHandler.CreateMood))
	mux.HandleFunc("GET /api/v1/home", auth.RequireAuth(sessions, homeHandler.GetHome))
	mux.HandleFunc("GET /api/v1/calendar", auth.RequireAuth(sessions, calendarHandler.GetCalendar))
	mux.HandleFunc("GET /api/v1/insights/daily", auth.RequireAuth(sessions, insightHandler.GetDaily))
	mux.HandleFunc("GET /api/v1/insights/weekly", auth.RequireAuth(sessions, insightHandler.GetWeekly))
	mux.HandleFunc("GET /api/v1/insights/monthly", auth.RequireAuth(sessions, insightHandler.GetMonthly))
	mux.HandleFunc("GET /api/v1/reminders/today", auth.RequireAuth(sessions, dailyReviewHandler.GetTodayReminder))
	mux.HandleFunc("POST /api/v1/daily-review", auth.RequireAuth(sessions, dailyReviewHandler.FinalizeDailyReview))
	mux.HandleFunc("GET /api/v1/me/profile", auth.RequireAuth(sessions, profileHandler.GetProfile))
	mux.HandleFunc("PATCH /api/v1/me/profile", auth.RequireAuth(sessions, profileHandler.UpdateProfile))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("ORBIT backend listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
