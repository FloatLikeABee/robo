"""
ScholarForge Enhanced — thesis-grade prompts, web-search reference gathering,
and post-generation polishing for academic documents.

Adds three capability layers on top of the base ScholarForge pipeline:

1. ENHANCED PROMPTS
   - Organizer: thesis-specific clarification (methodology, theory, research questions)
   - Writer: claim-evidence-warrant structure, thesis-level vocabulary, dense citations
   - Reviewer: thesis quality dimensions (argument, evidence, methodology, tone)

2. WEB-SEARCH REFERENCE GATHERING
   - RESEARCHER role: searches for seminal papers, recent studies, methodology refs
   - Injects real references into the writing context
   - Builds structured bibliography entries

3. POST-GENERATION POLISHING
   - POLISHER role: transitions, citation consistency, institutional formatting
   - Front matter: declaration, acknowledgements, abstract, TOC
   - Bibliography formatting per chosen citation style
   - Word-count verification, coherence check
"""

from __future__ import annotations

import json
import logging
import re
from typing import Any, Dict, Iterator, List, Optional, Tuple

from .models import ScholarForgeProfile, ScholarForgeStatus
from .scholar_forge_rag import query_project_rag

logger = logging.getLogger(__name__)

# ──────────────────────────────────────────────────────────────────────
# CONSTANTS
# ──────────────────────────────────────────────────────────────────────
ROLE_RESEARCHER = "researcher"
ROLE_POLISHER = "polisher"

RAG_QUERY_RESULTS = 5
MAX_PREV_CHARS = 20_000

CITATION_GUIDES: Dict[str, str] = {
    "APA 7": (
        "APA 7th edition: in-text (Author, Year); reference list: Author, A. A. (Year). "
        "Title. Publisher. DOI. For journal articles: Author, A. A. (Year). Title. "
        "Journal Name, Volume(Issue), pages. DOI."
    ),
    "APA 6": (
        "APA 6th edition: in-text (Author, Year); reference list: Author, A. A. (Year). "
        "Title. Publisher. DOI."
    ),
    "MLA 9": (
        'MLA 9th edition: in-text (Author Page); Works Cited: Author. Title. '
        'Publisher, Year. For journal: Author. "Title." Journal, vol., no., year, pp. Database, DOI.'
    ),
    "MLA 8": (
        "MLA 8th edition: in-text (Author Page); Works Cited: Author. Title. Publisher, Year."
    ),
    "Chicago": (
        "Chicago Manual of Style: notes-bibliography or author-date. "
        "Footnote: Author, Title (Publisher, Year), page. "
        "Bibliography: Author. Title. Publisher, Year."
    ),
    "Harvard": (
        "Harvard referencing: in-text (Author, Year: page); reference list: "
        "Author (Year) Title. Publisher. Available at: URL (Accessed: date)."
    ),
    "IEEE": (
        'IEEE: in-text [1], [2]; reference list: [1] A. Author, "Title," '
        "Journal, vol., no., pp., Year."
    ),
    "Vancouver": (
        "Vancouver: in-text (1), (2); reference list numbered in order of appearance: "
        "1. Author A. Title. Journal. Year;Volume(Issue):pages."
    ),
}


def _citation_guide(citation_style: str) -> str:
    style = (citation_style or "APA 7").strip()
    for key, guide in CITATION_GUIDES.items():
        if key.lower() in style.lower():
            return guide
    return CITATION_GUIDES["APA 7"]


def _doc_type_value(project: ScholarForgeProfile) -> str:
    dt = project.document_type
    return dt.value if hasattr(dt, "value") else str(dt)


def _is_thesis_type(project: ScholarForgeProfile) -> bool:
    return _doc_type_value(project) in ("thesis", "dissertation")


# ──────────────────────────────────────────────────────────────────────
# LAYER 1: ENHANCED THESIS-AWARE PROMPTS
# ──────────────────────────────────────────────────────────────────────


