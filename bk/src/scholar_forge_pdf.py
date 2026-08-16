"""
ScholarForge — export composed markdown to PDF.
"""
from __future__ import annotations

import logging
import os
import re
import uuid
from pathlib import Path
from typing import Optional, Tuple

from fpdf.enums import XPos, YPos
from fpdf.errors import FPDFException

from .config import settings

logger = logging.getLogger(__name__)

_FONT_FAMILY = "ScholarForgeSans"
_MAX_TOKEN_LEN = 120


def _output_dir() -> str:
    path = os.path.join(settings.data_directory, settings.scholar_forge_output_dir)
    os.makedirs(path, exist_ok=True)
    return path


def _fonts_dir() -> Path:
    return Path(settings.data_directory) / "fonts"


def _font_pairs() -> list[Tuple[Path, Path]]:
    """Candidate (regular, bold) TrueType font paths, first match wins."""
    fonts = _fonts_dir()
    return [
        (fonts / "DejaVuSans.ttf", fonts / "DejaVuSans-Bold.ttf"),
        (
            Path("/System/Library/Fonts/Supplemental/Arial Unicode.ttf"),
            Path("/System/Library/Fonts/Supplemental/Arial Bold.ttf"),
        ),
        (
            Path("/Library/Fonts/Arial Unicode.ttf"),
            Path("/Library/Fonts/Arial Bold.ttf"),
        ),
        (
            Path("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"),
            Path("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"),
        ),
        (
            Path("/usr/share/fonts/dejavu/DejaVuSans.ttf"),
            Path("/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf"),
        ),
        (
            Path("C:/Windows/Fonts/arial.ttf"),
            Path("C:/Windows/Fonts/arialbd.ttf"),
        ),
    ]


def _resolve_font_paths() -> Tuple[Path, Path]:
    for regular, bold in _font_pairs():
        if regular.is_file():
            bold_path = bold if bold.is_file() else regular
            return regular, bold_path
    raise RuntimeError(
        "No Unicode PDF font found. Install DejaVu Sans into "
        f"{_fonts_dir()} (DejaVuSans.ttf + DejaVuSans-Bold.ttf), or use macOS/Windows system fonts."
    )


def _sanitize_filename(title: str) -> str:
    safe = re.sub(r"[^\w\s-]", "", title or "document").strip().replace(" ", "_")
    return (safe[:60] or "scholar_forge")[:60]


def _break_long_tokens(text: str, max_len: int = _MAX_TOKEN_LEN) -> str:
    """Insert breaks in URLs and unbreakable tokens so fpdf2 can wrap lines."""
    parts: list[str] = []
    for token in re.split(r"(\s+)", text):
        if len(token) <= max_len or token.isspace():
            parts.append(token)
            continue
        chunk = token
        while len(chunk) > max_len:
            parts.append(chunk[:max_len])
            parts.append("\n")
            chunk = chunk[max_len:]
        parts.append(chunk)
    return "".join(parts)


def _strip_markdown_inline(text: str) -> str:
    clean = re.sub(r"\*\*(.+?)\*\*", r"\1", text)
    clean = re.sub(r"\*(.+?)\*", r"\1", clean)
    clean = re.sub(r"`(.+?)`", r"\1", clean)
    return _break_long_tokens(clean)


