import displayio
import array
import wifi
import terminalio
import adafruit_requests
import time
from adafruit_magtag.magtag import MagTag

PIXEL_BRIGHTNESS=4

magtag = MagTag()

magtag.add_text(
    text_font=terminalio.FONT,
    text_position=(
        50,
        (magtag.graphics.display.height // 2) - 1,
    ),
    text_scale=1,
)

magtag.peripherals.neopixel_disable = False
magtag.peripherals.neopixels[3] = (PIXEL_BRIGHTNESS,0,0)
magtag.set_text('Connecting...')

magtag.network.connect()

magtag.peripherals.neopixels[3] = (0,PIXEL_BRIGHTNESS,0)
magtag.peripherals.neopixels[2] = (PIXEL_BRIGHTNESS,0,0)
magtag.set_text('Retrieving cat...')

res = magtag.network.requests.get('http://nuc.lan:1337/raw')
magtag.peripherals.neopixels[2] = (0,PIXEL_BRIGHTNESS,0)
magtag.peripherals.neopixels[1] = (PIXEL_BRIGHTNESS,0,0)

cat = res.content
x = int.from_bytes(cat[0:2], 'big')
y = int.from_bytes(cat[2:4], 'big')
print('{} => {}'.format(cat[0:2], x))
print('{} => {}'.format(cat[2:4], y))

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
        bitmap[px+margin_x,py+margin_y] = cat[py*x+px+4]

sprite = displayio.TileGrid(bitmap, pixel_shader=palette)
group.append(sprite)

magtag.peripherals.neopixels[1] = (0,PIXEL_BRIGHTNESS,0)
magtag.peripherals.neopixels[0] = (PIXEL_BRIGHTNESS,0,0)
time.sleep(magtag.graphics.display.time_to_refresh)
magtag.graphics.display.show(group)
magtag.graphics.display.refresh()

magtag.peripherals.neopixels[0] = (0,PIXEL_BRIGHTNESS,0)
time.sleep(magtag.graphics.display.time_to_refresh)
magtag.peripherals.neopixel_disable = True
magtag.exit_and_deep_sleep(900)