def enhanced_clarification_prompt(
    project: ScholarForgeProfile,
    rag_ctx: str = "",
    images_block: str = "",
) -> str:
    """Organizer clarification prompt with thesis-specific intelligence."""
    doc_type = _doc_type_value(project)
    is_thesis = _is_thesis_type(project)
    meta = project.document_meta

    thesis_questions = ""
    if is_thesis:
        thesis_questions = (
            "\nFor THESES/DISSERTATIONS, probe these academic dimensions deeply:\n"
            "- Research question(s) / hypotheses: are they clearly stated and testable?\n"
            "- Theoretical framework: which theories underpin the work?\n"
            "- Methodology: quantitative, qualitative, mixed-methods? Data sources?\n"
            "- Literature review scope: which key works must be covered?\n"
            "- Contribution to knowledge: what gap does this fill?\n"
            "- Ethical considerations: IRB approval, data handling?\n"
            "- Delimitations / limitations: what is in scope and what is not?\n"
            "- Expected chapter structure: any specific requirements?\n"
        )

    doc_label = "thesis/dissertation" if is_thesis else "scholarly article"
    return (
        f"You are the ScholarForge ORGANIZER — an expert academic supervisor guiding "
        f"{doc_label} composition.\n\n"
        f"## Project Context\n"
        f"Document type: {doc_type}\n"
        f"Subject: {project.subject}\n"
        f"Title: {project.title}\n"
        f"Intro: {project.short_intro}\n"
        f"Detailed prompt: {project.detailed_prompt}\n"
        f"Citation style: {meta.citation_style or 'APA 7'}\n"
        f"Language: {meta.language or 'English'}\n"
        f"Author: {meta.author_name}\n"
        f"Affiliation: {meta.author_affiliation}\n"
        f"University: {meta.university}\n"
        f"Department: {meta.department}\n"
        f"Degree: {meta.degree_program}\n"
        f"Supervisor: {meta.supervisor}\n"
        f"Institutional requirements: {meta.thesis_requirements_notes}\n"
        + (f"\n## Uploaded material context\n{rag_ctx[:8000]}\n" if rag_ctx else "")
        + (images_block if images_block else "")
        + thesis_questions
        + "\n\n## Task\n"
        "Analyze whether requirements are SUFFICIENT for high-quality academic writing. "
        "If NOT sufficient, generate 4–10 targeted questions covering methodology, "
        "scope, theoretical grounding, and contribution. "
        "Be rigorous — thesis-level work demands clarity.\n\n"
        "Respond with ONLY valid JSON:\n"
        "{\n"
        '  "sufficient": true|false,\n'
        '  "title": "form title if insufficient",\n'
        '  "intro": "why more info is needed",\n'
        '  "fields": [\n'
        '    {"id":"field_id","label":"Question","field_type":"text|textarea|select|number",'
        '"required":true,"help_text":"...","options":[],"placeholder":"..."}\n'
        "  ]\n"
        "}\n"
    )


def enhanced_structure_prompt(
    project: ScholarForgeProfile,
    rag_ctx: str = "",
) -> str:
    """Structure generation with thesis-aware section planning."""
    doc_type = _doc_type_value(project)
    is_thesis = _is_thesis_type(project)
    meta = project.document_meta

    degree_label = "doctoral dissertation" if doc_type == "dissertation" else "master's thesis"

    if is_thesis:
        reqs_line = (
            f"Honor institutional requirements: {meta.thesis_requirements_notes}\n"
            if meta.thesis_requirements_notes
            else "Configure additional requirements in document metadata."
        )
        type_guide = (
            "## Thesis/Dissertation Structure Requirements\n"
            "Generate a COMPLETE academic structure appropriate for a "
            f"{degree_label}. "
            "Must include ALL of these standard elements:\n\n"
            "**Front Matter:**\n"
            "1. Title page\n"
            "2. Declaration / ethics statement\n"
            "3. Abstract (structured if required)\n"
            "4. Acknowledgements\n"
            "5. Table of Contents\n"
            "6. List of Figures / Tables\n"
            "7. List of Abbreviations (if needed)\n\n"
            "**Body Chapters:**\n"
            "1. Introduction — background, problem statement, research questions, "
            "objectives, significance, thesis outline\n"
            "2. Literature Review — theoretical framework, empirical review, "
            "research gap, conceptual framework\n"
            "3. Methodology — research design, population/sample, data collection, "
            "data analysis, validity/reliability, ethical considerations\n"
            "4. Results / Findings — presentation of data, analysis\n"
            "5. Discussion — interpretation, comparison with literature, implications\n"
            "6. Conclusion — summary, contributions, limitations, recommendations, "
            "future work\n\n"
            "**Back Matter:**\n"
            "- Bibliography / References\n"
            "- Appendices (if needed)\n\n"
            + reqs_line
        )
    else:
        type_guide = (
            "## Article Structure Requirements\n"
            "Standard journal article: Abstract, Introduction, Methods/Approach, "
            "Results, Discussion, Conclusion, References.\n"
            "Include IMRaD structure if the subject demands it."
        )

    return (
        f"You are the ScholarForge ORGANIZER designing a {doc_type} structure.\n\n"
        f"## Project\n"
        f"Subject: {project.subject}\n"
        f"Title: {project.title}\n"
        f"Detailed prompt: {project.detailed_prompt}\n"
        f"Short intro: {project.short_intro}\n"
        f"Citation style: {meta.citation_style or 'APA 7'}\n"
        f"Language: {meta.language or 'English'}\n"
        f"Word targets: {project.detailed_prompt[:500]}\n"
        + (f"\n## Materials\n{rag_ctx[:6000]}\n" if rag_ctx else "")
        + type_guide
        + '\n\n## Output Format — JSON ONLY\n'
        "{\n"
        f'  "document_title": "{project.title}",\n'
        '  "abstract_outline": "structured abstract plan",\n'
        '  "notes": "guidance for writer — tone, depth, key citations to pursue",\n'
        '  "sections": [\n'
        '    {"id":"sec_1","title":"Section Title","description":"detailed scope",'
        '"order":1,"word_target":2500}\n'
        "  ]\n"
        "}\n\n"
        "IMPORTANT: For theses, include ALL standard chapters listed above. "
        "For articles, include all IMRaD sections. "
        "Give meaningful word targets per section (>=1000 for thesis chapters). "
        "The 'notes' field must give concrete writing guidance: tone, voice, "
        "citation density expectations, and specific references to pursue."
    )


