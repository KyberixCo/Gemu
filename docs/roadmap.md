# Sprint siguiente · Console Core

Primer incremento completado: identidad visual, prototipo navegable, catálogo demo, favoritos, búsqueda, núcleo de lectura Steam y preparación de URI con validación.

1. Host Wails v2.16.0 y binding de consulta Steam implementados. Proceso genérico conectado al catálogo principal, con demo como modo explícito. Pendiente validar visualmente en Windows.
2. SQLite con migraciones: juegos, instalaciones, favoritos y ajustes. Probar escritura atómica y recuperación tras cierre.
3. Detección Steam por registro y rutas convencionales implementada; consulta real desde Gestión → Juegos → Steam. Pendiente importar ejecutables con validación backend.
4. Implementar máquina de estados: idle, preparing, launch requested, running, restoring, error. No inferir una sesión real a partir de la apertura del cliente.
5. XInput con acciones semánticas, repetición y desconexión. Pausar input mientras el juego esté activo.
6. Validar en Windows el ciclo Steam → juego → Gemu, recuperación del foco y conservación de la selección. Pruebas físicas con TV, mando, Alt+Tab y cliente sin sesión iniciada.

Criterio de salida: seleccionar y abrir un juego instalado con mando, confirmar su ejecución cuando la evidencia lo permita y regresar con selección y favoritos conservados. Epic, emuladores y modo shell quedan fuera de este sprint.

## Trabajo de movimiento y sonido

Especificación: [Movimiento y sonido de Gemu](feedback.md). Se implementa junto al corte vertical, sin retrasar la prueba de Steam → juego → retorno.

- **F1 · Base:** tokens de movimiento, preferencias Sistema/Reducido/Completo, sonido/volumen y coordinador de eventos. Salida: ajuste persistente y eventos deduplicados.
- **F2 · Navegación:** foco inmediato, fondo estabilizado con dos capas, banda y paneles interrumpibles. Salida: navegación sostenida y A/B rápido sin pérdida de foco ni colas.
- **F3 · Audio:** paquete original con procedencia, precarga, voces limitadas, activación de navegador y botón de prueba. Salida: confirmación coherente, silencio efectivo y ningún tick por restauración de foco.
- **F4 · Sesión nativa:** señales de lanzamiento/retorno basadas en estados Go; suspensión de audio y movimiento durante el juego. Depende de los puntos 4–6 del sprint.
- **F5 · Calibración:** TV, auriculares, Windows/WebView2, movimiento reducido y presupuesto de frames. Salida: criterios de aceptación de la especificación medidos en hardware de referencia.

F1–F3 ya tienen un primer corte sobre el catálogo demo: bienvenida, preferencias, sonidos originales, navegación audible y transiciones. Quedan deduplicación completa mediante IDs de acción, pulido de interrupciones del fondo y calibración visual/auditiva. No reproducir señales de lanzamiento ni retorno de sesión en la demo. La revisión incorpora ambiente continuo con iluminación dinámica sutil, controles independientes y suspensión por foco. Música y animación de geometría flotante quedan fuera. La bienvenida con sonido se incorporó por petición del producto, con control independiente y omisión.

## Validación de este incremento

- `go test ./...` y `go vet ./...`: correctos.
- `npm run build`: TypeScript y bundle de producción correctos.
- `npm test`: tres pruebas de filtros correctas.
- No fue posible inspeccionar la UI en navegador en este entorno: el servidor Vite no puede abrir `127.0.0.1:5173` (`EPERM`). Queda pendiente revisar visualmente y probar mandos físicos; no se afirma validación nativa en Windows.

## Música procedural

Implementado un primer compositor local sobre el texto del juego, con frases variables, descanso y cambio de carácter sin reiniciar transporte ni cortar voces. Especificación en [Composición procedural](procedural-music.md). Pendiente calibración auditiva prolongada y rendimiento en hardware de referencia.

## Primer corte del backend Steam · 11 de octubre de 2026

