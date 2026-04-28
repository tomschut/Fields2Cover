// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package journey builds the Journey envelope sibling carried by every
// journey-endpoint response. The pipeline edges are fixed and known
// from api-go/openapi.yaml; this package is a pure data builder with
// no I/O and no business logic.

package journey

import (
	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
)

// Step name constants — match the OpenAPI Journey.step_name enum.
const (
	StepParse     = "fields:parse"
	StepHeadlands = "headlands:generate"
	StepSwaths    = "swaths:generate"
	StepSort      = "swaths:sort"
	StepRoute     = "routes:generate"
	StepPlan      = "paths:plan"
)

const methodPOST api.JourneyLinkMethod = api.POST

// AfterParse builds the envelope returned by /fields:parse.
// Next  -> /headlands:generate
// Back  -> nil (entry point)
func AfterParse() api.Journey {
	return api.Journey{
		StepName: api.JourneyStepName(StepParse),
		Next: &api.JourneyLink{
			Uri:           "/headlands:generate",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/GenerateHeadlandsRequest",
		},
		Back: nil,
	}
}

// AfterHeadlands builds the envelope returned by /headlands:generate.
func AfterHeadlands() api.Journey {
	return api.Journey{
		StepName: api.JourneyStepName(StepHeadlands),
		Next: &api.JourneyLink{
			Uri:           "/swaths:generate",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/GenerateSwathsRequest",
		},
		Back: &api.JourneyLink{
			Uri:           "/fields:parse",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/ParseFieldRequest",
		},
	}
}

// AfterSwaths builds the envelope returned by /swaths:generate.
func AfterSwaths() api.Journey {
	return api.Journey{
		StepName: api.JourneyStepName(StepSwaths),
		Next: &api.JourneyLink{
			Uri:           "/swaths:sort",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/SortSwathsRequest",
		},
		Back: &api.JourneyLink{
			Uri:           "/headlands:generate",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/GenerateHeadlandsRequest",
		},
	}
}

// AfterSort builds the envelope returned by /swaths:sort.
func AfterSort() api.Journey {
	return api.Journey{
		StepName: api.JourneyStepName(StepSort),
		Next: &api.JourneyLink{
			Uri:           "/routes:generate",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/GenerateRouteRequest",
		},
		Back: &api.JourneyLink{
			Uri:           "/swaths:generate",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/GenerateSwathsRequest",
		},
	}
}

// AfterRoute builds the envelope returned by /routes:generate.
func AfterRoute() api.Journey {
	return api.Journey{
		StepName: api.JourneyStepName(StepRoute),
		Next: &api.JourneyLink{
			Uri:           "/paths:plan",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/PlanPathRequest",
		},
		Back: &api.JourneyLink{
			Uri:           "/swaths:sort",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/SortSwathsRequest",
		},
	}
}

// AfterPlan builds the envelope returned by /paths:plan (terminal step).
// Next is nil — clients have walked the full pipeline.
func AfterPlan() api.Journey {
	return api.Journey{
		StepName: api.JourneyStepName(StepPlan),
		Next:     nil,
		Back: &api.JourneyLink{
			Uri:           "/routes:generate",
			Method:        methodPOST,
			BodySchemaRef: "#/components/schemas/GenerateRouteRequest",
		},
	}
}