def enhanced_writer_paragraph_prompt(
    project: ScholarForgeProfile,
    section_title: str,
    section_description: str,
    section_order: int,
    word_target: int,
    para_idx: int,
    para_total: int,
    prior_in_section: str,
    prior_sections: str,
    rag_ctx: str,
    reference_context: str = "",
    images_catalog: str = "",
) -> str:
    """Writer prompt enforcing thesis-level academic writing standards."""
    doc_type = _doc_type_value(project)
    is_thesis = _is_thesis_type(project)
    meta = project.document_meta
    cite_style = meta.citation_style or "APA 7"

    thesis_quality = ""
    if is_thesis:
        thesis_quality = (
            "## Thesis-Level Writing Standards\n"
            "- Every claim must be backed by evidence or citation\n"
            "- Use the **Claim → Evidence → Warrant** argument structure\n"
            "- Maintain formal academic register — no colloquialisms\n"
            "- Define key terms on first use\n"
            "- Connect to the overarching research question(s)\n"
            "- Show awareness of scholarly debate — acknowledge counter-arguments\n"
            "- Use hedging language appropriately (suggest, indicate, may, appear)\n"
        )

    cite_density = (
        "aim for 2-4 citations per paragraph for thesis-level writing"
        if is_thesis
        else "cite key sources appropriately"
    )
    para_range = "150–250" if is_thesis else "120–180"

    return (
        f"You are an EXPERT ACADEMIC WRITER composing a {doc_type}. "
        f"Write with the rigor and depth expected of peer-reviewed scholarship.\n\n"
        f"## Document Plan\n"
        f"{(project.final_plan or project.detailed_prompt)[:12000]}\n\n"
        f"## Current Section ({section_order}): {section_title}\n"
        f"Section brief: {section_description}\n"
        f"Target for ENTIRE section: ~{word_target or 2000} words\n"
        f"Paragraph {para_idx + 1} of approximately {para_total} in this section\n\n"
        f"## Citation Requirements\n"
        f"Style: {cite_style}\n"
        f"Citation guide: {_citation_guide(cite_style)}\n"
        f"Density: {cite_density}\n\n"
        + thesis_quality
        + (f"## Prior Sections Summary\n{prior_sections}\n\n" if prior_sections else "")
        + (f"## Paragraphs Already Written in This Section\n{prior_in_section}\n\n" if prior_in_section else "")
        + (f"## Research Materials & References\n{rag_ctx[:8000]}\n\n" if rag_ctx else "")
        + (f"## Gathered References (use these!)\n{reference_context[:6000]}\n\n" if reference_context else "")
        + (images_catalog if images_catalog else "")
        + f"\n\n## Task\n"
        f"Write ONLY paragraph {para_idx + 1}. One cohesive scholarly paragraph "
        f"({para_range} words). "
        f"Do NOT repeat prior paragraphs. "
        f"Include in-text citations in {cite_style} format where you reference sources. "
        f"Ensure logical flow from previous paragraphs. "
        f"Return ONLY the paragraph text — no headings, labels, numbering, or commentary."
    )


