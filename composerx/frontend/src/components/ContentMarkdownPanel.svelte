<script>
  import { renderMarkdownHtml } from '../lib/contentMarkdown'
  import '@robo/platform-chat/message-body.css'

  /** @type {{ markdown?: string, mode?: 'preview' | 'source' }} */
  let { markdown = '', mode = 'preview' } = $props()

  const html = $derived(renderMarkdownHtml(markdown))
</script>

{#if mode === 'preview'}
  <div class="content-markdown-preview platform-message-body">
    {#if String(markdown || '').trim()}
      {@html html}
    {:else}
      <p class="content-markdown-empty">No content.</p>
    {/if}
  </div>
{:else}
  <pre class="content-markdown-source">{markdown || ''}</pre>
{/if}

<style>
  .content-markdown-preview {
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.75rem;
    padding: 0.85rem 1rem;
    background: var(--color-bg);
    max-height: min(68vh, 640px);
    overflow: auto;
  }

  .content-markdown-empty {
    margin: 0;
    font-size: 0.85rem;
    color: var(--color-text-subtle);
  }

  .content-markdown-source {
    margin: 0;
    padding: 0.85rem 1rem;
    border-radius: 0.75rem;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg-elevated);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
      monospace;
    font-size: 0.85rem;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: min(68vh, 640px);
    overflow: auto;
  }
</style>
