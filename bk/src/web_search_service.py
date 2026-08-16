"""Standalone web search utilities (moved out of the legacy tool system)."""

import logging
import random
import time
import requests
from urllib.parse import quote_plus, urljoin, urlparse
from bs4 import BeautifulSoup
from typing import List, Dict, Any


def _duckduckgo_search(query: str, max_results: int = 5) -> List[Dict[str, str]]:
    """Search DuckDuckGo HTML and return results."""
    from duckduckgo_search import DDGS
    try:
        with DDGS() as ddgs:
            results = []
            for r in ddgs.text(query, max_results=max_results):
                results.append({
                    "title": r.get("title", ""),
                    "link": r.get("href", ""),
                    "snippet": r.get("body", ""),
                })
            return results
    except Exception as e:
        logging.getLogger(__name__).warning(f"DuckDuckGo search failed: {e}")
        return []


def _brave_search(query: str, max_results: int = 5) -> List[Dict[str, str]]:
    """Fallback Brave search via HTML scraping."""
    try:
        url = f"https://search.brave.com/search?q={quote_plus(query)}"
        headers = {
            "User-Agent": (
                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
            )
        }
        response = requests.get(url, headers=headers, timeout=10)
        soup = BeautifulSoup(response.text, "html.parser")
        results = []
        for result in soup.select(".snippet")[:max_results]:
            title_elem = result.select_one(".title")
            link_elem = result.select_one("a")
            desc_elem = result.select_one(".description")
            title = title_elem.get_text(strip=True) if title_elem else ""
            link = link_elem.get("href", "") if link_elem else ""
            snippet = desc_elem.get_text(strip=True) if desc_elem else ""
            if title or snippet:
                results.append({"title": title, "link": link, "snippet": snippet})
        return results
    except Exception as e:
        logging.getLogger(__name__).warning(f"Brave search failed: {e}")
        return []


def web_search(query: str, max_results: int = 5) -> str:
    """Search the web and return a formatted string of results."""
    logger = logging.getLogger(__name__)
    if not query or not query.strip():
        return "Error: Empty search query"

    # Try DuckDuckGo first, then Brave fallback
    results = _duckduckgo_search(query, max_results=max_results)
    if not results:
        time.sleep(random.uniform(0.5, 1.5))
        results = _brave_search(query, max_results=max_results)

    if not results:
        return f"No web search results found for: {query}"

    lines = [f"Web search results for: {query}\n"]
    for i, r in enumerate(results, 1):
        lines.append(f"{i}. {r.get('title', '')}")
        lines.append(f"   URL: {r.get('link', '')}")
        lines.append(f"   {r.get('snippet', '')}\n")
    return "\n".join(lines)


def wikipedia_search(query: str) -> str:
    """Search Wikipedia and return a formatted summary."""
    try:
        from langchain_community.tools import WikipediaQueryRun
        from langchain_community.utilities import WikipediaAPIWrapper
        wiki = WikipediaQueryRun(api_wrapper=WikipediaAPIWrapper())
        return wiki.run(query)
    except Exception as e:
        logging.getLogger(__name__).error(f"Wikipedia search failed: {e}")
        return f"Error searching Wikipedia: {e}"
