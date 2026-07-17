package middleware

import "testing"

func TestCoresSupportsSameOriginWithoutCrossOriginConfiguration(t *testing.T) {
	t.Setenv("KUBEMANAGE_ALLOWED_ORIGINS", "")
	_ = Cores()
}
