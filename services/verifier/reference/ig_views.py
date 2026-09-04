"""Fetch Instagram reel/post view counts anonymously.

Method: shortcode GraphQL query -> owner pk; owner clips feed -> play_count.
Plain HTTP, anonymous cookies only. Verified 2026-09-04.
Reference: https://github.com/instaloader/instaloader (instaloader/structures.py).
"""

from __future__ import annotations

import json
import re
import time
import urllib.parse
from dataclasses import dataclass
from datetime import datetime, timezone

import requests

BASE = "https://www.instagram.com"
UA = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)
DOC_POST = "27128499623469141"
DOC_CLIPS = "27234427476213202"
APP_ID = "936619743392459"
ASBD_ID = "129477"


class IGViewsError(Exception):
    """Base error for IG views fetching."""


class NotFound(IGViewsError):
    """Post deleted, private, or shortcode invalid."""


class RateLimited(IGViewsError):
    """Instagram returned 429; back off."""


class DocIdRotated(IGViewsError):
    """GraphQL doc_id no longer valid; refresh from instaloader."""


class NotInFirstPage(IGViewsError):
    """Reel not in owner's first clips feed page."""


@dataclass
class ReelStats:
    shortcode: str
    views: int | None
    likes: int | None
    comments: int | None
    username: str
    user_pk: str
    taken_at: str | None
    fetched_at: str


def extract_shortcode(url_or_shortcode: str) -> str:
    """Accept full URLs (reels/, reel/, p/, tv/, ?igsh junk) or a bare shortcode."""
    s = url_or_shortcode.strip()
    if re.fullmatch(r"[A-Za-z0-9_-]{5,}", s):
        return s
    parsed = urllib.parse.urlparse(s)
    path = parsed.path.strip("/")
    for prefix in ("reels", "reel", "p", "tv"):
        m = re.fullmatch(rf"{prefix}/([A-Za-z0-9_-]+)", path)
        if m:
            return m.group(1)
    raise NotFound(f"cannot extract shortcode from {url_or_shortcode!r}")


def _new_session() -> requests.Session:
    """Bootstrap an anonymous session; returns it with csrftoken cookie set."""
    s = requests.Session()
    s.headers.update({"User-Agent": UA, "Accept-Language": "en-US"})
    r = s.get(BASE + "/", timeout=30)
    if r.status_code != 200:
        raise IGViewsError(f"bootstrap failed: HTTP {r.status_code}")
    if not s.cookies.get("csrftoken"):
        raise IGViewsError("no csrftoken cookie from bootstrap")
    return s


def _graphql(s: requests.Session, doc_id: str, variables: dict) -> dict:
    """POST /graphql/query with the persisted-query shape. Retries transient 5xx once."""
    body = {
        "variables": json.dumps(variables, separators=(",", ":")),
        "doc_id": doc_id,
        "server_timestamps": "true",
    }
    headers = {
        "X-CSRFToken": s.cookies.get("csrftoken") or "",
        "X-IG-App-ID": APP_ID,
        "X-ASBD-ID": ASBD_ID,
        "X-Requested-With": "XMLHttpRequest",
        "X-FB-Friendly-Name": "",
        "Origin": BASE,
        "Referer": BASE + "/",
        "Content-Type": "application/x-www-form-urlencoded",
        "Accept": "*/*",
    }
    for attempt in (1, 2):
        try:
            r = s.post(BASE + "/graphql/query", headers=headers, data=body, timeout=30)
        except requests.RequestException as e:
            if attempt == 2:
                raise IGViewsError(f"network error: {e}") from e
            time.sleep(3 * attempt)
            continue
        if r.status_code == 429:
            raise RateLimited("HTTP 429 from Instagram")
        if r.status_code >= 500:
            if attempt == 2:
                raise IGViewsError(f"HTTP {r.status_code} from Instagram")
            time.sleep(3 * attempt)
            continue
        if r.status_code != 200:
            raise IGViewsError(f"HTTP {r.status_code}: {r.text[:200]}")
        try:
            data = r.json()
        except ValueError as e:
            raise IGViewsError(f"non-JSON response: {r.text[:200]}") from e
        if data.get("errors"):
            msg = data["errors"][0].get("message", "")
            if "execution error" in msg:
                raise DocIdRotated(
                    f"doc_id {doc_id} failed; refresh doc_ids from instaloader"
                )
            raise IGViewsError(f"graphql error: {msg}")
        return data
    raise IGViewsError("unreachable")


def _post_metadata(s: requests.Session, shortcode: str) -> dict:
    """Call 2: shortcode -> media item with owner pk (view_count null here; ignore)."""
    variables = {
        "shortcode": shortcode,
        "__relay_internal__pv__PolarisAIGMMediaWebLabelEnabledrelayprovider": False,
    }
    data = _graphql(s, DOC_POST, variables)
    items = (
        data.get("data", {})
        .get("xdt_api__v1__media__shortcode__web_info", {})
        .get("items", [])
    )
    if not items:
        raise NotFound(f"post {shortcode} not found (private or deleted)")
    return items[0]


def _clips_feed(s: requests.Session, user_pk: str, page_size: int) -> list[dict]:
    """Call 3: owner clips feed edges."""
    variables = {
        "data": {
            "include_feed_video": True,
            "page_size": page_size,
            "target_user_id": str(user_pk),
        }
    }
    data = _graphql(s, DOC_CLIPS, variables)
    conn = data.get("data", {}).get("xdt_api__v1__clips__user__connection_v2", {})
    edges = conn.get("edges", [])
    return [e.get("node", {}).get("media", {}) for e in edges if e.get("node", {}).get("media")]


def _is_shortcode_valid(s: requests.Session, shortcode: str) -> bool:
    """Execution errors are ambiguous (rotated doc_id vs bad shortcode).
    Check a known-good control shortcode to tell them apart."""
    try:
        _post_metadata(s, "Db_JpvfO3DU")
    except (DocIdRotated, NotFound, RateLimited):
        return False
    return True


def get_reel_stats(url_or_shortcode: str) -> ReelStats:
    """Fetch views/likes/comments for any public reel/post. No login, no browser."""
    shortcode = extract_shortcode(url_or_shortcode)
    s = _new_session()

    try:
        item = _post_metadata(s, shortcode)
    except DocIdRotated:
        if _is_shortcode_valid(s, shortcode):
            raise NotInFirstPage(
                f"post {shortcode} unreachable (private, deleted, or invalid)"
            ) from None
        raise
    except NotFound:
        raise NotFound(f"post {shortcode} not found (private or deleted)") from None
    user = item.get("user") or {}
    user_pk = user.get("pk")
    if not user_pk:
        raise NotFound(f"post {shortcode} has no owner pk")

    for page_size in (12, 50):
        medias = _clips_feed(s, user_pk, page_size)
        for media in medias:
            if media.get("code") == shortcode:
                return ReelStats(
                    shortcode=shortcode,
                    views=media.get("play_count"),
                    likes=media.get("like_count"),
                    comments=media.get("comment_count"),
                    username=user.get("username", ""),
                    user_pk=str(user_pk),
                    taken_at=str(item.get("taken_at")) if item.get("taken_at") else None,
                    fetched_at=datetime.now(timezone.utc).isoformat(),
                )
    raise NotInFirstPage(
        f"reel {shortcode} not in owner's clips feed pages 12/50"
    )


if __name__ == "__main__":
    import sys

    stats = get_reel_stats(sys.argv[1])
    print(json.dumps(stats.__dict__, indent=2))
