# Servex
Servex simply provides a easy http server which serves the current directory. 
I needed a way which does not require me writing a server to testing out demos, 
web projects and was not in need of more larger or complex solutions.

## Using
Servex provides the ability to set certain options on startup.

```bash
> servex -h
  Usage of servex:
  -addrs string
      addrs: The address and port to use for the http server. (default ":4050")
  -assets string
      assets: sets the absolute path to use for assets.
   (default "/Users/apple/Labs/go/src/github.com/servi-io/web/assets")
  -base string
      base: This values sets the path to be loaded as the base path.
   (default "/Users/apple/Labs/go/src/github.com/servi-io/web")
  -withIndex
      withIndex: Indicates whether we should serve index.html as root path. (default true)
```

Starting servex is as simple as calling `servex` if you desire to start the server
on the default port `4050` else you can use the `-addr` option to set a custom `ip:port`
combo to be used.

Servex provides two routes:

  - `/files` which serves any file within the current directory and any subdirectory

  - `/assets` which serves any file specifically within the asset path provided.

If there is a `index.html` file found in the directory and the `-withIndex` option 
is not false then that will be loaded once you navigate to the `ip:port` address 
used by servex.

## Disclaimer
Servex is rather simple, its just to provide a easy solution to use and not a highly
featured server.