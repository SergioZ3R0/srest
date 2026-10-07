package ui

import (
	"strings"
	"testing"

	"github.com/SergioZ3R0/srest/internal/api"
)

func TestCategoriesExist(t *testing.T) {
	if len(categories) == 0 {
		t.Fatal("no categories defined")
	}
	for _, cat := range categories {
		if len(cat.endpoints) == 0 {
			t.Errorf("category %q has no endpoints", cat.name)
		}
	}
}

func TestEndpointsHaveValidBase(t *testing.T) {
	validBases := map[string]bool{"slurm": true, "slurmdb": true}
	for _, cat := range categories {
		for _, ep := range cat.endpoints {
			if !validBases[ep.base] {
				t.Errorf("endpoint %q has invalid base %q", ep.name, ep.base)
			}
		}
	}
}

func TestBuiltPathQueryParams(t *testing.T) {
	c := newComposer()
	c.setVersion(api.Version{Major: 0, Minor: 0, Micro: 45})

	// Select slurmdb category, then associations endpoint.
	c.selectCategory(1)
	for i, ep := range categories[1].endpoints {
		if ep.name == "associations" {
			c.selectEndpoint(i)
			break
		}
	}

	// Set query params.
	ep := c.currentEndpoint()
	for i := range ep.params {
		if ep.params[i].name == "account" {
			ep.params[i].value = "test-account"
		}
		if ep.params[i].name == "cluster" {
			ep.params[i].value = "linux"
		}
	}

	path := c.builtPath()
	if !strings.Contains(path, "/slurmdb/v0.0.45/associations") {
		t.Errorf("path missing slurmdb prefix: %s", path)
	}
	if !strings.Contains(path, "account=test-account") {
		t.Errorf("path missing account param: %s", path)
	}
	if !strings.Contains(path, "cluster=linux") {
		t.Errorf("path missing cluster param: %s", path)
	}
}

func TestBuiltPathPathParam(t *testing.T) {
	c := newComposer()
	c.setVersion(api.Version{Major: 0, Minor: 0, Micro: 45})

	// Select slurm category, then job detail endpoint.
	c.selectCategory(0)
	for i, ep := range categories[0].endpoints {
		if ep.name == "job detail" {
			c.selectEndpoint(i)
			break
		}
	}

	// Set path param.
	ep := c.currentEndpoint()
	for i := range ep.params {
		if ep.params[i].name == "job_id" {
			ep.params[i].value = "12345"
		}
	}

	path := c.builtPath()
	if path != "/job/12345" {
		t.Errorf("expected /job/12345, got %s", path)
	}
}

func TestBuiltPathSlurmdbPathParam(t *testing.T) {
	c := newComposer()
	c.setVersion(api.Version{Major: 0, Minor: 0, Micro: 45})

	// Select slurmdb category, then account detail endpoint.
	c.selectCategory(1)
	for i, ep := range categories[1].endpoints {
		if ep.name == "account detail" {
			c.selectEndpoint(i)
			break
		}
	}

	// Set path param.
	ep := c.currentEndpoint()
	for i := range ep.params {
		if ep.params[i].name == "account_name" {
			ep.params[i].value = "my-account"
		}
	}

	path := c.builtPath()
	if path != "/slurmdb/v0.0.45/account/my-account" {
		t.Errorf("expected /slurmdb/v0.0.45/account/my-account, got %s", path)
	}
}

func TestSelectCategory(t *testing.T) {
	c := newComposer()

	c.selectCategory(2) // actions
	if c.categoryIdx != 2 {
		t.Errorf("expected categoryIdx 2, got %d", c.categoryIdx)
	}
	if c.endpointIdx != 0 {
		t.Errorf("expected endpointIdx 0 after category change, got %d", c.endpointIdx)
	}

	c.selectCategory(99) // out of bounds
	if c.categoryIdx != 2 {
		t.Errorf("expected categoryIdx to stay at 2, got %d", c.categoryIdx)
	}
}

func TestSelectEndpoint(t *testing.T) {
	c := newComposer()

	c.selectCategory(1) // slurmdb
	eps := categories[1].endpoints
	c.selectEndpoint(3) // accounts
	if c.endpointIdx != 3 {
		t.Errorf("expected endpointIdx 3, got %d", c.endpointIdx)
	}
	_ = eps

	c.selectEndpoint(99) // out of bounds
	if c.endpointIdx != 3 {
		t.Errorf("expected endpointIdx to stay at 3, got %d", c.endpointIdx)
	}
}

func TestSetPartitionOptions(t *testing.T) {
	c := newComposer()
	names := []string{"cpu", "gpu", "debug"}
	c.setPartitionOptions(names)

	// Check that at least one endpoint has partition options.
	found := false
	for _, cat := range categories {
		for _, ep := range cat.endpoints {
			for _, p := range ep.params {
				if p.name == "partition" && len(p.options) == 3 {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("no endpoint has partition options set")
	}
}

func TestRenderSidebar(t *testing.T) {
	c := newComposer()
	sidebar := c.renderSidebar()

	if !strings.Contains(sidebar, "slurm") {
		t.Error("sidebar missing 'slurm' category")
	}
	if !strings.Contains(sidebar, "slurmdb") {
		t.Error("sidebar missing 'slurmdb' category")
	}
	if !strings.Contains(sidebar, "actions") {
		t.Error("sidebar missing 'actions' category")
	}
	if !strings.Contains(sidebar, "associations") {
		t.Error("sidebar missing 'associations' endpoint")
	}
}

func TestRenderBuilder(t *testing.T) {
	c := newComposer()
	c.setVersion(api.Version{Major: 0, Minor: 0, Micro: 45})
	c.selectCategory(1)
	for i, ep := range categories[1].endpoints {
		if ep.name == "associations" {
			c.selectEndpoint(i)
			break
		}
	}

	builder := c.renderBuilder()
	if !strings.Contains(builder, "GET") {
		t.Error("builder missing method")
	}
	if !strings.Contains(builder, "/slurmdb/v0.0.45/associations") {
		t.Error("builder missing path")
	}
}
