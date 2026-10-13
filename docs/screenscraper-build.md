# Acceso de aplicación a ScreenScraper

El usuario final configura únicamente `username` y `password` de su cuenta personal. El backend obtiene `devid`, `devpassword` y `softname` de la configuración incorporada al compilar; ignora valores de desarrollador de perfiles antiguos o enviados desde el frontend.

## Variables del build

| Variable de entorno | Parámetro API | Valor predeterminado |
| --- | --- | --- |
| `GEMU_SCREENSCRAPER_DEVID` | `devid` | Vacío |
| `GEMU_SCREENSCRAPER_DEVPASSWORD` | `devpassword` | Vacío |
| `GEMU_SCREENSCRAPER_SOFTNAME` | `softname` | `Gemu` |

El build carga automáticamente `.env` desde la raíz del repositorio. Copia `.env.example` a `.env` y completa el ID y contraseña asignados a Gemu. `.env` está ignorado por Git; el archivo de ejemplo contiene únicamente nombres de variables y campos vacíos.

También puedes proporcionar las variables del proceso o secretos enmascarados de CI, que tienen prioridad sobre `.env`. No escribas secretos en archivos versionados, argumentos de comandos ni logs.

Se admiten líneas `NOMBRE=valor`, prefijo `export`, comentarios y valores entre comillas simples o dobles. Los valores son literales: no se expanden variables, escapes ni comandos; no se admiten valores multilínea. Solo se leen las tres variables de ScreenScraper. Los errores de formato indican el número de línea sin mostrar su contenido.

Con las variables configuradas, ejecuta desde la raíz:

```powershell
./build.ps1
```

La inyección y validación son obligatorias por defecto, sin flags. Si falta el ID o la contraseña, el build se detiene antes de invocar Wails y muestra cómo configurar `.env`. No hay credenciales de ejemplo activadas por defecto.

Para desarrollo, `./dev.ps1` reutiliza la carga automática de `.env` y conserva su directorio de datos independiente. Usa estos scripts para compilar y desarrollar: los comandos directos `wails build` y `go build` no ejecutan la preparación de credenciales.

La carga de `.env`, prioridad de entorno, validación y manejo literal se verifican con `./scripts/test-build-env.ps1`.

## Manejo de secretos

`build.ps1` genera temporalmente `internal/providers/content/screenscraper_build_local.go`, ignorado por Git, y lo elimina al terminar, incluso si Wails falla. Un mutex de Windows por repositorio impide sesiones build/dev simultáneas y se libera si el proceso termina. Un archivo residual con la marca del generador se regenera automáticamente desde `.env` o las variables de entorno; no hace falta borrarlo a mano. Los archivos sin esa marca se conservan para evitar sobrescribir cambios propios. Usa los scripts para que todas las sesiones participen en este bloqueo.

No se usan `-ldflags -X`, ya que los argumentos de enlazado quedan registrados en la información de build de Go. Las credenciales no se publican por el puente Wails ni se incluyen en los archivos JavaScript del frontend. Los campos de desarrollador de almacenes antiguos se conservan por compatibilidad de lectura, pero no autentican consultas.

Esto evita publicar secretos en el repositorio; no impide extraerlos del ejecutable o de la caché de compilación. El acceso directo desde una aplicación de escritorio requiere esa limitación. Si las credenciales deben permanecer exclusivamente en infraestructura controlada, las consultas necesitan un servicio intermediario. Usa runners privados o efímeros para builds con credenciales.
