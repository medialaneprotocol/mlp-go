package mlp

import "testing"

func TestBWE(t *testing.T) {
	cfg := BweConfig{PaceFloorBps: 100000, RequestedFPS: 30, ViewerLive: true}
	e := NewBandwidthEstimator(cfg)
	est := e.Estimate()
	if est.PaceBps != 100000 {
		t.Fatal("pace")
	}
	e.SetRTT(100)
	if e.RTTMsValue() != 100 {
		t.Fatal("rtt")
	}
}