`App.DiscoverSteam(root)` expone una consulta de solo lectura. Una ruta vacía detecta Steam por registro y rutas convencionales de Windows; una ruta explícita debe ser absoluta y contener `steamapps`. Lee `libraryfolders.vdf` y los manifiestos ACF de todas sus bibliotecas. El resultado incluye IDs estables, títulos, AppIDs, rutas, estado de instalación, contador de juegos instalados y diagnósticos. Una biblioteca inaccesible o un manifiesto inválido no descartan los juegos válidos. Steamworks Common Redistributables (228980) se excluye del catálogo de juegos.

El panel Gestión → Juegos → Steam ofrece un asistente con detección automática como opción principal y configuración manual de carpeta como alternativa: consulta Steam, muestra juegos e incidencias y guarda la ruta/habilita la fuente al finalizar. Cancelar descarta el proceso y las respuestas tardías. El guardado afecta solo a Steam y conserva otros borradores. No persiste un catálogo, inicia Steam ni lanza juegos.

Validación local: tres juegos instalados detectados, sin diagnósticos. `wails build`, pruebas de Go, `go vet ./...` y 85 pruebas del frontend correctos. La validación visual de la ventana y del recorrido con mando sigue pendiente. El siguiente corte conecta esta biblioteca a Jugar y habilita la solicitud de lanzamiento con validación backend.
## Importación genérica · estado actual

Implementado el contrato `library.Source` y el asistente común `SourceWizard`: listado de instalaciones → elección de perfil → metadatos, arte y logros → revisión y publicación. Steam es el primer adaptador registrado. Las fuentes disponibles se consultan al backend; el asistente y los filtros de la biblioteca se reutilizan para nuevos adaptadores.

La biblioteca principal usa el catálogo publicado, incluye metadatos y arte local, muestra perfil/logros por título y ofrece un modo demo explícito. Persistencia inicial en JSON versionado con reemplazo mediante archivo temporal; SQLite sigue pendiente. La API tiene prioridad para perfil y logros cuando hay clave; el perfil local y el último progreso guardado del mismo perfil sirven como fallback. Sin evidencia de logros quedan pendientes. Las credenciales no se persisten.

Prueba real en esta máquina: tres juegos, metadatos, seis imágenes y perfil local, sin diagnósticos, publicados en una biblioteca de prueba del proyecto. La integración autenticada de logros se probó con HTTP simulado; pendiente validarla con una clave real. `wails build`, pruebas de Go, `go vet ./...` y 84 pruebas del frontend correctos. Pendiente revisión visual nativa y con mando.

Siguiente corte: lanzamiento validado, estados de sesión y recuperación del foco. Detalles y contratos en [Importación de bibliotecas](library-import.md).
## Perfiles locales de jugadores · 12 de octubre de 2026

Implementado el asistente inicial de creación, entrada automática tras el logo con un único perfil y selector cuando hay varios. Gestión → Jugadores permite crear perfiles y cambiar de jugador. Biblioteca, asociación Steam y logros se guardan por perfil; favoritos y ajustes del frontend usan claves separadas. Se comparte únicamente la caché de arte. Persistencia JSON atómica; SQLite sigue pendiente.

Desarrollo usa un almacén independiente en artifacts/development-data. El comando `.\dev.ps1 -ResetData` conserva una copia y reinicia el flujo de primera configuración. Contratos, compatibilidad con datos anteriores y recorrido de prueba en [Perfiles de jugadores](player-profiles.md).

Validación de perfiles: `go test ./...`, `go vet ./...`, 88 pruebas del frontend y `wails build` correctos. Se verifica también el cambio de jugador durante una importación en curso sin bloquear ni publicar datos en el otro perfil. Pendiente revisión visual en WebView2 y con mando físico.

## Limpieza visual · 12 de octubre de 2026

Aplicada una primera revisión transversal a perfiles, biblioteca, ajustes y gestión: superficies sin marcos, esquinas rectas en formularios y controles, hover sutil y foco con una sola marca de borde y glow suave, sin escalado. La selección conserva un glow de menor intensidad. Criterios en [Diseño de la experiencia](design.md). Compilación y 88 pruebas del frontend correctas; pendiente inspección visual en WebView2 y navegación con mando físico.