class _ScholarForgePDF:
    def __init__(self) -> None:
        try:
            from fpdf import FPDF
        except ImportError as e:
            raise RuntimeError(
                "PDF export requires fpdf2. Install with: pip install fpdf2"
            ) from e

        regular, bold = _resolve_font_paths()
        self.pdf = FPDF()
        self.pdf.set_auto_page_break(auto=True, margin=15)
        self.pdf.set_margins(15, 15, 15)
        self.pdf.add_page()
        self.pdf.add_font(_FONT_FAMILY, "", fname=str(regular))
        self.pdf.add_font(_FONT_FAMILY, "B", fname=str(bold))
        self.body_size = 11
        self.width = self.pdf.epw
        self._set_body()

    def _set_body(self) -> None:
        self.pdf.set_font(_FONT_FAMILY, size=self.body_size)

    def _write_block(self, text: str, *, size: int, bold: bool = False, line_h: float = 5) -> None:
        self.pdf.set_font(_FONT_FAMILY, "B" if bold else "", size=size)
        try:
            self.pdf.multi_cell(
                self.width,
                line_h,
                text,
                new_x=XPos.LMARGIN,
                new_y=YPos.NEXT,
            )
        except FPDFException as e:
            # Last-resort fallback for stubborn tokens.
            safe = text.encode("ascii", errors="replace").decode("ascii")
            self.pdf.multi_cell(
                self.width,
                line_h,
                safe,
                new_x=XPos.LMARGIN,
                new_y=YPos.NEXT,
            )
            if safe != text:
                logger.warning("PDF export replaced unsupported characters: %s", e)
        self._set_body()

    def write_markdown(self, markdown_text: str) -> None:
        in_code = False
        for line in (markdown_text or "").splitlines():
            stripped = line.rstrip()
            if stripped.startswith("```"):
                in_code = not in_code
                continue
            if in_code:
                self._write_block(stripped or " ", size=10, line_h=4.5)
                continue

            if stripped.startswith("# "):
                self._write_block(_strip_markdown_inline(stripped[2:].strip()), size=16, bold=True, line_h=8)
                self.pdf.ln(2)
            elif stripped.startswith("## "):
                self._write_block(_strip_markdown_inline(stripped[3:].strip()), size=13, bold=True, line_h=7)
                self.pdf.ln(1)
            elif stripped.startswith("### "):
                self._write_block(_strip_markdown_inline(stripped[4:].strip()), size=11, bold=True, line_h=6)
            elif stripped.startswith("- ") or stripped.startswith("* "):
                bullet = "- " + _strip_markdown_inline(stripped[2:].strip())
                self._write_block(bullet, size=self.body_size, line_h=5)
            elif stripped == "---":
                self.pdf.ln(2)
            elif stripped:
                self._write_block(_strip_markdown_inline(stripped), size=self.body_size, line_h=5)
            else:
                self.pdf.ln(3)

    def save(self, pdf_path: str) -> None:
        self.pdf.output(pdf_path)


def markdown_to_pdf(markdown_text: str, title: str) -> tuple[str, str]:
    """
    Convert markdown-ish text to PDF.
    Returns (file_id, display_filename).
    """
    file_id = str(uuid.uuid4())
    out_dir = _output_dir()
    pdf_path = os.path.join(out_dir, f"{file_id}.pdf")
    display_name = f"{_sanitize_filename(title)}_{file_id[:8]}.pdf"

    writer = _ScholarForgePDF()
    writer.write_markdown(markdown_text)
    writer.save(pdf_path)
    logger.info("ScholarForge PDF saved: %s", pdf_path)
    return file_id, display_name


def save_markdown(markdown_text: str, title: str) -> tuple[str, str]:
    """Save markdown source alongside PDF."""
    file_id = str(uuid.uuid4())
    out_dir = _output_dir()
    md_path = os.path.join(out_dir, f"{file_id}.md")
    display_name = f"{_sanitize_filename(title)}_{file_id[:8]}.md"
    with open(md_path, "w", encoding="utf-8") as f:
        f.write(markdown_text)
    return file_id, display_name


def get_output_path(file_id: str, ext: str) -> Optional[str]:
    fid = file_id.lower().strip()
    if not re.match(r"^[0-9a-f-]{36}$", fid):
        return None
    path = os.path.join(_output_dir(), f"{fid}.{ext.lstrip('.')}")
    return path if os.path.isfile(path) else None


def export_pdf_from_markdown_id(markdown_id: str, title: str) -> tuple[str, str]:
    """Build PDF from an existing ScholarForge markdown output file."""
    md_path = get_output_path(markdown_id, "md")
    if not md_path:
        raise FileNotFoundError(f"Markdown output not found: {markdown_id}")
    with open(md_path, encoding="utf-8") as f:
        body = f.read()
    return markdown_to_pdf(body, title)
