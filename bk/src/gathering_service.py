"""
Gathering Service: AI-powered data gathering using Wikipedia, Reddit (via web search),
and general web search. Uses preset system prompt and configurable limits.
"""
import logging
from typing import Dict, Any, Optional
from .config import settings
from .llm_factory import LLMFactory, LLMProvider
from .llm_langchain_wrapper import LangChainLLMWrapper
from .web_search_service import web_search, wikipedia_search
from .models import LLMProviderType


GATHERING_SYSTEM_PROMPT = """You are a research assistant that gathers information from multiple sources.

## Your Task
The user will give you a topic or question. You must gather comprehensive information using these resources IN ORDER:

1. **Wikipedia** (MOST IMPORTANT): Search Wikipedia first for factual overview, definitions, and established knowledge.
2. **Reddit** (IMPORTANT): Search Reddit for real-world discussions, opinions, and experiences. Use the Web Search tool with queries like "site:reddit.com [your topic]" to find Reddit content.
3. **Web Search**: Use general web search for additional sources, news, articles, and recent information.

## Instructions
- Use each resource type at least once when relevant.
- For Reddit: Always use Web Search with "site:reddit.com" in your query to find Reddit discussions.
- Synthesize findings into a clear, structured markdown report.
- Include key points, sources (Wikipedia, Reddit threads, URLs), and a brief summary.
- Be concise but thorough. Organize with headers (##, ###), bullet points, and numbered lists.
- If a source has no useful information, try a different search and move on. Do not repeat failed searches.
- When you have gathered enough to answer the user's question comprehensively, provide your final report.

## Output Format
Respond with a well-formatted markdown report. End with "Final Answer:" followed by your complete report.
"""


class GatheringService:
    """Service for AI-powered data gathering from Wikipedia, Reddit, and web."""

    def __init__(self):
        self.logger = logging.getLogger(__name__)

    async def gather(
        self,
        prompt: str,
        llm_provider: Optional[LLMProviderType] = None,
        model_name: Optional[str] = None,
        max_iterations: int = 10,
        max_tokens: int = 8192,
        temperature: float = 0.5,
    ) -> Dict[str, Any]:
        """
        Gather information using Wikipedia, Reddit (via web search), and web search.

        Args:
            prompt: User's topic or question to research
            llm_provider: Optional LLM provider (default: from settings)
            model_name: Optional model name (default: provider default)
            max_iterations: Max agent iterations (default: 10, prevents infinite search)
            max_tokens: Max response tokens (default: 8192)
            temperature: LLM temperature (default: 0.5 for balanced output)

        Returns:
            Dict with success, content (markdown), metadata, and optional error
        """
        try:
            provider_str = (
                llm_provider.value if llm_provider else (settings.default_llm_provider or "qwen")
            ).lower()

            if provider_str == "gemini":
                provider = LLMProvider.GEMINI
                api_key = settings.gemini_api_key
                model = model_name or settings.gemini_default_model
            elif provider_str == "qwen":
                provider = LLMProvider.QWEN
                api_key = settings.qwen_api_key
                model = model_name or settings.qwen_default_model
            elif provider_str == "mistral":
                provider = LLMProvider.MISTRAL
                api_key = settings.mistral_api_key
                model = model_name or settings.mistral_default_model
            elif provider_str == "groq":
                provider = LLMProvider.GROQ
                api_key = getattr(settings, "groq_api_key", "")
                model = model_name or getattr(settings, "groq_default_model", "llama-3.3-70b-versatile")
            else:
                provider = LLMProvider.QWEN
                api_key = settings.qwen_api_key
                model = model_name or settings.qwen_default_model

            if not api_key:
                return {
                    "success": False,
                    "error": f"{provider_str.capitalize()} API key not configured",
                    "content": "",
                }

            llm_caller = LLMFactory.create_caller(
                provider=provider,
                api_key=api_key,
                model=model,
                temperature=temperature,
                max_tokens=max_tokens,
            )
            llm = LangChainLLMWrapper(llm_caller=llm_caller)

            search_context_parts = []
            try:
                wiki_result = wikipedia_search(prompt)
                if wiki_result and not wiki_result.startswith("Error"):
                    search_context_parts.append(f"## Wikipedia\n\n{wiki_result}\n")
            except Exception as e:
                self.logger.warning(f"Wikipedia gathering failed: {e}")

            try:
                web_result = web_search(prompt, max_results=max_iterations)
                if web_result and not web_result.startswith("No web"):
                    search_context_parts.append(f"## Web Search\n\n{web_result}\n")
            except Exception as e:
                self.logger.warning(f"Web search gathering failed: {e}")

            try:
                reddit_result = web_search(f"site:reddit.com {prompt}", max_results=max_iterations)
                if reddit_result and not reddit_result.startswith("No web"):
                    search_context_parts.append(f"## Reddit Discussions\n\n{reddit_result}\n")
            except Exception as e:
                self.logger.warning(f"Reddit gathering failed: {e}")

            search_context = "\n\n".join(search_context_parts)
            full_prompt = (
                f"{GATHERING_SYSTEM_PROMPT}\n\n"
                f"Research topic: {prompt}\n\n"
                f"Gathered context:\n{search_context}\n\n"
                "Synthesize the above into a well-formatted markdown report. "
                "End with 'Final Answer:' followed by your complete report."
            )

            result = await llm_caller.generate_async(full_prompt)

            if isinstance(result, dict) and "output" in result:
                content = result["output"]
            elif isinstance(result, str):
                content = result
            else:
                content = str(result)

            if "Final Answer:" in content:
                content = content.split("Final Answer:")[-1].strip()
            elif search_context and not content.strip():
                # Fallback if LLM returns empty
                content = search_context

            return {
                "success": True,
                "content": content,
                "provider": provider_str,
                "model": model,
                "max_iterations": max_iterations,
                "metadata": {
                    "prompt": prompt[:200] + ("..." if len(prompt) > 200 else ""),
                },
            }

        except Exception as e:
            self.logger.error(f"Error in gathering: {e}")
            import traceback
            self.logger.error(traceback.format_exc())
            return {
                "success": False,
                "error": str(e),
                "content": "",
            }
