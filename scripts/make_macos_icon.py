#!/usr/bin/env python3
from pathlib import Path
import shutil
import subprocess
import tempfile

from PIL import Image, ImageDraw, ImageFilter


ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "logo-final.png"
OUT = ROOT / "yuanshu-ai.icns"

ICON_SIZES = [
    (16, "icon_16x16.png"),
    (32, "icon_16x16@2x.png"),
    (32, "icon_32x32.png"),
    (64, "icon_32x32@2x.png"),
    (128, "icon_128x128.png"),
    (256, "icon_128x128@2x.png"),
    (256, "icon_256x256.png"),
    (512, "icon_256x256@2x.png"),
    (512, "icon_512x512.png"),
    (1024, "icon_512x512@2x.png"),
]


def blue_bounds(image):
    pixels = image.load()
    points = []
    for y in range(image.height):
        for x in range(image.width):
            r, g, b, a = pixels[x, y]
            if a > 0 and r < 90 and g > 90 and b > 110:
                points.append((x, y))

    if not points:
        raise RuntimeError("could not find blue logo background in source image")

    return (
        min(x for x, _ in points),
        min(y for _, y in points),
        max(x for x, _ in points) + 1,
        max(y for _, y in points) + 1,
    )


def extract_mark(image, bounds):
    crop = image.crop(bounds).convert("RGBA")
    mask = Image.new("L", crop.size, 0)
    src = crop.load()
    dst = mask.load()

    for y in range(crop.height):
        for x in range(crop.width):
            r, g, b, a = src[x, y]
            if a > 0 and r > 235 and g > 235 and b > 235:
                dst[x, y] = 255

    bounds = mask.getbbox()
    if not bounds:
        raise RuntimeError("could not find white logo mark in source image")

    return mask.crop(bounds)


def build_master():
    source = Image.open(SOURCE).convert("RGBA")
    bounds = blue_bounds(source)
    mark = extract_mark(source, bounds)

    canvas = Image.new("RGBA", (1024, 1024), (0, 0, 0, 0))
    draw = ImageDraw.Draw(canvas)

    icon_size = 824
    left = (1024 - icon_size) // 2
    top = left
    right = left + icon_size
    bottom = top + icon_size
    radius = 185

    shadow = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    shadow_draw = ImageDraw.Draw(shadow)
    shadow_draw.rounded_rectangle(
        (left, top + 18, right, bottom + 18),
        radius=radius,
        fill=(0, 0, 0, 70),
    )
    canvas.alpha_composite(shadow.filter(ImageFilter.GaussianBlur(28)))

    draw.rounded_rectangle(
        (left, top, right, bottom),
        radius=radius,
        fill=(40, 143, 176, 255),
    )

    mark_target_w = int(icon_size * 0.66)
    scale = mark_target_w / mark.width
    mark_target_h = int(mark.height * scale)
    mark = mark.resize((mark_target_w, mark_target_h), Image.Resampling.LANCZOS)

    mark_layer = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    mark_x = (1024 - mark.width) // 2
    mark_y = (1024 - mark.height) // 2
    white = Image.new("RGBA", mark.size, (255, 255, 255, 255))
    mark_layer.paste(white, (mark_x, mark_y), mark)
    canvas.alpha_composite(mark_layer)

    return canvas


def main():
    if not SOURCE.exists():
        raise SystemExit(f"missing source image: {SOURCE}")
    if shutil.which("iconutil") is None:
        raise SystemExit("iconutil is required to generate .icns on macOS")

    master = build_master()
    with tempfile.TemporaryDirectory() as tmp:
        iconset = Path(tmp) / "yuanshu-ai.iconset"
        iconset.mkdir()
        for size, name in ICON_SIZES:
            resized = master.resize((size, size), Image.Resampling.LANCZOS)
            resized.save(iconset / name)

        subprocess.run(
            ["iconutil", "-c", "icns", str(iconset), "-o", str(OUT)],
            check=True,
        )

    print(f"generated {OUT}")


if __name__ == "__main__":
    main()
