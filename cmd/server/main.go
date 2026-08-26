package main

import (
	"flag"
	"fmt"
	"mural-biocare/internal/application"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/httpapi"
	"mural-biocare/internal/persistence"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:19081", "listen address")
	self := flag.Bool("self-check", false, "run self check")
	retestHours := flag.Int("baseline-retest-hours", 24, "环境复测最小间隔小时数")
	flag.Parse()
	if *addr == "127.0.0.1:19081" {
		if p := os.Getenv("PORT"); p != "" {
			*addr = "127.0.0.1:" + p
		}
	}
	if *self {
		if err := selfCheck(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("自检通过")
		return
	}
	if !validListen(*addr) {
		fmt.Fprintln(os.Stderr, "地址必须绑定 127.0.0.1 且端口不低于 1024")
		os.Exit(2)
	}
	dir := filepath.Join(os.TempDir(), "mural-biocare-data")
	st, err := persistence.New(dir)
	if err != nil {
		panic(err)
	}
	app := application.New(st)
	if *retestHours <= 0 {
		fmt.Fprintln(os.Stderr, "环境复测间隔必须为正数")
		os.Exit(2)
	}
	app.BaselineRetestInterval = time.Duration(*retestHours) * time.Hour
	srv := &http.Server{Addr: *addr, Handler: httpapi.New(app).Handler(), ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	fmt.Println("服务监听", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func validListen(addr string) bool {
	a, err := net.ResolveTCPAddr("tcp", addr)
	return err == nil && a.IP.IsLoopback() && a.Port >= 1024
}
func selfCheck() error {
	dir, err := os.MkdirTemp("", "mural-self-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	st, _ := persistence.New(dir)
	app := application.New(st)
	c, err := app.Create("敦煌", "A-01", "菌斑", "owner", 20, 50, "")
	if err != nil {
		return err
	}
	rev := c.Revision
	c, err = app.Assessment(c.CaseID, structAssessment(), application.Command{Actor: "detector", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	rev = c.Revision
	c, err = app.Plan(c.CaseID, plan(), application.Command{Actor: "author", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	rev = c.Revision
	c, err = app.Review(c.CaseID, "APPROVE", "ok", application.Command{Actor: "reviewer", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	rev = c.Revision
	c, err = app.Pilot(c.CaseID, pilot(), application.Command{Actor: "author", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	rev = c.Revision
	c, err = app.Start(c.CaseID, application.Command{Actor: "team", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	rev = c.Revision
	c, err = app.Checkpoint(c.CaseID, checkpoint(), application.Command{Actor: "team", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	rev = c.Revision
	c, err = app.Complete(c.CaseID, application.Command{Actor: "team", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	rev = c.Revision
	c, err = app.Outcome(c.CaseID, outcome(), application.Command{Actor: "verifier", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	rev = c.Revision
	_, err = app.Archive(c.CaseID, application.Command{Actor: "owner", ExpectedRevision: rev})
	if err != nil {
		return err
	}
	return app.Verify(c.CaseID)
}
func structAssessment() domain.ContaminationAssessment {
	return domain.ContaminationAssessment{SamplePoints: []domain.SamplePoint{{ID: "s1", Location: "A-01", Result: "positive", CollectedAt: time.Now()}}, OrganismFindings: "fungi", ActivityLevel: "high", SpreadBoundary: "A-01", Method: "culture", AssessorID: "detector"}
}
func plan() domain.TreatmentPlan {
	return domain.TreatmentPlan{MaterialName: "biocide-X", CompatibilityBasis: "lab", ApplicationParameters: "1%", ProtectionMeasures: "PPE", RollbackConditions: "rinse", RequiredObservationDays: 7}
}
func pilot() domain.PilotTrial {
	return domain.PilotTrial{BeforeActivity: 10, AfterActivity: 2, AfterColorDelta: 1, ColorThreshold: 2, ObservationDays: 7}
}
func checkpoint() domain.ExecutionCheckpoint {
	return domain.ExecutionCheckpoint{Sequence: 1, Phase: "TREATMENT", ExpectedCondition: "stable", ObservedValue: "stable", Result: "PASS", RecordedBy: "team"}
}
func outcome() domain.OutcomeVerification {
	return domain.OutcomeVerification{PostActivity: 1, ColorDelta: 1, SurfaceStability: 0.9, ActivityThreshold: 2, ColorThreshold: 2, StabilityThreshold: 0.8, ObservationDays: 14}
}

var _ = strconv.Itoa
