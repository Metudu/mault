# Mault – Secure Secret Management CLI
Mault is a lightweight, Go-powered command-line application for securely storing and retrieving secrets.
It uses an encrypted vault backed by a database, ensuring your sensitive data is never stored in plain text.

## Features
- Encrypted storage for secrets using a vault system
- Initialization step to securely generate a salt and master key
- No accidental data leakage — sensitive DB queries are silenced
- Cross-platform — works anywhere Go binaries run
- Simple CLI interface for listing, adding, and retrieving secrets

## Installation
### Releases
Check out the releases tab and find the suitable binary for you. 
### From source
```sh
git clone https://github.com/Metudu/mault.git
cd mault
go build -o mault .
```

## Usage 
> [!WARNING]
> You need to initialize the mault first by running `mault init`

```
mault
  -> init     - initialize the mault by providing the master key
  -> create   - create a new secret
  -> generate - let application generate a secret for you
  -> list     - list all the secrets without revealing
  -> get      - decrypt and reveal a secret
  -> delete   - delete a secret
```

> [!NOTE]
> The commands don't expect any arguments, everything will be asked to you in order to hide the secrets from your console history.

Run `mault help` for more.

## License
See [MIT LICENSE](./LICENSE) for details.