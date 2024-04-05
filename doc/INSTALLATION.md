# Install Prebuilt Binaries
Install prebuilt:
```sh
# macOS and Linux (and WSL)
curl -fsSL https://github.com/xhd2015/xgo/raw/master/install.sh | bash

# Windows
powershell -c "irm github.com/xhd2015/xgo/raw/master/install.ps1|iex"
```

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
```