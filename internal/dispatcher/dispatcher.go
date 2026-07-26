package dispatcher

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/tuantu/oj-web/internal/checkers"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
	"github.com/tuantu/oj-web/internal/judgepool"
)

var WakeupDispatcher = make(chan struct{}, 1)

type Dispatcher struct {
	Queries    *sqlcdb.Queries
	Ctx        context.Context
	MaxWorkers int
	jobs       chan sqlcdb.Submission
	wg         sync.WaitGroup
}

func NewDispatcher(queries *sqlcdb.Queries, ctx context.Context, maxWorkers int) *Dispatcher {
	return &Dispatcher{
		Queries:    queries,
		Ctx:        ctx,
		MaxWorkers: maxWorkers,
	}
}

func (d *Dispatcher) Start() {
	d.jobs = make(chan sqlcdb.Submission, d.MaxWorkers)

	for i := 0; i < d.MaxWorkers; i++ {
		d.wg.Add(1)
		go d.worker(i)
	}

	go d.producer()
}

func (d *Dispatcher) worker(id int) {
	defer d.wg.Done()
	log.Printf("Worker %d started", id)
	for sub := range d.jobs {
		d.evaluateSubmission(sub)
	}
	log.Printf("Worker %d stopped", id)
}

func (d *Dispatcher) producer() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Khi context cancelled, close channel jobs
	defer close(d.jobs)

	// Hàm hỗ trợ rút tất cả submission đang rảnh ra
	processJobs := func() {
		for {
			sub, err := d.Queries.DequeueSubmission(d.Ctx)
			if err != nil {
				// Hết bài nộp hoặc lỗi
				break
			}
			log.Printf("Dequeued submission #%d, dispatching...", sub.ID)
			select {
			case d.jobs <- sub:
			case <-d.Ctx.Done():
				// Trả lại db nếu shutdown giữa chừng
				d.Queries.RequeueSubmission(context.Background(), sub.ID)
				return
			}
		}
	}

	for {
		select {
		case <-d.Ctx.Done():
			log.Println("Dispatcher stopped, waiting for workers to finish...")
			d.wg.Wait()
			return
		case <-ticker.C:
			processJobs()
		case <-WakeupDispatcher:
			processJobs()
		}
	}
}

