"""
Video Story Generator — scene-by-scene pipeline:
  user prompt → prompt polish → shared cast/world sheets (once) → videos
"""
from __future__ import annotations

import base64
import json
import logging
import mimetypes
import os
import re
import shutil
import subprocess
import time
import uuid
from datetime import datetime
from typing import Any, Dict, List, Optional, Tuple
from urllib.parse import quote_plus

import requests

from .config import settings
from .llm_factory import LLMFactory, LLMProvider
from .llm_langchain_wrapper import LangChainLLMWrapper
from .models import (
    VideoStoryCastMember,
    VideoStoryImageAsset,
    VideoStoryProfile,
    VideoStoryScene,
    VideoStorySceneStatus,
    VideoStoryStatus,
    VideoStoryWorld,
)
from .video_story_manager import VideoStoryManager

logger = logging.getLogger(__name__)

_ROLE_PROMPT = "prompt"
_PROVIDER_ORDER = ("gemini", "qwen", "groq", "mistral")


def _api_key_for_provider(provider_str: str) -> str:
    p = provider_str.lower()
    if p == "gemini":
        return (settings.gemini_api_key or "").strip()
    if p == "qwen":
        return (settings.qwen_api_key or "").strip()
    if p == "mistral":
        return (settings.mistral_api_key or "").strip()
    if p == "groq":
        return (getattr(settings, "groq_api_key", "") or "").strip()
    return ""


def _create_prompt_llm() -> Tuple[Any, str, str]:
    api_key = (settings.video_story_prompt_api_key or "").strip()
    base_url = (settings.video_story_prompt_base_url or "").strip().rstrip("/")
    model = (settings.video_story_prompt_model or settings.gemini_default_model).strip()
    provider = (settings.video_story_prompt_provider or settings.default_llm_provider or "gemini").lower()

    if api_key and base_url:
        from qwen_caller import QwenCaller

        caller = QwenCaller(api_key=api_key, model=model, base_url=base_url, temperature=0.6, max_tokens=4096)
        return LangChainLLMWrapper(llm_caller=caller), model, f"openai-compatible@{base_url}"

    if api_key:
        prov = LLMProvider(provider)
        caller_class = LLMFactory._callers[prov]
        kwargs: Dict[str, Any] = {"temperature": 0.6, "max_tokens": 4096, "timeout": settings.api_timeout}
        if provider == "qwen":
            kwargs["base_url"] = settings.qwen_base_url
        caller = caller_class(api_key=api_key, model=model, **kwargs)
        return LangChainLLMWrapper(llm_caller=caller), model, provider

    for p in [provider] + [x for x in _PROVIDER_ORDER if x != provider]:
        key = _api_key_for_provider(p)
        if not key:
            continue
        prov = LLMProvider(p)
        caller_class = LLMFactory._callers[prov]
        m = model if p == provider else getattr(settings, f"{p}_default_model", settings.default_model)
        kwargs = {"temperature": 0.6, "max_tokens": 4096, "timeout": settings.api_timeout}
        if p == "qwen":
            kwargs["base_url"] = settings.qwen_base_url
        caller = caller_class(api_key=key, model=m, **kwargs)
        return LangChainLLMWrapper(llm_caller=caller), m, p

    raise ValueError(
        "No LLM configured for Video Story prompt polishing. "
        "Set VIDEO_STORY_PROMPT_API_KEY (+ BASE_URL) or global provider keys."
    )


def _invoke_llm(llm: Any, prompt: str) -> str:
    if hasattr(llm, "invoke"):
        out = llm.invoke(prompt)
        return out if isinstance(out, str) else str(out)
    if hasattr(llm, "generate"):
        return llm.generate(prompt)
    raise ValueError("LLM has no invoke/generate method")


def _extract_json(text: str) -> Optional[dict]:
    text = (text or "").strip()
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass
    start, end = text.find("{"), text.rfind("}")
    if start >= 0 and end > start:
        try:
            return json.loads(text[start : end + 1])
        except json.JSONDecodeError:
            return None
    return None


_IDENTITY_BLOCK_RE = re.compile(
    r"\n*=== LOCKED IDENTITY \(do not change\) ===\n.*?=== END LOCKED IDENTITY ===\n*",
    re.DOTALL,
)


def _scene_action_text(scene: VideoStoryScene) -> str:
    """Return the action portion of a polished prompt (strip locked identity block)."""
    text = (scene.polished_prompt or scene.user_prompt or "").strip()
    if not text:
        return ""
    text = _IDENTITY_BLOCK_RE.sub("\n", text).strip()
    # Repeated locking used to stack these headers; drop every leading copy.
    while text.upper().startswith("SCENE ACTION:"):
        text = text.split(":", 1)[-1].strip()
    return text


def _first_sentence(text: str, max_chars: int) -> str:
    flat = " ".join((text or "").split())
    if not flat:
        return ""
    first = re.split(r"(?<=[.!?])\s+", flat)[0]
    if len(first) > max_chars:
        first = first[:max_chars].rsplit(" ", 1)[0].rstrip(",;:") + "."
    return first


def _trim_to_words(text: str, max_words: int) -> str:
    words = text.split()
    if len(words) <= max_words:
        return text
    clipped = " ".join(words[:max_words])
    for sep in (". ", "! ", "? "):
        idx = clipped.rfind(sep)
        if idx > len(clipped) * 0.6:
            return clipped[: idx + 1].strip()
    return clipped


def _relevant_cast(
    project: VideoStoryProfile, scene: VideoStoryScene, action_text: str
) -> List[VideoStoryCastMember]:
    """Cast members actually present in this beat, so identity text stays short."""
    lowered = (action_text or "").lower()
    declared = {str(n).strip().lower() for n in (scene.character_descriptions or []) if str(n).strip()}
    selected: List[VideoStoryCastMember] = []
    for member in project.cast or []:
        keys = {
            n.strip().lower()
            for n in [member.name, *(member.aliases or [])]
            if n and n.strip()
        }
        if (keys & declared) or any(len(k) > 2 and k in lowered for k in keys):
            selected.append(member)
    if selected:
        return selected
    return [m for m in (project.cast or []) if m.is_primary]


def _video_prompt_max_words() -> int:
    return max(40, int(getattr(settings, "video_story_video_prompt_max_words", 100) or 100))


def _build_video_submit_prompt(project: VideoStoryProfile, scene: VideoStoryScene) -> str:
    """Compact single-paragraph prompt for the video API (hard word cap).

    Prefer action text. Append only a short primary-character cue if room remains —
    the I2V reference image already carries look/world for continuing scenes.
    """
    action = " ".join(_scene_action_text(scene).split())
    if not action:
        return ""

    max_words = _video_prompt_max_words()
    prompt = _trim_to_words(action, max_words)

    # Tiny identity cue only when budget remains (primary first).
    for member in sorted(
        (m for m in _relevant_cast(project, scene, action) if (m.canonical_description or "").strip()),
        key=lambda m: 0 if m.is_primary else 1,
    )[:2]:
        bit = f"{member.name}: {_first_sentence(member.canonical_description, 70).rstrip('.')}"
        candidate = f"{prompt} {bit}."
        if len(candidate.split()) > max_words:
            break
        prompt = candidate

    style = _first_sentence(project.style_bible or "", 60)
    if style:
        candidate = f"{prompt} Style: {style}"
        if len(candidate.split()) <= max_words:
            prompt = candidate
    return _trim_to_words(prompt, max_words)


def _previous_scene(project: VideoStoryProfile, scene: VideoStoryScene) -> Optional[VideoStoryScene]:
    ordered = sorted(project.scenes, key=lambda s: s.order)
    prev = None
    for s in ordered:
        if s.id == scene.id:
            return prev
        prev = s
    # Fallback by order if id not found yet
    earlier = [s for s in ordered if s.order < scene.order]
    return earlier[-1] if earlier else None


