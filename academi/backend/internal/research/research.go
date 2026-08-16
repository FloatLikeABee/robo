// Package research gathers lightweight public snippets (Wikipedia, DuckDuckGo, arXiv)
// to ground STEM document drafts. No API keys required; respect rate limits in production.
package research

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const userAgent = "AcademiBackend/1.0 (educational; local development)"

// Source is a citable reference line for the chat UI.
type Source struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url,omitempty"`
}

// Gather returns condensed research notes and sources for the given query.
func Gather(query string) (notes string, sources []Source) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", nil
	}
	if len(query) > 400 {
		query = query[:400]
	}

	client := &http.Client{Timeout: 14 * time.Second}
	var parts []string

	if title, excerpt, pageURL, err := wikipediaBest(client, query); err == nil && excerpt != "" {
		parts = append(parts, fmt.Sprintf("[Wikipedia — %s]\n%s", title, truncate(excerpt, 2200)))
		sources = append(sources, Source{Title: "Wikipedia: " + title, Type: "wiki", URL: pageURL})
	}

	if abstract, absURL, err := duckDuckGoInstant(client, query); err == nil && abstract != "" {
		parts = append(parts, fmt.Sprintf("[Web summary — DuckDuckGo instant answer]\n%s", truncate(abstract, 1400)))
		if absURL != "" {
			sources = append(sources, Source{Title: "Web (DuckDuckGo)", Type: "web", URL: absURL})
		}
	}

	if block, src, ok := arxivTop(client, query); ok {
		parts = append(parts, block)
		sources = append(sources, src)
	}

	return strings.Join(parts, "\n\n"), dedupeSources(sources)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func dedupeSources(in []Source) []Source {
	seen := make(map[string]bool)
	out := make([]Source, 0, len(in))
	for _, s := range in {
		key := s.URL
		if key == "" {
			key = s.Title + "|" + s.Type
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, s)
	}
	return out
}

type wikiSearchResp struct {
	Query struct {
		Search []struct {
			Title string `json:"title"`
		} `json:"search"`
	} `json:"query"`
}

type wikiSummaryResp struct {
	Title   string `json:"title"`
	Extract string `json:"extract"`
	ContentURLs struct {
		Desktop struct {
			Page string `json:"page"`
		} `json:"desktop"`
	} `json:"content_urls"`
}

func wikipediaBest(client *http.Client, query string) (title, excerpt, pageURL string, err error) {
	u := "https://en.wikipedia.org/w/api.php?action=query&format=json&list=search&srsearch=" +
		url.QueryEscape(query) + "&srlimit=1"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("wiki search: status %d", resp.StatusCode)
	}
	var ws wikiSearchResp
	if err := json.Unmarshal(body, &ws); err != nil {
		return "", "", "", err
	}
	if len(ws.Query.Search) == 0 {
		return "", "", "", fmt.Errorf("no wiki results")
	}
	title = ws.Query.Search[0].Title

	su := "https://en.wikipedia.org/api/rest_v1/page/summary/" + url.PathEscape(title)
	req2, err := http.NewRequest(http.MethodGet, su, nil)
	if err != nil {
		return "", "", "", err
	}
	req2.Header.Set("User-Agent", userAgent)
	resp2, err := client.Do(req2)
	if err != nil {
		return "", "", "", err
	}
	defer resp2.Body.Close()
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", "", "", err
	}
	if resp2.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("wiki summary: status %d", resp2.StatusCode)
	}
	var sum wikiSummaryResp
	if err := json.Unmarshal(body2, &sum); err != nil {
		return "", "", "", err
	}
	pageURL = sum.ContentURLs.Desktop.Page
	if pageURL == "" {
		pageURL = "https://en.wikipedia.org/wiki/" + url.PathEscape(strings.ReplaceAll(title, " ", "_"))
	}
	return sum.Title, sum.Extract, pageURL, nil
}

type ddgResp struct {
	Abstract     string `json:"Abstract"`
	AbstractText string `json:"AbstractText"`
	AbstractURL  string `json:"AbstractURL"`
	Answer       string `json:"Answer"`
	Definition   string `json:"Definition"`
}

func duckDuckGoInstant(client *http.Client, query string) (text, refURL string, err error) {
	u := "https://api.duckduckgo.com/?q=" + url.QueryEscape(query) + "&format=json&no_html=1&skip_disambig=1"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("ddg: status %d", resp.StatusCode)
	}
	var d ddgResp
	if err := json.Unmarshal(body, &d); err != nil {
		return "", "", err
	}
	text = strings.TrimSpace(d.AbstractText)
	if text == "" {
		text = strings.TrimSpace(d.Abstract)
	}
	if text == "" && d.Answer != "" {
		text = strings.TrimSpace(d.Answer)
	}
	if text == "" && d.Definition != "" {
		text = strings.TrimSpace(d.Definition)
	}
	refURL = strings.TrimSpace(d.AbstractURL)
	if text == "" {
		return "", "", fmt.Errorf("empty ddg")
	}
	return text, refURL, nil
}

var reArxivID = regexp.MustCompile(`<id>([^<]+)</id>`)
var reArxivTitle = regexp.MustCompile(`<title>([^<]+)</title>`)

func arxivTop(client *http.Client, query string) (block string, src Source, ok bool) {
	q := url.QueryEscape(truncate(query, 180))
	u := "http://export.arxiv.org/api/query?search_query=all:" + q + "&start=0&max_results=2"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", Source{}, false
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", Source{}, false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", Source{}, false
	}
	raw := string(body)
	idx := strings.Index(raw, "<entry>")
	if idx < 0 {
		return "", Source{}, false
	}
	chunk := raw[idx:]
	end := strings.Index(chunk, "</entry>")
	if end > 0 {
		chunk = chunk[:end]
	}
	idm := reArxivID.FindStringSubmatch(chunk)
	tm := reArxivTitle.FindStringSubmatch(chunk)
	if len(idm) < 2 || len(tm) < 2 {
		return "", Source{}, false
	}
	idURL := strings.TrimSpace(idm[1])
	title := strings.TrimSpace(strings.ReplaceAll(tm[1], "\n", " "))
	title = strings.Join(strings.Fields(title), " ")
	if title == "" || strings.Contains(strings.ToLower(title), "arxiv query") {
		return "", Source{}, false
	}
	block = fmt.Sprintf("[arXiv — related preprint]\n%s\n%s", title, idURL)
	src = Source{Title: "arXiv: " + truncate(title, 120), Type: "paper", URL: idURL}
	return block, src, true
}
