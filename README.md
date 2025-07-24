<img alt="gopasslogo" src="/assets/gopass.png" height="120" width="auto">

# 🛡️ Gopass – Simple Encrypted CLI Vault

**Gopass** is a lightweight command-line tool written in Go for securely storing, encrypting, and retrieving secrets. It uses AES-GCM encryption with password-derived keys and stores passwords in your system keyring for convenience.

Perfect for developers or sysadmins who need fast, local, password-protected secret storage.

> **DISCLAIMER:** I built this for fun, as a way to teach myself Go, and it is by no means a fully functional product. The code is quite trashy and bugs might appear. Use at own risk!

---

## Features
- 🔐 AES-256 GCM encryption
- 💾 Vault stored as a single encrypted file
- 📁 Export/import vaults easily (flattened key-value pair, JSON format)
- 🔑 Passwords stored securely in keyring (per vault)
- ❌ Clears cached password when switching vaults
- 🧠 Caches last used vault via `~/.gopassrc` config

---

## More info
- Vaults are standalone encrypted files — you can copy or move them freely as long as you remember the password.
- The password is stored in your keyring and retrieved automatically unless you remove it.
- Wrong password? It will detect decryption failure and re-prompt cleanly.

---

## Getting Started

### First-time setup
```bash
cd gopass
go build ./cmd/gopass
cp gopass ~/go/bin # or wherever you have your Go binaries!
```

```bash
gopass -config /path/to/vault.dat
```

> This will prompt for a password. If the file doesn't exist, it will be created.

---

## Commands
```bash
gopass vault
```
> Display currently loaded vault

```bash
gopass list
gopass list -expose
```
> List all stored keys (secret values hidden by default), use '-expose' flag to reveal secrets.

```bash
gopass add <key> <value>
```
> Add a new key–value pair to the vault

```bash
gopass get <key>
```
> Retrieve a stored value by key, the value is copied to clipboard automatically.

```bash
gopass export <filename> # filename example: 'workvault.json'
```
> Export vault contents to JSON file

```bash
gopass import <filename>
```
> Import entries from JSON file

```bash
gopass help
```
> Show usage information

---

## Switching Vaults
To switch between vaults:
```bash
gopass -config /path/to/another.dat
```

> Automatically clears the previously cached password to avoid conflicts.

---

## Requirements
- Go 1.22+
- Linux/macOS (keyring support)
- `libsecret` (Arch Linux distros):  
  ```bash
  sudo pacman -S libsecret
  ```
---

## License
MIT © prozod 2025

<img alt="gopass-ss1" src="/assets/ss1.png" height="500" width="auto">
<img alt="gopass-ss2" src="/assets/ss2.png" height="400" width="auto">
<img alt="gopass-ss3" src="/assets/ss3.png" height="400" width="auto">
