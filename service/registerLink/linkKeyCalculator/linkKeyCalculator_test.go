package linkKeyCalculator

import "testing"

func TestLinkKeyCalculator_GetLinkKey_LinkKeyStabilityFor0(t *testing.T) {
	var linkKeyCalculator = New(123123123)

	res1 := linkKeyCalculator.GetLinkKey(0)
	res2 := linkKeyCalculator.GetLinkKey(0)

	if res1 != res2 {
		t.Errorf("res1 = %s | res2 = %s", res1, res2)
	}
}

func TestLinkKeyCalculator_GetLinkKey_LinkKeyStabilityFor100(t *testing.T) {
	var linkKeyCalculator = New(123123123)

	res1 := linkKeyCalculator.GetLinkKey(100)
	res2 := linkKeyCalculator.GetLinkKey(100)

	if res1 != res2 {
		t.Errorf("res1 = %s | res2 = %s", res1, res2)
	}
}

func TestLinkKeyCalculator_GetLinkKey_LinkKeyStabilityFor1000(t *testing.T) {
	var linkKeyCalculator = New(123123123)

	res1 := linkKeyCalculator.GetLinkKey(1000)
	res2 := linkKeyCalculator.GetLinkKey(1000)

	if res1 != res2 {
		t.Errorf("res1 = %s | res2 = %s", res1, res2)
	}
}

func TestLinkKeyCalculator_GetLinkKey_LinkKeyStabilityFor10000(t *testing.T) {
	var linkKeyCalculator = New(123123123)

	res1 := linkKeyCalculator.GetLinkKey(10000)
	res2 := linkKeyCalculator.GetLinkKey(10000)

	if res1 != res2 {
		t.Errorf("res1 = %s | res2 = %s", res1, res2)
	}
}

func TestLinkKeyCalculator_GetLinkKey_LinkKeyLength(t *testing.T) {
	var linkKeyCalculator = New(123123123)

	res1 := linkKeyCalculator.GetLinkKey(1000000)

	if len(res1) != 4 {
		t.Errorf("len(res1) =  %d, want 4", len(res1))
	}
}
