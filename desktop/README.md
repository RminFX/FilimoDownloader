# Filimo Downloader (Desktop)

اپ صف‌دانلود برای مک، ویندوز و لینوکس. منطق دانلود همان پروژه Go است.

## مک (آماده)

اپ ساخته‌شده:

`desktop/build/bin/FilimoDownloader.app`

دوبار کلیک کنید. توکن `AuthV1` را از کوکی فیلیمو بگذارید.

## ساخت دوباره

از پوشه `desktop`:

```bash
export PATH="$PATH:$HOME/go/bin"
./build.sh mac-arm    # مک Apple Silicon
./build.sh windows    # روی ویندوز
./build.sh linux      # روی لینوکس
```

ویندوز و لینوکس را روی همان سیستم بسازید (CGO / WebView).

پیش‌نیاز: Go، Node، FFmpeg، و روی مک Xcode Command Line Tools.