Arte de Steam: las portadas quedan fuera de la galería de fondos. La carga y rotación del escenario comprueban dimensiones reales (mínimo 1280 × 720, horizontal), también para arte previamente importado; imágenes pequeñas o inválidas usan el respaldo procedural. No es necesario reimportar la biblioteca para aplicar esta protección.

## ScreenScraper y SteamGridDB · 12 de octubre de 2026

Implementados los adaptadores Go, comprobación de acceso, prioridades y credenciales de sesión en la importación Steam. Descarga de arte estático validada, galería local con procedencia y rotación solo de fondos; metadatos de ScreenScraper con coincidencia única para PC Windows. Los fallos de proveedor permiten importar Steam y reutilizar la galería anterior. Contratos, recorrido y límites en [Integraciones de contenido](content-providers.md).

Validación: pruebas Go y `go vet ./...`, 93 pruebas del frontend, compilación frontend y build Wails correctos. Pendiente consulta con credenciales reales e inspección en WebView2. Los sistemas retro, selección manual de candidatos e integración visual de logos/iconos quedan para un siguiente corte.

Compatibilidad con hosts abiertos antes de la integración: comprobar `CheckContentProvider` antes de probar acceso o importar con proveedores, con instrucciones de reinicio en lugar de un TypeError. ScreenScraper distingue credenciales de cuenta y de desarrollador en las etiquetas y aclara que Metadatos y Arte las comparten. Validación del ajuste: compilación frontend y 95 pruebas correctas.

## Gestor seguro de credenciales · 12 de octubre de 2026

Implementado el almacén cifrado con DPAPI de Windows, separado por perfil y entorno, con escritura atómica y fallo cerrado. Gestión → Credenciales y los editores de proveedores, Steam y asistente permiten guardar, reemplazar y eliminar. La interfaz consulta solo el estado; Go recupera las claves guardadas al probar acceso e importar, con prioridad de los valores temporales y sin guardado automático. Contratos y límites en [Gestor seguro](credentials.md).

Validación: pruebas con DPAPI real y credenciales ficticias bajo el usuario Windows, pruebas de Go, `go vet ./...`, 96 pruebas frontend, compilación frontend y build Wails. Pendiente revisión visual y consulta con cuentas reales; IGDB y llamadas al asistente siguen fuera de este incremento.

El gestor depende ahora de la interfaz genérica `credentials.Vault`; fábrica por sistema operativo y capacidades informadas a la UI. Windows conserva DPAPI. Los sistemas sin adaptador permiten credenciales temporales y rechazan la persistencia, sin fallback en texto plano. Verificada compilación del paquete para Linux y macOS; sus almacenes nativos y el soporte completo de esas plataformas quedan pendientes.

## Asistentes de configuración y acceso ScreenScraper · 12 de octubre de 2026

Implementados `WizardFrame` y `ProviderSetupWizard`, compartidos por Gestión → Información y arte, Credenciales y el maestro `FirstRunSetup`. El maestro se abre para perfiles sin biblioteca y se puede repetir desde Gestión → Configuración guiada; reutiliza el asistente existente de importación Steam. ScreenScraper distingue registro de Gemu, cuenta personal, comprobación, preferencias y revisión. Sin acceso comprobado se guarda pendiente, sin activar ni bloquear el inicio.

La comprobación verifica acceso al catálogo, además de la cuenta, y distingue errores de aplicación, servicio restringido, bloqueo y cuota sin mostrar secretos. Falta obtener/configurar las credenciales de desarrollador asignadas a Gemu para validar una consulta real; crear una cuenta personal no completa ese requisito. Detalles y alcance en [Configuración guiada](setup-wizards.md). Pendiente revisión visual en WebView2 y mando físico.

