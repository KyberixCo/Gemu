# Plan maestro de implementación

Launcher unificado para Windows · Wails + Go + TypeScript · Arquitectura y roadmap técnico

## 1\. Visión del producto

Propongo desarrollar un **frontend de videojuegos para Windows que transforme un PC convencional en una experiencia semejante a una consola**, capaz de iniciar juegos de Steam, Epic Games, otras tiendas y emuladores desde una única interfaz diseñada para televisores y gamepads.

El objetivo no es crear otro administrador de bibliotecas. Es construir una alternativa a Steam Big Picture, Playnite y BigBox, poniendo especial atención en tres cualidades\:

- **Experiencia de consola\:** encender el PC, encontrar los juegos y comenzar a jugar sin utilizar ratón ni teclado.
- **Independencia de los proveedores\:** administrar una biblioteca unificada sin depender de una tienda específica, respetando los clientes oficiales y sus sistemas DRM.
- **Fiabilidad como shell de Windows\:** el launcher debe sobrevivir a errores, recuperar el foco al salir de los juegos y permitir administrar el equipo sin depender de Explorer.

ARQUITECTURA OBJETIVO

**UI de consola — TypeScript + React**

Pantalla de inicio · Biblioteca · Búsqueda · Detalles · Configuración

Bindings tipados · Eventos Wails

**Backend — Go**

Servicios de aplicación · Estado · Orquestación

**SQLite**

Juegos, fuentes, historial y ajustes

**Windows API**

Procesos, ventanas, mandos y shell

**Adaptadores externos**

## 2\. Decisiones tecnológicas

| Componente | Elección propuesta |
|--------|--------|
| Aplicación de escritorio | Wails v3 |
| Backend | Go |
| Frontend | React + TypeScript + Vite |
| Estilos | Tailwind CSS + componentes propios |
| Estado de interfaz | Zustand |
| Navegación | React Router + motor de foco espacial |
| Base de datos | SQLite con `modernc.org/sqlite` |
| Consultas | `sqlc` y migraciones SQL |
| Controladores | XInput inicialmente, adaptadores HID después |
| Interoperabilidad Windows | Win32 mediante `golang.org/x/sys/windows` |
| Datos y metadatos | Proveedores intercambiables y caché local |
| Distribución | Ejecutable Windows + instalador NSIS |
| Automatización | GitHub Actions |

### ¿Wails v2 o v3?

