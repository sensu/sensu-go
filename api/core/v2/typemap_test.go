package v2

import "testing"

func TestResolveResource_GH1565(t *testing.T) {
	v1, err := ResolveResource("Check")
	if err != nil {
		t.Fatal(err)
	}
	v2, err := ResolveResource("Check")
	if err != nil {
		t.Fatal(err)
	}

	c1, c2 := v1.(*Check), v2.(*Check)

	if c1 == c2 {
		t.Fatal("pointer values should differ")
	}

	if &c1.Subscriptions == &c2.Subscriptions {
		t.Fatal("internal pointer values should differ")
	}
}