Validación del incremento: 101 pruebas frontend, build frontend, pruebas Go, `go vet ./...` y compilación Wails correctos. El registro y acceso reales de Gemu en ScreenScraper no se dan por comprobados.

## Acceso de aplicación a ScreenScraper en el build · 13 de octubre de 2026

Reservados `GEMU_SCREENSCRAPER_DEVID`, `GEMU_SCREENSCRAPER_DEVPASSWORD` y `GEMU_SCREENSCRAPER_SOFTNAME` para inyección durante la compilación del backend. `build.ps1` genera una configuración temporal ignorada por Git y la elimina al terminar; `dev.ps1` reutiliza el mecanismo. El ajuste posterior descrito abajo hace obligatoria la configuración sin flags.

El usuario final configura únicamente su cuenta personal. El proveedor ignora credenciales de aplicación enviadas por perfiles o frontend; los almacenes anteriores siguen siendo legibles. Parámetros, uso y límites de protección del ejecutable en [Acceso de aplicación a ScreenScraper](screenscraper-build.md).

Verificación: 101 pruebas frontend, build frontend, pruebas Go, `go vet ./...` y compilación Wails correctos. Comprobadas inyección con valores ficticios, limpieza del archivo temporal tras éxito/fallo y rechazo de builds que exigen credenciales ausentes. El ejecutable final se compiló sin credenciales de aplicación; el acceso general a la API con credenciales propias sigue pendiente.

### Carga automática de `.env` · 13 de octubre de 2026

Eliminado el flag de activación: `build.ps1` y `dev.ps1` cargan `.env` desde la raíz y validan el ID y contraseña antes de invocar Wails. Las variables de entorno tienen prioridad para CI. `.env.example` reserva los nombres sin incluir secretos; `.env` permanece ignorado por Git. El lector acepta valores literales, comillas y comentarios sin ejecutar ni expandir el contenido.

Verificación con `scripts/test-build-env.ps1`: carga, prioridad de entorno, configuración solo de CI, credenciales ausentes/incompletas, errores redactados, inyección automática sin flags y limpieza tras éxito/fallo mediante Wails simulado en un directorio aislado. La recompilación real de este ajuste queda pendiente mientras la sesión Wails activa ocupa su archivo temporal de configuración.


### Corrección del catálogo de sistemas ScreenScraper · 13 de octubre de 2026

Resuelto «sistema PC Windows no encontrado»: el adaptador esperaba `nom`, pero `systemesListe.php` devuelve nombres regionales en el objeto `noms`. Ahora identifica PC Windows mediante esos nombres, conserva el formato simple y utiliza el ID del catálogo. Los casos ausentes, sin ID o ambiguos no se guardan en caché.

Verificación: pruebas Go y `go vet ./...` correctos; pruebas de cuenta y acceso anónimo con el formato real, distinción de DOS/Windows antiguos y caché. Una consulta temporal con el adaptador corregido resolvió PC Windows como ID `138` en la lista pública real, sin usar ni mostrar el `.env`. Esto verifica la lectura del catálogo, no el acceso a búsquedas de juegos con las credenciales configuradas.

## Diseño del flujo de arte con mando · 13 de octubre de 2026

Especificado el recorrido de [Obtener y elegir arte con mando](artwork-flow.md). Accesos propuestos: Cambiar arte desde la ficha y Gestión → Arte para completar faltantes o elegir juegos. Búsqueda con proveedores activados, miniaturas paginadas, comparación, selección por tipo y publicación explícita, con cancelación, recuperación de foco, incidencias y deshacer. También se documentó la ruta actual de actualización Steam para obtener arte con las integraciones existentes.

Implementación pendiente: operación de arte independiente de la importación, consulta de candidatos, progreso/cancelación en el host, persistencia de selecciones y asociación de juego, actualización inmediata del catálogo y revisión con mando físico. El commit actual reemplaza la fuente completa; no debe reutilizarse para guardar arte de un único juego. Orden previsto: recorrido individual, completar biblioteca y resolución de coincidencias. Este incremento es de diseño/documentación y no incorpora aún las pantallas propuestas.