def polish_scene(project: VideoStoryProfile, scene: VideoStoryScene) -> VideoStoryScene:
    """Expand a short user query into a clip-length video prompt with continuity."""
    llm, _, _ = _create_prompt_llm()
    default_dur = float(getattr(settings, "video_story_clip_seconds", 5.0) or 5.0)
    duration = float(scene.duration_seconds or default_dur)
    if duration <= 0:
        duration = default_dur
    scene.duration_seconds = duration

    prev = _previous_scene(project, scene)
    continue_from_prev = bool(scene.continue_from_previous) and prev is not None
    if scene.order <= 1:
        continue_from_prev = False
        scene.continue_from_previous = False

    locked_cast = ""
    if project.cast:
        locked_cast = "\n".join(
            f"- {m.name}: {m.canonical_description}" for m in project.cast if m.canonical_description
        )
    locked_world = (project.world.canonical_description if project.world else "") or ""
    locked_style = (project.style_bible or "").strip()

    prev_block = ""
    if prev:
        prev_action = _trim_to_words(_scene_action_text(prev), _video_prompt_max_words())
        prev_query = (prev.user_prompt or "").strip()
        if continue_from_prev:
            prev_block = (
                "CONTINUITY MODE — this beat continues directly from the previous scene.\n"
                "Keep the same characters, wardrobe, location continuity, and camera energy "
                "unless the short query clearly requires a small change.\n"
                f"Previous scene {prev.order} short query: {prev_query or '(none)'}\n"
                f"Previous scene polished action (end state to continue from):\n{prev_action or '(none)'}\n"
                "Start this clip as a seamless continuation of that end state "
                "(same subject still in frame / immediately after).\n"
            )
        else:
            cut_note = (scene.cut_note or "").strip()
            prev_block = (
                "HARD CUT — user requested a break from the previous scene.\n"
                f"Previous scene {prev.order} short query (context only): {prev_query or '(none)'}\n"
                f"Cut note: {cut_note or '(none — new beat / location / focus)'}\n"
                "Do not seamlessly continue the previous shot; establish the new beat clearly.\n"
            )

    max_words = _video_prompt_max_words()
    prompt = (
        "You are a video story prompt engineer for short clip generation.\n"
        f"TARGET CLIP LENGTH: {duration:g} seconds. The polished prompt MUST describe only what can "
        f"realistically happen in about {duration:g} seconds — one clear beat, not a long montage.\n"
        f"HARD WORD LIMIT: polished_prompt MUST be at most {max_words} words. Count carefully. "
        "Prefer concrete action + camera over adjectives. One paragraph only.\n"
        "The user provides a VERY SHORT query (e.g. \"he runs on the street\"). Expand it into a "
        "tight video-generation prompt with camera, motion, and timing that fit the clip length.\n"
        "Do NOT invent new character looks if a locked cast is provided — reuse those identities "
        "for face/hair/outfit only by brief name cues (e.g. 'Tiger-man'), not full bios.\n"
        "Do NOT paste locked cast/world blocks into polished_prompt; only describe scene ACTION "
        "and camera. Identity is injected later separately.\n\n"
        f"Overall story context:\n{(project.story_context or '')[:2000]}\n\n"
        + (f"Locked cast names (reference only):\n{locked_cast}\n\n" if locked_cast else "")
        + (f"Locked world (reference only):\n{locked_world[:800]}\n\n" if locked_world else "")
        + (f"Locked style bible:\n{locked_style[:400]}\n\n" if locked_style else "")
        + (prev_block + "\n" if prev_block else "This is the first scene — establish the opening beat.\n\n")
        + f"Scene {scene.order} title: {scene.title or f'Scene {scene.order}'}\n"
        f"User short query: {scene.user_prompt.strip()}\n"
        f"continue_from_previous: {continue_from_prev}\n"
        f"duration_seconds: {duration:g}\n"
        f"max_words: {max_words}\n\n"
        "Respond with JSON only:\n"
        "{\n"
        f'  "polished_prompt": "ONE paragraph, ≤{max_words} words, {duration:g}s action+camera only",\n'
        '  "character_names": ["cast names appearing in this beat"],\n'
        '  "scenery_notes": "brief location notes for this beat only",\n'
        '  "continuity_summary": "one sentence: how this connects to previous (or hard cut)"\n'
        "}"
    )
    raw = _invoke_llm(llm, prompt)
    data = _extract_json(raw) or {}
    polished = str(data.get("polished_prompt") or scene.user_prompt).strip()
    polished = " ".join(polished.split())
    # Ensure duration cue is present for the video model, then enforce the hard cap.
    if f"{duration:g}-second" not in polished.lower() and f"{duration:g} second" not in polished.lower():
        polished = f"A {duration:g}-second continuous shot. {polished}"
    polished = _trim_to_words(polished, max_words)
    scene.polished_prompt = polished
    names = data.get("character_names") or []
    scene.character_descriptions = [str(n) for n in names if n]
    scene.scenery_description = str(data.get("scenery_notes") or "").strip() or None
    continuity = str(data.get("continuity_summary") or "").strip()
    if continuity:
        scene.notes = continuity
    scene.images = [img for img in (scene.images or []) if img.filename]
    scene.status = VideoStorySceneStatus.PROMPT_POLISHED
    scene.error = None
    return scene


def _polish_plain_text(llm: Any, instruction: str, text: str) -> str:
    prompt = (
        f"{instruction}\n\n"
        f"Source text:\n{text.strip()}\n\n"
        "Return only the improved text — no JSON, no markdown fences, no preamble."
    )
    out = _invoke_llm(llm, prompt).strip()
    if out.startswith("```"):
        out = re.sub(r"^```[a-z]*\n?", "", out)
        out = re.sub(r"\n?```$", "", out).strip()
    return out


def polish_content_field(
    project: VideoStoryProfile,
    field: str,
    *,
    cast_member_id: Optional[str] = None,
    source_text: Optional[str] = None,
) -> VideoStoryProfile:
    """AI-polish optional story / cast / world copy before image or video generation."""
    llm, _, _ = _create_prompt_llm()
    field = (field or "").strip().lower()

    if field == "story_context":
        text = (source_text if source_text is not None else project.story_context or "").strip()
        if not text:
            raise ValueError("Story context is empty")
        project.story_context = _polish_plain_text(
            llm,
            "Expand this video story premise into 2–5 vivid sentences for AI video generation. "
            "Keep the same characters, tone, and intent. Be concrete about setting and conflict.",
            text,
        )
        return project

    if field == "description":
        text = (source_text if source_text is not None else project.description or "").strip()
        if not text:
            raise ValueError("Description is empty")
        project.description = _polish_plain_text(
            llm,
            "Polish this short project description for a video story series. "
            "One short paragraph, clear and production-ready.",
            text,
        )
        return project

    if field == "style_bible":
        text = (source_text if source_text is not None else project.style_bible or "").strip()
        if not text:
            raise ValueError("Style bible is empty")
        project.style_bible = _polish_plain_text(
            llm,
            "Polish this visual style bible for consistent AI-generated video. "
            "Cover medium, palette, lighting, and camera language in 2–4 sentences.",
            text,
        )
        return project

    if field == "cast_description":
        if not cast_member_id:
            raise ValueError("cast_member_id required")
        member = next((m for m in project.cast if m.id == cast_member_id), None)
        if not member:
            raise ValueError("Cast member not found")
        text = (source_text if source_text is not None else member.canonical_description or "").strip()
        if not text:
            raise ValueError("Character description is empty")
        member.canonical_description = _polish_plain_text(
            llm,
            f"Polish this locked character look description for '{member.name or 'character'}'. "
            "Same identity every scene: face, hair, outfit, body type. 2–4 sentences, no scene action.",
            text,
        )
        if member.image and not (member.image.prompt or "").strip():
            member.image.prompt = (
                f"Character reference sheet, full body and clear face, neutral pose, plain background. "
                f"{member.name}: {member.canonical_description}"
            )
        return project

    if field == "world_description":
        world = project.world or VideoStoryWorld()
        text = (source_text if source_text is not None else world.canonical_description or "").strip()
        if not text:
            raise ValueError("World description is empty")
        world.canonical_description = _polish_plain_text(
            llm,
            "Polish this locked world / environment description for consistent AI video. "
            "Establishing location only — no characters, no plot. 2–4 sentences.",
            text,
        )
        if world.image and not (world.image.prompt or "").strip():
            world.image.prompt = (
                f"Environment establishing plate, no readable text. {world.canonical_description}"
            )
        project.world = world
        return project

    raise ValueError(f"Unknown field: {field}")


def _build_identity_block(project: VideoStoryProfile) -> str:
    cast_lines = []
    for m in project.cast or []:
        desc = (m.canonical_description or "").strip()
        if not desc:
            continue
        cast_lines.append(f"- {m.name}: {desc}")
    world = (project.world.canonical_description if project.world else "") or ""
    style = (project.style_bible or "").strip()
    parts = ["=== LOCKED IDENTITY (do not change) ==="]
    if cast_lines:
        parts.append("CAST (same face/hair/outfit/body across every scene):")
        parts.extend(cast_lines)
    if world.strip():
        parts.append("WORLD / ENVIRONMENT (keep palette and architecture consistent):")
        parts.append(world.strip())
    if style:
        parts.append("STYLE BIBLE:")
        parts.append(style)
    parts.append(
        "Hard rules: do not redesign characters; do not change face, hair, clothing, age, or body type; "
        "do not invent alternate outfits unless the scene explicitly requires a temporary prop; "
        "keep the same world palette and materials."
    )
    parts.append("=== END LOCKED IDENTITY ===")
    return "\n".join(parts)


