# Web Visualizer

`web-visualizer` is a Go-based CLI tool and web-crawler that can:

- Start an HTTP server exposing a web-crawling API.
- Crawl a target URL from the CLI and pretty print its endpoints as a graph.

## Installation

- **From source (Go)**:
  - Ensure Go is installed (version matching the `go.mod` in this repo).
  - Clone this repository and build the binary:

```bash
git clone https://github.com/RupenderSinghRathore/web-visualizer.git
cd web-visualizer
go build -o app ./cmd/app
```

- **Using Docker**:

```bash
docker pull kami0sama/web-visualizer:latest
```

## Usage

After building (or obtaining) the `app` executable, you can run it in two primary modes.

### Server Mode

Start the web-crawler API server:

```bash
$ export PORT=8080
$ ./app server
2026-03-01 11:31:36 INFO starting the server addr=:8080 env=development
```

It uses port 8080 if not specified

```bash
$ BODY='{"url":"https://some-site.com/"}'
$ curl -d "$BODY" localhost:8080/graph
{
	"graph": {
		"/": {
			"visited": 30,
			"status": 200,
			"links": [
				"/about",
                "/contact",
                ...

```

This runs a web server that exposes crawling functionality over HTTP.

### Client Mode

Run the crawler from the command line and pretty print the discovered endpoints as a graph:

```bash
❯ ./app client -url "https://some-site.com"
── /(200, 30)
   ├─ /about(200, 16)
   ├─ /contact(200, 10)
   │  └─ /success(200, 5)
   ├─ /blog(200, 45)
   │  ├─ /post-1(200, 120)
   │  └─ ...
   ╰─ /assets
      └─ ...
```

This will crawl the given URL and display its endpoints in a graph-like representation.

## Docker Usage

You can also run the tool via Docker. For example:

```bash
docker run  -e PORT=8000 -p 8080:8000 kami0sama/web-visualizer:latest server
2026-03-01 11:31:36 INFO starting the server addr=:8000 env=development
```

or to run the client mode against a URL:

```bash
docker run --rm kami0sama/web-visualizer:latest client -url "https://some-site.com"
── /(200, 30)
   ├─ /about(200, 16)
   ├─ /contact(200, 10)
   │  └─ /success(200, 5)
   ├─ /blog(200, 45)
   │  ├─ /post-1(200, 120)
   │  └─ ...
   ╰─ /assets
      └─ ...
```

Adjust additional flags, environment variables, and volumes as needed for your setup.

## License
Distributed under the MIT License. See `LICENSE` for more information.
