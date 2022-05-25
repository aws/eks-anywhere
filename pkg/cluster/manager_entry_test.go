package cluster_test

import (
	"testing"

	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestConfigManagerEntryMerge(t *testing.T) {
	g := NewWithT(t)
	kind1 := "kind1"
	kind2 := "kind2"
	kind3 := "kind3"
	generator := func() cluster.APIObject { return &anywherev1.Cluster{} }
	processor := func(*cluster.Config, cluster.ObjectLookup) {}
	validator := func(*cluster.Config) error { return nil }
	defaulter := func(*cluster.Config) error { return nil }

	c := cluster.NewConfigManagerEntry()

	c2 := cluster.NewConfigManagerEntry()
	g.Expect(c2.RegisterMapping(kind1, generator)).To(Succeed())
	g.Expect(c2.RegisterMapping(kind2, generator)).To(Succeed())
	c2.RegisterProcessors(processor)
	c2.RegisterDefaulters(defaulter)
	c2.RegisterValidations(validator)

	c3 := cluster.NewConfigManagerEntry()
	g.Expect(c3.RegisterMapping(kind3, generator)).To(Succeed())
	c2.RegisterProcessors(processor)
	c3.RegisterDefaulters(defaulter)
	c3.RegisterValidations(validator)

	g.Expect(c.Merge(c2, c3)).To(Succeed())

	g.Expect(len(c.APIObjectMapping)).To(Equal(3))
	g.Expect(c.APIObjectMapping[kind1]).To(Not(BeNil()))
	g.Expect(c.APIObjectMapping[kind2]).To(Not(BeNil()))
	g.Expect(c.APIObjectMapping[kind3]).To(Not(BeNil()))
	g.Expect(c.APIObjectMapping["kind4"]).To(BeNil())
	g.Expect(len(c.Processors)).To(Equal(2))
	g.Expect(len(c.Defaulters)).To(Equal(2))
	g.Expect(len(c.Validations)).To(Equal(2))
}

func TestConfigManagerEntryRegisterMappingError(t *testing.T) {
	g := NewWithT(t)
	kind1 := "kind1"
	generator := func() cluster.APIObject { return &anywherev1.Cluster{} }

	c := cluster.NewConfigManagerEntry()
	g.Expect(c.RegisterMapping(kind1, generator)).To(Succeed())
	g.Expect(c.RegisterMapping(kind1, generator)).To(MatchError(ContainSubstring("mapping for api object kind1 already registered")))
}
