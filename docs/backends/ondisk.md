# `ondisk` storage backend

This is an experimental on disk K/V backend. It stores the encrypted data in the filesystem in a content adressable manner. It is fully encrypted, including metadata. Content can be encrypted using any of the supported encryption backend but it's only being tested with age. Metadata is always encrypted with age.

This might become the default storage and RCS backend in gopass 2.x.

**WARNING**: The disk format is still experimental and will change. **DO NOT USE** unless you want to help with the implementation.

This backend can be fully decrypted and parsed without gopass. The index is
age encrypted serialized JSON. It maps the keys (secret names) to content
addressable blobs on the filesystem. Those are usually encrypted with age.
The age keyring itself is also age encrypted serialized JSON.

#### Background: How do access ondisk secrets without gopass

This section assumes `age` and `jq` are properly installed.

```
# Decrypt the gopass-age keyring
age -d -o /tmp/keyring ~/.config/gopass/age-keyring.age
# Extract the private key
cat /tmp/keyring | jq ".[1].identity" | cut -d'"' -f2 > /tmp/private-key
# Decrypt the index
# TODO
# Locate the latest secrets
# TODO
# Decrypt it
age -d -i /tmp/private-key -o /tmp/plaintext ~/.local/share/gopass/stores/root/foo.age
```