def extract_story_bible(project: VideoStoryProfile, *, force: bool = False) -> VideoStoryProfile:
    """Extract a stable cast + world + style bible from the whole story (once)."""
    if (
        not force
        and project.cast
        and project.world
        and (project.world.canonical_description or "").strip()
        and (project.style_bible or "").strip()
    ):
        return project

    llm, _, _ = _create_prompt_llm()
    scene_block = "\n\n".join(
        f"Scene {s.order} — {s.title}\n"
        f"User: {s.user_prompt}\n"
        f"Polished: {(s.polished_prompt or '')[:800]}"
        for s in sorted(project.scenes, key=lambda x: x.order)
        if (s.user_prompt or "").strip()
    )
    prompt = (
        "You are a showrunner locking visual continuity for an animated/live-action video series.\n"
        "From the story context and ALL scenes, extract ONE stable cast list and ONE world description.\n"
        "Rules:\n"
        "- Same person across scenes = ONE cast member (dedupe nicknames/aliases).\n"
        "- Each character gets one detailed, reusable visual description (face, hair, age, body, "
        "signature outfit, colors). Do not write scene-specific actions into the description.\n"
        "- World description covers setting, architecture, palette, lighting mood — not one-off props.\n"
        "- style_bible is short: medium, color palette, camera language.\n"
        "- Mark the main protagonist with is_primary=true (exactly one if possible).\n\n"
        f"Story title: {project.title}\n"
        f"Story context:\n{project.story_context[:4000]}\n\n"
        f"Scenes:\n{scene_block[:12000]}\n\n"
        "Respond with JSON only:\n"
        "{\n"
        '  "style_bible": "short locked look",\n'
        '  "world": {"canonical_description": "..."},\n'
        '  "cast": [\n'
        '    {"name": "...", "aliases": ["..."], "canonical_description": "...", "is_primary": true}\n'
        "  ]\n"
        "}"
    )
    raw = _invoke_llm(llm, prompt)
    data = _extract_json(raw) or {}

    # Preserve already-generated image filenames when re-extracting.
    prev_by_name = {
        (m.name or "").strip().lower(): m for m in (project.cast or []) if (m.name or "").strip()
    }
    prev_world_image = project.world.image if project.world else None

    cast: List[VideoStoryCastMember] = []
    for i, item in enumerate(data.get("cast") or []):
        if not isinstance(item, dict):
            continue
        name = str(item.get("name") or f"Character {i + 1}").strip()
        key = name.lower()
        prev = prev_by_name.get(key)
        aliases = [str(a) for a in (item.get("aliases") or []) if a]
        member = VideoStoryCastMember(
            id=prev.id if prev else str(uuid.uuid4()),
            name=name,
            canonical_description=str(item.get("canonical_description") or "").strip(),
            aliases=aliases,
            image=prev.image if prev else None,
            is_primary=bool(item.get("is_primary")),
        )
        cast.append(member)

    if cast and not any(m.is_primary for m in cast):
        cast[0].is_primary = True

    world_data = data.get("world") if isinstance(data.get("world"), dict) else {}
    world_desc = str(world_data.get("canonical_description") or "").strip()
    project.cast = cast
    project.world = VideoStoryWorld(
        canonical_description=world_desc,
        image=prev_world_image,
    )
    project.style_bible = str(data.get("style_bible") or project.style_bible or "").strip()
    return project


def lock_scene_prompts(project: VideoStoryProfile) -> VideoStoryProfile:
    """Inject the locked identity block into every scene polished prompt."""
    block = _build_identity_block(project)
    for scene in project.scenes:
        base = (scene.polished_prompt or scene.user_prompt or "").strip()
        if not base:
            continue
        base = _IDENTITY_BLOCK_RE.sub("\n", base).strip()
        # Idempotent: re-locking must not stack another SCENE ACTION header.
        while base.upper().startswith("SCENE ACTION:"):
            base = base.split(":", 1)[-1].strip()
        scene.polished_prompt = f"{block}\n\nSCENE ACTION:\n{base}"
        # Drop pending per-scene image jobs; shared assets are the source of truth.
        scene.images = [img for img in (scene.images or []) if img.filename]
        has_sheets = bool(_primary_cast_member(project)) or bool(
            project.world and project.world.image and project.world.image.filename
        )
        if has_sheets and scene.status in (
            VideoStorySceneStatus.DRAFT,
            VideoStorySceneStatus.PROMPT_POLISHED,
            VideoStorySceneStatus.FAILED,
        ):
            scene.status = VideoStorySceneStatus.IMAGES_READY
    return project


def generate_shared_assets(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    *,
    include_characters: bool = True,
    include_scenery: bool = True,
    force_regenerate: bool = False,
    cast_member_id: Optional[str] = None,
    regen_world: bool = False,
) -> VideoStoryProfile:
    """Generate one sheet per cast member + one world sheet (resume-friendly).

    force_regenerate clears existing filenames for the selected scope before generating.
    cast_member_id / regen_world narrow the scope to a single asset.
    """
    project.status = VideoStoryStatus.GENERATING_IMAGES
    shared_dir = os.path.join(manager.project_dir(project.id), "shared")
    delay = float(getattr(settings, "video_story_image_request_delay", 1.5) or 0)
    errors: List[str] = []
    generated = 0

    only_one = bool(cast_member_id) or regen_world
    do_chars = include_characters and (not only_one or bool(cast_member_id))
    do_world = include_scenery and (not only_one or regen_world)

    def _clear_image(asset: Optional[VideoStoryImageAsset]) -> None:
        if not asset or not asset.filename:
            return
        old = os.path.join(shared_dir, asset.filename)
        asset.filename = None
        try:
            if os.path.isfile(old):
                os.remove(old)
        except OSError:
            logger.warning("Could not remove old shared asset file %s", old)

    if force_regenerate or cast_member_id or regen_world:
        if do_chars:
            for member in project.cast or []:
                if cast_member_id and member.id != cast_member_id:
                    continue
                if force_regenerate or cast_member_id:
                    _clear_image(member.image)
        if do_world and project.world and (force_regenerate or regen_world):
            _clear_image(project.world.image)

    jobs: List[Tuple[str, VideoStoryImageAsset, str]] = []
    # (label, asset, dest_subdir_filename_prefix)

    if do_chars:
        for member in project.cast or []:
            if cast_member_id and member.id != cast_member_id:
                continue
            desc = (member.canonical_description or "").strip()
            custom_prompt = ((member.image.prompt if member.image else "") or "").strip()
            if not desc and not custom_prompt:
                continue
            if member.image and member.image.filename:
                continue
            prompt = custom_prompt or (
                f"Character reference sheet, full body and clear face, neutral pose, "
                f"consistent design bible, plain background. {member.name}: {desc}"
            )
            asset = VideoStoryImageAsset(
                id=(member.image.id if member.image else str(uuid.uuid4())),
                asset_type="character",
                prompt=prompt,
                description=member.name,
            )
            member.image = asset
            jobs.append((f"cast:{member.name}", asset, f"cast_{_slug(member.name)}"))

    if do_world and project.world:
        world_desc = (project.world.canonical_description or "").strip()
        custom_prompt = ((project.world.image.prompt if project.world.image else "") or "").strip()
        if (world_desc or custom_prompt) and not (project.world.image and project.world.image.filename):
            prompt = custom_prompt or (
                f"Environment / world establishing plate, no readable text, "
                f"consistent series location. {world_desc}"
            )
            asset = VideoStoryImageAsset(
                id=(project.world.image.id if project.world.image else str(uuid.uuid4())),
                asset_type="scenery",
                prompt=prompt,
                description="World",
            )
            project.world.image = asset
            jobs.append(("world", asset, "world"))

    for i, (label, asset, prefix) in enumerate(jobs):
        if i > 0 and delay > 0:
            time.sleep(delay)
        fname = f"{prefix}_{uuid.uuid4().hex[:8]}.png"
        path = os.path.join(shared_dir, fname)
        result = _generate_image_to_path(asset.prompt, path)
        if result.get("success"):
            asset.filename = fname
            generated += 1
            try:
                manager.save_project(project)
            except Exception:
                logger.exception("Failed to persist shared asset %s", label)
        else:
            err = str(result.get("error") or "Shared image generation failed")
            errors.append(f"{label}: {err}")
            logger.warning("Shared asset failed (%s): %s", label, err)
            if "429" in err or "rate" in err.lower():
                time.sleep(max(delay, 5.0))

    cast_ready = sum(1 for m in (project.cast or []) if m.image and m.image.filename)
    world_ready = bool(project.world and project.world.image and project.world.image.filename)
    if errors and generated == 0:
        project.status = VideoStoryStatus.FAILED
        project.metadata = dict(project.metadata or {})
        project.metadata["shared_assets_error"] = errors[-1]
    else:
        project.status = VideoStoryStatus.GENERATING_IMAGES if (cast_ready or world_ready) else project.status
        if errors:
            project.metadata = dict(project.metadata or {})
            project.metadata["shared_assets_error"] = "; ".join(errors[-3:])
        elif "shared_assets_error" in (project.metadata or {}):
            project.metadata = dict(project.metadata or {})
            project.metadata.pop("shared_assets_error", None)

    # Mark scenes that can proceed once sheets exist.
    if cast_ready or world_ready:
        for scene in project.scenes:
            if scene.user_prompt.strip() and scene.status != VideoStorySceneStatus.VIDEO_READY:
                if scene.status != VideoStorySceneStatus.FAILED or not scene.video_filename:
                    scene.status = VideoStorySceneStatus.IMAGES_READY
                    if scene.error and "rate" in (scene.error or "").lower():
                        scene.error = None
    return project


