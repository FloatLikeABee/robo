from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import Field
from typing import Optional, List
import os


def _read_local_key(filename: str) -> str:
  """
  Read a secret key from a local file (one-line), returning empty string if missing.
  The file is expected to live at the project root and is gitignored.
  """
  try:
    # config.py lives in src/, project root is one level up
    root_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    path = os.path.join(root_dir, filename)
    with open(path, "r", encoding="utf-8") as f:
      return f.read().strip()
  except OSError:
    return ""


class Settings(BaseSettings):
    """Loads env vars from the process environment and, if present, `.env` then `env` at the project root."""
    model_config = SettingsConfigDict(
        env_file=(".env", "env"),
        env_file_encoding="utf-8",
        extra="ignore",
    )

    # Database settings
    chroma_persist_directory: str = "./chroma_db"
    data_directory: str = "./data"

    # LLM Provider settings
    # Options: gemini, qwen, mistral, groq
    default_llm_provider: str = "gemini"
    default_model: str = "gemini-2.5-flash"

    # Gemini settings
    # Default: read from local gitignored file `gemini_key` at project root.
    # Can still be overridden via environment variable GEMINI_API_KEY / .env.
    gemini_api_key: str = _read_local_key("gemini_key")
    gemini_default_model: str = "gemini-2.5-flash"

    # Qwen settings (set QWEN_API_KEY in .env or environment)
    qwen_api_key: str = ""
    qwen_base_url: str = "https://dashscope.aliyuncs.com/compatible-mode/v1"
    qwen_default_model: str = "qwen3.7-plus"

    # Mistral settings (set MISTRAL_API_KEY in .env or environment)
    mistral_api_key: str = ""
    mistral_default_model: str = "mistral-large-latest"

    # Groq settings — set GROQ_API_KEY in `.env`, `env`, or the environment (do not commit real keys).
    groq_api_key: str = ""
    groq_default_model: str = "llama-3.3-70b-versatile"

    # Embedding settings
    embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2"
    hf_ssl_verify: bool = True  # Set to False to disable SSL verification for HuggingFace downloads (development only)
    hf_download_timeout: int = 300  # Timeout in seconds for HuggingFace model downloads
    hf_proxy: Optional[str] = None  # Proxy URL for HuggingFace downloads (e.g., "http://proxy.example.com:8080" or "https://proxy.example.com:8080")
    hf_mirror: Optional[str] = None  # Mirror endpoint for HuggingFace (e.g., "https://hf-mirror.com" for China mirror)

    # API settings
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    debug_mode: bool = Field(True, validation_alias="DEBUG")  # .env can use DEBUG=true (or DEBUG_MODE=true)
    api_timeout: int = 180  # Timeout in seconds for LLM API calls (increased for Qwen compatibility)

    # CORS settings - allow all by default
    # You can override these via environment variables if needed.
    # `cors_origins` is kept for compatibility but we primarily rely on `cors_origin_regex`
    # so that the server echoes back the actual Origin instead of "*".
    cors_origins: List[str] = Field(
        default_factory=lambda: [
            "http://localhost:3000",
            "http://127.0.0.1:3000",
            "http://localhost:3031",
            "http://127.0.0.1:3031",
            "http://localhost:3040",
            "http://127.0.0.1:3040",
        ]
    )
    cors_allow_credentials: bool = True
    cors_origin_regex: str = ".*"

    # Email settings
    smtp_server: Optional[str] = None
    smtp_port: int = 587
    smtp_username: Optional[str] = None
    smtp_password: Optional[str] = None

    # Pollinations image generation (Graphic Document Generator, /images). Optional Bearer token.
    pollinations_api_key: str = ""

    # Financial API settings (set ALPHA_VANTAGE_API_KEY in .env)
    alpha_vantage_api_key: Optional[str] = None

    # Web Search API settings (optional - for Tavily API)
    # Tavily has a free tier: 128 searches/month
    # If not set, will use free search engines (unlimited)
    tavily_api_key: Optional[str] = None

    # Text-to-SQL default SQL Server connection (used when no db_tool_id / connection_config / connection_string provided)
    # Override via env: TEXT_TO_SQL_DEFAULT_HOST, TEXT_TO_SQL_DEFAULT_PORT, etc.
    text_to_sql_default_host: str = ""
    text_to_sql_default_port: int = 1433
    text_to_sql_default_database: str = ""
    text_to_sql_default_username: str = ""
    text_to_sql_default_password: str = ""

    # ScholarForge — academic article & thesis composer (three dedicated model roles)
    # Per-role overrides: set API_KEY (+ BASE_URL for OpenAI-compatible hosts e.g. SiliconFlow).
    # When API_KEY is set, that role uses these credentials instead of global provider keys.
    scholar_forge_organizer_provider: str = "gemini"
    scholar_forge_organizer_model: str = "gemini-2.5-flash"
    scholar_forge_organizer_api_key: str = ""
    scholar_forge_organizer_base_url: str = ""
    scholar_forge_writer_provider: str = "gemini"
    scholar_forge_writer_model: str = "gemini-2.5-flash"
    scholar_forge_writer_api_key: str = ""
    scholar_forge_writer_base_url: str = ""
    scholar_forge_image_provider: str = "qwen"
    scholar_forge_image_model: str = "qwen-vl-ocr-2025-11-20"
    scholar_forge_image_api_key: str = ""
    scholar_forge_image_base_url: str = ""
    scholar_forge_min_prompt_words: int = 100
    scholar_forge_output_dir: str = "scholar_forge_output"
    # Reviewer role — critiques each paragraph; falls back to organizer credentials when unset
    scholar_forge_reviewer_provider: str = "gemini"
    scholar_forge_reviewer_model: str = "gemini-2.5-flash"
    scholar_forge_reviewer_api_key: str = ""
    scholar_forge_reviewer_base_url: str = ""
    scholar_forge_max_review_rounds: int = 2
    scholar_forge_review_pass_score: int = 75
    # Polisher role — post-generation polish for thesis/article
    scholar_forge_polisher_provider: str = "gemini"
    scholar_forge_polisher_model: str = "gemini-2.5-flash"
    scholar_forge_polisher_api_key: str = ""
    scholar_forge_polisher_base_url: str = ""
    # Enhanced features
    scholar_forge_web_search_enabled: bool = True
    scholar_forge_use_enhanced_prompts: bool = True
    scholar_forge_polish_enabled: bool = True
    scholar_forge_max_reference_searches: int = 6

    # Video Story Generator — three dedicated model roles
    video_story_prompt_provider: str = "gemini"
    video_story_prompt_model: str = "gemini-2.5-flash"
    video_story_prompt_api_key: str = ""
    video_story_prompt_base_url: str = ""
    video_story_image_provider: str = "pollinations"
    video_story_image_model: str = "flux"
    video_story_image_api_key: str = ""
    video_story_image_base_url: str = ""
    # SiliconFlow text-to-image (POST /v1/images/generations)
    video_story_image_size: str = "1024x1024"
    video_story_image_batch_size: int = 1
    video_story_image_num_inference_steps: int = 20
    video_story_image_guidance_scale: float = 7.5
    video_story_image_request_delay: float = 1.5
    video_story_image_max_retries: int = 5
    video_story_video_provider: str = "pollinations"
    # SiliconFlow I2V default; T2V auto-selected when no scene image exists.
    video_story_video_model: str = "Wan-AI/Wan2.2-I2V-A14B"
    video_story_video_api_key: str = ""
    video_story_video_api_url: str = ""
    # SiliconFlow (async submit/status) video provider
    video_story_video_base_url: str = "https://api.siliconflow.cn/v1"
    video_story_video_image_size: str = "1280x720"
    video_story_video_negative_prompt: str = ""
    video_story_video_seed: int = 0
    video_story_video_poll_interval: int = 5
    video_story_video_poll_timeout: int = 600
    # Default clip length used when polishing short scene queries into video prompts
    video_story_clip_seconds: float = 5.0
    # Pause between scene video API calls (helps long stories avoid rate limits)
    video_story_video_request_delay: float = 3.0
    # Submit retries per scene; the provider intermittently returns Failed with no reason
    video_story_video_max_attempts: int = 3
    # SiliconFlow: keep video prompts short; hard cap for polish + submit
    video_story_video_prompt_max_words: int = 100
    # Extra passes over still-failed scenes at the end of a batch
    video_story_video_retry_rounds: int = 2
    # When true, continuing scenes seed I2V from the previous clip's last frame (PNG).
    # JPEG-encoded last frames are rejected by SiliconFlow; cast fallback still applies.
    video_story_i2v_use_last_frame: bool = True
    # Optional: prepend duplicate reference frames at start of continuing scenes (0 = off)
    video_story_continuity_hold_frames: int = 0
    video_story_output_dir: str = "video_stories"


settings = Settings()
