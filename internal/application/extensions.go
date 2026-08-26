package application

import (
	"fmt"
	"mural-biocare/internal/audit"
	"mural-biocare/internal/domain"
	"sort"
	"strings"
)

type BlockingIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Ref     string `json:"ref,omitempty"`
}

type StageBlockingIssues struct {
	Stage  string          `json:"stage"`
	Issues []BlockingIssue `json:"issues"`
}

type ArchiveReadiness struct {
	Ready           bool                  `json:"ready"`
	CheckedRevision int                   `json:"checked_revision"`
	BlockingIssues  []StageBlockingIssues `json:"blocking_issues"`
}

func (r ArchiveReadiness) Summary() string {
	parts := []string{}
	for _, stage := range r.BlockingIssues {
		for _, issue := range stage.Issues {
			if issue.Ref != "" {
				parts = append(parts, issue.Code+":"+issue.Ref)
			} else {
				parts = append(parts, issue.Code)
			}
		}
	}
	return strings.Join(parts, ",")
}

func archiveReadiness(c *domain.TreatmentCase, events []audit.Event) ArchiveReadiness {
	r := ArchiveReadiness{CheckedRevision: c.Revision}
	group := map[string][]BlockingIssue{}
	add := func(stage, code, message, ref string) {
		group[stage] = append(group[stage], BlockingIssue{Code: code, Message: message, Ref: ref})
	}
	if c.Status != domain.StatusOutcomeVerified && c.Status != domain.StatusArchived {
		add("封存", "INVALID_STATUS", "个案尚未达到成效复验通过状态", string(c.Status))
	}
	if c.Assessment == nil {
		add("评估", "ASSESSMENT_MISSING", "缺少污染评估", "")
	}
	if c.Plan == nil || c.Plan.ReviewDecision != "APPROVE" {
		add("方案", "APPROVED_PLAN_MISSING", "缺少当前版本批准清单", "")
	}
	if c.Pilot == nil || !c.Pilot.Passed {
		add("试验", "PILOT_NOT_PASSED", "小区试验门禁未通过", "")
	}
	if len(c.Checkpoints) == 0 || (len(c.PlannedCheckpoints) > 0 && len(c.Checkpoints) < len(c.PlannedCheckpoints)) {
		add("执行", "CHECKPOINT_PLAN_INCOMPLETE", "锁定检查点计划未全部核销", "")
	}
	for _, d := range c.Deviations {
		if !d.Resolved {
			add("执行", "OPEN_DEVIATION", "存在开放偏差", d.ID)
		}
		if d.Resolved && len(d.EvidenceRefs) == 0 {
			add("执行", "EMPTY_EVIDENCE_REFERENCE", "已解决偏差缺少证据引用", d.ID)
		}
		for _, ref := range d.EvidenceRefs {
			if strings.TrimSpace(ref) == "" {
				add("执行", "EMPTY_EVIDENCE_REFERENCE", "偏差包含空证据引用", d.ID)
			}
		}
	}
	for _, item := range c.RectificationItems {
		if item.Status == "OPEN" {
			add("复验", "OPEN_RECTIFICATION_ITEM", "整改任务尚未关闭", item.ID)
		}
	}
	if c.Outcome == nil || !c.Outcome.Passed {
		add("复验", "OUTCOME_NOT_PASSED", "成效复验未通过", "")
	}
	if err := audit.Verify(events); err != nil {
		add("审计", "AUDIT_CHAIN_INVALID", "事件摘要链校验失败", err.Error())
	}
	if len(events) == 0 {
		add("审计", "AUDIT_CHAIN_EMPTY", "事件链为空", "")
	} else if events[len(events)-1].Revision != c.Revision {
		add("审计", "REVISION_DISCONTINUITY", "快照与事件修订号不一致", fmt.Sprintf("event=%d", events[len(events)-1].Revision))
	}
	order := []string{"建档", "评估", "方案", "试验", "执行", "复验", "审计", "封存"}
	for _, stage := range order {
		issues := group[stage]
		if len(issues) == 0 {
			continue
		}
		sort.SliceStable(issues, func(i, j int) bool {
			if issues[i].Code == issues[j].Code {
				return issues[i].Ref < issues[j].Ref
			}
			return issues[i].Code < issues[j].Code
		})
		r.BlockingIssues = append(r.BlockingIssues, StageBlockingIssues{Stage: stage, Issues: issues})
	}
	r.Ready = len(r.BlockingIssues) == 0
	return r
}

func (s *Service) ArchiveReadiness(id string) (ArchiveReadiness, error) {
	unlock := s.Locker.Lock(id)
	defer unlock()
	c, ok := s.Store.Get(id)
	if !ok {
		return ArchiveReadiness{}, ErrNotFound
	}
	return archiveReadiness(c, s.Store.Events(id)), nil
}

func (s *Service) AssessmentDiff(id string, fromVersion, toVersion int) (domain.AssessmentDiff, error) {
	if fromVersion <= 0 || toVersion <= 0 || fromVersion >= toVersion {
		return domain.AssessmentDiff{}, fmt.Errorf("%w: assessment version order", domain.ErrInvalid)
	}
	unlock := s.Locker.Lock(id)
	defer unlock()
	c, ok := s.Store.Get(id)
	if !ok {
		return domain.AssessmentDiff{}, ErrNotFound
	}
	from, ok := c.AssessmentVersion(fromVersion)
	if !ok {
		return domain.AssessmentDiff{}, ErrNotFound
	}
	to, ok := c.AssessmentVersion(toVersion)
	if !ok {
		return domain.AssessmentDiff{}, ErrNotFound
	}
	return domain.BuildAssessmentDiff(from, to), nil
}
