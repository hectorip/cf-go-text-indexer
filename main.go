package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Los tags `json:"nombre"` mapean campos Go a nombres JSON personalizados

// Estructura para el índice
type Index struct {
	Dir       string      `json:"dir"`
	Generated time.Time   `json:"generated"`
	Model     string      `json:"model"`
	Items     []IndexItem `json:"items"`
}

// Estructura para un ítem del índice
type IndexItem struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	Summary  string    `json:"summary"`
	Keywords []string  `json:"keywords"`
	Error    string    `json:"error,omitempty"`
}

type Summarizer interface {
	Summarize(ctx context.Context, model, filename, preview string) (summary string, keywords []string, err error)
}

func main() {
	dir := flag.String("dir", "", "Directorio a indexar")
	out := flag.String("out", "index.json", "Archivo JSON de salida")
	maxBytes := flag.Int("max", 64*1024, "Máximo de bytes a leer por archivo")
	include := flag.String("include", ".txt,.md,.log,.rst,.json,.yaml,.yml,.toml,.go,.py,.js,.ts", "Extensiones de texto (coma separadas)")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout por archivo para llamada al LLM")
	flag.Parse()

	// Elegir summarizer
	provider := strings.ToLower(env("LLM_PROVIDER", "openai"))
	model := env("LLM_MODEL", "gpt-4o-mini")
	var s Summarizer
	switch provider {
	case "ollama":
		s = &OllamaSummarizer{Base: env("OLLAMA_BASE", "http://localhost:11434")}
	default: // openai compatible
		apikey := os.Getenv("LLM_API_KEY")
		if apikey == "" {
			fmt.Fprintln(os.Stderr, "WARN: LLM_API_KEY vacío; se generará índice SIN resumen/keywords")
			s = NoopSummarizer{}
		} else {
			s = &OpenAICompat{Base: env("OPENAI_BASE", "https://api.openai.com"), APIKey: apikey}
		}
	}

	exts := toSet(*include)
	var items []IndexItem // make()

	root, _ := filepath.Abs(*dir)
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !exts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		info, e := os.Stat(path)
		item := IndexItem{Path: filepath.ToSlash(rel)}
		if e != nil {
			item.Error = e.Error()
			items = append(items, item)
			return nil
		}
		item.Size = info.Size()
		item.ModTime = info.ModTime()

		// Leer hasta maxBytes
		f, e := os.Open(path)
		if e != nil {
			item.Error = e.Error()
			items = append(items, item)
			return nil
		}
		defer f.Close()
		lr := io.LimitedReader{R: f, N: int64(*maxBytes)}
		b, e := io.ReadAll(&lr)
		if e != nil {
			item.Error = e.Error()
			items = append(items, item)
			return nil
		}
		preview := string(b)

		// LLM (con timeout por archivo)
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		sum, kws, e := s.Summarize(ctx, model, rel, preview)
		if e != nil {
			item.Error = e.Error()
		}
		item.Summary = sum
		item.Keywords = kws
		items = append(items, item)
		return nil
	})

	idx := Index{
		Dir:       root,
		Generated: time.Now(),
		Model:     model,
		Items:     items,
	}
	if err := writeJSON(*out, idx); err != nil {
		fmt.Fprintln(os.Stderr, "write error:", err)
		os.Exit(1)
	}
	fmt.Println("OK →", *out, "items:", len(items))
}


type NoopSummarizer struct{}

func (NoopSummarizer) Summarize(ctx context.Context, model, filename, preview string) (string, []string, error) {
	p := strings.Fields(preview)
	if len(p) > 50 {
		p = p[:50]
	}
	s := strings.Join(p, " ")
	return s, []string{"texto", "sin-llm"}, nil
}

// OpenAI compatible (Chat Completions)
type OpenAICompat struct {
	Base   string
	APIKey string
}

func (c *OpenAICompat) Summarize(ctx context.Context, model, filename, preview string) (string, []string, error) {
	body := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "Responde SOLO un JSON: {\"summary\": \"...\", \"keywords\": [\"...\"]}"},
			{"role": "user", "content": prompt(filename, preview)},
		},
		"temperature": 0.2,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(c.Base, "/")+"/v1/chat/completions", strings.NewReader(string(b)))
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		d, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(d)))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", nil, err
	}
	if len(out.Choices) == 0 {
		return "", nil, errors.New("sin choices")
	}
	return parseJSON(out.Choices[0].Message.Content)
}


type OllamaSummarizer struct{ Base string }

func (o *OllamaSummarizer) Summarize(ctx context.Context, model, filename, preview string) (string, []string, error) {
	if model == "" {
		model = "llama3.1:8b"
	}
	body := map[string]any{"model": model, "prompt": prompt(filename, preview), "stream": false}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(o.Base, "/")+"/api/generate", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		d, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(d)))
	}
	var out struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", nil, err
	}
	return parseJSON(out.Response)
}


func prompt(filename, preview string) string {
	if len(preview) > 6000 {
		preview = preview[:6000]
	}
	return fmt.Sprintf(`Archivo: %s
Devuelve SOLO:
{"summary":"resumen en 1-2 frases, 40-80 palabras, sin saltos","keywords":["5-10 en minúsculas"]}
Texto:
%s`, filename, preview)
}

func parseJSON(s string) (string, []string, error) {
	s = strings.TrimSpace(s)
	// recortar fences ```json ... ```
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			s = s[i : j+1]
		}
	}

	// Unmarshal el JSON en una estructura temporal
	var tmp struct {
		Summary  string   `json:"summary"`
		Keywords []string `json:"keywords"`
	}
	if err := json.Unmarshal([]byte(s), &tmp); err != nil {
		return "", nil, err
	}
	return tmp.Summary, tmp.Keywords, nil
}


func toSet(csv string) map[string]bool {
	m := map[string]bool{}
	for _, e := range strings.Split(csv, ",") {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		m[strings.ToLower(e)] = true
	}
	return m
}

// Utilidad para obtener variables de entorno
func env(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

// Escribe un JSON en un archivo temporal y lo renombra
func writeJSON(path string, v any) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	f.Close()
	return os.Rename(tmp, path)
}
