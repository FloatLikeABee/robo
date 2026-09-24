// Package htmldoc holds shared dark HTML document styling for MorphNotes exports.
package htmldoc

// Dark document palette: deep blue base, sky-blue links, purple gradient wash.
// Keep in sync with morph-engi/backend/src/api/project_docs.rs build_project_html.

const darkRootVars = `--ink:#e8eef7;--muted:#94a3b8;--line:#1e293b;--bg:#0b1220;--card:#111827;--accent:#38bdf8;--accent-purple:#818cf8`

const darkBodyGradient = `radial-gradient(1200px 600px at 10% -10%,#1e1b4b 0%,var(--bg) 55%)`

// DarkProseCSS is shared prose typography for generated HTML documents.
func DarkProseCSS() string {
	return `.prose{font-size:1.05rem;color:var(--ink)}
.prose>:first-child{margin-top:0}
.prose h1,.prose h2,.prose h3,.prose h4{line-height:1.25;margin:1.35em 0 .55em;font-weight:700;color:#f8fafc}
.prose h1{font-size:1.55rem}.prose h2{font-size:1.3rem}.prose h3{font-size:1.12rem}
.prose p,.prose ul,.prose ol,.prose blockquote,.prose pre,.prose table{margin:.85em 0}
.prose ul,.prose ol{padding-left:1.4em}
.prose li{margin:.35em 0}
.prose blockquote{padding:.35em 0 .35em 1em;border-left:3px solid var(--accent);color:var(--muted)}
.prose code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em;background:var(--card);padding:.12em .35em;border-radius:4px;border:1px solid var(--line)}
.prose pre{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}
.prose a{color:var(--accent)}
.prose hr{border:0;border-top:1px solid var(--line);margin:1.5em 0}
.prose strong{font-weight:700}
.prose table{border-collapse:collapse;width:100%;font-size:.95em}
.prose th,.prose td{border:1px solid var(--line);padding:.45em .6em;text-align:left}
.prose th{background:var(--card)}`
}

// DarkDocumentCSS returns root, layout, and prose rules for a standard document page.
func DarkDocumentCSS() string {
	return `:root{` + darkRootVars + `}
*{box-sizing:border-box}body{margin:0;font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:` + darkBodyGradient + `;line-height:1.55}
.wrap{max-width:760px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
` + DarkProseCSS() + `
.foot{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`
}

// DarkCaseTaskDocumentCSS includes scrollbars and case-task detail UI hooks.
func DarkCaseTaskDocumentCSS(extra string) string {
	return `:root{color-scheme:dark;` + darkRootVars + `;--scrollbar-thumb:#64748b;--scrollbar-track:#0b1220;scrollbar-width:thin;scrollbar-color:var(--scrollbar-thumb) var(--scrollbar-track)}
html,body{height:100%;margin:0}
*{box-sizing:border-box}body{font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:` + darkBodyGradient + `;line-height:1.55}
.wrap{max-width:760px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
` + DarkProseCSS() + `
` + extra + `
.foot{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`
}

// DarkBigNoteDocumentCSS is the dark Big note page including optional form blocks.
func DarkBigNoteDocumentCSS() string {
	return `:root{` + darkRootVars + `}
*{box-sizing:border-box}body{margin:0;font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:` + darkBodyGradient + `;line-height:1.55}
.wrap{max-width:720px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
` + `.prose{font-size:1.05rem;color:var(--ink)}
.prose>:first-child{margin-top:0}
.prose h1,.prose h2,.prose h3,.prose h4{line-height:1.25;margin:1.35em 0 .55em;font-weight:700}
.prose h1{font-size:1.55rem}.prose h2{font-size:1.3rem}.prose h3{font-size:1.12rem}
.prose p,.prose ul,.prose ol,.prose blockquote,.prose pre,.prose table{margin:.85em 0}
.prose ul,.prose ol{padding-left:1.4em}
.prose li{margin:.25em 0}
.prose blockquote{padding:.35em 0 .35em 1em;border-left:3px solid var(--accent);color:var(--muted)}
.prose code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em;background:var(--card);padding:.12em .35em;border-radius:4px;border:1px solid var(--line)}
.prose pre{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}
.prose pre code{background:transparent;border:0;padding:0}
.prose a{color:var(--accent)}
.prose hr{border:0;border-top:1px solid var(--line);margin:1.5em 0}
.prose strong{font-weight:700}
.prose table{border-collapse:collapse;width:100%;font-size:.95em}
.prose th,.prose td{border:1px solid var(--line);padding:.45em .6em;text-align:left}
.prose th{background:var(--card)}` + `
.prose h1,.prose h2,.prose h3,.prose h4{color:#f8fafc}
.q{margin:1.25rem 0;padding:1rem 1.1rem;border:1px solid var(--line);border-radius:12px;background:var(--card)}
.q label{display:block;font-family:system-ui,sans-serif;font-weight:600;margin-bottom:.55rem;color:#e2e8f0}
.q input[type=text],.q textarea,.q select{width:100%;padding:.65rem .75rem;border:1px solid #334155;border-radius:8px;font:inherit;background:#0f172a;color:#e2e8f0}
.q textarea{min-height:96px;resize:vertical}
.opts{display:grid;gap:.4rem;font-family:system-ui,sans-serif}
.opts label{font-weight:500;display:flex;gap:.5rem;align-items:flex-start;color:#cbd5e1}
button.primary{margin-top:1.25rem;padding:.7rem 1.1rem;border:0;border-radius:999px;background:var(--accent);color:#0b1220;font:600 14px system-ui,sans-serif;cursor:pointer}
.theme{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`
}
