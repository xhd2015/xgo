# Limitation On Stdlib Functions
Stdlib functions like `os.ReadFile`, `io.Read` are widely used by go code. So installing a trap on these functions may have big impact on performance and security.

And as compiler treats stdlib from ordinary module differently, current implementation to support stdlib function is based on source code injection, which may causes build time to slow down.

So only a limited list of stdlib functions can be mocked. However, if there lacks some functions you may want to use, you can leave a comment in [Issue#6](https://github.com/xhd2015/xgo/issues/6) or fire an issue to let us know and add it.

# Supported List
## `os`
- `Getenv`
- `Getwd`

## `time`
- `Now`
- `Time.Format`

## `os/exec`
- `Command`
- `(*Cmd).Run`
- `(*Cmd).Output`
- `(*Cmd).Start`

# `net/http`
- `Get`
- `Head`
- `Post`
- `Serve`
- `(*Server).Close`
- `Handle`

# `net`
- `Dial`
- `DialIP`
- `DialUDP`
- `DialUnix`
- `DialTimeout`