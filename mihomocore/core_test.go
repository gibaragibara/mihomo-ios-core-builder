package mihomocore

import "testing"

func TestValidateConfigAcceptsAnywhereContract(t *testing.T) {
	config := []byte(`
mixed-port: 7890
allow-lan: false
proxies: []
proxy-groups: []
rules:
  - MATCH,DIRECT
`)

	if err := validateConfig(config); err != nil {
		t.Fatalf("expected config to validate: %v", err)
	}
}

func TestValidateConfigRejectsConflictingOptions(t *testing.T) {
	tests := map[string]string{
		"mixed-port": "mixed-port: 7891\n",
		"allow-lan":  "mixed-port: 7890\nallow-lan: true\n",
		"tun":        "mixed-port: 7890\ntun:\n  enable: true\n",
		"dns":        "mixed-port: 7890\ndns:\n  enable: true\n",
	}

	for name, config := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateConfig([]byte(config)); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

