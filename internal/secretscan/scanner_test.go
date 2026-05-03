package secretscan

import "testing"

func TestScanEntriesPathRules(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		entryType string
		rule      string
	}{
		{name: "dotenv exact filename", path: ".env", rule: "dotenv file"},
		{name: "dotenv suffix filename", path: "services/api/.env.production", rule: "dotenv file"},
		{name: "pem glob", path: "deploy/prod.pem", rule: "private key/certificate file"},
		{name: "key glob", path: "keys/server.key", rule: "private key/certificate file"},
		{name: "ssh rsa filename", path: "id_rsa", rule: "SSH private key filename"},
		{name: "ssh ed25519 filename", path: "home/.ssh/id_ed25519", rule: "SSH private key filename"},
		{name: "secrets directory entry", path: "secrets", entryType: "tree", rule: "secrets directory"},
		{name: "secrets directory", path: "app/secrets/config.yml", rule: "secrets directory"},
		{name: "credentials exact filename", path: "credentials", rule: "credentials file"},
		{name: "credentials yml", path: "config/credentials.yml", rule: "credentials file"},
		{name: "p12 glob", path: "certs/app.p12", rule: "PKCS12 certificate bundle"},
		{name: "pfx glob", path: "certs/app.pfx", rule: "PKCS12 certificate bundle"},
		{name: "kubeconfig exact filename", path: "clusters/prod/kubeconfig", rule: "kubeconfig file"},
		{name: "npmrc exact filename", path: ".npmrc", rule: "npm credentials file"},
		{name: "pypirc exact filename", path: "release/.pypirc", rule: "PyPI credentials file"},
		{name: "terraform tfvars exact filename", path: "infra/terraform.tfvars", rule: "Terraform variable secrets file"},
	}

	scanner := New(Options{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entryType := tt.entryType
			if entryType == "" {
				entryType = "blob"
			}
			findings := scanner.ScanEntry(Entry{Path: tt.path, Type: entryType, Size: 128})
			if !hasFinding(findings, tt.rule) {
				t.Fatalf("expected rule %q for %q, got %#v", tt.rule, tt.path, findings)
			}
		})
	}
}

func TestScanContentRules(t *testing.T) {
	tests := []struct {
		name string
		data string
		rule string
	}{
		{name: "AWS access key", data: "AWS_ACCESS_KEY_ID=AKIA1234567890ABCDEF", rule: "AWS access key ID"},
		{name: "GitHub classic token prefix", data: "token=ghp_exampleplaceholder", rule: "GitHub token prefix"},
		{name: "GitHub fine grained token prefix", data: "token=github_pat_exampleplaceholder", rule: "GitHub token prefix"},
		{name: "GitHub app token prefix", data: "token=ghs_exampleplaceholder", rule: "GitHub token prefix"},
		{name: "private key header", data: "-----BEGIN OPENSSH PRIVATE KEY-----\nplaceholder", rule: "private key block header"},
		{name: "generic password assignment", data: "password=placeholder", rule: "generic token assignment"},
		{name: "generic secret assignment", data: "secret = placeholder", rule: "generic token assignment"},
		{name: "generic api key assignment", data: "api_key=placeholder", rule: "generic token assignment"},
		{name: "generic token assignment", data: "token = placeholder", rule: "generic token assignment"},
	}

	scanner := New(Options{ScanContent: true})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := scanner.ScanContent("config/settings.yml", []byte(tt.data))
			if !hasFinding(findings, tt.rule) {
				t.Fatalf("expected rule %q, got %#v", tt.rule, findings)
			}
		})
	}
}

func TestScanContentSkipsBinary(t *testing.T) {
	scanner := New(Options{ScanContent: true})

	findings := scanner.ScanContent("blob.dat", []byte{0xff, 0x00, 'A', 'K', 'I', 'A'})
	if len(findings) != 0 {
		t.Fatalf("expected binary content to be skipped, got %#v", findings)
	}
}

func TestShouldScanContent(t *testing.T) {
	scanner := New(Options{ScanContent: true, MaxContentBytes: 10})

	if !scanner.ShouldScanContent("config.yml", 10) {
		t.Fatal("expected small text-looking path to be eligible")
	}
	if scanner.ShouldScanContent("config.yml", 11) {
		t.Fatal("expected oversized file to be skipped")
	}
	if scanner.ShouldScanContent("image.png", 10) {
		t.Fatal("expected binary extension to be skipped")
	}
}

func TestDisabledScanner(t *testing.T) {
	scanner := New(Options{Disabled: true, ScanContent: true})

	if got := scanner.ScanEntries([]Entry{{Path: ".env", Type: "blob"}}); len(got) != 0 {
		t.Fatalf("expected disabled path scanner to return no findings, got %#v", got)
	}
	if got := scanner.ScanContent(".env", []byte("token=placeholder")); len(got) != 0 {
		t.Fatalf("expected disabled content scanner to return no findings, got %#v", got)
	}
	if scanner.ShouldScanContent(".env", 10) {
		t.Fatal("expected disabled scanner to reject content scanning")
	}
}

func hasFinding(findings []Finding, rule string) bool {
	for _, finding := range findings {
		if finding.Rule == rule {
			return true
		}
	}
	return false
}
