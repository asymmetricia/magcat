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

func proxyCat(apiKey string, raw bool) func(rw http.ResponseWriter, req *http.Request) {
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

			grey := image.NewPaletted(resized.Bounds(), palette)

			draw.FloydSteinberg.Draw(grey, resized.Bounds(), resized, image.Point{})

			if raw {
				sendRaw(rw, grey)
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

var greys = []color.Gray16{
	color.Black,
	{110 << 8},
	{150 << 8},
	color.White,
}
var palette = color.Palette{
	greys[0], greys[1], greys[2], greys[3],
}

func sendRaw(rw http.ResponseWriter, img *image.Paletted) {
	x := img.Bounds().Dx()
	y := img.Bounds().Dy()

	rw.Write([]byte{
		byte(x & 0xFF00 >> 8),
		byte(x & 0xFF),
		byte(y & 0xFF00 >> 8),
		byte(y & 0xFF),
	})

	for _, c := range greys {
		rw.Write([]byte{byte(c.Y >> 8)})
	}

	rw.Write(img.Pix)
}

func testPattern(raw bool) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var err error

		grey := image.NewPaletted(image.Rect(0, 0, 296, 128), palette)
		draw.Draw(grey, grey.Bounds(), image.White, image.Point{}, draw.Over)
		draw.Draw(grey, image.Rect(0, 0, 128, 32), image.Black, image.Point{}, draw.Over)
		draw.Draw(grey, image.Rect(0, 32, 128, 64), image.NewUniform(greys[1]), image.Point{}, draw.Over)
		draw.Draw(grey, image.Rect(0, 64, 128, 96), image.NewUniform(greys[2]), image.Point{}, draw.Over)

		gradient := image.NewGray(image.Rect(0, 0, 296-128, 128))
		for y := 0; y < 128; y++ {
			draw.Draw(gradient, image.Rect(0, y, 296-128, y+1),
				image.NewUniform(color.Gray{Y: uint8(y) << 1}), image.Point{}, draw.Over)
		}
		draw.FloydSteinberg.Draw(grey, image.Rect(128, 0, 296, 128), gradient, image.Point{})

		if raw {
			sendRaw(rw, grey)
		} else {
			err = bmp.Encode(rw, grey)
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

	http.HandleFunc("/", proxyCat(*apiKey, false))
	http.HandleFunc("/raw", proxyCat(*apiKey, true))
	http.HandleFunc("/test", testPattern(false))
	http.HandleFunc("/test/raw", testPattern(true))
	log.Fatal(http.ListenAndServe(":1337", nil))
}
