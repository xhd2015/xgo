# Standard installation
```sh
go install github.com/xhd2015/xgo/cmd/xgo@latest
```

# Install Prebuilt Binaries
For environments like CI, xgo can be installed from pre-built binaries:
```sh
# macOS and Linux (and WSL)
curl -fsSL https://github.com/xhd2015/xgo/raw/master/install.sh | bash

# Windows
powershell -c "irm github.com/xhd2015/xgo/raw/master/install.ps1|iex"
```

After installation, `~/.xgo/bin/xgo` will be available.

# Upgrade if you've already installed
If you've already installed `xgo`, you can upgrade it with:

```sh
xgo upgrade
```

# Build from source
If you want to build from source, run with:

```sh
git clone https://github.com/xhd2015/xgo
cd xgo
go run ./script/build-release --local

# check build version
~/.xgo/bin/xgo version
```