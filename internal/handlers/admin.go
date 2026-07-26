package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

// Admin Dashboard
func (env *Env) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_dashboard.html", nil)
}

func parseExamplesFromForm(r *http.Request) []byte {
	inputs := r.Form["example_input[]"]
	outputs := r.Form["example_output[]"]

	var examples []map[string]string
	for i := 0; i < len(inputs) && i < len(outputs); i++ {
		examples = append(examples, map[string]string{
			"input":  inputs[i],
			"output": outputs[i],
		})
	}
	if len(examples) == 0 {
		return []byte("[]")
	}
	bytes, err := json.Marshal(examples)
	if err != nil {
		return []byte("[]")
	}
	return bytes
}

// Problems
func (env *Env) AdminCreateProblemGetHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	var classes []sqlcdb.Class
	if user.Role == "admin" {
		classes, _ = env.Queries.ListClasses(r.Context())
	} else {
		userClasses, _ := env.Queries.GetUserClasses(r.Context(), user.ID)
		for _, c := range userClasses {
			if c.Role == "teacher" {
				classes = append(classes, sqlcdb.Class{
					ID:          c.ID,
					Name:        c.Name,
					Description: c.Description,
					ScheduleMd:  c.ScheduleMd,
					NoticeMd:    c.NoticeMd,
					CreatedAt:   c.CreatedAt,
				})
			}
		}
	}
	render(w, r, "admin_create_problem_new.html", map[string]interface{}{
		"Classes": classes,
	})
}

func (env *Env) AdminCreateProblemPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	slug := r.FormValue("code")
	title := r.FormValue("name")
	statement := r.FormValue("description")

	timeLimitSec, _ := strconv.ParseFloat(r.FormValue("time_limit"), 64)
	timeLimit := int(timeLimitSec * 1000)
	if timeLimit <= 0 {
		timeLimit = 1000
	}
	memoryLimitKb, _ := strconv.Atoi(r.FormValue("memory_limit"))
	memoryLimit := memoryLimitKb / 1024
	if memoryLimit <= 0 {
		memoryLimit = 256
	}

	if slug == "" || title == "" {
		render(w, r, "admin_create_problem_new.html", map[string]string{"Error": "Slug and Title are required"})
		return
	}

	// Check if slug exists
	_, err := env.Queries.GetProblemBySlug(r.Context(), slug)
	if err == nil {
		render(w, r, "admin_create_problem_new.html", map[string]string{"Error": "Slug already exists"})
		return
	}

	tags := r.FormValue("tags")
	if tags == "" {
		tags = "[]"
	}

	isHidden := r.FormValue("is_hidden") == "on"
	
	problem, err := env.Queries.CreateProblem(r.Context(), sqlcdb.CreateProblemParams{
		Slug:              slug,
		Title:             title,
		StatementMd:       statement,
		TimeLimitMs:       int32(timeLimit),
		MemoryLimitMb:     int32(memoryLimit),
		CheckerType:       "diff",
		Category:          "Uncategorized",
		InputDesc:         r.FormValue("input_desc"),
		OutputDesc:        r.FormValue("output_desc"),
		ConstraintsDesc:   r.FormValue("constraints_desc"),
		Examples:          parseExamplesFromForm(r),
		CustomCheckerCode: "",
		EditorialContent:  r.FormValue("editorial_content"),
		Tags:              []byte(tags),
		TestcaseVisibility: r.FormValue("testcase_visibility"),
		MirrorFrom:        r.FormValue("mirror_from"),
		IsHidden:          isHidden,
	})
	if err != nil {
		render(w, r, "admin_create_problem_new.html", map[string]string{"Error": "Failed to create problem: " + err.Error()})
		return
	}

	classIDStr := r.FormValue("class_id")
	if classIDStr != "" {
		if classID, err := strconv.ParseInt(classIDStr, 10, 64); err == nil {
			env.Queries.AddProblemToClass(r.Context(), sqlcdb.AddProblemToClassParams{
				ClassID:   classID,
				ProblemID: problem.ID,
			})
		}
	}

	http.Redirect(w, r, "/problems/"+slug, http.StatusSeeOther)
}

