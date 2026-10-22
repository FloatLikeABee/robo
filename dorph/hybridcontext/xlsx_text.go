package hybridcontext

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var xmlnsStrip = regexp.MustCompile(`\s+xmlns="[^"]*"`)

// XLSXFromBytes extracts tab-separated text from the first worksheet (best-effort, no external deps).
func XLSXFromBytes(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("not a valid xlsx zip: %w", err)
	}
	var shared []string
	for _, f := range zr.File {
		if f.Name == "xl/sharedStrings.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return "", err
			}
			shared = parseSharedStrings(b)
			break
		}
	}
	var sheetFile *zip.File
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			sheetFile = f
			break
		}
	}
	if sheetFile == nil {
		return "", fmt.Errorf("no worksheet found in xlsx")
	}
	rc, err := sheetFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	sheetBytes, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return unmarshalSheet(sheetBytes, shared)
}

type sstXML struct {
	SI []struct {
		T []string `xml:"t"`
	} `xml:"si"`
}

func parseSharedStrings(raw []byte) []string {
	raw = xmlnsStrip.ReplaceAll(raw, nil)
	var doc sstXML
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return nil
	}
	out := make([]string, 0, len(doc.SI))
	for _, si := range doc.SI {
		out = append(out, strings.TrimSpace(strings.Join(si.T, "")))
	}
	return out
}

type sheetXML struct {
	SheetData struct {
		Rows []struct {
			Cells []struct {
				T string `xml:"t,attr"`
				V string `xml:"v"`
			} `xml:"c"`
		} `xml:"row"`
	} `xml:"sheetData"`
}

func unmarshalSheet(raw []byte, shared []string) (string, error) {
	raw = xmlnsStrip.ReplaceAll(raw, nil)
	var doc sheetXML
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return "", fmt.Errorf("parse worksheet: %w", err)
	}
	var b strings.Builder
	for _, r := range doc.SheetData.Rows {
		var cells []string
		for _, c := range r.Cells {
			val := strings.TrimSpace(c.V)
			switch c.T {
			case "s":
				if i, err := strconv.Atoi(val); err == nil && i >= 0 && i < len(shared) {
					val = shared[i]
				}
			}
			cells = append(cells, val)
		}
		nonEmpty := false
		for _, x := range cells {
			if strings.TrimSpace(x) != "" {
				nonEmpty = true
				break
			}
		}
		if !nonEmpty {
			continue
		}
		b.WriteString(strings.Join(cells, "\t"))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String()), nil
}
