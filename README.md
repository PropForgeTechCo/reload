# Reload

![Tests](https://github.com/aarol/reload/actions/workflows/test.yml/badge.svg)

Reload is a Go package, which enables "soft reloading" of web server assets and templates, 
reloading the browser almost instantly via Websockets. The strength of Reload lies in its 
simple API and easy integration to any Go projects.

This package is a fork of https://github.com/aarol/reload with the following changes:

  - Does not attempt to watch directories if there are none

## Installation

`go get github.com/PropForgeTechCo/reload`

## Usage

1. Create a new Reloader and insert the middleware to your handler chain:

   ```go
   
    // HotReload is a middleware handler that injects a hot refresh script into the page.
    // The script connects back to a websocket (handled by the middleware) and if the server
    // is restarted, via air, or otherwise, then the browser will lose the connection to the websocket
    // and when it reconnects, it will refresh the page. It retries the connection every second.
    func HotReload(next http.Handler) http.Handler {

	// Call `New()` with a list of directories to recursively watch, here we don't
	// watch any directories, because we need to rebuild the service to refresh the templates.
	// However, the reloader will complain about not having any directories to watch, so we
	// pass in the current directory.
    refresh := reload.New()
   
	refresh.OnReload = func() {
	}
	return refresh.Handle(next)
}
   ```

2. Run your application, make changes to files in the specified directories, and see the updated page instantly!

## How it works

When added to the top of the middleware chain, `(*Reloader).Handle()` will inject a small `<script/>` at the end of any HTML file sent by your application. This script will instruct the browser to open a WebSocket connection back to your server, which will be also handled by the middleware.

The injected script is at the bottom of [this file](https://github.com/aarol/reload/blob/main/reload.go).

You can also do the injection yourself, as the package also exposes the methods `(*Reloader).ServeWS` and `(*Reloader).WatchDirectories`, which are used by the `(*Reloader).Handle` middleware.

> Currently, injecting the script is done by appending to the end of the document, even after the \</html\> tag.
> This makes the library code _much_ simpler, but may break older/less forgiving browsers.

## Caveats

- Reload works with everything that the server sends to the client (HTML,CSS,JS etc.), but it cannot restart the server itself,
  since it's just a middleware running on the server.

  To reload the entire server, you can use another file watcher on top, like [watchexec](https://github.com/watchexec/watchexec):

  `watchexec -r --exts .go -- go run .`

  When the websocket connection to the server is lost, the browser will try to reconnect every second.
  This means that when the server comes back, the browser will still reload, although not as fast :)

- Reload will not work for embedded assets, since all go:embed files are baked into the executable at build time.
