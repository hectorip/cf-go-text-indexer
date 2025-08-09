# Sistema Operativo

Explicación del código de `main.go` que usa bibliotecas del sistema operativo.

## Imports de Bibliotecas del Sistema Operativo

```go
import (
    "io"           // Entrada/salida básica
    "net/http"     // Comunicación de red HTTP
    "os"           // Interfaz del sistema operativo
    "path/filepath"// Manipulación de rutas de archivos
)
```

## Explicación Detallada por Líneas

### Manejo de Variables de Entorno

```go
provider := strings.ToLower(env("LLM_PROVIDER", "openai"))
```

- Llama a la función `env()` que internamente usa `os.Getenv()`
- Obtiene la variable de entorno `LLM_PROVIDER` del sistema operativo

```go
fmt.Fprintln(os.Stderr, "WARN: LLM_API_KEY vacío; se generará índice SIN resumen/keywords")
```

- **Acceso directo al SO:** Usa `os.Stderr` que representa el flujo de error estándar del sistema operativo
- Escribe un mensaje de advertencia al stderr del proceso

### Manipulación de Rutas y Archivos

```go
root, _ := filepath.Abs(*dir)
```

- **Acceso al SO:** `filepath.Abs()` resuelve la ruta absoluta consultando el directorio de trabajo actual del sistema operativo
- Convierte una ruta relativa en absoluta usando información del filesystem

```go
filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
```

- **Acceso directo al SO:** `filepath.WalkDir()` recorre recursivamente el sistema de archivos
- `os.DirEntry` es una interfaz que representa entradas de directorio del sistema operativo

```go
if err != nil || d.IsDir() {
```

- `d.IsDir()` consulta el tipo de entrada del filesystem (directorio vs archivo)

```go
if !exts[strings.ToLower(filepath.Ext(path))] {
```

- `filepath.Ext()` extrae la extensión del archivo analizando la ruta

```go
rel, _ := filepath.Rel(root, path)
```

- `filepath.Rel()` calcula la ruta relativa entre dos rutas del filesystem

```go
info, e := os.Stat(path)
```

- **Acceso directo al SO:** `os.Stat()` hace una llamada al sistema para obtener información detallada del archivo
- Retorna metadatos como tamaño, fecha de modificación, permisos, etc.

```go
item := IndexItem{Path: filepath.ToSlash(rel)}
```

- `filepath.ToSlash()` convierte separadores de ruta específicos del SO (\ en Windows) a barras (/)

```go
item.Size = info.Size()
item.ModTime = info.ModTime()
```

- `info.Size()` y `info.ModTime()` extraen información del archivo obtenida del sistema operativo

### Operaciones de Lectura de Archivos

```go
f, e := os.Open(path)
```

- **Acceso directo al SO:** `os.Open()` abre un archivo para lectura

```go
defer f.Close()
```

- `f.Close()` cierra el descriptor de archivo cuando la función termina
- Libera recursos del sistema operativo

```go
lr := io.LimitedReader{R: f, N: int64(*maxBytes)}
```

- `io.LimitedReader` es un wrapper que limita cuántos bytes se pueden leer del archivo

```go
b, e := io.ReadAll(&lr)
```

- **Acceso al SO:** `io.ReadAll()` lee datos del archivo hasta EOF o límite
- Hace llamadas al sistema para leer bytes del descriptor de archivo

### Salida a Stderr y Control de Procesos

```go
fmt.Fprintln(os.Stderr, "write error:", err)
```

- **Acceso directo al SO:** Escribe al flujo de error estándar del proceso

```go
os.Exit(1)
```

- **Acceso directo al SO:** `os.Exit()` termina el proceso con código de salida 1

```go
fmt.Println("OK →", *out, "items:", len(items))
```

- `fmt.Println()` internamente escribe a `os.Stdout` (salida estándar)

### Comunicación de Red HTTP

```go
req, _ := http.NewRequestWithContext(ctx, "POST", ...)
```

- `http.NewRequestWithContext()` crea una solicitud HTTP que eventualmente usará sockets del SO

```go
resp, err := http.DefaultClient.Do(req)
```

- **Acceso al SO:** `http.DefaultClient.Do()` realiza la solicitud HTTP
- Internamente abre sockets de red y hace llamadas al sistema para comunicación TCP/IP

```go
defer resp.Body.Close()
```

- Cierra la conexión HTTP y libera recursos de red del sistema operativo

```go
d, _ := io.ReadAll(resp.Body)
```

- Lee datos de la respuesta HTTP desde el socket de red

## Funciones Auxiliares que Acceden al SO

```go
f, err := os.Create(tmp)
```

- **Acceso directo al SO:** `os.Create()` crea un nuevo archivo en el filesystem
- Hace llamadas al sistema para crear el archivo
- Diferencia con os.Open():

  - `os.Open()` abre un archivo para lectura o escritura
  - `os.Create()` crea un archivo nuevo para escritura

```go
return os.Rename(tmp, path)
```

- **Acceso directo al SO:** `os.Rename()` renombra/mueve un archivo en el filesystem
- Operación atómica que hace una llamada al sistema para mover el archivo
