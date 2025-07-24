package vault

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	"github.com/zalando/go-keyring"
	"golang.org/x/term"

	"github.com/prozod/gopass/internal/common"
)

var lastCachedVault string

const (
	service          = "gopass"
	saltSize         = 16
	nonceSize        = 12
	keyLen           = 32 // 32bytes, 256bit AES key
	pbkdf2Iterations = 100_000
)

func deriveKey(password, salt []byte) ([]byte, error) {
	key := pbkdf2.Key(password, salt, pbkdf2Iterations, keyLen, sha256.New)
	return key, nil
}

func LoadWithReader(filepath string, reader PasswordReader) (*Vault, error) {
	data, err := os.ReadFile(filepath)
	keyID := "vault:" + filepath
	var password string

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(common.Blue + "Vault file not found. Creating new vault." + common.Reset)
			passBytes, err := reader.Read(common.Green + "Enter password for new vault: " + common.Reset)
			fmt.Println()
			if err != nil {
				return nil, fmt.Errorf("failed to read password: %v", err)
			}
			password = string(passBytes)
			if err := keyring.Set(service, keyID, password); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to store password in keyring: %v\n", err)
			}
			vault := &Vault{Entries: make(map[string]string)}
			if err := vault.Save(filepath); err != nil {
				return nil, fmt.Errorf("failed to save initial vault: %v", err)
			}
			return vault, nil
		}
		return nil, fmt.Errorf("failed to read vault file: %v", err)
	}

tryDecrypt:
	password, err = keyring.Get(service, keyID)
	if err != nil {
		fmt.Print("Enter password to decrypt vault: ")
		passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %v", err)
		}
		password = string(passBytes)
		_ = keyring.Set(service, keyID, password)
	}

	if len(data) < saltSize+nonceSize {
		return nil, fmt.Errorf("vault file is too short or corrupted")
	}
	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	key, err := deriveKey([]byte(password), salt)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %v", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Decryption failed. Possibly wrong password.")
		_ = keyring.Delete(service, keyID)

		fmt.Print("Enter password to decrypt vault: ")
		passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %v", err)
		}
		password = string(passBytes)
		_ = keyring.Set(service, keyID, password)
		goto tryDecrypt
	}

	var v Vault
	v.Entries = make(map[string]string)
	decoder := gob.NewDecoder(bytes.NewReader(plaintext))
	if err := decoder.Decode(&v.Entries); err != nil {
		return nil, fmt.Errorf("failed to decode vault data: %v", err)
	}

	return &v, nil
}

func Load(filepath string) (*Vault, error) {
	return LoadWithReader(filepath, TerminalPasswordReader{})
}

func (v *Vault) Save(filepath string) error {
	// pw from keyring
	keyID := "vault:" + filepath
	password, err := keyring.Get(service, keyID)
	if err != nil {
		return fmt.Errorf("no password found in keyring for %s: %v", filepath, err)
	}

	var salt []byte

	// if file exists, extract the salt for reuse
	if existingData, err := os.ReadFile(filepath); err == nil && len(existingData) >= saltSize {
		salt = existingData[:saltSize]
	} else {
		salt = make([]byte, saltSize)
		if _, err := rand.Read(salt); err != nil {
			return fmt.Errorf("salt error: %v", err)
		}
	}

	// encode entries to plaintext
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err = encoder.Encode(v.Entries)
	if err != nil {
		return fmt.Errorf("failed to encode vault: %v", err)
	}
	plaintext := buf.Bytes()

	key, err := deriveKey([]byte(password), salt)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(salt); err != nil {
		return err
	}
	if _, err := f.Write(nonce); err != nil {
		return err
	}
	if _, err := f.Write(ciphertext); err != nil {
		return err
	}

	return nil
}

func SaveVaultAccessToConfig(vaultPath string) error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not determine current user: %v", err)
	}

	configPath := filepath.Join(usr.HomeDir, ".gopassrc")
	content := fmt.Sprintf("vault=%s\n", vaultPath)

	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}
	return nil
}

func GetVaultPathFromConfig() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("cannot get current user from system")
	}

	configPath := filepath.Join(usr.HomeDir, ".gopassrc")
	file, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf("cannot open .gopassrc")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if after, ok := strings.CutPrefix(line, "vault="); ok {
			path := after
			return string(path), nil
		}
	}

	return "", fmt.Errorf("cannot get vault path from .gopassrc")
}

/* reset keyring on each vault change */
func GetLastVaultFilePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(usr.HomeDir, ".gopassrc_previous"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func SetLastVaultFilePath(vaultPath string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(usr.HomeDir, ".gopassrc_previous"), []byte(vaultPath), 0o600)
}

func ClearOldVaultPasswordIfNeeded(oldVault, newVault string) error {
	if oldVault != "" && oldVault != newVault {
		oldKeyID := "vault:" + oldVault
		fmt.Println(common.Yellow+"Clearing cached password for old vault:"+common.Reset, oldVault)
		err := keyring.Delete(service, oldKeyID)
		if err != nil && !errors.Is(err, keyring.ErrNotFound) {
			return fmt.Errorf("failed to clear old vault password: %w", err)
		}
	}
	return nil
}
