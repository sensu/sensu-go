package schedulerd

import (
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceSet(t *testing.T) {
	AinA := corev2.FixtureCheckConfig("A")
	AinA.Namespace = "A"
	BinA := corev2.FixtureCheckConfig("B")
	BinA.Namespace = "A"
	AinB := corev2.FixtureCheckConfig("A")
	AinB.Namespace = "B"
	BinB := corev2.FixtureCheckConfig("B")
	BinB.Namespace = "B"

	set := make(namespacedChecks)

	added, changed, removed := set.Update([]*corev2.CheckConfig{
		AinA, BinA,
	})
	assert.Equal(t, []*corev2.CheckConfig{AinA, BinA}, added, "expected to add checks to namespace A")
	assert.Equal(t, []*corev2.CheckConfig(nil), changed, "expected no changes")
	assert.Equal(t, []*corev2.CheckConfig(nil), removed, "expected no deletions")

	added, changed, removed = set.Update([]*corev2.CheckConfig{
		AinA, BinA,
	})
	assert.Equal(t, []*corev2.CheckConfig(nil), added, "expected no additions")
	assert.Equal(t, []*corev2.CheckConfig(nil), changed, "expected no changes")
	assert.Equal(t, []*corev2.CheckConfig(nil), removed, "expected no deletions")

	tmp := *AinA
	AinA_1 := &tmp
	AinA_1.Cron = "* * * *"
	added, changed, removed = set.Update([]*corev2.CheckConfig{
		AinA_1, BinA,
	})
	assert.Equal(t, []*corev2.CheckConfig(nil), added, "expected no additions")
	assert.Equal(t, []*corev2.CheckConfig{AinA}, changed, "expected check A in namespace A to have changed")
	assert.Equal(t, []*corev2.CheckConfig(nil), removed, "expected no deletions")

	added, changed, removed = set.Update([]*corev2.CheckConfig{
		AinB, BinA, BinB,
	})
	assert.Equal(t, []*corev2.CheckConfig{AinB, BinB}, added, "expected to add checks A and B in namespace B")
	assert.Equal(t, []*corev2.CheckConfig(nil), changed, "expected no changes")
	assert.Equal(t, []*corev2.CheckConfig{AinA_1}, removed, "expected check A in namespace A to be deleted")
}