def prepare_shared_cast_and_world(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    *,
    include_characters: bool = True,
    include_scenery: bool = True,
    force_bible: bool = False,
    force_regenerate: bool = False,
    cast_member_id: Optional[str] = None,
    regen_world: bool = False,
) -> VideoStoryProfile:
    """Main image step: extract bible → generate shared sheets → lock scene prompts."""
    # Ensure scenes have at least a base polished prompt before bible extraction.
    for scene in sorted(project.scenes, key=lambda s: s.order):
        if scene.user_prompt.strip() and not (scene.polished_prompt or "").strip():
            polish_scene(project, scene)

    # Do not re-extract bible when regenerating from user-edited prompts,
    # unless explicitly requested or the bible is still missing.
    needs_bible = force_bible or not project.cast or not (
        project.world and (project.world.canonical_description or "").strip()
    )
    skip_extract = (cast_member_id or regen_world or force_regenerate) and not force_bible and not needs_bible
    if not skip_extract:
        extract_story_bible(project, force=force_bible or needs_bible)
        manager.save_project(project)

    if cast_member_id and not any(m.id == cast_member_id for m in (project.cast or [])):
        raise ValueError(f"Cast member not found: {cast_member_id}")

    generate_shared_assets(
        manager,
        project,
        include_characters=include_characters,
        include_scenery=include_scenery,
        force_regenerate=force_regenerate,
        cast_member_id=cast_member_id,
        regen_world=regen_world,
    )
    lock_scene_prompts(project)
    manager.save_project(project)
    return project


def _slug(text: str) -> str:
    base = re.sub(r"[^a-zA-Z0-9]+", "_", (text or "").strip().lower()).strip("_")
    return (base or "asset")[:40]


def _shared_image_path(manager: VideoStoryManager, project_id: str, filename: Optional[str]) -> Optional[str]:
    if not filename:
        return None
    path = os.path.join(manager.project_dir(project_id), "shared", filename)
    return path if os.path.isfile(path) else None


def _primary_cast_member(project: VideoStoryProfile) -> Optional[VideoStoryCastMember]:
    cast = project.cast or []
    for m in cast:
        if m.is_primary and m.image and m.image.filename:
            return m
    for m in cast:
        if m.image and m.image.filename:
            return m
    return None


def _video_frames_dir(manager: VideoStoryManager, project_id: str) -> str:
    d = os.path.join(manager.project_dir(project_id), "videos", "frames")
    os.makedirs(d, exist_ok=True)
    return d


def _run_ffprobe_field(video_path: str, entries: str) -> Optional[str]:
    if not shutil.which("ffprobe"):
        return None
    cmd = [
        "ffprobe",
        "-v",
        "error",
        "-select_streams",
        "v:0",
        "-show_entries",
        entries,
        "-of",
        "default=noprint_wrappers=1:nokey=1",
        video_path,
    ]
    try:
        proc = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if proc.returncode != 0:
            return None
        line = (proc.stdout or "").strip().splitlines()
        return line[0].strip() if line else None
    except (subprocess.TimeoutExpired, OSError):
        return None


def _probe_video_fps(video_path: str) -> Optional[float]:
    raw = _run_ffprobe_field(video_path, "stream=r_frame_rate")
    if not raw or raw == "0/0":
        return None
    if "/" in raw:
        num, den = raw.split("/", 1)
        try:
            den_f = float(den)
            return float(num) / den_f if den_f else None
        except ValueError:
            return None
    try:
        return float(raw)
    except ValueError:
        return None


def _probe_video_size(video_path: str) -> Tuple[Optional[int], Optional[int]]:
    w_raw = _run_ffprobe_field(video_path, "stream=width")
    h_raw = _run_ffprobe_field(video_path, "stream=height")
    try:
        w = int(w_raw) if w_raw else None
        h = int(h_raw) if h_raw else None
        return w, h
    except ValueError:
        return None, None


def _extract_last_frame(video_path: str, dest_path: str) -> bool:
    """Grab the true final frame of an mp4.

    Uses ``-update 1`` so every decoded frame overwrites the same PNG and the
    final one wins. Seeking to a computed timestamp is unreliable: a target past
    the last frame's PTS makes ffmpeg exit 0 while writing nothing.
    """
    if not shutil.which("ffmpeg"):
        logger.error("ffmpeg not found — install ffmpeg for video scene continuity")
        return False
    parent = os.path.dirname(dest_path)
    if parent:
        os.makedirs(parent, exist_ok=True)

    # Tail decode first, then whole-file decode if the container resists seeking.
    attempts = [
        ["ffmpeg", "-y", "-sseof", "-1", "-i", video_path, "-update", "1", dest_path],
        ["ffmpeg", "-y", "-i", video_path, "-update", "1", dest_path],
    ]
    for cmd in attempts:
        try:
            if os.path.isfile(dest_path):
                os.remove(dest_path)
        except OSError:
            pass
        try:
            proc = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
        except (subprocess.TimeoutExpired, OSError) as exc:
            logger.warning("ffmpeg last-frame extract error for %s: %s", video_path, exc)
            continue
        if proc.returncode == 0 and os.path.isfile(dest_path) and os.path.getsize(dest_path) > 0:
            return True
        logger.warning(
            "ffmpeg last-frame extract produced no frame for %s (rc=%s): %s",
            video_path,
            proc.returncode,
            (proc.stderr or "")[-400:],
        )
    return False


def _prepend_reference_hold(video_path: str, ref_image_path: str, *, hold_frames: int) -> bool:
    """Prepend exact reference frames so the clip joins pixel-perfect with the previous scene."""
    if hold_frames <= 0 or not shutil.which("ffmpeg"):
        return True
    w, h = _probe_video_size(video_path)
    fps = _probe_video_fps(video_path) or 24.0
    if not w or not h:
        logger.warning("Could not probe video size for continuity hold: %s", video_path)
        return False

    hold_seconds = hold_frames / max(fps, 1.0)
    tmp_out = f"{video_path}.hold.mp4"
    filter_complex = (
        f"[0:v]scale={w}:{h}:force_original_aspect_ratio=increase,crop={w}:{h},"
        f"fps={fps:.3f},trim=duration={hold_seconds:.6f},setpts=PTS-STARTPTS,format=yuv420p[hold];"
        f"[1:v]fps={fps:.3f},format=yuv420p,setpts=PTS-STARTPTS[main];"
        f"[hold][main]concat=n=2:v=1:a=0[out]"
    )
    cmd = [
        "ffmpeg",
        "-y",
        "-loop",
        "1",
        "-i",
        ref_image_path,
        "-i",
        video_path,
        "-filter_complex",
        filter_complex,
        "-map",
        "[out]",
        "-c:v",
        "libx264",
        "-preset",
        "fast",
        "-crf",
        "18",
        "-pix_fmt",
        "yuv420p",
        "-an",
        tmp_out,
    ]
    try:
        proc = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
        if proc.returncode != 0 or not os.path.isfile(tmp_out):
            logger.warning(
                "Continuity hold prepend failed for %s: %s",
                video_path,
                (proc.stderr or "")[-500:],
            )
            try:
                if os.path.isfile(tmp_out):
                    os.remove(tmp_out)
            except OSError:
                pass
            return False
        os.replace(tmp_out, video_path)
        return True
    except (subprocess.TimeoutExpired, OSError) as exc:
        logger.warning("Continuity hold prepend error for %s: %s", video_path, exc)
        try:
            if os.path.isfile(tmp_out):
                os.remove(tmp_out)
        except OSError:
            pass
        return False


def _remove_last_frame_file(manager: VideoStoryManager, project_id: str, filename: Optional[str]) -> None:
    if not filename:
        return
    path = os.path.join(_video_frames_dir(manager, project_id), filename)
    try:
        if os.path.isfile(path):
            os.remove(path)
    except OSError:
        logger.warning("Could not remove last-frame file %s", path)


def _ensure_scene_last_frame(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    scene: VideoStoryScene,
    video_path: Optional[str] = None,
) -> Optional[str]:
    """Extract and cache the last frame of a scene video; return the PNG path."""
    video_dir = os.path.join(manager.project_dir(project.id), "videos")
    if not video_path:
        if not scene.video_filename:
            return None
        video_path = os.path.join(video_dir, scene.video_filename)
    if not os.path.isfile(video_path):
        return None

    frames_dir = _video_frames_dir(manager, project.id)
    video_stem = os.path.splitext(os.path.basename(video_path))[0]
    expected_name = f"scene_{scene.order}_last_{video_stem}.png"

    if scene.last_frame_filename == expected_name:
        existing = os.path.join(frames_dir, expected_name)
        if os.path.isfile(existing) and os.path.getsize(existing) > 0:
            video_mtime = os.path.getmtime(video_path)
            frame_mtime = os.path.getmtime(existing)
            if frame_mtime >= video_mtime:
                return existing

    old_frame = scene.last_frame_filename
    dest = os.path.join(frames_dir, expected_name)
    if not _extract_last_frame(video_path, dest):
        return None

    scene.last_frame_filename = expected_name
    if old_frame and old_frame != expected_name:
        _remove_last_frame_file(manager, project.id, old_frame)
    return dest


