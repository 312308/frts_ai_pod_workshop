#!/usr/bin/env python3
"""Capture a reference website screenshot for /generate-ui-design.

Usage:
  python3 tools/design/capture-reference.py --url https://example.com --out artifacts/design/capture.png
  python3 tools/design/capture-reference.py --url https://example.com --path /pricing --viewport mobile

Requires: playwright (`pip install playwright && playwright install chromium`)
Fallback: exit 2 if Playwright unavailable — operator supplies screenshot_path.
"""
from __future__ import annotations

import argparse
import sys
from pathlib import Path
from urllib.parse import urljoin


def main() -> int:
    parser = argparse.ArgumentParser(description="Capture reference site screenshot")
    parser.add_argument("--url", required=True, help="Base reference URL")
    parser.add_argument("--path", default="/", help="Path on site")
    parser.add_argument(
        "--viewport",
        choices=("desktop", "tablet", "mobile"),
        default="desktop",
        help="Viewport preset",
    )
    parser.add_argument("--out", required=True, help="Output PNG path")
    args = parser.parse_args()

    viewports = {
        "desktop": {"width": 1440, "height": 900},
        "tablet": {"width": 768, "height": 1024},
        "mobile": {"width": 390, "height": 844},
    }

    target = urljoin(args.url.rstrip("/") + "/", args.path.lstrip("/"))
    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)

    try:
        from playwright.sync_api import sync_playwright
    except ImportError:
        print("Playwright not installed. Use screenshot_path fallback.", file=sys.stderr)
        return 2

    vp = viewports[args.viewport]
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport=vp)
        page.goto(target, wait_until="networkidle", timeout=60000)
        page.screenshot(path=str(out_path), full_page=True)
        browser.close()

    print(f"Captured {target} -> {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
