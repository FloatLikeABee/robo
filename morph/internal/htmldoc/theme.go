// Package htmldoc holds shared dark HTML document styling for MorphNotes exports.
package htmldoc

// Dark document palette: near-black base, cyan / violet / mint / amber accents.
// Keep in sync with morph-engi/backend/src/api/project_docs.rs build_project_html
// and morph/frontend case-task preview CSS.

const sansStack = `Arial, "Helvetica Neue", Helvetica, sans-serif`

const darkRootVars = `--ink:#e8eef7;--muted:#94a3b8;--line:#1e293b;--bg:#070b14;--card:#111827;--accent:#38bdf8;--accent-2:#a78bfa;--accent-3:#34d399;--accent-4:#fbbf24;--accent-purple:#a78bfa`

const darkBodyGradient = `radial-gradient(900px 480px at 8% -12%, rgba(167,139,250,.55) 0%, transparent 58%), radial-gradient(720px 420px at 100% 0%, rgba(56,189,248,.35) 0%, transparent 52%), var(--bg)`

// DarkProseCSS is shared prose typography for generated HTML documents.
func DarkProseCSS() string {
	return `.prose{font-size:1.05rem;color:var(--ink)}
.prose>:first-child{margin-top:0}
.prose h1,.prose h2,.prose h3,.prose h4{line-height:1.25;margin:1.35em 0 .55em;font-weight:700}
.prose h1{font-size:1.55rem;color:#f8fafc}.prose h2{font-size:1.3rem;color:var(--accent-2)}.prose h3{font-size:1.12rem;color:var(--accent-3)}.prose h4{color:var(--accent-4)}
.prose p,.prose ul,.prose ol,.prose blockquote,.prose pre,.prose table{margin:.85em 0}
.prose ul,.prose ol{padding-left:1.4em}
.prose li{margin:.35em 0}
.prose blockquote{padding:.35em 0 .35em 1em;border-left:3px solid var(--accent-4);color:var(--muted)}
.prose code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em;background:var(--card);padding:.12em .35em;border-radius:4px;border:1px solid var(--line)}
.prose pre{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}
.prose a{color:var(--accent)}
.prose hr{border:0;border-top:1px solid var(--line);margin:1.5em 0}
.prose strong{font-weight:700}
.prose table{border-collapse:collapse;width:100%;font-size:.95em}
.prose th,.prose td{border:1px solid var(--line);padding:.45em .6em;text-align:left}
.prose th{background:var(--card)}
` + VisualBlockCSS()
}

// VisualBlockCSS styles section cards, milestone rails, callouts, stats, and bar charts.
func VisualBlockCSS() string {
	return `.doc-card{margin:1.1rem 0;padding:1rem 1.15rem 1.15rem;border:1px solid var(--line);border-radius:16px;background:rgba(17,24,39,.78);box-shadow:0 0 0 1px rgba(56,189,248,.12)}
.doc-card>h2{margin:.1rem 0 .45rem;font-size:1.2rem}
.rail{list-style:none;margin:.4rem 0 0;padding:0}
.rail>li{position:relative;margin:0;padding:.4rem 0 .4rem 1.35rem;border-left:2px solid rgba(52,211,153,.55)}
.rail>li:before{content:"";position:absolute;left:-5px;top:.65rem;width:8px;height:8px;border-radius:50%;background:#34d399}
.callout{margin:1rem 0;padding:.8rem 1rem;border-radius:12px;background:rgba(167,139,250,.16);border:1px solid rgba(167,139,250,.45);color:#e2e8f0}
.stat-row{display:flex;flex-wrap:wrap;gap:.6rem;margin:.8rem 0}
.stat{min-width:6.5rem;padding:.65rem .8rem;border-radius:12px;background:#0f172a;border:1px solid var(--line)}
.stat b{display:block;font:700 1.3rem/1.1 Arial,"Helvetica Neue",Helvetica,sans-serif;color:var(--accent)}
.stat span{font:11px/1.3 system-ui,sans-serif;color:var(--muted);letter-spacing:.04em;text-transform:uppercase}
.chart{margin:.35rem 0 .2rem}
.chart svg{width:100%;height:auto;display:block}`
}

// DarkDocumentCSS returns root, layout, and prose rules for a standard document page.
func DarkDocumentCSS() string {
	return `:root{color-scheme:dark;` + darkRootVars + `}
*{box-sizing:border-box}body{margin:0;font-family:` + sansStack + `;color:var(--ink);background:` + darkBodyGradient + `;line-height:1.55}
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
*{box-sizing:border-box}body{font-family:` + sansStack + `;color:var(--ink);background:` + darkBodyGradient + `;line-height:1.55}
.wrap{max-width:760px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
` + DarkProseCSS() + `
` + extra + `
.foot{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`
}

// DarkBigNoteDocumentCSS is the dark Big note page including optional form blocks.
func DarkBigNoteDocumentCSS() string {
	return `:root{color-scheme:dark;` + darkRootVars + `}
*{box-sizing:border-box}body{margin:0;font-family:` + sansStack + `;color:var(--ink);background:` + darkBodyGradient + `;line-height:1.55}
.wrap{max-width:720px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
` + `.prose{font-size:1.05rem;color:var(--ink)}
.prose>:first-child{margin-top:0}
.prose h1,.prose h2,.prose h3,.prose h4{line-height:1.25;margin:1.35em 0 .55em;font-weight:700}
.prose h1{font-size:1.55rem;color:#f8fafc}.prose h2{font-size:1.3rem;color:var(--accent-2)}.prose h3{font-size:1.12rem;color:var(--accent-3)}.prose h4{color:var(--accent-4)}
.prose p,.prose ul,.prose ol,.prose blockquote,.prose pre,.prose table{margin:.85em 0}
.prose ul,.prose ol{padding-left:1.4em}
.prose li{margin:.25em 0}
.prose blockquote{padding:.35em 0 .35em 1em;border-left:3px solid var(--accent-4);color:var(--muted)}
.prose code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em;background:var(--card);padding:.12em .35em;border-radius:4px;border:1px solid var(--line)}
.prose pre{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}
.prose pre code{background:transparent;border:0;padding:0}
.prose a{color:var(--accent)}
.prose hr{border:0;border-top:1px solid var(--line);margin:1.5em 0}
.prose strong{font-weight:700}
.prose table{border-collapse:collapse;width:100%;font-size:.95em}
.prose th,.prose td{border:1px solid var(--line);padding:.45em .6em;text-align:left}
.prose th{background:var(--card)}` + `
.q{margin:1.25rem 0;padding:1rem 1.1rem;border:1px solid var(--line);border-radius:12px;background:var(--card)}
.q label{display:block;font-family:system-ui,sans-serif;font-weight:600;margin-bottom:.55rem;color:#e2e8f0}
.q input[type=text],.q textarea,.q select{width:100%;padding:.65rem .75rem;border:1px solid #334155;border-radius:8px;font:inherit;background:#0f172a;color:#e2e8f0}
.q textarea{min-height:96px;resize:vertical}
.opts{display:grid;gap:.4rem;font-family:system-ui,sans-serif}
.opts label{font-weight:500;display:flex;gap:.5rem;align-items:flex-start;color:#cbd5e1}
button.primary{margin-top:1.25rem;padding:.7rem 1.1rem;border:0;border-radius:999px;background:var(--accent);color:#0b1220;font:600 14px system-ui,sans-serif;cursor:pointer}
.theme{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}
` + VisualBlockCSS()
}
