package case_cache_race_test

import (
	"mural-biocare/internal/application"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
	"sync"
	"testing"
)

func TestCaseCacheSynchronizesReadAndInvalidation(t *testing.T) {
	store, err := persistence.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	c, err := app.Create("石窟现场", "东壁", "黑色菌斑", "owner-1", 20, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := app.Get(c.CaseID); !ok {
		t.Fatal("预热详情缓存失败")
	}

	start := make(chan struct{})
	errs := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-start
			app.Get(c.CaseID)
		}()
	}
	go func() {
		defer wg.Done()
		<-start
		nextSection := "东壁修正区"
		_, updateErr := app.CorrectProfile(c.CaseID, domain.ProfileCorrectionInput{
			MuralSection: &nextSection,
			Reason:       "现场复核定位修正",
		}, application.Command{Actor: "owner-1", ExpectedRevision: c.Revision})
		errs <- updateErr
	}()
	close(start)
	wg.Wait()
	if err := <-errs; err != nil {
		t.Fatal(err)
	}

	latest, ok := app.Get(c.CaseID)
	if !ok || latest.Revision != c.Revision+1 || latest.MuralSection != "东壁修正区" {
		t.Fatalf("写入后读取到陈旧详情: %+v", latest)
	}
}
