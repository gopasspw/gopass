# Entropy

Generating cryptographic keys needs a lot of entropy. Especially `gnupg --gen-key`
depletes the kernel entropy pool (`/dev/random`) quite fast and may appear to be
stuck when it's waiting for new entropy.

If you wonder how to speed this up consider installing `rng-tools`
if this is available on your platform.

After installing `rng-tools` please make sure `rngd` is actually running and
replenishing your entropy pool.

You can do so by keeping a watch on your available entropy and running an entropy
consuming process like so.

```
watch -n1 cat /proc/sys/kernel/random/entropy_avail
# switch to another terminal / screen
cat /dev/random | rngtest -c 1000
```

The second command should complete within a few seconds and report no errors.
If it takes much longer you probably don't have an hardware RNG and will have
to generate some entropy by triggering some network activity and input.

You should avoid `havaged`.

### Debian / Ubuntu

```
sudo apt-get install rng-tools
```

### CentOS / Fedora / Red Hat

```
sudo yum install rng-tools
```

## Further Information

* [RNG-Tools on the Arch Linux Wiki](https://wiki.archlinux.org/index.php/Rng-tools)
* [gopass Issue #486](https://github.com/justwatchcom/gopass/issues/486)