def _previous_scene_last_frame_path(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    scene: VideoStoryScene,
) -> Tuple[Optional[str], Optional[VideoStoryScene]]:
    """Nearest earlier scene that has a usable video, plus its last frame.

    Walks backwards so one failed scene does not reset the whole chain back to
    the cast sheet.
    """
    video_dir = os.path.join(manager.project_dir(project.id), "videos")
    earlier = sorted(
        [s for s in project.scenes if s.order < scene.order],
        key=lambda s: s.order,
        reverse=True,
    )
    for candidate in earlier:
        if not candidate.video_filename:
            continue
        candidate_video = os.path.join(video_dir, candidate.video_filename)
        if not os.path.isfile(candidate_video):
            continue
        frame = _ensure_scene_last_frame(manager, project, candidate, candidate_video)
        if frame:
            return frame, candidate
    return None, None


def _cast_or_world_image_path(manager: VideoStoryManager, project: VideoStoryProfile) -> Optional[str]:
    primary = _primary_cast_member(project)
    if primary and primary.image:
        path = _shared_image_path(manager, project.id, primary.image.filename)
        if path:
            return path
    if project.world and project.world.image:
        path = _shared_image_path(manager, project.id, project.world.image.filename)
        if path:
            return path
    return None


def _resolve_i2v_image_path(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    scene: VideoStoryScene,
    *,
    prefer_last_frame: Optional[bool] = None,
) -> Optional[str]:
    """Pick I2V seed image.

    Default (and the path that worked before continuity work): shared cast/world sheet.
    Optional last-frame seeding is behind ``VIDEO_STORY_I2V_USE_LAST_FRAME`` because
    SiliconFlow frequently returns Failed with an empty reason on video-frame inputs.
    """
    use_last = (
        bool(prefer_last_frame)
        if prefer_last_frame is not None
        else bool(getattr(settings, "video_story_i2v_use_last_frame", False))
    )
    if use_last and scene.order > 1 and scene.continue_from_previous:
        frame_path, source = _previous_scene_last_frame_path(manager, project, scene)
        if frame_path and source:
            logger.info(
                "Scene %s I2V: using last frame from scene %s", scene.order, source.order
            )
            return frame_path
        logger.warning(
            "Scene %s last-frame unavailable — falling back to cast/world",
            scene.order,
        )

    image_path = _cast_or_world_image_path(manager, project)
    if image_path:
        logger.info("Scene %s I2V: using shared cast/world sheet", scene.order)
        return image_path

    ordered_images = sorted(
        [img for img in (scene.images or []) if img.filename],
        key=lambda img: 0 if img.asset_type == "character" else 1 if img.asset_type == "scenery" else 2,
    )
    for img in ordered_images:
        candidate = os.path.join(manager.project_dir(project.id), "images", img.filename)
        if os.path.isfile(candidate):
            logger.info("Scene %s I2V: using legacy per-scene image", scene.order)
            return candidate
    return None


def _normalize_siliconflow_images_url(base_url: str) -> str:
    """Accept host, /v1, or full /images/generations and return the generations URL."""
    base = (base_url or "https://api.siliconflow.cn/v1").strip().rstrip("/")
    if base.endswith("/images/generations"):
        return base
    if base.endswith("/v1"):
        return f"{base}/images/generations"
    return f"{base}/v1/images/generations"


def _extract_siliconflow_image(data: Dict[str, Any]) -> Tuple[Optional[str], Optional[str]]:
    """Return (url, b64) from a SiliconFlow / OpenAI-style images response."""
    candidates = data.get("images") or data.get("data") or []
    if isinstance(candidates, dict):
        candidates = [candidates]
    for item in candidates:
        if not isinstance(item, dict):
            continue
        url = item.get("url") or item.get("image_url")
        if url:
            return str(url), None
        b64 = item.get("b64_json") or item.get("base64")
        if b64:
            return None, str(b64)
    url = data.get("url") or data.get("image_url")
    if url:
        return str(url), None
    return None, None


def _generate_image_siliconflow(prompt: str, dest_path: str) -> Dict[str, Any]:
    """SiliconFlow text-to-image via POST /v1/images/generations."""
    api_key = (settings.video_story_image_api_key or "").strip()
    if not api_key:
        return {"success": False, "error": "Set VIDEO_STORY_IMAGE_API_KEY for the SiliconFlow image provider"}

    api_url = _normalize_siliconflow_images_url(settings.video_story_image_base_url)
    configured_model = (settings.video_story_image_model or "").strip()
    model = configured_model if configured_model and configured_model != "flux" else "Kwai-Kolors/Kolors"
    payload: Dict[str, Any] = {
        "model": model,
        "prompt": prompt,
        "image_size": settings.video_story_image_size or "1024x1024",
        "batch_size": max(1, int(settings.video_story_image_batch_size or 1)),
        "num_inference_steps": int(settings.video_story_image_num_inference_steps or 20),
        "guidance_scale": float(settings.video_story_image_guidance_scale or 7.5),
    }
    headers = {"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"}
    max_retries = max(1, int(getattr(settings, "video_story_image_max_retries", 5) or 5))
    last_error = ""

    for attempt in range(max_retries):
        try:
            resp = requests.post(api_url, json=payload, headers=headers, timeout=180)
        except requests.RequestException as e:
            last_error = f"SiliconFlow image request failed ({api_url}): {e}"
            # brief backoff for transient network/DNS issues
            time.sleep(min(30.0, 2.0 * (attempt + 1)))
            continue

        if resp.status_code == 429:
            # SiliconFlow IPM / rate limit — wait and retry.
            retry_after = resp.headers.get("Retry-After")
            try:
                wait_s = float(retry_after) if retry_after else min(60.0, 5.0 * (attempt + 1))
            except ValueError:
                wait_s = min(60.0, 5.0 * (attempt + 1))
            last_error = f"SiliconFlow image rate-limited (HTTP 429): {resp.text[:200]}"
            logger.warning("%s — retrying in %.1fs (%s/%s)", last_error, wait_s, attempt + 1, max_retries)
            time.sleep(wait_s)
            continue

        if resp.status_code != 200:
            return {
                "success": False,
                "error": f"SiliconFlow image error: HTTP {resp.status_code}: {resp.text[:300]}",
            }

        try:
            data = resp.json() if resp.content else {}
        except ValueError:
            return {"success": False, "error": "SiliconFlow image API returned invalid JSON"}

        url, b64 = _extract_siliconflow_image(data if isinstance(data, dict) else {})
        if url:
            try:
                img = requests.get(url, timeout=120)
            except requests.RequestException as e:
                return {"success": False, "error": f"SiliconFlow image download failed: {e}"}
            if img.status_code != 200:
                return {"success": False, "error": f"SiliconFlow image download failed: HTTP {img.status_code}"}
            with open(dest_path, "wb") as f:
                f.write(img.content)
            return {"success": True, "url": url}

        if b64:
            if b64.startswith("data:"):
                b64 = b64.split(",", 1)[-1]
            try:
                raw = base64.b64decode(b64)
            except ValueError:
                return {"success": False, "error": "SiliconFlow image API returned invalid base64 data"}
            with open(dest_path, "wb") as f:
                f.write(raw)
            return {"success": True}

        return {"success": False, "error": "SiliconFlow image API returned no image URL or base64 data"}

    return {"success": False, "error": last_error or "SiliconFlow image request failed after retries"}


def _generate_image_to_path(prompt: str, dest_path: str) -> Dict[str, Any]:
    provider = (settings.video_story_image_provider or "pollinations").lower()
    if provider in ("siliconflow", "silicon", "silicon_flow"):
        return _generate_image_siliconflow(prompt, dest_path)

    if provider == "pollinations":
        model = settings.video_story_image_model or "flux"
        encoded = quote_plus(prompt.strip())
        image_url = f"https://gen.pollinations.ai/image/{encoded}?model={model}"
        headers = {"User-Agent": "GroundControl/1.0"}
        key = (settings.video_story_image_api_key or settings.pollinations_api_key or "").strip()
        if key:
            headers["Authorization"] = f"Bearer {key}"
        try:
            resp = requests.get(image_url, headers=headers, timeout=120)
        except requests.RequestException as e:
            return {"success": False, "error": f"Image request failed: {e}"}
        if resp.status_code != 200:
            return {"success": False, "error": f"Image download failed: HTTP {resp.status_code}"}
        ctype = resp.headers.get("content-type", "")
        if "image" not in ctype:
            return {
                "success": False,
                "error": f"Image provider returned non-image response ({ctype or 'unknown'}): "
                f"{resp.text[:200]}",
            }
        with open(dest_path, "wb") as f:
            f.write(resp.content)
        return {"success": True, "url": image_url}

    api_url = (settings.video_story_image_base_url or "").strip()
    api_key = (settings.video_story_image_api_key or "").strip()
    if not api_url:
        return {"success": False, "error": "Set VIDEO_STORY_IMAGE_BASE_URL for custom image provider"}
    headers = {"Content-Type": "application/json"}
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"
    payload = {"prompt": prompt, "model": settings.video_story_image_model or "flux"}
    try:
        resp = requests.post(api_url, json=payload, headers=headers, timeout=180)
    except requests.RequestException as e:
        return {"success": False, "error": f"Image API request failed: {e}"}
    if resp.status_code != 200:
        return {"success": False, "error": f"Image API error: HTTP {resp.status_code}: {resp.text[:200]}"}
    ctype = resp.headers.get("content-type", "")
    if "image" in ctype:
        with open(dest_path, "wb") as f:
            f.write(resp.content)
        return {"success": True}
    try:
        data = resp.json() if resp.content else {}
    except ValueError:
        return {"success": False, "error": "Image API returned neither an image nor valid JSON"}
    url = data.get("url") or data.get("image_url")
    if url:
        img = requests.get(url, timeout=120)
        if img.status_code == 200:
            with open(dest_path, "wb") as f:
                f.write(img.content)
            return {"success": True, "url": url}
    return {"success": False, "error": "Image API returned no image"}


