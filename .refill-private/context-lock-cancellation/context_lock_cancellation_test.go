package context_lock_cancellation_test

import (
	"context"
	"errors"
	"mural-biocare/internal/application"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
	"sync"
	"testing"
	"time"
)

type controlledContext struct {
	done    chan struct{}
	checked chan struct{}
	once    sync.Once
}

func newControlledContext() *controlledContext {
	return &controlledContext{done: make(chan struct{}), checked: make(chan struct{})}
}

func (c *controlledContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *controlledContext) Done() <-chan struct{}       { return c.done }
func (c *controlledContext) Value(any) any               { return nil }
func (c *controlledContext) Err() error {
	select {
	case <-c.done:
		return context.Canceled
	default:
		c.once.Do(func() { close(c.checked) })
		return nil
	}
}

func (c *controlledContext) cancel() { close(c.done) }

type mutationResult struct {
	err error
}

func TestContextCancellationInterruptsCaseLockWait(t *testing.T) {
	store, err := persistence.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := application.New(store)
	c, err := service.Create("一号窟", "东壁", "菌斑扩散", "owner-1", 20, 55, "")
	if err != nil {
		t.Fatal(err)
	}

	unlock := service.Locker.Lock(c.CaseID)
	var unlockOnce sync.Once
	release := func() { unlockOnce.Do(unlock) }
	defer release()

	ctx := newControlledContext()
	result := make(chan mutationResult, 1)
	go func() {
		_, mutateErr := service.CorrectProfile(c.CaseID, applicationProfileCorrection(), application.Command{
			Context:          ctx,
			Actor:            c.OwnerID,
			ExpectedRevision: c.Revision,
		})
		result <- mutationResult{err: mutateErr}
	}()

	<-ctx.checked
	ctx.cancel()
	select {
	case got := <-result:
		if !contextError(got.err) {
			t.Fatalf("取消后返回了非 context 错误: %v", got.err)
		}
	case <-time.After(time.Second):
		release()
		got := <-result
		persisted, _ := store.Get(c.CaseID)
		t.Fatalf("context 取消未中断个案锁等待: mutation_err=%v revision=%d", got.err, persisted.Revision)
	}
}

func applicationProfileCorrection() domain.ProfileCorrectionInput {
	section := "东壁下部"
	return domain.ProfileCorrectionInput{MuralSection: &section, Reason: "复核定位后修正区段"}
}

func contextError(err error) bool {
	return err != nil && errors.Is(err, context.Canceled)
}
