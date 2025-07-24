package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/prozod/gopass/internal/vault"
	"github.com/zalando/go-keyring"
)

func TestVaultLoadWithStaticPassword(t *testing.T) {
	vaultPath := "testvault.dat"
	defer os.Remove(vaultPath)

	v := &vault.Vault{Entries: map[string]string{"gmail": "pass123"}}
	reader := vault.StaticPasswordReader{Password: "test123"}
	_ = keyring.Set("gopass", "vault:"+vaultPath, "test123")
	_ = v.Save(vaultPath)

	loaded, err := vault.LoadWithReader(vaultPath, reader)
	if err != nil {
		t.Fatalf("failed to load vault: %v", err)
	}

	if loaded.Entries["gmail"] != "pass123" {
		t.Fatalf("expected value/password not found")
	}
}

func TestVaultExportJSON_Success(t *testing.T) {
	v := &vault.Vault{
		Entries: map[string]string{
			"email":  "abc@example.com",
			"github": "token123",
		},
	}
	tempFile := "export_test.json"
	defer os.Remove(tempFile)

	err := v.Export(tempFile)
	if err != nil {
		t.Fatalf("expected no error on export, got %v", err)
	}

	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	if !bytes.Contains(data, []byte("email")) || !bytes.Contains(data, []byte("abc@example.com")) {
		t.Fatal("exported JSON does not contain expected data")
	}
}

func TestVaultExportJSON_InvalidPath(t *testing.T) {
	v := &vault.Vault{
		Entries: map[string]string{"foo": "bar"},
	}

	err := v.Export("/invalid/path/export.json")
	fmt.Println("ERR: ", err)
	if err == nil {
		t.Fatal("expected error due to invalid export path")
	}
}

func TestVaultLoad_CorruptedFile(t *testing.T) {
	filename := "corrupted.dat"
	defer os.Remove(filename)

	err := os.WriteFile(filename, []byte("thisisnotvalidvaultdata"), 0o600)
	if err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	os.Setenv("GOPASS_TEST_PASSWORD", "wrongpass")
	defer os.Unsetenv("GOPASS_TEST_PASSWORD")

	_, err = vault.Load(filename)
	if err == nil {
		t.Fatal("expected error while loading corrupted file")
	}
}

func TestVaultOverwriteExistingKey(t *testing.T) {
	v := &vault.Vault{Entries: map[string]string{"gmail": "oldpass"}}
	err := v.Add("gmail", "newpass", "testvault.dat")
	if err == nil {
		t.Fatal("expected error when overwriting existing key (should skip it)")
	}

	if v.Entries["gmail"] != "oldpass" {
		t.Fatalf("expected 'oldpass' for gmail, got %s", v.Entries["gmail"])
	}
}

func TestVaultLoadWrongPassword(t *testing.T) {
	vaultPath := "secure.dat"
	defer os.Remove(vaultPath)

	v := &vault.Vault{Entries: map[string]string{"test": "123"}}
	_ = keyring.Set("gopass", "vault:"+vaultPath, "correctpass")
	_ = v.Save(vaultPath)

	_ = keyring.Delete("gopass", "vault:"+vaultPath)
	reader := vault.StaticPasswordReader{Password: "wrongpass"}
	_, err := vault.LoadWithReader(vaultPath, reader)
	if err == nil {
		t.Fatal("expected failure due to wrong password, got success")
	}
}

func TestLoadNonexistentVaultFile(t *testing.T) {
	_, err := vault.Load("thisfiledoesnotexist.dat")
	if err == nil {
		t.Fatal("expected error loading nonexistent vault file")
	}
}

func TestAddEmptyKeyOrValue(t *testing.T) {
	v := &vault.Vault{Entries: make(map[string]string)}
	dummyPath := "dummy.dat"
	_ = keyring.Set("gopass", "vault:"+dummyPath, "testpass")
	defer keyring.Delete("gopass", "vault:"+dummyPath)
	defer os.Remove(dummyPath)

	v.Add("", "somepass", dummyPath)
	if _, exists := v.Entries[""]; exists {
		t.Error("should not add empty key")
	}

	err := v.Add("somekey", "", dummyPath)
	if err == nil {
		t.Error("expected error when adding entry with empty value")
	}
}

func TestTwoVaultsSameNameDifferentPaths(t *testing.T) {
	os.Mkdir("dir1", 0o755)
	os.Mkdir("dir2", 0o755)
	defer os.RemoveAll("dir1")
	defer os.RemoveAll("dir2")

	v1Path := "dir1/work.dat"
	v2Path := "dir2/work.dat"

	v1 := &vault.Vault{Entries: map[string]string{"site1": "abc"}}
	v2 := &vault.Vault{Entries: map[string]string{"site2": "def"}}

	_ = keyring.Set("gopass", "vault:"+v1Path, "pass1")
	_ = keyring.Set("gopass", "vault:"+v2Path, "pass2")

	_ = v1.Save(v1Path)
	_ = v2.Save(v2Path)

	loaded1, _ := vault.LoadWithReader(v1Path, vault.StaticPasswordReader{Password: "pass1"})
	loaded2, _ := vault.LoadWithReader(v2Path, vault.StaticPasswordReader{Password: "pass2"})

	if _, ok := loaded1.Entries["site1"]; !ok {
		t.Fatal("vault1 did not load correctly")
	}
	if _, ok := loaded2.Entries["site2"]; !ok {
		t.Fatal("vault2 did not load correctly")
	}
}

func TestVaultImport_InvalidJSON(t *testing.T) {
	v := &vault.Vault{
		Entries: make(map[string]string),
	}

	invalidJSON := []byte(`{invalid-json:}`)

	err := v.Import(invalidJSON, "testvault.dat")
	if err == nil {
		t.Fatal("expected error due to invalid JSON, got nil")
	}
}

func TestVaultImport_NestedJSON(t *testing.T) {
	v := &vault.Vault{
		Entries: make(map[string]string),
	}

	nestedJSON := []byte(`{ "key1": { "nested": "value" } }`)

	err := v.Import(nestedJSON, "testvault.dat")
	if err == nil {
		t.Fatal("expected error for nested JSON, got nil")
	}
}
