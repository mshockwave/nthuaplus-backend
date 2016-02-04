package public

import "testing"

func TestFormalIdVerifier(t *testing.T) {
	id := Config.GetString("test.legalFormalId")
	if len(id) > 0 {
		t.Logf("Legal formal ID to test: %s\n", id)
		pass := FormalIdVerifier(id)
		if !pass { t.FailNow() }
	}else{
		t.Error("Required legal formal ID(test.legalFormalId) in config file")
	}
}
