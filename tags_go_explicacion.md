# Sintaxis de Tags en Go: Explicación Completa

## ¿Qué son los Tags en Go?

Los **tags** (etiquetas) en Go son metadatos que se pueden asociar a los campos de una estructura (`struct`). Se escriben como cadenas literales después de la declaración del campo y proporcionan información adicional que puede ser utilizada por bibliotecas en tiempo de ejecución.

## Sintaxis Básica

```go
type Estructura struct {
    Campo tipo `tag:"valor"`
}
```

### Ejemplo del código:

```go
type Index struct {
    Dir       string      `json:"dir"`
    Generated time.Time   `json:"generated"`
    Model     string      `json:"model"`
    Items     []IndexItem `json:"items"`
}

type IndexItem struct {
    Path     string    `json:"path"`
    Size     int64     `json:"size"`
    ModTime  time.Time `json:"mod_time"`
    Summary  string    `json:"summary"`
    Keywords []string  `json:"keywords"`
    Error    string    `json:"error,omitempty"`
}
```

## Explicación de los Tags JSON

### 1. **Tag Básico**: `json:"nombre"`

```go
Dir string `json:"dir"`
```

- **Función**: Mapea el campo `Dir` de Go al campo `"dir"` en JSON
- **Sin tag**: El campo aparecería como `"Dir"` en JSON (nombre original)
- **Con tag**: El campo aparece como `"dir"` en JSON (nombre personalizado)

### 2. **Tag con Opción**: `json:"nombre,omitempty"`

```go
Error string `json:"error,omitempty"`
```

- **`omitempty`**: Si el campo está vacío (valor cero), se omite del JSON
- **Valores considerados "vacíos"**:
  - `""` para strings
  - `0` para números
  - `nil` para punteros/slices/maps
  - `false` para booleanos

### Ejemplo Práctico de Serialización

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"
)

type Ejemplo struct {
    Nombre   string    `json:"name"`
    Edad     int       `json:"age"`
    Email    string    `json:"email,omitempty"`
    Activo   bool      `json:"active"`
    Creado   time.Time `json:"created_at"`
}

func main() {
    // Con todos los campos
    e1 := Ejemplo{
        Nombre: "Juan",
        Edad:   30,
        Email:  "juan@email.com",
        Activo: true,
        Creado: time.Now(),
    }
    
    // Sin email (será omitido)
    e2 := Ejemplo{
        Nombre: "María",
        Edad:   25,
        Activo: false,
        Creado: time.Now(),
    }
    
    json1, _ := json.Marshal(e1)
    json2, _ := json.Marshal(e2)
    
    fmt.Println("Con email:", string(json1))
    fmt.Println("Sin email:", string(json2))
}
```

**Salida:**

```json
Con email: {"name":"Juan","age":30,"email":"juan@email.com","active":true,"created_at":"2024-01-15T10:30:00Z"}
Sin email: {"name":"María","age":25,"active":false,"created_at":"2024-01-15T10:30:00Z"}
```

## Otros Usos de Tags en Go

### 1. **Tags de Validación** (con bibliotecas como `validator`)

```go
import "github.com/go-playground/validator/v10"

type Usuario struct {
    Nombre   string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Edad     int    `json:"age" validate:"min=18,max=120"`
    Password string `json:"-" validate:"required,min=8"`
}
```

**Opciones de validación:**

- `required`: Campo obligatorio
- `min=N`: Valor mínimo
- `max=N`: Valor máximo
- `email`: Formato de email válido
- `oneof=red green blue`: Valor debe ser uno de los especificados

### 2. **Tags de Base de Datos** (GORM, sqlx)

```go
type Producto struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Nombre      string    `json:"name" gorm:"size:100;not null"`
    Precio      float64   `json:"price" gorm:"type:decimal(10,2)"`
    Descripcion string    `json:"description" gorm:"type:text"`
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
```

**Opciones GORM:**

- `primaryKey`: Clave primaria
- `size:N`: Tamaño máximo
- `not null`: No puede ser nulo
- `type:tipo`: Tipo específico de base de datos
- `autoCreateTime`/`autoUpdateTime`: Timestamps automáticos

### 3. **Tags YAML**

```go
import "gopkg.in/yaml.v2"