def generate_scene_images(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    scene: VideoStoryScene,
    include_characters: bool = True,
    include_scenery: bool = True,
) -> VideoStoryScene:
    if not scene.polished_prompt and scene.user_prompt:
        scene = polish_scene(project, scene)

    base = os.path.join(manager.project_dir(project.id), "images")
    assets: List[VideoStoryImageAsset] = list(scene.images or [])

    if not assets:
        if include_characters:
            for i, desc in enumerate(scene.character_descriptions or []):
                assets.append(
                    VideoStoryImageAsset(
                        id=str(uuid.uuid4()),
                        asset_type="character",
                        prompt=desc,
                        description=f"Character {i + 1}",
                    )
                )
        if include_scenery and scene.scenery_description:
            assets.append(
                VideoStoryImageAsset(
                    id=str(uuid.uuid4()),
                    asset_type="scenery",
                    prompt=scene.scenery_description,
                    description="Scene environment",
                )
            )

    delay = float(getattr(settings, "video_story_image_request_delay", 1.5) or 0)
    pending = [
        a
        for a in assets
        if not a.filename
        and a.prompt.strip()
        and not (a.asset_type == "character" and not include_characters)
        and not (a.asset_type == "scenery" and not include_scenery)
    ]
    errors: List[str] = []
    for i, asset in enumerate(pending):
        if i > 0 and delay > 0:
            time.sleep(delay)
        fname = f"img_{scene.order}_{asset.asset_type}_{uuid.uuid4().hex[:8]}.png"
        path = os.path.join(base, fname)
        result = _generate_image_to_path(asset.prompt, path)
        if result.get("success"):
            asset.filename = fname
        else:
            err = str(result.get("error") or "Image generation failed")
            errors.append(err)
            logger.warning("Scene %s asset %s failed: %s", scene.order, asset.asset_type, err)
            # Keep going so later assets / resume can still succeed; stop only on hard failures.
            if "429" in err or "rate-limited" in err.lower() or "rate limiting" in err.lower():
                # After retries still rate-limited — pause briefly then continue remaining.
                time.sleep(max(delay, 5.0))

    scene.images = assets
    ready = [a for a in assets if a.filename]
    if pending and not ready:
        scene.error = errors[-1] if errors else "Image generation failed"
        scene.status = VideoStorySceneStatus.FAILED
    elif errors:
        scene.error = f"Partial image generation ({len(ready)} ready): {errors[-1]}"
        scene.status = VideoStorySceneStatus.IMAGES_READY if ready else VideoStorySceneStatus.FAILED
    else:
        scene.error = None
        scene.status = VideoStorySceneStatus.IMAGES_READY if ready or not pending else VideoStorySceneStatus.FAILED
    return scene


def _image_to_data_uri(
    image_path: str,
    *,
    max_side: int = 0,
    jpeg_quality: int = 0,
    as_png: bool = False,
) -> Optional[str]:
    """Encode a local image as a ``data:<mime>;base64,<...>`` URI.

    JPEG recompression is fine for cast sheets, but SiliconFlow often rejects
    last-frame video stills when sent as JPEG (empty Failed reason). Those
    should be sent as PNG via ``as_png=True``.
    """
    try:
        if as_png or max_side > 0 or jpeg_quality > 0:
            from io import BytesIO
            from PIL import Image

            with Image.open(image_path) as im:
                im = im.convert("RGB")
                if max_side > 0 and max(im.size) > max_side:
                    im.thumbnail((max_side, max_side), Image.Resampling.LANCZOS)
                buf = BytesIO()
                if as_png:
                    im.save(buf, format="PNG", optimize=True)
                    mime = "image/png"
                else:
                    im.save(buf, format="JPEG", quality=jpeg_quality or 85, optimize=True)
                    mime = "image/jpeg"
                raw = buf.getvalue()
        else:
            with open(image_path, "rb") as f:
                raw = f.read()
            mime = mimetypes.guess_type(image_path)[0] or "image/png"
    except OSError as e:
        logger.warning("Could not read image for base64 encoding (%s): %s", image_path, e)
        return None
    except Exception as e:
        logger.warning("Could not compress image for base64 encoding (%s): %s", image_path, e)
        try:
            with open(image_path, "rb") as f:
                raw = f.read()
            mime = mimetypes.guess_type(image_path)[0] or "image/png"
        except OSError as e2:
            logger.warning("Fallback image read failed (%s): %s", image_path, e2)
            return None
    b64 = base64.b64encode(raw).decode("utf-8")
    return f"data:{mime};base64,{b64}"


def _siliconflow_failure_reason(data: Dict[str, Any]) -> str:
    """Extract the most useful failure text from a SiliconFlow status payload."""
    for key in ("reason", "message", "error", "detail"):
        val = data.get(key)
        if isinstance(val, str) and val.strip() and val.strip().lower() not in ("string", "null", "none"):
            return val.strip()
    results = data.get("results")
    if isinstance(results, dict):
        for key in ("reason", "message", "error"):
            val = results.get(key)
            if isinstance(val, str) and val.strip():
                return val.strip()
    return (
        "provider returned Failed with no reason (usually a transient capacity or "
        "content-filter rejection) — retry this scene"
    )


