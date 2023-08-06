# vaultsg

This is a proof-of-concept Vault KV secrets generator and manager.

`vaultsg` is used to quickly and declaratively define Vault KV secrets at the
desired paths. It supports static and generated secrets with specific conditions.

## Usage

```bash
# generate/add all secrets
$ vaultsg add all

# generate/add secret(s) at foo/bar/*
$ vaultsg add foo/bar

# generate plan to show dry run
$ vaultsg plan all

# list all secret paths
# this shows the secret version or nil if it has not been added
$ vaultsg list
foo/baz      v1
foo/bar/bat  v2
foo/bar/baf  nil

# list all secret paths foo/bar/*
$ vaultsg list foo/bar
```

A `secrets.yml` file is required to run `vaultsg add`

```yml
# secrets.yml
foo:
  baz:
    data:
      # generate a password of length 24 with special chars
      - key: password
        length: 24
        special: true
  bar:
    bat:
      data:
        # set the following static secrets
        - key: username
          secret: foo
        - key: password
          secret: password
```

The `secrets.yml` file is intended to be declarative. An `--overwrite` flag must
be passed to re-generate any secrets.

```bash
# overwrite the secret at foo/baz with a new version
$ vaultsg add --overwrite foo/baz

# without the --overwrite flag, no change will made
$ vaultsg add foo/baz
```

For static secrets, `vaultsg` will compare the defined secret to that stored in
Vault and make any necessary changes.

The user will be prompted for `y/n` when overwriting any secret. Pass the
`--force` flag to skip the prompt:

```bash
$ vaultsg add --overwrite --force foo/baz
```

## State

To reduce complexity, `vaultsg` does not manage any state file. Any changes made
outside of `vaultsg` will be overridden if they differ with the defined secrets.

Any secrets defined outside `vaultsg` and not in `secrets.yml` will not be
managed.

## Configuration

A Vault address and token is required to run `vaultsg`:

```yml
# config.yml
vault_address: "https://localhost:8200"
vault_token: "foo"

vault_secrets: "secrets.yml"
```

Ideally, the token should only have the capabilities to manage the required KV
secrets.

## Encryption

Because `secrets.yml` can contain static secrets, it will need to be encrypted
and decrypted with `age`.
