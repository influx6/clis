package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/influx6/faux/context"
	"github.com/influx6/fractals/fhttp"
)

func main() {

	var (
		addrs        string
		hasIndexFile bool
		basePath     string
		assetPath    string
		extraFiles   string
		files        []string
	)

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	defaultAssets := filepath.Join(pwd, "assets")

	flag.StringVar(&extraFiles, "files", "", "files: Provides a argument to contain comma seperate paths to serve from directory\n\tExample: servex -files app.js:./app.js, db.svg:./assets/db.svg")
	flag.StringVar(&addrs, "addrs", ":4050", "addrs: The address and port to use for the http server.")
	flag.StringVar(&basePath, "base", pwd, "base: This values sets the path to be loaded as the base path.\n\t")
	flag.StringVar(&assetPath, "assets", defaultAssets, "assets: sets the absolute path to use for assets.\n\t")
	flag.BoolVar(&hasIndexFile, "withIndex", true, "withIndex: Indicates whether we should serve index.html as root path.")
	flag.Parse()

	files = strings.Split(extraFiles, ",")
	for index, fl := range files {
		files[index] = strings.TrimSpace(fl)
	}

	basePath = filepath.Clean(basePath)
	assetPath = filepath.Clean(assetPath)

	apphttp := fhttp.Drive(fhttp.MW(fhttp.RequestLogger(os.Stdout)))(fhttp.MW(fhttp.ResponseLogger(os.Stdout)))

	approuter := fhttp.Route(apphttp)

	approuter(fhttp.Endpoint{
		Path:    "/assets/*",
		Method:  "GET",
		Action:  func(ctx context.Context, rw *fhttp.Request) error { return nil },
		LocalMW: fhttp.DirFileServer(assetPath, "/assets/"),
	})

	approuter(fhttp.Endpoint{
		Path:    "/files/*",
		Method:  "GET",
		Action:  func(ctx context.Context, rw *fhttp.Request) error { return nil },
		LocalMW: fhttp.DirFileServer(basePath, "/files/"),
	})

	if hasIndexFile {
		approuter(fhttp.Endpoint{
			Path:    "/",
			Method:  "GET",
			Action:  func(ctx context.Context, rw *fhttp.Request) error { return nil },
			LocalMW: fhttp.IndexServer(basePath, "index.html", ""),
		})
	}

	for _, fl := range files {
		flset := strings.Split(fl, ":")

		if len(flset) < 2 {
			fmt.Printf("Unable to split extra file path: Path: %q, Splits: %+q\n", fl, flset)
			continue
		}

		approuter(fhttp.Endpoint{
			Path:    flset[0],
			Method:  "GET",
			Action:  func(ctx context.Context, rw *fhttp.Request) error { return nil },
			LocalMW: fhttp.IndexServer(basePath, flset[1], ""),
		})
	}

	apphttp.Serve(addrs)
}
