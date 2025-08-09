# text-indexer (mini)

Indexa **archivos de texto** en un directorio y genera `index.json` con:

- `summary` y `keywords` por archivo (vía LLM opcional)
- tamaño y fecha de modificación por archivo

## Requisitos

- Go 1.22+
- (Opcional) API compatible con OpenAI **o** Ollama local

## Build

```bash
make build       # bin/text-indexer
make cross       # bin/* para varias plataformas
```

## Uso

OpenAI-compatible (OpenAI, OpenRouter, etc.):

```bash
export LLM_PROVIDER=openai
export LLM_API_KEY=sk-...
export LLM_MODEL=gpt-4o-mini
./bin/text-indexer -dir ~/Notas -out index.json
```

Ollama local:

```bash
export LLM_PROVIDER=ollama
export LLM_MODEL=llama3.1:8b
./bin/text-indexer -dir ~/Notas -out index.json
```

Sin token (modo rápido, sin llamadas LLM):

```bash
./bin/text-indexer -dir . -out index.json
```

Flags útiles:

- `--include` extensiones: `.txt,.md,.log,.rst,.json,.yaml,.yml,.toml,.go,.py,.js,.ts`
- `--max` bytes máximos a leer por archivo (default 65536)
- `--timeout` timeout por archivo para la llamada LLM

## Notas

- Solo archivos de texto (por extensión).
- Si no defines `LLM_API_KEY` (modo openai), el resumen es básico (sin LLM) - las primeras 50 palabras.
