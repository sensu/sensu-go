package etcdstore_test

import (
	"context"
	"fmt"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sirupsen/logrus"
)

func make100Processes() []*corev2.Process {
	result := make([]*corev2.Process, 0, 100)
	for i := 0; i < 100; i++ {
		result = append(result, &corev2.Process{Name: fmt.Sprintf("process-%d", i)})
	}
	return result
}

var fixtureV2Entity = &corev2.Entity{
	ObjectMeta: corev2.ObjectMeta{
		Namespace:   "default",
		Name:        "default",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	},
	EntityClass:   "host",
	User:          "agent1",
	Subscriptions: []string{"linux", corev2.GetEntitySubscription("default")},
	Redact: []string{
		"password",
	},
	System: corev2.System{
		Arch:           "amd64",
		OS:             "linux",
		Platform:       "Gentoo",
		PlatformFamily: "lol",
		Network: corev2.Network{
			Interfaces: []corev2.NetworkInterface{
				{
					Name: "eth0",
					MAC:  "return of the",
					Addresses: []string{
						"127.0.0.1",
					},
				},
			},
		},
		LibCType:      "glibc",
		VMSystem:      "kvm",
		VMRole:        "host",
		CloudProvider: "aws",
		FloatType:     "hard",
		Processes:     make100Processes(),
	},
	LastSeen:          12345,
	SensuAgentVersion: "0.0.1",
}

var fixtureV3EntityConfig = &corev3.EntityConfig{
	Metadata: &corev2.ObjectMeta{
		Namespace:   "default",
		Name:        "default",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	},
	EntityClass:   "host",
	User:          "agent1",
	Subscriptions: []string{"linux", corev2.GetEntitySubscription("default")},
	Redact: []string{
		"password",
	},
}

var fixtureV3EntityState = &corev3.EntityState{
	Metadata: &corev2.ObjectMeta{
		Namespace:   "default",
		Name:        "default",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	},
	System: corev2.System{
		Arch:           "amd64",
		OS:             "linux",
		Platform:       "Gentoo",
		PlatformFamily: "lol",
		Network: corev2.Network{
			Interfaces: []corev2.NetworkInterface{
				{
					Name: "eth0",
					MAC:  "return of the",
					Addresses: []string{
						"127.0.0.1",
					},
				},
			},
		},
		LibCType:      "glibc",
		VMSystem:      "kvm",
		VMRole:        "host",
		CloudProvider: "aws",
		FloatType:     "hard",
		Processes:     make100Processes(),
	},
	LastSeen:          12345,
	SensuAgentVersion: "0.0.1",
}