Al 9 de octubre de 2026, **Wails v2 sigue siendo la versión estable y Wails v3 continúa en beta**, con una versión beta.28 publicada el 5 de octubre. Wails v3 aporta servicios tipados, eventos y capacidades modernas de empaquetado, pero requiere validación para un uso tan exigente como reemplazar el shell del sistema. [v3.wails.io](https://v3.wails.io/changelog/)[v3.wails.io](https://v3.wails.io/faq/)[v3.wails.io](https://v3.wails.io/guides/build/windows/)

Elegiría **Wails v3 con una versión fijada**, sujeto a una prueba técnica inicial de estabilidad. Mantendría el dominio y los adaptadores Go independientes del framework para conservar una ruta de migración a Wails v2 si fuera necesario.

### Dos correcciones importantes a la investigación original

**Compatibilidad con Windows.** Shell Launcher está oficialmente disponible en Windows Enterprise, Education e IoT Enterprise, no en Windows Pro. Para Pro conviene contemplar un modo fullscreen con inicio automático, independiente del shell oficial, o validar una alternativa de shell personalizado sin presentarla como equivalente a Shell Launcher. [learn.microsoft.com](https://learn.microsoft.com/en-us/windows/configuration/shell-launcher/)[learn.microsoft.com](https://learn.microsoft.com/en-us/windows/client-management/mdm/assignedaccess-csp)

**Steamworks y Epic Online Services.** Ninguno debe considerarse requisito para importar o abrir juegos existentes. Steam Input puede actuar sobre los juegos ejecutados mediante Steam, mientras que su API nativa está orientada a la integración en aplicaciones que la implementan. La lectura de bibliotecas instaladas y la ejecución mediante los clientes oficiales son estrategias más apropiadas para la primera versión. [partner.steamgames.com](https://partner.steamgames.com/doc/features/steam_controller/concepts)[partner.steamgames.com](https://partner.steamgames.com/doc/features/steam_controller/getting_started_for_devs?l=malay\&language=english)

## 3\. Arquitectura interna

Recomiendo un **monolito modular**, no microservicios. Todo el sistema puede ejecutarse dentro del proceso principal de Wails, con módulos Go desacoplados y adaptadores para servicios externos.

Wails v3 permite exponer servicios de Go con bindings TypeScript generados y transmitir eventos desde el backend. Esa será la frontera entre lógica de negocio e interfaz. [v3.wails.io](https://v3.wails.io/features/bindings/methods/)[v3.wails.io](https://v3.wails.io/features/bindings/services/)[v3.wails.io](https://v3.wails.io/features/events/system/)

### 3.1 Módulos del backend

| Módulo | Responsabilidad |
|--------|--------|
| `library` | Catálogo unificado, favoritos, búsquedas y colecciones |
| `discovery` | Descubrir clientes, instalaciones, archivos y ROMs |
| `providers` | Adaptadores Steam, Epic, emuladores y ejecutables |
| `metadata` | Portadas, descripciones, identificaciones y caché |
| `launcher` | Preparar y ejecutar juegos |
| `sessions` | Detectar inicio, ejecución, finalización y duración |
| `input` | Lectura de gamepads, mapeos, repeticiones y eventos |
| `windows` | Gestión de ventanas, foco, pantalla completa y DPI |
| `system` | Inicio automático, diagnóstico, energía y modo kiosco |
| `settings` | Configuración persistente y perfiles |
| `updates` | Versiones, integridad y recuperación ante actualizaciones |

Separaría los proveedores en dos contratos: `LibraryProvider`, para descubrir juegos, y `LaunchAdapter`, para ejecutarlos. No todas las tiendas permiten consultar bibliotecas completas, y algunos títulos importados manualmente deben poder ejecutarse sin implementar un proveedor.

### 3.2 Estructura propuesta del repositorio

```

launcher/
├── cmd/
│   └── launcher/                 # Entrada Wails
├── internal/
│   ├── app/                      # Servicios Wails y composición
│   ├── domain/
│   │   ├── game.go
│   │   ├── installation.go
│   │   ├── launch.go
│   │   └── session.go
│   ├── library/
│   ├── discovery/
│   ├── launcher/
│   ├── sessions/
│   ├── input/
│   ├── metadata/
│   ├── settings/
│   ├── persistence/
│   ├── providers/
│   │   ├── steam/
│   │   ├── epic/
│   │   ├── retroarch/
│   │   ├── dolphin/
│   │   └── executable/
│   └── platform/
│       └── windows/
│           ├── gamepad/
│           ├── processes/
│           ├── windows/
│           ├── registry/
│           └── kiosk/
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── features/
│   │   │   ├── library/
│   │   │   ├── game-details/
│   │   │   ├── gamepad/
│   │   │   ├── onboarding/
│   │   │   └── settings/
│   │   ├── navigation/
│   │   ├── stores/
│   │   ├── themes/
│   │   └── bindings/            # Generados por Wails
│   └── package.json
├── migrations/
├── tests/
│   ├── fixtures/
│   └── integration/
├── build/
│   └── windows/
├── scripts/
├── docs/
└── .github/workflows/

```



### 3.3 Contratos principales

En vez de permitir que React ejecute comandos del sistema, todos los procesos deben pasar por el backend de Go.

```

type LibraryProvider interface {
    ID() string
    Discover(ctx context.Context) ([]GameInstallation, error)
}

type LaunchAdapter interface {
    Supports(game GameInstallation) bool
    Prepare(ctx context.Context, game GameInstallation) (LaunchSpec, error)
    Launch(ctx context.Context, spec LaunchSpec) (LaunchHandle, error)
}

type SessionTracker interface {
    Track(ctx context.Context, handle LaunchHandle) (*GameSession, error)
}

```



Los modelos `GameInstallation`, `LaunchSpec`, `LaunchHandle` y `GameSession` serán estructuras propias del dominio. El proveedor prepara una especificación de ejecución; el módulo `launcher` valida la operación y coordina su ciclo de vida.

Una instalación física no debe confundirse con un juego. El mismo título podría existir en Steam, Epic o como ejecutable independiente, pero aparecer una sola vez en el catálogo con varias opciones de lanzamiento.

### 3.4 Modelo de datos

El esquema inicial de SQLite debería incluir\:

| Tabla | Datos |
|--------|--------|
| `games` | Identidad canónica, título, descripción, plataforma |
| `installations` | Juego, proveedor, ruta, identificador externo, estado |
| `providers` | Tipo, configuración, habilitado |
| `assets` | Portadas, fondos, iconos, hashes y procedencia |
| `collections` | Colecciones personalizadas |
| `collection_games` | Relaciones colección-juego |
| `play_sessions` | Inicio, final, duración y resultado |
| `emulators` | Ejecutable, plataforma, parámetros y configuración |
| `rom_sources` | Directorios, extensiones y plataformas |
| `settings` | Preferencias persistentes |
| `profiles` | Perfiles locales, para una etapa posterior |

La biblioteca debe mantenerse disponible sin conexión. Utilizaría `%LOCALAPPDATA%` para los datos de usuario, la base de datos y las portadas cacheadas, y `%PROGRAMFILES%` para los binarios instalados.

## 4\. El subsistema más importante: gestión de juegos

Un launcher puede mostrar una biblioteca correctamente y aun así fallar como consola si no consigue iniciar, supervisar y abandonar una sesión de juego sin intervención.

Por eso desarrollaría este módulo antes de invertir demasiado esfuerzo en animaciones.

### 4.1 Máquina de estados de ejecución

**IDLE**

**PREPARING**

**LAUNCH\_REQUESTED**

**RUNNING**

**RESTORING**

**IDLE**

---

Desde PREPARING, LAUNCH\_REQUESTED o RUNNING pueden producirse errores, cancelaciones o tiempos de espera. El sistema debe resolverlos sin dejar la interfaz bloqueada.

El proceso de lanzamiento será\:

1. Validar que el título esté instalado y que su proveedor sea accesible.
2. Construir el comando o URI apropiado y ejecutar mediante el adaptador.
3. Confirmar que el juego realmente se inició, cuando sea posible.
4. Ceder el foco a la ventana del juego sin mantener el launcher encima.
5. Supervisar el proceso o la ventana relevante.
6. Al terminar, registrar la sesión y recuperar el fullscreen y el foco del launcher.

**Una consideración crítica\:** abrir `steam://run/...` o un URI de Epic no proporciona necesariamente un proceso hijo que represente el juego. No podemos llamar a `Wait()` y asumir que, al retornar, la partida terminó.

Cada proveedor necesitará una estrategia de seguimiento y un nivel de confianza\:

- **Alta\:** proceso que controlamos y cuyo ciclo de vida conocemos.
- **Media\:** proceso o ventana identificados por el juego y sus ejecutables conocidos.
- **Baja\:** lanzamiento entregado al cliente externo sin confirmación suficiente.

Con confianza baja, no se deben inventar tiempos de juego ni declarar que el título terminó. Además, no debe forzarse el cierre de Steam, Epic o un emulador cuando el usuario cierra el launcher.

## 5\. Gamepads: arquitectura controller-first

La investigación identifica Xbox, PlayStation y otros mandos como entradas prioritarias. Mi propuesta consiste en **capturar la entrada en Go y convertirla en acciones semánticas**, sin obligar a los componentes React a conocer las particularidades del dispositivo.

```

Gamepad físico
    ↓
Windows Input Adapter
    ↓
Normalización de botones y ejes
    ↓
Dead zones + debounce + key repeat
    ↓
Navigation Actions
    ├── NAV_UP / NAV_DOWN
    ├── NAV_LEFT / NAV_RIGHT
    ├── CONFIRM / BACK
    ├── CONTEXT_MENU
    └── HOME
    ↓
Wails Events
    ↓
Spatial Focus Engine (TypeScript)
    ↓
UI React

```



Para la primera versión implementaría XInput de manera nativa desde Go mediante `XInputGetState`, adecuado para mandos compatibles con XInput. Microsoft documenta un máximo de cuatro controladores bajo esta API. Para DualSense, Switch Pro y dispositivos HID genéricos, añadiría un segundo adaptador —evaluando SDL3— en la siguiente fase. [learn.microsoft.com](https://learn.microsoft.com/en-us/windows/win32/xinput/getting-started-with-xinput?redirectedfrom=MSDN)[learn.microsoft.com](https://learn.microsoft.com/en-us/windows/win32/api/xinput/nf-xinput-xinputgetstate)

El comportamiento requerido incluye navegación en cuatro direcciones, repetición controlada al mantener pulsado un botón, zonas muertas de sticks, conexión y desconexión en caliente, varios mandos, y gestión independiente del foco del navegador embebido.

No usaría el foco DOM convencional como único mecanismo de navegación. Desarrollaría un `SpatialFocusManager` que resuelva el siguiente elemento según su posición en pantalla y tenga reglas explícitas para modales, carruseles, pestañas y pantallas virtualizadas.

Cuando el juego esté activo, el launcher debe **dejar de interpretar botones de navegación**, evitando interferencias y aperturas accidentales. Al regresar, recuperará el elemento seleccionado previamente.

## 6\. Integración con tiendas y emuladores

Implementaría los proveedores progresivamente, priorizando los mecanismos locales y evitando depender de APIs privadas.

| Proveedor | Descubrimiento | Ejecución | Prioridad |
|--------|--------|--------|--------|
| Steam | VDF + manifiestos ACF | Cliente Steam / URI | P0 |
| Ejecutables Win32 | Importación manual | Ejecutable con argumentos | P0 |
| Epic Games | Manifiestos locales | Cliente Epic / URI validado | P1 |
| RetroArch | Directorios y ROMs | CLI + core | P1 |
| Dolphin | Directorios y ROMs | CLI | P1 |
| PCSX2 / PPSSPP | Directorios y ROMs | Adaptador particular | P2 |
| GOG / EA / Ubisoft | Integración local por proveedor | Cliente oficial | P2 |
| Heroic / Legendary | Configuración externa | CLI | P3 |

### Steam

Leer `libraryfolders.vdf` y `appmanifest_*.acf` para descubrir instalaciones en múltiples unidades. Esta estrategia no necesita autenticación adicional ni Steamworks para reconocer los juegos instalados. [github.com](https://github.com/beallio/game-install-finder)[github.com](https://github.com/caseyrmorrison/thelearner-depot)

El adaptador debe detectar instalaciones incompletas, actualizar el catálogo tras cambios, ejecutar por AppID y no confundir la apertura del cliente Steam con el inicio del juego.

### Epic Games

Primero leería los manifiestos locales del Epic Games Launcher, validando sus campos y rutas según las versiones reales instaladas. Para ejecutar, usaría los mecanismos del cliente que hayan sido verificados mediante pruebas.

Como integración opcional posterior, estudiaría Legendary\. Su CLI permite consultar instalaciones y ejecutar juegos con autenticación, pero añade dependencias y requisitos de mantenimiento. [github.com](https://github.com/legendary-gl/legendary/blob/master/README.md)

No implementaría un sistema OAuth propio para Epic durante el MVP.

### Emuladores

En RetroArch manejaría asociaciones entre plataformas, extensiones y cores instalados. En Dolphin utilizaría los parámetros documentados `--batch` y `--exec`, en vez de asumir que un comando genérico sirve para todas sus versiones. [github.com](https://github.com/dolphin-emu/dolphin/blob/master/Readme.md)

Los ROMs, BIOS y emuladores se configurarían como recursos externos; el launcher no distribuiría contenido protegido.

## 7\. Diseño de la experiencia visual

La aplicación debe tener dos modos de interacción construidos sobre el mismo sistema.



**Console Mode**

Pantalla completa, navegación mediante gamepad, portadas grandes, interfaz minimalista y transiciones fluidas.



**Desktop Mode**

Configuración avanzada de bibliotecas, emuladores, rutas, metadatos y mantenimiento con teclado y ratón.

Referencias visuales conceptuales, no diseños definitivos del producto.

Esto resuelve una dificultad importante: configurar un emulador o corregir metadatos usando exclusivamente un gamepad puede resultar frustrante, mientras que navegar por la biblioteca debería ser completamente natural desde un sofá.

### Pantallas principales

Arquitectura de navegación (8 vistas)

La interfaz utilizará imágenes de alta resolución con degradados que garanticen legibilidad, escala para 1080p y 4K, animaciones discretas y carga diferida de portadas. El efecto visual debe construirse alrededor del arte de cada juego, evitando fondos animados pesados que consuman recursos cuando el PC debería estar ejecutando un videojuego.

## 8\. Roadmap de implementación

Planificaría el desarrollo para una persona trabajando principalmente a tiempo completo, con apoyo de agentes de programación y pruebas continuas.

Estimación orientativa: MVP funcional en 6–8 semanas; beta ampliada en 12–16 semanas; primera versión robusta para kiosco en 18–24 semanas, según el hardware y las incompatibilidades encontradas.

La implementación de Shell Launcher real quedaría detrás de un instalador administrativo y un procedimiento reversible. El usuario de juegos seguiría sin permisos de administrador.

## 9\. Requisitos de calidad y pruebas

Estos objetivos deben tratarse como criterios de aceptación, no como prestaciones ya demostradas.

| Indicador | Objetivo |
|--------|--------|
| Arranque del launcher | Interfaz utilizable en 3 segundos o menos en hardware objetivo, una vez iniciado el proceso |
| Navegación | Respuesta percibida inferior a 100 ms |
| Animaciones | 60 FPS a 1080p en hardware objetivo |
| Biblioteca | 1.000 juegos sin bloqueos perceptibles |
| Controlador | 100 % de los flujos esenciales accesibles |
| Modo offline | Catálogo local plenamente funcional |
| Recuperación del juego | Retorno automático a la interfaz en casos verificables |
| Persistencia | Sin pérdida de favoritos o configuraciones tras cierre inesperado |
| Despliegue | Sin elevación requerida durante uso normal |

### Pruebas imprescindibles

Además de pruebas unitarias Go y Vitest para TypeScript, desarrollaría pruebas de integración con manifiestos reales anonimizados, comandos de ejecución falsos y simulaciones de desconexión.

La validación física deberá cubrir cambio de foco, Alt+Tab, apagado inesperado, fallos de WebView2, pérdida de Internet, desconexión del mando, clientes externos que muestran ventanas de inicio de sesión y juegos que necesitan privilegios especiales.

Para WebView2, utilizaría por defecto el runtime Evergreen con una verificación explícita durante la instalación. Microsoft recomienda mantenerlo actualizado y gestionar sus fallos de proceso. [learn.microsoft.com](https://learn.microsoft.com/en-us/microsoft-edge/webview2/concepts/evergreen-vs-fixed-version)[learn.microsoft.com](https://learn.microsoft.com/en-us/microsoft-edge/webview2/concepts/developer-guide)

## 10\. Riesgos que deben resolverse temprano

| Riesgo | Impacto | Mitigación |
|--------|--------|--------|
| Wails v3 beta | Alto | Fijar versión y probar recuperación del WebView |
| Input PlayStation/Switch | Alto | Abstracción de input y segundo adaptador |
| Foco tras salir de un juego | Crítico | Gestor de ventanas nativo + pruebas físicas |
| Epic cambia mecanismos internos | Medio | Adaptador aislado y fallback manual |
| Cliente externo no expone PID | Alto | Seguimiento con niveles de confianza |
| Fallo del launcher como shell | Crítico | Política de reinicio y acceso administrativo de recuperación |
| Actualización defectuosa | Alto | Instalación transaccional y rollback |
| Metadatos y portadas | Medio | Caché, atribución y proveedores con licencias verificadas |

No intentaría bloquear combinaciones protegidas de Windows ni deshabilitar su seguridad. Shell Launcher por sí mismo tampoco constituye un sistema de aislamiento: Microsoft indica que deben utilizarse controles complementarios cuando se requiere restringir otras aplicaciones. [learn.microsoft.com](https://learn.microsoft.com/en-us/windows/configuration/shell-launcher/)

## 11\. Backlog del primer sprint

Para comenzar el repositorio de forma que un agente de IA pueda implementar el trabajo sin improvisar la arquitectura, definiría las siguientes tareas\:

**Checklist técnico inicial**  de

[Copiar backlog]

## 12\. Definición exacta del MVP

## Versión 0.1 — Console Core

(Objetivo inicial)

El primer producto entregable debe permitir que una persona encienda su PC, encuentre su biblioteca de Steam, seleccione un juego utilizando un gamepad, lo inicie y, al terminar, regrese al launcher conservando su navegación y preferencias.

El MVP incluirá biblioteca local, detección de Steam, importación manual, portada por juego, favoritos, interfaz fullscreen, navegación XInput, ejecución y gestión básica de sesiones. No incluirá aún sincronización de cuentas, perfiles infantiles, temas descargables ni un sistema de plugins de terceros.

Mi recomendación de ejecución es **no empezar por los emuladores ni por un dashboard visualmente sofisticado**\. Comenzaría con un prototipo vertical que conecte Wails, Windows, XInput, Steam y el regreso desde el juego. Una vez demostrada esa cadena en el PC de destino, los módulos restantes podrán desarrollarse en paralelo y con un riesgo considerablemente menor.

Ese orden permite construir una aplicación real desde las primeras semanas, en lugar de acumular componentes individuales que podrían fallar al integrarlos en el entorno kiosco.