def enhanced_reviewer_prompt(
    project: ScholarForgeProfile,
    section_title: str,
    paragraph_text: str,
    para_idx: int,
) -> str:
    """Reviewer prompt with thesis-quality evaluation dimensions."""
    is_thesis = _is_thesis_type(project)
    meta = project.document_meta
    cite_style = meta.citation_style or "APA 7"

    thesis_criteria = ""
    if is_thesis:
        thesis_criteria = (
            "7. **Methodology alignment** — consistent with the research approach?\n"
            "8. **Contribution clarity** — does this advance the thesis argument?\n"
        )

    citation_line = (
        "2. **Citation correctness** — proper " + cite_style
        + " format, relevant sources?\n"
    )

    return (
        "You are the ScholarForge REVIEWER — a strict academic peer reviewer. "
        "Critique this paragraph for scholarly publication quality.\n\n"
        f"Document type: {_doc_type_value(project)}\n"
        f"Citation style: {cite_style}\n"
        f"Section: {section_title}\n"
        f"Plan excerpt:\n{(project.final_plan or '')[:4000]}\n\n"
        f"## Paragraph {para_idx + 1} DRAFT\n{paragraph_text}\n\n"
        "## Evaluation Criteria (score 0-100 each, composite final score)\n"
        "1. **Argument quality** — clear claim, evidence, reasoning?\n"
        + citation_line
        + "3. **Academic tone** — formal register, precise vocabulary, no fluff?\n"
        "4. **Coherence** — flows from prior context, transitions smooth?\n"
        "5. **Evidence density** — claims supported, not asserted?\n"
        "6. **Originality** — synthesis not just summary?\n"
        + thesis_criteria
        + '\n\n## Output — JSON ONLY\n'
        "{\n"
        '  "approved": true/false,\n'
        '  "quality_score": 0-100,\n'
        '  "issues": ["concrete issue 1", "issue 2"],\n'
        '  "suggestions": ["actionable fix 1", "actionable fix 2"],\n'
        '  "summary": "one-paragraph review summary"\n'
        "}\n\n"
        "Set approved=true if quality_score >= 75 AND no major issues remain. "
        "Be precise in issues and suggestions — quote the text when pointing to problems. "
        "For thesis work, demand higher standards: citation density, argument rigor, "
        "methodological awareness."
    )


def enhanced_revise_prompt(
    project: ScholarForgeProfile,
    section_title: str,
    paragraph_text: str,
    review_summary: str,
    issues: List[str],
    suggestions: List[str],
    para_idx: int,
) -> str:
    """Revision prompt that enforces academic quality improvements."""
    issues_text = "\n".join(f"- {i}" for i in issues) or "(none)"
    suggestions_text = "\n".join(f"- {s}" for s in suggestions) or "(none)"

    return (
        "You are the ScholarForge WRITER revising for academic publication. "
        "Address EVERY reviewer issue and suggestion. "
        "Preserve factual content and citations while improving quality.\n\n"
        f"Section: {section_title}\n"
        f"Citation style: {project.document_meta.citation_style or 'APA 7'}\n"
        f"Paragraph {para_idx + 1} current draft:\n{paragraph_text}\n\n"
        f"## Reviewer Summary\n{review_summary}\n\n"
        f"## Issues to Fix\n{issues_text}\n\n"
        f"## Specific Changes Needed\n{suggestions_text}\n\n"
        "Apply ALL feedback. If a suggestion would remove a citation, keep the citation "
        "and rephrase. If the reviewer asks for more evidence, add a supporting sentence. "
        "Return ONLY the revised paragraph — no commentary."
    )


# ──────────────────────────────────────────────────────────────────────
# LAYER 2: WEB-SEARCH REFERENCE GATHERING
# ──────────────────────────────────────────────────────────────────────


def _build_research_queries(project: ScholarForgeProfile) -> List[str]:
    """Generate targeted academic search queries from project context."""
    queries: List[str] = []
    subject = project.subject

    title_words = project.title.split()
    query_terms = " ".join(w for w in title_words if len(w) > 4)

    # Core topic search
    queries.append(f"scholarly articles {subject} research paper")

    # Seminal / key works
    if query_terms:
        queries.append(f"seminal papers key studies {query_terms}")

    # Recent publications (last 5 years)
    queries.append(f"recent research 2022 2023 2024 2025 {subject}")

    # Methodology references
    queries.append(f"research methodology {subject} methods")

    # Literature review
    queries.append(f"literature review {subject} state of the art")

    # Citation-format specific queries
    queries.append(f"academic reference {subject} doi")

    # Deduplicate and trim
    seen: set = set()
    unique: List[str] = []
    for q in queries:
        if q not in seen:
            seen.add(q)
            unique.append(q)
    return unique[:8]


