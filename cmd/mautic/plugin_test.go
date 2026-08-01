package main

import (
	"testing"

	"github.com/ideamans/go-llm-cli-kit/skillcheck"

	"github.com/ideamans/mautic-cli/cmd"
)

// TestPluginSkills is the whole enforcement mechanism for the distributed
// plugin: the version match, the Agent Skills frontmatter rules, and the
// discovery keywords that make an agent reach for the skill at all.
func TestPluginSkills(t *testing.T) {
	report := skillcheck.CheckDir("../../plugins/mautic-cli", skillcheck.Options{
		// PluginVersion, not the ldflags-injected version, which is "dev"
		// locally and would make this test pass for the wrong reason.
		Version: cmd.PluginVersion,

		// Terms a user is likely to say when they need this tool. Rewording a
		// skill description without updating this list is exactly the silent
		// discovery regression the test exists to catch.
		Keywords: []string{"mautic", "contact", "segment", "campaign", "email"},

		RequireInstallSkill: true,
	})

	for _, problem := range report.Problems {
		t.Error(problem)
	}
}