def _download_video(url: str, dest_path: str, headers: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
    try:
        vid = requests.get(url, headers=headers or {}, timeout=600)
    except requests.RequestException as e:
        return {"success": False, "error": f"Video download failed: {e}"}
    if vid.status_code != 200:
        return {"success": False, "error": f"Video download failed: HTTP {vid.status_code}"}
    with open(dest_path, "wb") as f:
        f.write(vid.content)
    return {"success": True, "url": url}


_SUPPORTED_VIDEO_SIZES = {"1280x720": 16 / 9, "720x1280": 9 / 16, "960x960": 1.0}


def _image_size_for_reference(image_path: Optional[str]) -> Optional[str]:
    """Match output size to the reference image aspect, as the provider does."""
    if not image_path or not os.path.isfile(image_path):
        return None
    try:
        from PIL import Image

        with Image.open(image_path) as im:
            width, height = im.size
    except Exception as exc:
        logger.warning("Could not read reference image size (%s): %s", image_path, exc)
        return None
    if not width or not height:
        return None
    ratio = width / height
    return min(_SUPPORTED_VIDEO_SIZES, key=lambda key: abs(ratio - _SUPPORTED_VIDEO_SIZES[key]))


def _normalize_siliconflow_video_base_url(base_url: str) -> str:
    base = (base_url or "https://api.siliconflow.cn/v1").strip().rstrip("/")
    if base.endswith("/video/submit"):
        return base[: -len("/video/submit")]
    if base.endswith("/v1"):
        return base
    return f"{base}/v1"


def _resolve_siliconflow_video_model(configured: str, has_image: bool) -> str:
    """Pick a valid SiliconFlow Wan model for T2V vs I2V.

    Docs support:
      - Wan-AI/Wan2.2-I2V-A14B (image required)
      - Wan-AI/Wan2.2-T2V-A14B (text only)
    """
    model = (configured or "").strip()
    lower = model.lower()
    is_i2v = "i2v" in lower
    is_t2v = "t2v" in lower
    if has_image:
        if is_t2v and not is_i2v:
            # Scene images exist — use I2V even if env still points at a T2V id.
            return "Wan-AI/Wan2.2-I2V-A14B"
        return model or "Wan-AI/Wan2.2-I2V-A14B"
    if is_i2v and not is_t2v:
        return "Wan-AI/Wan2.2-T2V-A14B"
    if model in ("turbo", "default", "flux"):
        return "Wan-AI/Wan2.2-T2V-A14B"
    return model or "Wan-AI/Wan2.2-T2V-A14B"


def _generate_video_siliconflow(prompt: str, dest_path: str, image_path: Optional[str] = None) -> Dict[str, Any]:
    """SiliconFlow video via async submit → poll status → download.

    Uses I2V when a scene image is available (base64 data URI), otherwise T2V.
    Retries once on provider Failed / rate-limit style responses.
    """
    api_key = (settings.video_story_video_api_key or "").strip()
    if not api_key:
        return {"success": False, "error": "Set VIDEO_STORY_VIDEO_API_KEY for the SiliconFlow video provider"}

    configured_base = (settings.video_story_video_base_url or "").strip()
    if not configured_base:
        # Allow VIDEO_STORY_VIDEO_API_URL to hold the API host when it is not an image URL.
        api_url_fallback = (settings.video_story_video_api_url or "").strip()
        lower = api_url_fallback.lower()
        looks_like_image = lower.startswith("data:image/") or any(
            lower.rstrip("?/").endswith(ext) for ext in (".png", ".jpg", ".jpeg", ".webp", ".gif")
        )
        if api_url_fallback and not looks_like_image:
            configured_base = api_url_fallback
    base_url = _normalize_siliconflow_video_base_url(configured_base)

    image_ref: Optional[str] = None
    if image_path and os.path.isfile(image_path):
        is_last_frame = "/frames/" in image_path.replace("\\", "/") and "_last_" in os.path.basename(
            image_path
        )
        if is_last_frame:
            # Video stills fail as JPEG on SiliconFlow; PNG is accepted.
            image_ref = _image_to_data_uri(image_path, max_side=768, as_png=True)
        else:
            image_ref = _image_to_data_uri(image_path, max_side=768, jpeg_quality=85)
    else:
        fallback = (settings.video_story_video_api_url or "").strip()
        lower = fallback.lower()
        if lower.startswith("data:image/") or (
            lower.startswith(("http://", "https://"))
            and any(lower.rstrip("?/").endswith(ext) for ext in (".png", ".jpg", ".jpeg", ".webp", ".gif"))
        ):
            image_ref = fallback

    model = _resolve_siliconflow_video_model(settings.video_story_video_model or "", bool(image_ref))
    is_i2v = "i2v" in model.lower()

    # Prefer explicit config (pre-continuity behavior used 1280x720); aspect match is fallback.
    configured_size = (settings.video_story_video_image_size or "").strip()
    if configured_size not in _SUPPORTED_VIDEO_SIZES:
        configured_size = ""
    image_size = configured_size or _image_size_for_reference(image_path) or "1280x720"

    safe_prompt = (prompt or "").strip()

    headers = {"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"}
    max_attempts = max(1, int(getattr(settings, "video_story_video_max_attempts", 3) or 3))
    last_error = ""

    for attempt in range(max_attempts):
        payload: Dict[str, Any] = {
            "model": model,
            "prompt": safe_prompt,
            "image_size": image_size,
        }
        if settings.video_story_video_negative_prompt:
            payload["negative_prompt"] = settings.video_story_video_negative_prompt
        if settings.video_story_video_seed:
            payload["seed"] = settings.video_story_video_seed

        if is_i2v:
            if not image_ref:
                return {
                    "success": False,
                    "error": (
                        f"SiliconFlow model {model} requires an image. "
                        "Generate Cast & world sheets first, or set "
                        "VIDEO_STORY_VIDEO_MODEL=Wan-AI/Wan2.2-T2V-A14B."
                    ),
                }
            payload["image"] = image_ref

        try:
            resp = requests.post(f"{base_url}/video/submit", json=payload, headers=headers, timeout=120)
        except requests.RequestException as e:
            last_error = f"SiliconFlow submit failed: {e}"
            time.sleep(3.0 * (attempt + 1))
            continue
        if resp.status_code == 429:
            last_error = f"SiliconFlow submit rate-limited: {resp.text[:200]}"
            time.sleep(min(60.0, 8.0 * (attempt + 1)))
            continue
        if resp.status_code != 200:
            return {
                "success": False,
                "error": (
                    f"SiliconFlow submit error: HTTP {resp.status_code} "
                    f"(model={model}): {resp.text[:300]}"
                ),
            }
        try:
            request_id = (resp.json() or {}).get("requestId")
        except ValueError:
            return {"success": False, "error": "SiliconFlow submit returned invalid JSON"}
        if not request_id:
            return {"success": False, "error": "SiliconFlow submit returned no requestId"}

        interval = max(2, int(settings.video_story_video_poll_interval or 5))
        deadline = time.monotonic() + max(interval, int(settings.video_story_video_poll_timeout or 600))
        last_status = ""
        while time.monotonic() < deadline:
            time.sleep(interval)
            try:
                status_resp = requests.post(
                    f"{base_url}/video/status", json={"requestId": request_id}, headers=headers, timeout=60
                )
            except requests.RequestException as e:
                logger.warning("SiliconFlow status poll failed (retrying): %s", e)
                continue
            if status_resp.status_code != 200:
                logger.warning("SiliconFlow status HTTP %s: %s", status_resp.status_code, status_resp.text[:200])
                continue
            try:
                data = status_resp.json() or {}
            except ValueError:
                continue
            last_status = str(data.get("status") or "")
            if last_status == "Succeed":
                results = data.get("results") or {}
                videos = results.get("videos") or []
                url = videos[0].get("url") if videos and isinstance(videos[0], dict) else None
                url = url or data.get("url")
                if not url:
                    return {"success": False, "error": "SiliconFlow reported success but returned no video URL"}
                return _download_video(url, dest_path)
            if last_status == "Failed":
                reason = _siliconflow_failure_reason(data if isinstance(data, dict) else {})
                last_error = f"SiliconFlow video generation failed: {reason}"
                logger.warning(
                    "SiliconFlow video Failed (attempt %s/%s, requestId=%s): %s",
                    attempt + 1,
                    max_attempts,
                    request_id,
                    last_error,
                )
                break

        else:
            last_error = f"SiliconFlow video timed out (last status: {last_status or 'unknown'})"
            break

        if attempt + 1 < max_attempts:
            time.sleep(min(45.0, 8.0 * (attempt + 1)))

    return {"success": False, "error": last_error or "SiliconFlow video generation failed"}


def _generate_video_to_path(prompt: str, dest_path: str, image_path: Optional[str] = None) -> Dict[str, Any]:
    provider = (settings.video_story_video_provider or "pollinations").lower()

    if provider in ("siliconflow", "silicon", "silicon_flow"):
        return _generate_video_siliconflow(prompt, dest_path, image_path=image_path)

    headers = {"User-Agent": "GroundControl/1.0"}
    api_key = (settings.video_story_video_api_key or settings.pollinations_api_key or "").strip()
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"

    if provider == "pollinations":
        model = settings.video_story_video_model or "turbo"
        encoded = quote_plus(prompt.strip())
        video_url = f"https://gen.pollinations.ai/video/{encoded}?model={model}"
        try:
            resp = requests.get(video_url, headers=headers, timeout=600)
        except requests.RequestException as e:
            return {"success": False, "error": f"Video request failed: {e}"}
        ctype = resp.headers.get("content-type", "")
        if resp.status_code != 200 or "video" not in ctype:
            return {
                "success": False,
                "error": (
                    f"Pollinations video request failed (HTTP {resp.status_code}, {ctype or 'unknown'}). "
                    "Pollinations does not offer a stable video endpoint — set "
                    "VIDEO_STORY_VIDEO_PROVIDER=http and VIDEO_STORY_VIDEO_API_URL to a real "
                    "video generation API."
                ),
            }
        with open(dest_path, "wb") as f:
            f.write(resp.content)
        return {"success": True, "url": video_url}

    api_url = (settings.video_story_video_api_url or "").strip()
    if not api_url:
        return {
            "success": False,
            "error": "Set VIDEO_STORY_VIDEO_API_URL for custom video provider (or use pollinations)",
        }

    if image_path and os.path.isfile(image_path):
        with open(image_path, "rb") as img_f:
            files = {"image": img_f}
            data = {"prompt": prompt, "model": settings.video_story_video_model or "default"}
            resp = requests.post(api_url, data=data, files=files, headers=headers, timeout=600)
    else:
        payload = {"prompt": prompt, "model": settings.video_story_video_model or "default"}
        headers["Content-Type"] = "application/json"
        resp = requests.post(api_url, json=payload, headers=headers, timeout=600)

    if resp.status_code != 200:
        return {"success": False, "error": f"Video API error: HTTP {resp.status_code}: {resp.text[:300]}"}

    ctype = resp.headers.get("content-type", "")
    if "video" in ctype or "octet-stream" in ctype:
        with open(dest_path, "wb") as f:
            f.write(resp.content)
        return {"success": True}

    try:
        data = resp.json()
    except ValueError:
        return {
            "success": False,
            "error": f"Video API returned unexpected content-type '{ctype}' and no JSON",
        }
    output = data.get("output")
    output_url = output.get("url") if isinstance(output, dict) else None
    url = data.get("url") or data.get("video_url") or output_url
    if url:
        vid = requests.get(url, headers=headers, timeout=600)
        if vid.status_code == 200:
            with open(dest_path, "wb") as f:
                f.write(vid.content)
            return {"success": True, "url": url}
    return {"success": False, "error": "Video API returned no video content"}


def _scene_video_path(manager: VideoStoryManager, project_id: str, filename: Optional[str]) -> Optional[str]:
    if not filename:
        return None
    path = os.path.join(manager.project_dir(project_id), "videos", filename)
    return path if os.path.isfile(path) and os.path.getsize(path) > 0 else None


def _latest_on_disk_scene_video(
    manager: VideoStoryManager, project_id: str, scene_order: int
) -> Optional[str]:
    """Recover the newest scene_N_*.mp4 left on disk after a failed regen."""
    video_dir = os.path.join(manager.project_dir(project_id), "videos")
    if not os.path.isdir(video_dir):
        return None
    prefix = f"scene_{scene_order}_"
    candidates = [
        name
        for name in os.listdir(video_dir)
        if name.startswith(prefix) and name.endswith(".mp4")
    ]
    existing = [
        name
        for name in candidates
        if os.path.isfile(os.path.join(video_dir, name))
        and os.path.getsize(os.path.join(video_dir, name)) > 0
    ]
    if not existing:
        return None
    existing.sort(key=lambda name: os.path.getmtime(os.path.join(video_dir, name)), reverse=True)
    return existing[0]


def sanitize_scene_video_refs(manager: VideoStoryManager, project: VideoStoryProfile) -> VideoStoryProfile:
    """Drop or repair video_filename entries that no longer exist on disk."""
    for scene in project.scenes or []:
        if not scene.video_filename:
            recovered = _latest_on_disk_scene_video(manager, project.id, scene.order)
            if recovered:
                scene.video_filename = recovered
                scene.status = VideoStorySceneStatus.VIDEO_READY
                if scene.error and "no credit" not in (scene.error or "").lower():
                    scene.error = None
            continue
        if _scene_video_path(manager, project.id, scene.video_filename):
            if scene.status != VideoStorySceneStatus.VIDEO_READY:
                scene.status = VideoStorySceneStatus.VIDEO_READY
            continue
        recovered = _latest_on_disk_scene_video(manager, project.id, scene.order)
        if recovered:
            logger.warning(
                "Scene %s video %s missing — recovered %s",
                scene.order,
                scene.video_filename,
                recovered,
            )
            scene.video_filename = recovered
            scene.status = VideoStorySceneStatus.VIDEO_READY
            continue
        logger.warning(
            "Scene %s video %s missing on disk — clearing reference",
            scene.order,
            scene.video_filename,
        )
        scene.video_filename = None
        scene.last_frame_filename = None
        if scene.status == VideoStorySceneStatus.VIDEO_READY:
            scene.status = VideoStorySceneStatus.IMAGES_READY
    return project


def generate_scene_video(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    scene: VideoStoryScene,
    *,
    force: bool = False,
) -> VideoStoryScene:
    video_dir = os.path.join(manager.project_dir(project.id), "videos")
    sanitize_scene_video_refs(manager, project)

    # Resume: keep an existing good video unless force regenerate.
    if not force and scene.video_filename:
        existing = _scene_video_path(manager, project.id, scene.video_filename)
        if existing:
            _ensure_scene_last_frame(manager, project, scene, existing)
            scene.status = VideoStorySceneStatus.VIDEO_READY
            scene.error = None
            return scene

    # Keep the previous filename until the new clip succeeds so Preview never 404s.
    previous_video = scene.video_filename
    previous_path = _scene_video_path(manager, project.id, previous_video)
    if force:
        scene.error = None

    # Ensure locked identity is present before video submit.
    if project.cast or (project.world and (project.world.canonical_description or "").strip()):
        if "=== LOCKED IDENTITY" not in (scene.polished_prompt or ""):
            lock_scene_prompts(project)

    # The stored polished prompt keeps the full locked-identity scaffolding for the UI;
    # the video model gets a compact, single-paragraph version instead.
    prompt = _build_video_submit_prompt(project, scene)
    if not prompt:
        scene.error = "Scene has no prompt — polish first"
        scene.status = VideoStorySceneStatus.FAILED
        return scene
    logger.info(
        "Scene %s video prompt: %s words, %s chars",
        scene.order,
        len(prompt.split()),
        len(prompt),
    )

    fname = f"scene_{scene.order}_{uuid.uuid4().hex[:8]}.mp4"
    dest = os.path.join(video_dir, fname)

    use_last = bool(getattr(settings, "video_story_i2v_use_last_frame", False))
    image_path = _resolve_i2v_image_path(manager, project, scene, prefer_last_frame=use_last)
    used_last_frame = bool(
        image_path
        and "/frames/" in image_path.replace("\\", "/")
        and "_last_" in os.path.basename(image_path)
    )

    result = _generate_video_to_path(prompt, dest, image_path=image_path)
    # Last-frame I2V is frequently rejected by SiliconFlow with an empty Failed reason.
    # Fall back once to the cast/world sheet (the path that worked before continuity).
    if not result.get("success") and used_last_frame:
        cast_path = _resolve_i2v_image_path(manager, project, scene, prefer_last_frame=False)
        if cast_path and cast_path != image_path:
            logger.warning(
                "Scene %s last-frame I2V failed (%s) — retrying with cast/world sheet",
                scene.order,
                result.get("error"),
            )
            try:
                if os.path.isfile(dest):
                    os.remove(dest)
            except OSError:
                pass
            image_path = cast_path
            used_last_frame = False
            result = _generate_video_to_path(prompt, dest, image_path=image_path)

    if not result.get("success"):
        # Drop a half-written destination if the provider failed after creating it.
        try:
            if os.path.isfile(dest) and os.path.getsize(dest) == 0:
                os.remove(dest)
        except OSError:
            pass
        scene.error = result.get("error")
        # Keep Preview usable: restore the previous on-disk video when present.
        if previous_path:
            scene.video_filename = previous_video
            scene.status = VideoStorySceneStatus.VIDEO_READY
            scene.error = f"Retry failed (kept previous video): {scene.error}"
            return scene
        scene.video_filename = None
        scene.status = VideoStorySceneStatus.FAILED
        return scene

    if not os.path.isfile(dest) or os.path.getsize(dest) <= 0:
        scene.error = "Video provider reported success but no file was written"
        if previous_path:
            scene.video_filename = previous_video
            scene.status = VideoStorySceneStatus.VIDEO_READY
            scene.error = f"Retry failed (kept previous video): {scene.error}"
            return scene
        scene.video_filename = None
        scene.status = VideoStorySceneStatus.FAILED
        return scene

    scene.video_filename = fname
    scene.status = VideoStorySceneStatus.VIDEO_READY
    scene.error = None
    hold_frames = int(getattr(settings, "video_story_continuity_hold_frames", 0) or 0)
    if hold_frames > 0 and used_last_frame:
        if not _prepend_reference_hold(dest, image_path, hold_frames=hold_frames):
            logger.warning("Scene %s: optional continuity hold prepend failed", scene.order)
    _ensure_scene_last_frame(manager, project, scene, dest)
    if force and previous_video and previous_video != fname and previous_path:
        try:
            os.remove(previous_path)
        except OSError:
            logger.warning("Could not remove old scene video %s", previous_path)
    return scene


def polish_all_scenes(project: VideoStoryProfile) -> VideoStoryProfile:
    project.status = VideoStoryStatus.POLISHING
    for scene in sorted(project.scenes, key=lambda s: s.order):
        if scene.user_prompt.strip():
            polish_scene(project, scene)
    return project


def generate_all_images(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    include_characters: bool = True,
    include_scenery: bool = True,
    **kwargs: Any,
) -> VideoStoryProfile:
    """Backward-compatible entry: build shared cast/world sheets once (not per scene)."""
    return prepare_shared_cast_and_world(
        manager,
        project,
        include_characters=include_characters,
        include_scenery=include_scenery,
        **kwargs,
    )


def generate_all_videos(
    manager: VideoStoryManager,
    project: VideoStoryProfile,
    *,
    force: bool = False,
) -> VideoStoryProfile:
    project.status = VideoStoryStatus.GENERATING_VIDEOS
    sanitize_scene_video_refs(manager, project)
    if project.cast or (project.world and (project.world.canonical_description or "").strip()):
        lock_scene_prompts(project)
    delay = float(getattr(settings, "video_story_video_request_delay", 3.0) or 0)
    ordered = sorted(
        [s for s in project.scenes if (s.user_prompt or "").strip()],
        key=lambda s: s.order,
    )

    def _run(scenes: List[VideoStoryScene], *, force_pass: bool) -> None:
        for idx, scene in enumerate(scenes):
            generate_scene_video(manager, project, scene, force=force_pass)
            try:
                manager.save_project(project)
            except Exception:
                logger.exception("Failed to persist after scene %s video", scene.order)
            if delay > 0 and idx < len(scenes) - 1:
                time.sleep(delay)

    _run(ordered, force_pass=force)

    # The provider fails intermittently with no reason, so sweep the stragglers.
    # Re-running in order also lets continuity re-link once a gap fills in.
    rounds = max(0, int(getattr(settings, "video_story_video_retry_rounds", 2) or 0))
    for round_no in range(rounds):
        pending = [
            s
            for s in ordered
            if s.status != VideoStorySceneStatus.VIDEO_READY
            or not _scene_video_path(manager, project.id, s.video_filename)
        ]
        if not pending:
            break
        logger.info(
            "Video retry sweep %s/%s for scenes %s",
            round_no + 1,
            rounds,
            [s.order for s in pending],
        )
        if delay > 0:
            time.sleep(delay)
        _run(pending, force_pass=False)

    ready = all(
        s.status == VideoStorySceneStatus.VIDEO_READY
        and _scene_video_path(manager, project.id, s.video_filename)
        for s in ordered
    )
    project.status = VideoStoryStatus.COMPLETED if ready else VideoStoryStatus.FAILED
    return project