def _parse_search_results(search_output: str) -> List[Dict[str, str]]:
    """Parse web search output string back into structured results."""
    if not search_output or "No search results" in search_output:
        return []

    results: List[Dict[str, str]] = []
    # Search output pattern: "N. Title\n   URL: url\n   snippet"
    pattern = re.compile(
        r"(\d+)\.\s+(.+?)\n\s+URL:\s+(.+?)\n\s+(.+)",
        re.MULTILINE | re.DOTALL,
    )
    for match in pattern.finditer(search_output):
        results.append({
            "title": match.group(2).strip(),
            "url": match.group(3).strip(),
            "snippet": match.group(4).strip()[:500],
        })
    return results[:15]


def gather_references(
    project: ScholarForgeProfile,
    web_search_func,
) -> Tuple[str, str]:
    """
    Run web searches to gather academic references.

    Args:
        project: ScholarForge project profile
        web_search_func: callable(query: str) -> str (e.g., web_search_service.web_search)

    Returns:
        (reference_context_block, bibliography_block)
    """
    queries = _build_research_queries(project)
    all_results: List[Dict[str, str]] = []
    seen_urls: set = set()

    for query in queries:
        try:
            raw = web_search_func(query)
            parsed = _parse_search_results(raw)
            for r in parsed:
                if r["url"] not in seen_urls:
                    seen_urls.add(r["url"])
                    all_results.append(r)
        except Exception as e:
            logger.warning("ScholarForge reference search failed for '%s': %s", query, e)

    if not all_results:
        return "", ""

    # Build reference context for the writer (concise)
    ref_ctx_lines = ["## Gathered Academic References\n"]
    for i, r in enumerate(all_results[:20], 1):
        ref_ctx_lines.append(
            f"{i}. **{r['title']}**\n"
            f"   Source: {r['url']}\n"
            f"   Summary: {r['snippet'][:300]}\n"
        )
    reference_context = "\n".join(ref_ctx_lines)

    # Build preliminary bibliography
    bib_lines = ["## Preliminary Bibliography\n"]
    for i, r in enumerate(all_results[:25], 1):
        bib_lines.append(
            f"{i}. {r['title']}. Retrieved from {r['url']}. {r['snippet'][:100]}..."
        )
    bibliography = "\n".join(bib_lines)

    return reference_context, bibliography


# ──────────────────────────────────────────────────────────────────────
# LAYER 3: POST-GENERATION POLISHING
# ──────────────────────────────────────────────────────────────────────


def _build_polishing_context(project: ScholarForgeProfile, full_body: str) -> str:
    """Build the context block for the polisher agent."""
    meta = project.document_meta
    structure = project.structure
    cite_style = meta.citation_style or "APA 7"

    sections_summary = ""
    if structure and structure.sections:
        sections_summary = "\n".join(
            f"{s.order}. {s.title} (~{s.word_target or '?'} words)"
            for s in sorted(structure.sections, key=lambda x: x.order)
        )

    return (
        f"## Document Info\n"
        f"Type: {_doc_type_value(project)}\n"
        f"Title: {project.title}\n"
        f"Subject: {project.subject}\n"
        f"Author: {meta.author_name}\n"
        f"University: {meta.university}\n"
        f"Department: {meta.department}\n"
        f"Degree: {meta.degree_program}\n"
        f"Supervisor: {meta.supervisor}\n"
        f"Citation style: {cite_style}\n"
        f"Citation guide: {_citation_guide(cite_style)}\n"
        f"Language: {meta.language or 'English'}\n"
        f"Keywords: {', '.join(meta.keywords) if meta.keywords else '(none)'}\n"
        f"Abstract word limit: {meta.abstract_word_limit or 'not specified'}\n"
        f"Institutional requirements: {meta.thesis_requirements_notes}\n"
        f"Submission date: {meta.submission_date}\n"
        f"\n## Section Structure\n{sections_summary}\n"
        f"\n## Full Document Body\n{full_body[:30000]}\n"
    )


