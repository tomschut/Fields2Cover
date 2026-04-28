// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package journey

import (
	"testing"
)

func TestAfterParse(t *testing.T) {
	t.Parallel()
	j := AfterParse()
	if string(j.StepName) != StepParse {
		t.Errorf("StepName = %q, want %q", j.StepName, StepParse)
	}
	if j.Next == nil {
		t.Fatal("Next is nil, want non-nil")
	}
	if j.Next.Uri != "/headlands:generate" {
		t.Errorf("Next.Uri = %q, want /headlands:generate", j.Next.Uri)
	}
	if j.Next.BodySchemaRef != "#/components/schemas/GenerateHeadlandsRequest" {
		t.Errorf("Next.BodySchemaRef = %q", j.Next.BodySchemaRef)
	}
	if j.Back != nil {
		t.Errorf("Back = %v, want nil", j.Back)
	}
}

func TestAfterPlan_Terminal(t *testing.T) {
	t.Parallel()
	j := AfterPlan()
	if string(j.StepName) != StepPlan {
		t.Errorf("StepName = %q, want %q", j.StepName, StepPlan)
	}
	if j.Next != nil {
		t.Errorf("Next = %v, want nil (terminal step)", j.Next)
	}
	if j.Back == nil {
		t.Fatal("Back = nil, want non-nil")
	}
}

func TestAfterHeadlands_BothLinks(t *testing.T) {
	t.Parallel()
	j := AfterHeadlands()
	if j.Next == nil || j.Back == nil {
		t.Fatalf("expected both links non-nil, got Next=%v Back=%v", j.Next, j.Back)
	}
	if j.Next.Uri != "/swaths:generate" {
		t.Errorf("Next.Uri = %q, want /swaths:generate", j.Next.Uri)
	}
	if j.Back.Uri != "/fields:parse" {
		t.Errorf("Back.Uri = %q, want /fields:parse", j.Back.Uri)
	}
}

func TestAfterSwaths_AfterSort_AfterRoute(t *testing.T) {
	t.Parallel()

	sw := AfterSwaths()
	if sw.Next == nil || sw.Next.Uri != "/swaths:sort" {
		t.Errorf("AfterSwaths Next.Uri = %v", sw.Next)
	}
	if sw.Back == nil || sw.Back.Uri != "/headlands:generate" {
		t.Errorf("AfterSwaths Back = %v", sw.Back)
	}

	so := AfterSort()
	if so.Next == nil || so.Next.Uri != "/routes:generate" {
		t.Errorf("AfterSort Next = %v", so.Next)
	}
	if so.Back == nil || so.Back.Uri != "/swaths:generate" {
		t.Errorf("AfterSort Back = %v", so.Back)
	}

	ro := AfterRoute()
	if ro.Next == nil || ro.Next.Uri != "/paths:plan" {
		t.Errorf("AfterRoute Next = %v", ro.Next)
	}
	if ro.Back == nil || ro.Back.Uri != "/swaths:sort" {
		t.Errorf("AfterRoute Back = %v", ro.Back)
	}
}

func TestAllStepsHavePostMethod(t *testing.T) {
	t.Parallel()
	if AfterParse().Next.Method != methodPOST {
		t.Error("AfterParse Next not POST")
	}
	if AfterHeadlands().Next.Method != methodPOST {
		t.Error("AfterHeadlands Next not POST")
	}
	if AfterHeadlands().Back.Method != methodPOST {
		t.Error("AfterHeadlands Back not POST")
	}
	if AfterPlan().Back.Method != methodPOST {
		t.Error("AfterPlan Back not POST")
	}
}
