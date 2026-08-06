# ☕️ coffee

Keep your Mac awake even though you close the lid of your Mac, `coffee` toggles macOS sleep prevention
(`pmset -a disablesleep`) from the command line.

## Install

### Homebrew

```sh
brew tap mrthiti/coffee https://github.com/mrthiti/coffee
brew install coffee
```

This builds `coffee` from source (requires Xcode Command Line Tools, which
Homebrew already needs) and installs it to your Homebrew `bin`. Homebrew
doesn't set up the passwordless-sudo rule that `install.sh` does, so run
`sudo coffee` unless you add the sudoers rule yourself — see "Usage" below.

### install.sh

```sh
curl -fsSL https://raw.githubusercontent.com/mrthiti/coffee/main/install.sh | sh
```

This downloads the latest build binary for your Mac (Apple Silicon or
Intel) from [GitHub Releases](https://github.com/mrthiti/coffee/releases),
installs it to `/usr/local/bin/coffee`, and grants your user passwordless
sudo for exactly `pmset -a disablesleep 0`/`1` — nothing else — via
`/etc/sudoers.d/coffee`, so plain `coffee` never needs a password
afterwards. If that last step fails (e.g. you decline the password
prompt), the binary is still installed — just use `sudo coffee` instead.

## Update

```sh
curl -fsSL https://raw.githubusercontent.com/mrthiti/coffee/main/install.sh | sh
```

Same command as install — it overwrites the existing binary with the latest
release. Run `coffee --version` before and after to confirm it updated.

## Uninstall

```sh
sudo pmset -a disablesleep 0
sudo rm /usr/local/bin/coffee /etc/sudoers.d/coffee
```

If `coffee` is currently running in the foreground, press Ctrl+C first — it
disables sleep prevention on exit automatically. Drop the
`/etc/sudoers.d/coffee` part if install.sh never got to set it up (e.g. you
declined the password prompt). `coffee` installs a single binary with no
config file, LaunchAgent, or background daemon, so removing it is enough.

## Usage

```
coffee                Run in the foreground: enables sleep prevention and
                      keeps re-checking it (macOS can silently reset it,
                      e.g. when switching between AC and battery power),
                      then disables it again on exit (press Ctrl+C to stop).
coffee status         Show the current sleep-prevention status.
coffee --help         Show this help message.
coffee --version      Show the installed version.
```

install.sh (see "Install" above) sets up passwordless sudo for you, so
plain `coffee` normally just works. If you built from source instead and
want the same thing, add this to a file under `/etc/sudoers.d/` yourself
(validate with `visudo -c -f <file>` before it takes effect):

```
<your-username> ALL=(root) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1
```

Without that, plain `coffee` only works if you already have a cached sudo
credential; otherwise run it as `sudo coffee` instead. `status` never needs
a password.

## Requirements

- macOS (Intel or Apple Silicon)

## Building from source

```sh
git clone https://github.com/mrthiti/coffee.git
cd coffee
go build -o coffee .
```

## License

[MIT](LICENSE)
