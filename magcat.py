import displayio
import array
import wifi
import terminalio
import adafruit_requests
import time
from adafruit_magtag.magtag import MagTag

magtag = MagTag()

magtag.add_text(
    text_font=terminalio.FONT,
    text_position=(
        50,
        (magtag.graphics.display.height // 2) - 1,
    ),
    text_scale=1,
)

magtag.set_text('Connecting...')

magtag.network.connect()
magtag.set_text('Retrieving cat...')

res = magtag.network.requests.get('http://192.168.1.193:1337/raw')
body = res.iter_content()

x = int.from_bytes(next(body), 'big') << 8 | int.from_bytes(next(body), 'big')
y = int.from_bytes(next(body), 'big') << 8 | int.from_bytes(next(body), 'big')

print('{} x {}'.format(x,y))
palette = displayio.Palette(4)
palette[0] = 0x000000
palette[1] = 0x606060
palette[2] = 0xC0C0C0
palette[3] = 0xFFFFFF
group = displayio.Group()
bitmap = displayio.Bitmap(296, 128, 4)

margin_x = int((296-x)/2)
margin_y = int((128-y)/2)
for py in range(0,y):
    for px in range(0,x):
        bitmap[px+margin_x,py+margin_y] = next(body)[0]

sprite = displayio.TileGrid(bitmap, pixel_shader=palette)
group.append(sprite)
magtag.graphics.display.show(group)

while True:
    try:
        magtag.graphics.display.refresh()
    except RuntimeError(e):
        print(e)
        time.sleep(1)
        continue
    break

time.sleep(5)
magtag.exit_and_deep_sleep(900)