// Contests
func (env *Env) AdminCreateContestGetHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	var classes []sqlcdb.Class
	if user.Role == "admin" {
		classes, _ = env.Queries.ListClasses(r.Context())
	} else {
		userClasses, _ := env.Queries.GetUserClasses(r.Context(), user.ID)
		for _, c := range userClasses {
			if c.Role == "teacher" {
				classes = append(classes, sqlcdb.Class{
					ID:          c.ID,
					Name:        c.Name,
					Description: c.Description,
					ScheduleMd:  c.ScheduleMd,
					NoticeMd:    c.NoticeMd,
					CreatedAt:   c.CreatedAt,
				})
			}
		}
	}
	render(w, r, "admin_create_contest.html", map[string]interface{}{
		"Classes": classes,
	})
}

func (env *Env) AdminCreateContestPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("name")
	startAtStr := r.FormValue("start_time")
	endAtStr := r.FormValue("end_time")

	if title == "" || startAtStr == "" || endAtStr == "" {
		render(w, r, "admin_create_contest.html", map[string]interface{}{"Error": "All fields are required"})
		return
	}

	startAt, errStart := time.Parse("2006-01-02T15:04", startAtStr)
	endAt, errEnd := time.Parse("2006-01-02T15:04", endAtStr)

	if errStart != nil || errEnd != nil {
		render(w, r, "admin_create_contest.html", map[string]interface{}{"Error": "Invalid time format"})
		return
	}

	if endAt.Before(startAt) {
		render(w, r, "admin_create_contest.html", map[string]interface{}{"Error": "End time must be after start time"})
		return
	}

	rankingType := r.FormValue("ranking_type")
	if rankingType != "IOI" && rankingType != "ICPC" {
		rankingType = "IOI" // default
	}

	isHidden := r.FormValue("is_hidden") == "on"

	contest, err := env.Queries.CreateContest(r.Context(), sqlcdb.CreateContestParams{
		Title:       title,
		StartAt:     pgtype.Timestamptz{Time: startAt, Valid: true},
		EndAt:       pgtype.Timestamptz{Time: endAt, Valid: true},
		RankingType: rankingType,
		IsHidden:    isHidden,
	})
	if err != nil {
		render(w, r, "admin_create_contest.html", map[string]interface{}{"Error": "Failed to create contest: " + err.Error()})
		return
	}

	classIDStr := r.FormValue("class_id")
	if classIDStr != "" {
		if classID, err := strconv.ParseInt(classIDStr, 10, 64); err == nil {
			env.Queries.AddContestToClass(r.Context(), sqlcdb.AddContestToClassParams{
				ClassID:    classID,
				ContestID: contest.ID,
			})
		}
	}

	http.Redirect(w, r, "/contests", http.StatusSeeOther)
}

// Blogs
func (env *Env) AdminCreateBlogGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_create_blog.html", nil)
}

// Groups
func (env *Env) AdminCreateGroupGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_create_group.html", nil)
}

// Announcements
func (env *Env) AdminCreateAnnouncementGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_create_announcement.html", nil)
}

// Test cases
func (env *Env) AdminCreateTestGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_create_test.html", nil)
}

