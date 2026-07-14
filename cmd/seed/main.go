package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tuantu/oj-web/internal/config"
	"github.com/tuantu/oj-web/internal/database"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	queries := sqlcdb.New(db.Pool)

	log.Println("Starting database seed...")

	// 1. Create Admin User
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("toplearn@admin"), bcrypt.DefaultCost)
	admin, err := queries.CreateUser(ctx, sqlcdb.CreateUserParams{
		Username:     "toplearn",
		PasswordHash: string(hashedPassword),
		DisplayName:  "TopLearn",
		Role:         "admin",
	})
	if err != nil {
		log.Printf("Skip creating admin (already exists or error): %v", err)
	} else {
		log.Printf("Created admin user: %s", admin.Username)
	}

	// 2. Create A Sample Problem
	problem, err := queries.CreateProblem(ctx, sqlcdb.CreateProblemParams{
		Slug:            "a-plus-b",
		Title:           "A + B Problem",
		Category:        "Introductory",
		TimeLimitMs:     1000,
		MemoryLimitMb:   256,
		StatementMd:     "Calculate the sum of two integers A and B.",
		InputDesc:       "The only line of input contains two integers A and B.",
		OutputDesc:      "Print the sum A + B.",
		ConstraintsDesc: "-10^9 <= A, B <= 10^9",
		Examples:        []byte(`[{"input":"1 2","output":"3"}]`),
		CheckerType:     "diff",
		Tags:               []byte("[]"),
		TestcaseVisibility: "all",
		MirrorFrom:         "",
		EditorialContent:   "",
	})
	if err != nil {
		log.Printf("Skip creating problem: %v", err)
	} else {
		log.Printf("Created problem: %s", problem.Title)

		// Create Test Cases
		queries.CreateTestCase(ctx, sqlcdb.CreateTestCaseParams{
			ProblemID:      problem.ID,
			OrderIndex:     1,
			Input:          "1 2",
			ExpectedOutput: "3",
			IsSample:       true,
		})
		queries.CreateTestCase(ctx, sqlcdb.CreateTestCaseParams{
			ProblemID:      problem.ID,
			OrderIndex:     2,
			Input:          "100 -50",
			ExpectedOutput: "50",
			IsSample:       false,
		})
		log.Println("Created test cases for A + B Problem")
	}

	// 3. Create Judge Node
	node, err := queries.CreateJudgeNode(ctx, sqlcdb.CreateJudgeNodeParams{
		Name:            "local-judge-1",
		BaseUrl:         "http://localhost:8081",
		ApiKeyEncrypted: "",
		IsActive:        true,
		MaxConcurrent:   3,
	})
	if err != nil {
		log.Printf("Skip creating judge node: %v", err)
	} else {
		// Update health manually to make it healthy
		queries.UpdateJudgeNodeHealth(ctx, sqlcdb.UpdateJudgeNodeHealthParams{
			ID:                  node.ID,
			IsHealthy:           true,
			SupportedLanguages:  []byte(`{"cpp": true, "python": true}`),
			MaxConcurrent:       3,
			LastHealthCheckAt:   pgtype.Timestamptz{Time: node.CreatedAt.Time, Valid: true},
			ConsecutiveFailures: 0,
		})
		log.Printf("Created judge node: %s", node.Name)
	}

	log.Println("Seeding complete.")
}
