import adafruit_requests
import array
import displayio
import terminalio
import time
import wifi
from adafruit_magtag.magtag import MagTag

PIXEL_BRIGHTNESS = 8

magtag = MagTag()

magtag.add_text(
    text_font=terminalio.FONT,
    text_position=(50, (magtag.graphics.display.height // 2) - 1, ),
    text_scale=1, )

magtag.peripherals.neopixel_disable = False
magtag.peripherals.neopixels[3] = (PIXEL_BRIGHTNESS, 0, 0)
magtag.set_text(
    "Connecting...\nBattery: {}V".format(magtag.peripherals.battery))

magtag.network.connect()

magtag.peripherals.neopixels[3] = (0, PIXEL_BRIGHTNESS, 0)
magtag.peripherals.neopixels[2] = (PIXEL_BRIGHTNESS, 0, 0)
magtag.set_text(
    "Retrieving cat...\nBattery: {}V".format(magtag.peripherals.battery))

res = magtag.network.requests.get('http://192.168.1.193:1337/raw')
magtag.peripherals.neopixels[2] = (0, PIXEL_BRIGHTNESS, 0)
magtag.peripherals.neopixels[1] = (PIXEL_BRIGHTNESS, 0, 0)

cat = res.content

# Two 2-byte integer dimensions
x = int.from_bytes(cat[0:2], 'big')
y = int.from_bytes(cat[2:4], 'big')
print('{} => {}'.format(cat[0:2], x))
print('{} => {}'.format(cat[2:4], y))
print('{} x {}'.format(x, y))

# Four one-byte grey values
palette = displayio.Palette(4)
palette[0] = (cat[4], cat[4], cat[4])
palette[1] = (cat[5], cat[5], cat[5])
palette[2] = (cat[6], cat[6], cat[6])
palette[3] = (cat[7], cat[7], cat[7])

# A display can only Show a Group
group = displayio.Group()

# Create a bitmap
bitmap = displayio.Bitmap(296, 128, 4)

# Make it into a sprite
sprite = displayio.TileGrid(bitmap, pixel_shader=palette)

# Add the sprite to the group
group.append(sprite)

# Initialize bitmap to white
bitmap.fill(3)

# x * y bytes of image data, each byte is an index into the above palette (just like displayio wants)
margin_x = int((296 - x) / 2)
margin_y = int((128 - y) / 2)
for py in range(0, y):
    for px in range(0, x):
        bitmap[px + margin_x, py + margin_y] = cat[py * x + px + 8]


magtag.peripherals.neopixels[1] = (0, PIXEL_BRIGHTNESS, 0)
magtag.peripherals.neopixels[0] = (PIXEL_BRIGHTNESS, 0, 0)

time.sleep(magtag.graphics.display.time_to_refresh)
magtag.graphics.display.show(group)

while True:
    try:
        magtag.graphics.display.refresh()
    except:
        time.sleep(0.1)
        continue
    break

magtag.peripherals.neopixels[0] = (0, PIXEL_BRIGHTNESS, 0)
time.sleep(magtag.graphics.display.time_to_refresh)
magtag.peripherals.neopixel_disable = True
magtag.exit_and_deep_sleep(900)