func (d *Dispatcher) evaluateSubmission(sub sqlcdb.Submission) {
	// 1. Fetch Problem
	problem, err := d.Queries.GetProblemByID(d.Ctx, sub.ProblemID)
	if err != nil {
		log.Printf("Failed to get problem for sub %d: %v", sub.ID, err)
		d.Queries.UpdateSubmissionFailed(d.Ctx, sqlcdb.UpdateSubmissionFailedParams{
			ID:            sub.ID,
			CompileOutput: "Internal Error: Problem not found",
		})
		return
	}

	// 2. Fetch Test Cases
	testCases, err := d.Queries.ListTestCasesByProblem(d.Ctx, problem.ID)
	if err != nil {
		log.Printf("Failed to get test cases for sub %d: %v", sub.ID, err)
		d.Queries.UpdateSubmissionFailed(d.Ctx, sqlcdb.UpdateSubmissionFailedParams{
			ID:            sub.ID,
			CompileOutput: "Internal Error: Test cases not found",
		})
		return
	}

	runAllTests := true
	if sub.ContestID != nil {
		contest, err := d.Queries.GetContestByID(d.Ctx, *sub.ContestID)
		if err == nil && contest.RankingType == "ICPC" {
			runAllTests = false
		}
	}

	checkerType := problem.CheckerType
	customCheckerCode := problem.CustomCheckerCode

	if checkerType != "diff" {
		// Both 'custom' and standard checkers (fcmp, wcmp...) go here
		encodedCode, err := checkers.PrepareCheckerSource(checkerType, customCheckerCode)
		if err != nil {
			log.Printf("Failed to prepare checker source for sub %d: %v", sub.ID, err)
			d.Queries.UpdateSubmissionFailed(d.Ctx, sqlcdb.UpdateSubmissionFailedParams{
				ID:            sub.ID,
				CompileOutput: "Internal Error: Checker source preparation failed",
			})
			return
		}
		checkerType = "custom"
		customCheckerCode = encodedCode
	}

	req := judgepool.JudgeRequest{
		Language:          sub.Language,
		SourceCode:        sub.SourceCode,
		TimeLimitMs:       problem.TimeLimitMs,
		MemoryLimitMb:     problem.MemoryLimitMb,
		CheckerType:       checkerType,
		CustomCheckerCode: customCheckerCode,
		RunAllTests:       runAllTests,
	}

	for _, tc := range testCases {
		req.TestCases = append(req.TestCases, judgepool.TestCase{
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
		})
	}

	// 3. Find available judge node and acquire it atomically
	node, err := d.Queries.AcquireJudgeNode(d.Ctx)
	if err != nil {
		// sql.ErrNoRows will be returned if no node is available
		log.Printf("No healthy judge nodes available for sub %d (active jobs maxed out)", sub.ID)
		// Requeue the submission to try again later
		d.Queries.RequeueSubmission(d.Ctx, sub.ID)
		
		// Optional: Add jitter to prevent thundering herd when many workers wake up
		jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
		time.Sleep(1500*time.Millisecond + jitter)
		return
	}

	// Mark as judging
	d.Queries.UpdateSubmissionJudging(d.Ctx, sqlcdb.UpdateSubmissionJudgingParams{
		ID:          sub.ID,
		JudgeNodeID: &node.ID,
	})
	
	// Ensure we release the job slot when done, regardless of outcome
	defer d.Queries.DecrementNodeActiveJobs(context.Background(), node.ID)

	// 4. Send to Judge Node
	client := judgepool.NewClient(node.BaseUrl, node.ApiKeyEncrypted)
	resp, err := client.SubmitJudge(d.Ctx, req)

	if err != nil {
		log.Printf("Judge request failed for sub %d: %v", sub.ID, err)
		d.Queries.UpdateSubmissionFailed(d.Ctx, sqlcdb.UpdateSubmissionFailedParams{
			ID:            sub.ID,
			CompileOutput: "Judge Node Connection Error: " + err.Error(),
		})
		return
	}

	// 5. Process results
	if resp.Verdict == "CE" {
		d.Queries.UpdateSubmissionResult(d.Ctx, sqlcdb.UpdateSubmissionResultParams{
			ID:            sub.ID,
			Verdict:       "CE",
			TimeMs:        0,
			MemoryKb:      0,
			CompileOutput: resp.CompileOutput,
		})
		return
	}

	if resp.Verdict == "SYSTEM_ERROR" {
		d.Queries.UpdateSubmissionResult(d.Ctx, sqlcdb.UpdateSubmissionResultParams{
			ID:            sub.ID,
			Verdict:       "SE",
			TimeMs:        0,
			MemoryKb:      0,
			CompileOutput: "System Error on Judge Node",
		})
		return
	}

	// Save individual test results and determine final verdict
	maxTime := 0
	maxMemory := 0

	for _, tr := range resp.TestResults {
		// Insert into submission_test_results
		d.Queries.CreateTestResult(d.Ctx, sqlcdb.CreateTestResultParams{
			SubmissionID: sub.ID,
			TestIndex:    int32(tr.TestIndex),
			Verdict:      tr.Verdict,
			TimeMs:       int32(tr.TimeMs),
			MemoryKb:     int32(tr.MemoryKb),
			VszKb:        int32(tr.VszKb),
			Stdout:       tr.Stdout,
			Stderr:       tr.Stderr,
		})

		if int(tr.TimeMs) > maxTime {
			maxTime = int(tr.TimeMs)
		}
		if int(tr.MemoryKb) > maxMemory {
			maxMemory = int(tr.MemoryKb)
		}
	}

	d.Queries.UpdateSubmissionResult(d.Ctx, sqlcdb.UpdateSubmissionResultParams{
		ID:            sub.ID,
		Verdict:       resp.Verdict,
		TimeMs:        int32(maxTime),
		MemoryKb:      int32(maxMemory),
		CompileOutput: resp.CompileOutput,
	})

	log.Printf("Finished sub %d with verdict %s", sub.ID, resp.Verdict)
}
