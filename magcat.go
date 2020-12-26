package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	url2 "net/url"
	"os"
	"strings"

	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
)

func getCat(ctx context.Context, apiKey string) (img io.ReadCloser, err error) {
	url := url2.URL{
		Scheme: "https",
		Host:   "api.thecatapi.com",
		Path:   "v1/images/search",
	}
	url.Query().Add("mime_types", "jpg")
	url.Query().Add("mime_types", "png")
	catReq, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)

	var res *http.Response
	if err == nil {
		catReq.Header.Add("x-api-key", apiKey)
		res, err = http.DefaultClient.Do(catReq)
	}

	if err == nil {
		defer res.Body.Close()
		if res.StatusCode/100 != 2 {
			err = fmt.Errorf("non-2XX %d from the cat API", res.StatusCode)
		}
	}

	var payload []struct {
		Url string
	}

	if err == nil {
		dec := json.NewDecoder(res.Body)
		err = dec.Decode(&payload)
	}

	var cat *http.Response
	if err == nil {
		cat, err = http.Get(payload[0].Url)
	}

	if err == nil {
		return cat.Body, nil
	}

	return nil, fmt.Errorf("retrieving cat: %w", err)
}

func proxyCat(apiKey string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var img image.Image
		var err error
		for try := 0; try < 5; try++ {
			var cat io.ReadCloser
			cat, err = getCat(req.Context(), apiKey)
			if err == nil {
				defer cat.Close()
				img, _, err = image.Decode(cat)
			}
			if err == nil {
				break
			}
		}

		if err == nil {
			bounds := img.Bounds()
			x := uint(296)
			y := uint(bounds.Dy() * 296 / bounds.Dx())
			if y > 128 {
				y = 128
				x = uint(bounds.Dx() * 128 / bounds.Dy())
			}
			log.Printf("(%d,%d) => (%d,%d)", bounds.Dx(), bounds.Dy(), x, y)
			// 296x128
			resized := resize.Resize(x, y, img, resize.Bicubic)
			grey := image.NewPaletted(resized.Bounds(), color.Palette{
				color.Black, color.Gray{Y: 96}, color.Gray{Y: 192}, color.White,
			})
			draw.FloydSteinberg.Draw(grey, resized.Bounds(), resized, image.Point{})

			if strings.HasSuffix(req.URL.Path, "raw") {
				rw.Write([]byte{
					byte(x & 0xFF00 >> 8),
					byte(x & 0xFF),
					byte(y & 0xFF00 >> 8),
					byte(y & 0xFF),
				})
				rw.Write(grey.Pix)
			} else {
				err = bmp.Encode(rw, grey)
			}
		}

		if err != nil {
			log.Printf("error while responding: %v", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}
}

func main() {
	apiKey := flag.String("api-key", os.Getenv("API_KEY"), "api key for thecatapi.com; env: API_KEY")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("-api-key (or env API_KEY) is required")
	}

	http.HandleFunc("/", proxyCat(*apiKey))
	log.Fatal(http.ListenAndServe(":1337", nil))
}