func BenchmarkCreateOrUpdateV2Entity(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithEtcdStore(b, func(s *etcdstore.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		entity := fixtureV2Entity
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				req := storev2.NewResourceRequestFromV2Resource(entity)
				wrapper, err := wrap.V2Resource(entity, wrap.CompressSnappy)
				if err != nil {
					b.Fatal(err)
				}
				if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkCreateOrUpdateV3EntityState(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithEtcdStore(b, func(s *etcdstore.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		eState := fixtureV3EntityState
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				req := storev2.NewResourceRequestFromResource(eState)
				wrapper, err := wrap.Resource(eState)
				if err != nil {
					b.Fatal(err)
				}
				if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkCreateOrUpdateV3EntityConfig(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithEtcdStore(b, func(s *etcdstore.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		eConfig := fixtureV3EntityConfig
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				req := storev2.NewResourceRequestFromResource(eConfig)
				wrapper, err := wrap.Resource(eConfig)
				if err != nil {
					b.Fatal(err)
				}
				if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkUpdateEntityV2(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithV1Store(b, func(s store.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		if err := s.CreateNamespace(ctx, ns); err != nil {
			b.Fatal(err)
		}
		entity := fixtureV2Entity
		ctx = store.NamespaceContext(ctx, "default")
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := s.UpdateEntity(ctx, entity); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkGetEntity(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithV1Store(b, func(s store.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		if err := s.CreateNamespace(ctx, ns); err != nil {
			b.Fatal(err)
		}
		entity := fixtureV2Entity
		ctx = store.NamespaceContext(ctx, "default")
		if err := s.UpdateEntity(ctx, entity); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if _, err := s.GetEntityByName(ctx, entity.Name); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkGetV3EntityState(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithEtcdStore(b, func(s *etcdstore.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		eState := fixtureV3EntityState
		req = storev2.NewResourceRequestFromResource(eState)
		wrapper, err = wrap.Resource(eState)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wrapper, err := s.Get(ctx, req)
				if err != nil {
					b.Fatal(err)
				}
				if _, err := wrapper.Unwrap(); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkGetV3EntityConfig(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithEtcdStore(b, func(s *etcdstore.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		eCfg := fixtureV3EntityConfig
		req = storev2.NewResourceRequestFromResource(eCfg)
		wrapper, err = wrap.Resource(eCfg)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				wrapper, err := s.Get(ctx, req)
				if err != nil {
					b.Fatal(err)
				}
				if _, err := wrapper.Unwrap(); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkListEntities(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithV1Store(b, func(s store.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		if err := s.CreateNamespace(ctx, ns); err != nil {
			b.Fatal(err)
		}
		entity := *fixtureV2Entity
		ctx = store.NamespaceContext(ctx, "default")
		eName := entity.Name
		for i := 0; i < 1000; i++ {
			entity.Name = fmt.Sprintf("%s-%d", eName, i)
			if err := s.UpdateEntity(ctx, &entity); err != nil {
				b.Fatal(err)
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				entities, err := s.GetEntities(ctx, &store.SelectionPredicate{})
				if err != nil {
					b.Fatal(err)
				}
				if got, want := len(entities), 1000; got != want {
					b.Fatalf("wrong number of entities: got %d, want %d", got, want)
				}
			}
		})
	})
}

func BenchmarkListV3EntityState(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithEtcdStore(b, func(s *etcdstore.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		for i := 0; i < 1000; i++ {
			eState := *fixtureV3EntityState
			eState.Metadata.Name = fmt.Sprintf("%s-%d", fixtureV3EntityState.Metadata.Name, i)
			req := storev2.NewResourceRequestFromResource(&eState)
			wrapper, err := wrap.Resource(&eState, wrap.CompressSnappy)
			if err != nil {
				b.Fatal(err)
			}
			if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
				b.Fatal(err)
			}
		}
		listReq := storev2.NewResourceRequestFromResource(fixtureV3EntityState)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				list, err := s.List(ctx, listReq, nil)
				if err != nil {
					b.Fatal(err)
				}
				if got, want := list.Len(), 1000; got != want {
					b.Fatalf("wrong number of entities: got %d, want %d", got, want)
				}
				entities := make([]corev3.EntityState, 0, 1000)
				if err := list.UnwrapInto(&entities); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkListV3EntityStateNoUnwrap(b *testing.B) {
	logrus.SetLevel(logrus.ErrorLevel)
	testWithEtcdStore(b, func(s *etcdstore.Store) {
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			b.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			b.Fatal(err)
		}
		for i := 0; i < 1000; i++ {
			eState := *fixtureV3EntityState
			eState.Metadata.Name = fmt.Sprintf("%s-%d", fixtureV3EntityState.Metadata.Name, i)
			req := storev2.NewResourceRequestFromResource(&eState)
			wrapper, err := wrap.Resource(&eState, wrap.CompressSnappy)
			if err != nil {
				b.Fatal(err)
			}
			if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
				b.Fatal(err)
			}
		}
		listReq := storev2.NewResourceRequestFromResource(fixtureV3EntityState)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				list, err := s.List(ctx, listReq, nil)
				if err != nil {
					b.Fatal(err)
				}
				if got, want := list.Len(), 1000; got != want {
					b.Fatalf("wrong number of entities: got %d, want %d", got, want)
				}
			}
		})
	})
}
