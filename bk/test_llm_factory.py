"""
Manual test script for LLM Factory (run with: python test_llm_factory.py).
API keys must be supplied via environment variables — never commit real keys.
"""
import os

from src.llm_factory import LLMFactory, LLMProvider

# Import all callers to register them with the factory
import gemini_caller
import groq_caller
import qwen_caller


def test_gemini():
    """Test Gemini caller"""
    key = os.environ.get("GEMINI_API_KEY", "")
    if not key:
        print("Skipping Gemini: set GEMINI_API_KEY")
        return
    print("\n=== Testing Gemini ===")
    caller = LLMFactory.create_caller(
        provider=LLMProvider.GEMINI,
        api_key=key,
        model="gemini-2.5-flash",
    )

    response = caller.generate("What is artificial intelligence? Answer in one sentence.")
    print(f"Response: {response}")


def test_qwen():
    """Test Qwen caller"""
    key = os.environ.get("QWEN_API_KEY", "")
    if not key:
        print("Skipping Qwen: set QWEN_API_KEY")
        return
    print("\n=== Testing Qwen ===")
    caller = LLMFactory.create_caller(
        provider=LLMProvider.QWEN,
        api_key=key,
        model="qwen3-max",
    )

    response = caller.generate("What is artificial intelligence? Answer in one sentence.")
    print(f"Response: {response}")


def test_groq():
    """Test Groq caller"""
    key = os.environ.get("GROQ_API_KEY", "")
    if not key:
        print("Skipping Groq: set GROQ_API_KEY")
        return
    print("\n=== Testing Groq ===")
    caller = LLMFactory.create_caller(
        provider=LLMProvider.GROQ,
        api_key=key,
        model="llama-3.3-70b-versatile",
    )

    response = caller.generate("What is artificial intelligence? Answer in one sentence.")
    print(f"Response: {response}")


def test_chat():
    """Test chat functionality"""
    key = os.environ.get("GEMINI_API_KEY", "")
    if not key:
        print("Skipping chat: set GEMINI_API_KEY")
        return
    print("\n=== Testing Chat with Gemini ===")
    caller = LLMFactory.create_caller(
        provider=LLMProvider.GEMINI,
        api_key=key,
        model="gemini-2.5-flash",
    )

    messages = [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "What is 2+2?"},
    ]

    response = caller.chat(messages)
    print(f"Response: {response}")


def test_streaming():
    """Test streaming functionality"""
    key = os.environ.get("QWEN_API_KEY", "")
    if not key:
        print("Skipping streaming: set QWEN_API_KEY")
        return
    print("\n=== Testing Streaming with Qwen ===")
    caller = LLMFactory.create_caller(
        provider=LLMProvider.QWEN,
        api_key=key,
        model="qwen3-max",
    )

    print("Streaming response: ", end="", flush=True)
    for chunk in caller.stream("Tell me a short joke."):
        print(chunk, end="", flush=True)
    print()


if __name__ == "__main__":
    print("LLM Factory Test Suite")
    print("=" * 50)

    # List available providers
    providers = LLMFactory.get_available_providers()
    print(f"\nAvailable providers: {providers}")

    # Test each provider
    try:
        test_gemini()
    except Exception as e:
        print(f"Gemini test failed: {e}")

    try:
        test_qwen()
    except Exception as e:
        print(f"Qwen test failed: {e}")

    try:
        test_groq()
    except Exception as e:
        print(f"Groq test failed: {e}")

    try:
        test_chat()
    except Exception as e:
        print(f"Chat test failed: {e}")

    try:
        test_streaming()
    except Exception as e:
        print(f"Streaming test failed: {e}")

    print("\n" + "=" * 50)
    print("Test suite completed!")