POLISHER_SYSTEM = (
    "You are the ScholarForge POLISHER — a senior academic editor who prepares manuscripts "
    "for final submission. You polish but never rewrite from scratch. Your job: make this "
    "document publication-ready."
)


def build_polisher_prompt(
    project: ScholarForgeProfile,
    full_body: str,
    reference_context: str = "",
    bibliography: str = "",
) -> str:
    """Build the complete polishing prompt."""
    is_thesis = _is_thesis_type(project)
    meta = project.document_meta
    cite_style = meta.citation_style or "APA 7"

    thesis_specific = ""
    if is_thesis:
        thesis_specific = (
            "\n## Thesis-Specific Requirements\n"
            "- Verify ALL front-matter elements are present (title page, declaration, "
            "abstract, acknowledgements, TOC, lists of figures/tables/abbreviations)\n"
            "- Include a formal declaration/ethics statement if missing\n"
            "- Add acknowledgements section placeholder if missing\n"
            "- Ensure chapter numbering is consistent (Chapter 1, 2, 3...)\n"
            "- Verify word count targets per chapter\n"
            "- Add 'Table of Contents' placeholder with section references\n"
            "- Add 'List of Figures' and 'List of Tables' if figures/tables referenced\n"
            "- Add 'List of Abbreviations' if acronyms used frequently\n"
            "- Honor ALL institutional formatting requirements from document metadata\n"
        )

    return (
        f"{POLISHER_SYSTEM}\n\n"
        f"{_build_polishing_context(project, full_body)}\n"
        + (f"\n## Reference Context\n{reference_context[:6000]}\n" if reference_context else "")
        + (f"\n## Preliminary Bibliography\n{bibliography[:5000]}\n" if bibliography else "")
        + thesis_specific
        + "\n\n## Polishing Tasks (perform in order)\n\n"
        "### 1. TRANSITIONS & FLOW\n"
        "- Add smooth transitional sentences BETWEEN sections where flow is abrupt\n"
        "- Ensure each chapter/section opens with an orienting paragraph\n"
        "- Verify logical progression: Introduction -> Literature -> Method -> Results -> "
        "Discussion -> Conclusion\n\n"
        "### 2. CITATION CONSISTENCY\n"
        f"- Verify ALL in-text citations follow {cite_style} format exactly\n"
        f"- Check citation guide: {_citation_guide(cite_style)}\n"
        "- Flag any citations without corresponding reference entries\n"
        "- Add '(Author, Year)' placeholders for claims that need citations\n"
        "- Ensure citation density is appropriate for academic work\n\n"
        "### 3. ACADEMIC LANGUAGE & TONE\n"
        "- Replace colloquialisms with formal academic equivalents\n"
        "- Remove hedging where the evidence is strong; add hedging where speculative\n"
        "- Eliminate redundant phrases and wordiness\n"
        "- Ensure consistent terminology throughout\n"
        "- Verify key terms are defined on first use\n\n"
        "### 4. FORMATTING & STRUCTURE\n"
        "- Ensure consistent heading hierarchy (# -> ## -> ###)\n"
        "- Add missing front-matter sections for theses\n"
        "- Format any tables to be markdown-readable\n"
        "- Add figure placement markers where images are referenced\n"
        "- Ensure abstract meets word limit requirements\n\n"
        "### 5. BIBLIOGRAPHY\n"
        f"- Format all references in proper {cite_style} style\n"
        "- Sort alphabetically by first author surname\n"
        "- Add DOIs where available\n"
        "- Flag any reference that is incomplete\n\n"
        "### 6. FINAL CHECK\n"
        "- Verify document title matches metadata\n"
        "- Check author name, affiliation, and submission date on title page\n"
        "- Ensure keywords are listed after abstract if required\n"
        "- Remove any placeholder text or [TODO] markers\n\n"
        "## Output\n"
        "Return the COMPLETE polished document in markdown. "
        "Do NOT remove content — only improve, format, and complete. "
        "Prefix your output with `---POLISHED---` on its own line."
    )


def extract_polished_document(polisher_output: str) -> str:
    """Extract the polished document markdown from polisher LLM output."""
    marker = "---POLISHED---"
    idx = polisher_output.find(marker)
    if idx >= 0:
        return polisher_output[idx + len(marker):].strip()
    # Fallback: return whole output, stripping common prefix artifacts
    cleaned = re.sub(r'^[#\s]*POLISHED\s+', '', polisher_output).strip()
    return cleaned or polisher_output