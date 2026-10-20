package catalog_test

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	assetCodePattern    = regexp.MustCompile(`^[A-Z0-9]{2,10}$`)
	networkCodePattern  = regexp.MustCompile(`^[a-z0-9_]{2,32}$`)
	providerCodePattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,32}$`)
)

var knownProviders = map[string]struct{}{
	"coingate":    {},
	"nowpayments": {},
	"bitpay":      {},
	"btcpay":      {},
}

type providerCode struct {
	Asset   string `json:"asset"`
	Network string `json:"network"`
}

type network struct {
	Code          string                  `json:"code"`
	Name          string                  `json:"name"`
	Kind          string                  `json:"kind"`
	Chain         string                  `json:"chain"`
	Status        string                  `json:"status"`
	ProviderCodes map[string]providerCode `json:"provider_codes"`
}

type asset struct {
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Decimals int       `json:"decimals"`
	Status   string    `json:"status"`
	Networks []network `json:"networks"`
}

type catalog struct {
	SchemaVersion  int     `json:"schema_version"`
	CatalogVersion int     `json:"catalog_version"`
	UpdatedAt      string  `json:"updated_at"`
	Assets         []asset `json:"assets"`
}

func loadArtifact(t *testing.T) (catalog, []byte) {
	t.Helper()
	raw, err := os.ReadFile("assets.json")
	if err != nil {
		t.Fatalf("read the artifact: %v", err)
	}
	var parsed catalog
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil {
		t.Fatalf("decode the artifact: %v", err)
	}
	return parsed, raw
}

func TestSchemaFile(t *testing.T) {
	raw, err := os.ReadFile("assets.schema.json")
	if err != nil {
		t.Fatalf("read the schema: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("decode the schema: %v", err)
	}
	if draft, _ := schema["$schema"].(string); !strings.Contains(draft, "2020-12") {
		t.Fatalf("schema draft = %q, want 2020-12", draft)
	}
	if additional, ok := schema["additionalProperties"]; !ok || additional != false {
		t.Fatalf("the schema must close additional properties")
	}
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("the schema has no required list")
	}
	fields := make(map[string]bool, len(required))
	for _, field := range required {
		name, _ := field.(string)
		fields[name] = true
	}
	for _, expected := range []string{"schema_version", "catalog_version", "updated_at", "assets"} {
		if !fields[expected] {
			t.Fatalf("schema required list misses %q", expected)
		}
	}
}

func TestArtifactContract(t *testing.T) {
	parsed, raw := loadArtifact(t)

	if parsed.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", parsed.SchemaVersion)
	}
	if parsed.CatalogVersion < 1 {
		t.Fatalf("catalog_version = %d, want >= 1", parsed.CatalogVersion)
	}
	if _, err := time.Parse("2006-01-02", parsed.UpdatedAt); err != nil {
		t.Fatalf("updated_at %q is not a date", parsed.UpdatedAt)
	}
	if len(parsed.Assets) == 0 {
		t.Fatal("the artifact has no assets")
	}

	seenAssets := make(map[string]bool)
	seenNetworks := make(map[string]bool)
	providerAssets := make(map[string]string)
	providerNetworks := make(map[string]bool)
	seenProviders := make(map[string]bool)
	activeAssets := 0
	for _, item := range parsed.Assets {
		if !assetCodePattern.MatchString(item.Code) {
			t.Fatalf("invalid asset code %q", item.Code)
		}
		if seenAssets[item.Code] {
			t.Fatalf("duplicate asset code %q", item.Code)
		}
		seenAssets[item.Code] = true
		if item.Name == "" || strings.TrimSpace(item.Name) != item.Name || len(item.Name) > 100 {
			t.Fatalf("asset %q has invalid name %q", item.Code, item.Name)
		}
		if item.Decimals < 0 || item.Decimals > 18 {
			t.Fatalf("asset %q decimals %d out of range", item.Code, item.Decimals)
		}
		if item.Status != "active" && item.Status != "deprecated" {
			t.Fatalf("asset %q has unknown status %q", item.Code, item.Status)
		}
		if len(item.Networks) == 0 {
			t.Fatalf("asset %q has no networks", item.Code)
		}
		activeNetworks := 0
		for _, network := range item.Networks {
			if !networkCodePattern.MatchString(network.Code) {
				t.Fatalf("invalid network code %q", network.Code)
			}
			if seenNetworks[network.Code] {
				t.Fatalf("duplicate network code %q", network.Code)
			}
			seenNetworks[network.Code] = true
			if network.Name == "" || strings.TrimSpace(network.Name) != network.Name || len(network.Name) > 100 {
				t.Fatalf("network %q has invalid name %q", network.Code, network.Name)
			}
			switch network.Kind {
			case "native", "token", "layer":
			default:
				t.Fatalf("network %q has unknown kind %q", network.Code, network.Kind)
			}
			if !networkCodePattern.MatchString(network.Chain) {
				t.Fatalf("network %q has invalid chain %q", network.Code, network.Chain)
			}
			if network.Status != "active" && network.Status != "deprecated" {
				t.Fatalf("network %q has unknown status %q", network.Code, network.Status)
			}
			if network.Status == "active" {
				activeNetworks++
				if len(network.ProviderCodes) == 0 {
					t.Fatalf("active network %q has no provider codes", network.Code)
				}
			}
			for provider, code := range network.ProviderCodes {
				if _, known := knownProviders[provider]; !known {
					t.Fatalf("network %q maps unknown provider %q", network.Code, provider)
				}
				if !providerCodePattern.MatchString(code.Asset) {
					t.Fatalf("network %q provider %q has invalid asset code %q", network.Code, provider, code.Asset)
				}
				if code.Network != "" && !providerCodePattern.MatchString(code.Network) {
					t.Fatalf("network %q provider %q has invalid network code %q", network.Code, provider, code.Network)
				}
				if strings.TrimSpace(code.Asset) != code.Asset || strings.TrimSpace(code.Network) != code.Network {
					t.Fatalf("network %q provider %q has padded codes", network.Code, provider)
				}
				key := provider + "\x00" + code.Asset
				if existing, ok := providerAssets[key]; ok && existing != item.Code {
					t.Fatalf("network %q provider %q code %q maps both %q and %q", network.Code, provider, code.Asset, existing, item.Code)
				}
				providerAssets[key] = item.Code
				if code.Network != "" {
					networkKey := key + "\x00" + code.Network
					if providerNetworks[networkKey] {
						t.Fatalf("network %q duplicates provider %q code %q/%q", network.Code, provider, code.Asset, code.Network)
					}
					providerNetworks[networkKey] = true
				}
				seenProviders[provider] = true
			}
		}
		if item.Status == "active" {
			activeAssets++
			if activeNetworks == 0 {
				t.Fatalf("active asset %q has no active network", item.Code)
			}
		}
	}
	if activeAssets == 0 {
		t.Fatal("the artifact has no active asset")
	}
	for provider := range knownProviders {
		if !seenProviders[provider] {
			t.Fatalf("provider %q is not mapped by any network", provider)
		}
	}

	if bytes.Contains(raw, []byte{0xe2, 0x80, 0x94}) {
		t.Fatal("the artifact contains an em dash")
	}
	if bytes.Contains(raw, []byte{'\t'}) {
		t.Fatal("the artifact contains a tab")
	}
}
