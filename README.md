This is an Adafruit MagTag project in two parts.

The github code handles talking to a web service ([The Cat API](https://thecatapi.com/)]), retrieving a random cat picture, scaling and dithering/quantizing it to greyscale for the MagTag, and then feeding the MagTag the idea in a sort of super stupid bitmap format.

To use:

1. Get an API token for The Cat API.
2. Download a release
3. Run it somewhere MagTag can access with `-api-token=YOUR-TOKEN`
4. Edit `magcat.py` to update the URL (keep the `/raw` at the end!), save it and the contents of `lib/` as `code.py` to a UF2-using MagTag

Questions/comments, open an issue.

![MagCat](magcat.bmp)