func (env *Env) AdminCreateTestPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	problemIDStr := r.FormValue("problem_id")
	problemID, err := strconv.ParseInt(problemIDStr, 10, 64)
	if err != nil {
		render(w, r, "admin_create_test.html", map[string]interface{}{"Error": "Invalid problem ID"})
		return
	}

	input := r.FormValue("input")
	expectedOutput := r.FormValue("expected_output")
	isSample := r.FormValue("is_sample") == "on"

	if input == "" || expectedOutput == "" {
		render(w, r, "admin_create_test.html", map[string]interface{}{
			"Error":     "Input and Expected Output are required",
			"ProblemID": problemID,
		})
		return
	}

	// Determine the order_index
	testCases, err := env.Queries.ListTestCasesByProblem(r.Context(), problemID)
	var nextOrderIndex int32 = 1
	if err == nil && len(testCases) > 0 {
		nextOrderIndex = testCases[len(testCases)-1].OrderIndex + 1
	}

	_, err = env.Queries.CreateTestCase(r.Context(), sqlcdb.CreateTestCaseParams{
		ProblemID:      problemID,
		OrderIndex:     nextOrderIndex,
		Input:          input,
		ExpectedOutput: expectedOutput,
		IsSample:       isSample,
	})

	if err != nil {
		render(w, r, "admin_create_test.html", map[string]interface{}{
			"Error":     "Failed to create test case: " + err.Error(),
			"ProblemID": problemID,
		})
		return
	}

	// Show success message and keep the problem ID filled in for easy addition of multiple test cases
	render(w, r, "admin_create_test.html", map[string]interface{}{
		"Success":   "Test case #" + strconv.Itoa(int(nextOrderIndex)) + " created successfully!",
		"ProblemID": problemID,
	})
}

// Judge Nodes Management
func (env *Env) AdminJudgesListHandler(w http.ResponseWriter, r *http.Request) {
	nodes, err := env.Queries.ListJudgeNodes(r.Context())
	if err != nil {
		http.Error(w, "Failed to load judge nodes", http.StatusInternalServerError)
		return
	}
	render(w, r, "admin_judges.html", map[string]interface{}{
		"Nodes": nodes,
	})
}

func (env *Env) AdminJudgeCreateGetHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_judge_form.html", nil)
}

func (env *Env) AdminJudgeCreatePostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	baseUrl := r.FormValue("base_url")
	apiKey := r.FormValue("api_key")
	isActive := r.FormValue("is_active") == "on"

	if name == "" || baseUrl == "" {
		render(w, r, "admin_judge_form.html", map[string]interface{}{
			"Error": "Name and Base URL are required.",
		})
		return
	}

	_, err := env.Queries.CreateJudgeNode(r.Context(), sqlcdb.CreateJudgeNodeParams{
		Name:            name,
		BaseUrl:         baseUrl,
		ApiKeyEncrypted: apiKey,
		IsActive:        isActive,
		MaxConcurrent:   3, // Default to 3
	})
	if err != nil {
		render(w, r, "admin_judge_form.html", map[string]interface{}{
			"Error": "Failed to create node: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/admin/judges", http.StatusSeeOther)
}

func (env *Env) AdminJudgeEditGetHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	node, err := env.Queries.GetJudgeNodeByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	render(w, r, "admin_judge_form.html", map[string]interface{}{
		"Node": node,
	})
}

func (env *Env) AdminJudgeEditPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	name := r.FormValue("name")
	baseUrl := r.FormValue("base_url")
	apiKey := r.FormValue("api_key")
	isActive := r.FormValue("is_active") == "on"

	node, _ := env.Queries.GetJudgeNodeByID(r.Context(), id)

	if name == "" || baseUrl == "" {
		render(w, r, "admin_judge_form.html", map[string]interface{}{
			"Error": "Name and Base URL are required.",
			"Node":  node,
		})
		return
	}

	err = env.Queries.UpdateJudgeNode(r.Context(), sqlcdb.UpdateJudgeNodeParams{
		ID:              id,
		Name:            name,
		BaseUrl:         baseUrl,
		ApiKeyEncrypted: apiKey,
		IsActive:        isActive,
		MaxConcurrent:   node.MaxConcurrent,
	})
	if err != nil {
		node, _ := env.Queries.GetJudgeNodeByID(r.Context(), id)
		render(w, r, "admin_judge_form.html", map[string]interface{}{
			"Error": "Failed to update node: " + err.Error(),
			"Node":  node,
		})
		return
	}

	http.Redirect(w, r, "/admin/judges", http.StatusSeeOther)
}

// Edit Problem
func (env *Env) AdminEditProblemGetHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	problem, err := env.Queries.GetProblemBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)
		return
	}

	// Create a wrapper struct to pass converted values to the template
	type ProblemView struct {
		Problem       sqlcdb.Problem
		TimeLimitSec  float64
		MemoryLimitKb int32
		Error         string
	}

	pv := ProblemView{
		Problem:       problem,
		TimeLimitSec:  float64(problem.TimeLimitMs) / 1000.0,
		MemoryLimitKb: problem.MemoryLimitMb * 1024,
	}

	render(w, r, "admin_edit_problem.html", pv)
}