type Config struct {
    Servidor string `yaml:"server" json:"server"`
    Puerto   int    `yaml:"port" json:"port"`
    Debug    bool   `yaml:"debug" json:"debug"`
}
```

### 4. **Tags XML**

```go
import "encoding/xml"

type Libro struct {
    XMLName xml.Name `xml:"libro"`
    Titulo  string   `xml:"titulo"`
    Autor   string   `xml:"autor"`
    ISBN    string   `xml:"isbn,attr"`
}
```

### 5. **Tags Personalizados**

```go
type Usuario struct {
    Nombre string `json:"name" custom:"secret"`
    Email  string `json:"email" custom:"public"`
}

// Función para leer tags personalizados
func leerTagsPersonalizados(v interface{}) {
    t := reflect.TypeOf(v)
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        if tag := field.Tag.Get("custom"); tag != "" {
            fmt.Printf("Campo %s tiene tag custom: %s\n", field.Name, tag)
        }
    }
}
```

### 6. **Tags Múltiples Combinados**

```go
type APIResponse struct {
    ID       int    `json:"id" xml:"id" yaml:"id" gorm:"primaryKey" validate:"required"`
    Message  string `json:"message" xml:"message" yaml:"message" gorm:"size:500" validate:"required,min=1"`
    Success  bool   `json:"success" xml:"success" yaml:"success" gorm:"default:false"`
    Data     string `json:"data,omitempty" xml:"data,omitempty" yaml:"data,omitempty"`
}
```

## Opciones Comunes de Tags JSON

### **Ignorar Campos Completamente**

```go
type Usuario struct {
    Nombre   string `json:"name"`
    Password string `json:"-"`          // Nunca aparece en JSON
    Email    string `json:"email"`
}
```

### **Usar Nombre Original del Campo**

```go
type Data struct {
    UserID string `json:",omitempty"`   // Usa "UserID", omite si vacío
    Name   string `json:","`            // Usa "Name", siempre incluye
}
```

### **Campos como String Aunque Sean Números**

```go
type Response struct {
    ID    int `json:"id,string"`        // Convierte número a string en JSON
    Count int `json:"count"`            // Mantiene como número
}
```

## Acceso a Tags en Tiempo de Ejecución

```go
import "reflect"

func analizarTags(v interface{}) {
    t := reflect.TypeOf(v)
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        jsonTag := field.Tag.Get("json")
        validateTag := field.Tag.Get("validate")
        
        fmt.Printf("Campo: %s\n", field.Name)
        fmt.Printf("  JSON tag: %s\n", jsonTag)
        fmt.Printf("  Validate tag: %s\n", validateTag)
    }
}
```

## Beneficios de Usar Tags

1. **Flexibilidad**: Personalizar nombres de campos sin cambiar el código
2. **Compatibilidad**: Adaptar estructuras Go a formatos externos (JSON, XML, YAML)
3. **Validación**: Aplicar reglas de validación declarativamente
4. **Configuración**: Especificar comportamientos de bibliotecas ORM
5. **Documentación**: Los tags sirven como documentación del mapeo de datos

## Buenas ideas para usar tags

1. **Consistencia**: Usar el mismo estilo de naming en todos los tags
2. **Omitempty**: Usar cuando sea apropiado para reducir tamaño del JSON
3. **Validación**: Combinar tags JSON con tags de validación
4. **Documentación**: Documentar tags personalizados

Los tags son una característica poderosa de Go que permite metaprogramación de manera declarativa y limpia, facilitando la integración con sistemas externos y bibliotecas de terceros.