## Flujo de arte implementado y portadas en la rotación · 13 de octubre de 2026

Implementado Gestión → Metadatos → juego → Arte → Buscar imágenes, con selección por tipo, vista previa, revisión y guardado independiente de la importación. Acceso también desde Cambiar arte en Jugar. Completar arte faltante permite elegir juegos, consultar con progreso/cancelación y revisar resultados parciales. Persistencia de elecciones manuales, fondos de rotación, logotipo y deshacer del último cambio; conflictos de revisión/deshacer no sobrescriben cambios posteriores. Las coincidencias ambiguas quedan como incidencias pendientes, sin selección automática.

La portada elegida se incorpora al escenario cuando sus dimensiones reales cumplen 1280 × 720 y formato horizontal; se alterna con los fondos y conserva su uso de portada. Se recuperan las dimensiones de portadas locales antiguas sin volver a descargarlas. No se duplican imágenes por URL. Las portadas aptas también se pueden añadir explícitamente a la rotación; se mantienen las transiciones y preferencias de movimiento existentes.

Verificación: pruebas de guardado por juego, conservación de otros datos, cancelación efectiva, conflictos, deshacer tras reinicio, elecciones manuales y rotación. Pruebas frontend de portada apta, rechazo de baja resolución, exclusión de imágenes no elegidas y deduplicación; build frontend y compilación Wails del flujo correctos. El usuario comprobó el flujo de búsqueda/guardado. No se realizó validación visual automatizada ni con mando físico: el servicio nativo de computer-use no estaba disponible.

## Uso completo de grids, logotipos e iconos · 13 de octubre de 2026

Eliminada la restricción de grids a 920×430 y 460×215 en SteamGridDB. Las tarjetas muestran la imagen completa conservando su proporción. El selector de arte incluye Grid / Portada, Fondo, Logotipo e Icono. Los iconos recorren búsqueda, validación, selección, persistencia, completar faltantes y deshacer; se recuperan los ya guardados en galerías anteriores.

El logotipo elegido sustituye el título visual en el escenario y la ficha, con respaldo textual si falta o falla la imagen. El icono identifica el juego junto al nombre en el escenario, las tarjetas del catálogo y Gestión. Ambos permanecen separados de la rotación de fondos y las elecciones manuales sobreviven a la importación.

Verificación: 108 pruebas frontend, build frontend, pruebas Go y `go vet ./...` correctos. Pruebas específicas de consulta de grids sin restricción de dimensiones, recepción de logos/icons, persistencia/restauración de iconos, roles de presentación y título accesible. Revisión visual con mando físico pendiente.

## Recuperación del build tras una interrupción · 13 de octubre de 2026

`build.ps1` ya no bloquea por la mera existencia de `screenscraper_build_local.go`. Recupera automáticamente archivos residuales del generador y conserva archivos no reconocidos. Un mutex por repositorio impide sesiones simultáneas; la limpieza y liberación se ejecutan incluso si Wails falla. Windows libera el bloqueo cuando termina el proceso.

Verificadas recuperación de archivo residual, exclusión de builds concurrentes sin modificar la configuración activa, liberación tras éxito/fallo y conservación de archivos propios. El build real recuperó el archivo que bloqueaba `dev.ps1`, compiló Gemu y eliminó su configuración temporal.

## Filtros MIME de logos e iconos SteamGridDB · 13 de octubre de 2026

Corregido HTTP 400 en logos e iconos: se enviaba `image/jpeg` como filtro común para todas las categorías, aunque esas dos no lo admiten. Se solicita PNG para logos/icons y se mantiene PNG/JPEG para grids/heroes. La prueba del adaptador reproduce el rechazo HTTP 400 de los filtros inválidos y exige recibir las cuatro categorías sin incidencias.

Verificación: pruebas de proveedores/contenido y biblioteca, y `go vet` de ambos paquetes correctos. No se modifican credenciales ni se necesita volver a configurar la conexión. Repite la búsqueda de imágenes después de cargar el host actualizado.