func (env *Env) AdminEditProblemPostHandler(w http.ResponseWriter, r *http.Request) {
	oldSlug := chi.URLParam(r, "slug")
	problem, err := env.Queries.GetProblemBySlug(r.Context(), oldSlug)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)
		return
	}

	r.ParseForm()
	slug := r.FormValue("code")
	title := r.FormValue("name")
	statement := r.FormValue("description")

	timeLimitSec, _ := strconv.ParseFloat(r.FormValue("time_limit"), 64)
	timeLimitMs := int32(timeLimitSec * 1000)
	if timeLimitMs <= 0 {
		timeLimitMs = 1000
	}
	memoryLimitKb, _ := strconv.Atoi(r.FormValue("memory_limit"))
	memoryLimitMb := int32(memoryLimitKb / 1024)
	if memoryLimitMb <= 0 {
		memoryLimitMb = 256
	}

	if slug == "" || title == "" {
		render(w, r, "admin_edit_problem.html", map[string]interface{}{
			"Error":   "All fields are required",
			"Problem": problem,
		})
		return
	}

	tags := r.FormValue("tags")
	if tags == "" {
		tags = "[]"
	}

	err = env.Queries.UpdateProblem(r.Context(), sqlcdb.UpdateProblemParams{
		ID:                problem.ID,
		Slug:              slug,
		Title:             title,
		Category:          problem.Category,
		TimeLimitMs:       timeLimitMs,
		MemoryLimitMb:     memoryLimitMb,
		StatementMd:       statement,
		InputDesc:         r.FormValue("input_desc"),
		OutputDesc:        r.FormValue("output_desc"),
		ConstraintsDesc:   r.FormValue("constraints_desc"),
		Examples:          parseExamplesFromForm(r),
		CheckerType:       r.FormValue("checker_type"),
		CustomCheckerCode: r.FormValue("custom_checker_code"),
		EditorialContent:  r.FormValue("editorial_content"),
		Tags:              []byte(tags),
		TestcaseVisibility: r.FormValue("testcase_visibility"),
		MirrorFrom:        r.FormValue("mirror_from"),
	})

	if err != nil {
		render(w, r, "admin_edit_problem.html", map[string]interface{}{
			"Error":   "Failed to update problem: " + err.Error(),
			"Problem": problem,
		})
		return
	}

	http.Redirect(w, r, "/admin/problems/"+slug+"/edit", http.StatusSeeOther)
}

// Edit Test
func (env *Env) AdminEditTestGetHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	problem, err := env.Queries.GetProblemBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)
		return
	}

	testCases, _ := env.Queries.ListTestCasesByProblem(r.Context(), problem.ID)

	render(w, r, "admin_edit_test.html", map[string]interface{}{
		"Problem":   problem,
		"TestCases": testCases,
	})
}

func (env *Env) AdminEditTestPostHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	problem, err := env.Queries.GetProblemBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)
		return
	}

	// Max upload size 100MB
	err = r.ParseMultipartForm(100 << 20)
	if err != nil {
		render(w, r, "admin_edit_test.html", map[string]interface{}{
			"Problem": problem,
			"Error":   "Failed to parse form: " + err.Error(),
		})
		return
	}

	var successMsg string
	file, header, err := r.FormFile("problem-data-zipfile")
	if err != nil {
		// No ZIP file, handle manual test case saving
		successMsg, err = env.HandleManualTestSave(r.Context(), r, problem.ID)
	} else {
		defer file.Close()
		successMsg, err = env.HandleZipUpload(r.Context(), file, header.Size, problem.ID)
	}

	// Reload testcases to show on page
	testCases, _ := env.Queries.ListTestCasesByProblem(r.Context(), problem.ID)

	if err != nil {
		render(w, r, "admin_edit_test.html", map[string]interface{}{
			"Problem":   problem,
			"TestCases": testCases,
			"Error":     err.Error(),
		})
		return
	}

	render(w, r, "admin_edit_test.html", map[string]interface{}{
		"Problem":   problem,
		"TestCases": testCases,
		"Success":   successMsg,
	})
}

func (env *Env) AdminJudgeDeletePostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = env.Queries.DeleteJudgeNode(r.Context(), id)
	if err != nil {
		// Just redirect back to the list with an error query param or similar, but for simplicity we will just log and redirect.
		http.Redirect(w, r, "/admin/judges?error=true", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/judges", http.StatusSeeOther)
}

func (env *Env) HandleManualTestSave(ctx context.Context, r *http.Request, problemID int64) (string, error) {
	testIDs := r.PostForm["test_ids[]"]
	inputs := r.PostForm["test_inputs[]"]
	outputs := r.PostForm["test_outputs[]"]
	descriptions := r.PostForm["test_descriptions[]"]
	orders := r.PostForm["test_order[]"]
	pretests := r.PostForm["test_pretests[]"]
	deletes := r.PostForm["test_deletes[]"]
	
	tx, err := env.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	qtx := env.Queries.WithTx(tx)

	validIds := make(map[int64]bool)

	for i := 0; i < len(testIDs); i++ {
		// skip if index out of bounds
		if i >= len(inputs) || i >= len(outputs) {
			continue
		}

		// check if deleted
		if i < len(deletes) && deletes[i] == "1" {
			id, _ := strconv.ParseInt(testIDs[i], 10, 64)
			if id > 0 {
				qtx.DeleteTestCase(ctx, id)
			}
			continue
		}

		id, _ := strconv.ParseInt(testIDs[i], 10, 64)
		isSample := false
		if i < len(pretests) && pretests[i] == "1" {
			isSample = true
		}
		
		desc := ""
		if i < len(descriptions) {
			desc = descriptions[i]
		}
		
		orderIdx := i + 1
		if i < len(orders) {
			o, err := strconv.Atoi(orders[i])
			if err == nil && o > 0 {
				orderIdx = o
			}
		}

		cleanInput := strings.ReplaceAll(inputs[i], "\r\n", "\n")
		cleanOutput := strings.ReplaceAll(outputs[i], "\r\n", "\n")

		var dbDesc *string
		if desc != "" {
			dbDesc = &desc
		}

		if id > 0 {
			// Update existing
			existingTC, err := env.Queries.GetTestCase(ctx, id)
			if err == nil {
				if strings.HasSuffix(cleanInput, "... [Data too large, editing disabled]") {
					cleanInput = existingTC.Input
				}
				if strings.HasSuffix(cleanOutput, "... [Data too large, editing disabled]") {
					cleanOutput = existingTC.ExpectedOutput
				}
			}

			err = qtx.UpdateTestCase(ctx, sqlcdb.UpdateTestCaseParams{
				ID:             id,
				OrderIndex:     int32(orderIdx),
				Input:          cleanInput,
				ExpectedOutput: cleanOutput,
				IsSample:       isSample,
				Description:    dbDesc,
			})
			if err != nil {
				return "", fmt.Errorf("failed to update test case %d: %v", id, err)
			}
			validIds[id] = true
		} else {
			// create new
			tc, err := qtx.CreateTestCase(ctx, sqlcdb.CreateTestCaseParams{
				ProblemID:      problemID,
				OrderIndex:     int32(orderIdx),
				Input:          cleanInput,
				ExpectedOutput: cleanOutput,
				IsSample:       isSample,
				Description:    dbDesc,
			})
			if err != nil {
				return "", fmt.Errorf("failed to create test case: %v", err)
			}
			validIds[tc.ID] = true
		}
	}

	// Delete cases not in validIds
	oldCases, _ := qtx.ListTestCasesByProblem(ctx, problemID)
	for _, oc := range oldCases {
		if !validIds[oc.ID] {
			qtx.DeleteTestCase(ctx, oc.ID)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return "Lưu cấu hình Test Cases thành công!", nil
}
