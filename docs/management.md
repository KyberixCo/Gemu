# Centro de gestión

**Estado actual de Juegos:** Steam importa juegos, metadatos, arte y perfiles mediante el backend. RetroArch incorpora detección, instalación portable y núcleos, configuración asistida, escaneo y lanzamiento de ROMs en Windows x64. Ver [Importación de bibliotecas](library-import.md) y [RetroArch](retroarch.md). Epic, GOG, la fuente manual y los demás emuladores siguen pendientes.

Acceso desde **Gestión** en la barra lateral, con navegación por teclado y mando, diseño adaptable y formularios. Las configuraciones generales se guardan explícitamente por jugador en SQLite bajo la clave `gemu.management.v1`; RetroArch se confirma dentro de su asistente junto al catálogo. La previsualización web utiliza localStorage para los borradores y requiere el host para RetroArch. Las claves API no forman parte del objeto persistido y se gestionan en el almacén seguro. Las fichas editadas y las rutas manuales de arte son borradores separados del catálogo visible. Steam, RetroArch y la búsqueda de arte usan el host; los demás emuladores y los proveedores LLM siguen pendientes.

## Contratos para el backend

- Fuentes: listar conectores y estado real; guardar rutas; descubrir instalaciones con progreso, cancelación y diagnósticos. Steam, Epic Games, GOG y manual tienen perfiles independientes. Habilitar un perfil en la demo no implica estar conectado.
- Emuladores: validar ejecutables y versiones, enumerar núcleos disponibles, configurar ROMs y pantalla. RetroArch añade núcleo libretro, shader y rebobinado. El backend debe traducir estos valores a argumentos y configuración real; no ejecutar cadenas de shell procedentes de formularios.
- Metadatos: leer/actualizar ficha por ID estable; buscar candidatos en proveedores; importar y validar arte; mostrar procedencia y vista previa antes de aplicar. Mantener catálogo e instalaciones separados. Guardar archivos bajo control del host.
- LLM: proveedor, URL base, modelo y presupuesto deseado. El host gestiona credenciales seguras, streaming, cancelación, límites efectivos y disponibilidad. Nunca insertar claves en localStorage. Ollama requiere modelo instalado; OpenRouter gratuito tiene límites y disponibilidad variable. El presupuesto del formulario todavía no limita consumo.
- Agente: herramientas para consultar biblioteca, inspeccionar perfiles, buscar metadatos y generar propuestas. Presentar diferencias y procedencia; aplicar solo propuestas aprobadas. Las solicitudes del primer corte generan únicamente una vista previa estática identificada como simulación.

## SQLite y FTS5

SQLite está integrado para perfiles, biblioteca y preferencias; contratos y migraciones en [Persistencia](persistence.md). La migración 2 incorpora FTS5, comprueba CREATE/INSERT/MATCH y reconstruye el índice de juegos existentes. La búsqueda de biblioteca ya consulta título, descripción y género, con filtros, ranking y paginación. Uso y API en [Búsqueda de biblioteca](search.md).

`library_documents` conserva propietario, juego, fuente y campos indexados. Su tabla FTS5 usa contenido externo y triggers para altas, cambios y bajas. MATCH recibe términos literales construidos desde el texto del usuario, con límites y ranking bm25. FTS5 es búsqueda textual, no semántica. Las notas de fuentes/emuladores y la conexión del asistente siguen pendientes; la columna `notes` está reservada y vacía. La API devuelve IDs y procedencia para permitir esa integración futura.

## Referencias de proveedores

Verificadas el 9 de octubre de 2026:

- MiniMax M3 y URL compatible `https://api.minimax.io/v1`: https://platform.minimax.io/docs/guides/text-generation
- Router gratuito `openrouter/free`: https://openrouter.ai/openrouter/free/apps
- API local de Ollama: https://github.com/ollama/ollama/blob/main/docs/api/introduction.mdx

Los modelos y endpoints permanecen editables para adaptar disponibilidad y coste. Este corte no valida credenciales ni conectividad.

## Mapeador visual de controles

**Gestión → Controles** ofrece plantillas Xbox, Nintendo Switch (Pro / Joy-Con combinados), DualShock 4 y DualSense. El dibujo seleccionable y el resumen permiten asignar botones físicos a acciones lógicas y atajos de emulación. Nintendo usa B/A/Y/X en posiciones sur/este/oeste/norte; PlayStation usa cruz/círculo/cuadrado/triángulo y una distribución simétrica de sticks. Los botones de sistema del dispositivo no se presentan como reasignables.

Perfiles independientes por destino (Global, RetroArch, Dolphin, PCSX2, PPSSPP) y familia. Zona muerta, sensibilidad e inversión Y se guardan junto con las asignaciones en `gemu.management.v1.controllers`. Configuraciones antiguas migran a un conjunto de perfiles vacío. Cada emulador tiene acceso **Mapear controles**. Guardar es explícito mediante el botón superior. Restaurar solo modifica el perfil seleccionado. La prueba es una simulación por clic sobre el diagrama, sin captura física.

Backend requerido: identificar dispositivos estables y capacidades USB/Bluetooth; capturar entrada sin activar la navegación de Gemu durante aprendizaje/prueba; definir correspondencia entre botones/ejes nativos y posiciones lógicas; aplicar perfil específico o heredar el global si no hay override. La calibración debe admitir sticks por separado. Traducir acciones a configuración/autoconfig real de cada emulador con vista previa, copia de recuperación y diagnóstico de acciones incompatibles. Validar dispositivos reales por sistema operativo: las plantillas visuales no garantizan compatibilidad de hardware. Joy-Con individuales, touchpad, giroscopio, vibración y gatillos adaptativos requieren contratos y capacidades adicionales.

## Refinamiento de gestión

La entrada abre **Metadatos**. La navegación separa Fuentes de juegos, Controles, Metadatos, Fuentes de metadatos, Fuentes de arte y Asistente. El encabezado señala cambios pendientes y habilita el guardado solo cuando hay modificaciones. Las fichas editadas se identifican en el listado.

Los gestores incluyen IGDB y ScreenScraper para metadatos, SteamGridDB, ScreenScraper y archivos propios para arte. Activación y prioridad se guardan dentro de `gemu.management.v1.contentProviders`; las configuraciones antiguas reciben valores por defecto. ScreenScraper comparte preferencias y credenciales entre ambos gestores. Cada proveedor enlaza su documentación oficial. Los valores nuevos viven en memoria; el gestor seguro permite persistirlos cifrados con DPAPI por perfil y eliminarlos por proveedor. La interfaz recibe solo su estado, no los secretos guardados. Una configuración habilitada nunca se presenta como conexión verificada. ScreenScraper y SteamGridDB ya se consultan durante la importación y ofrecen Probar acceso; detalles en [Integraciones de contenido](content-providers.md).

**Mejorar con el agente → Preparar mejora · demo** abre una comparación del valor actual y la propuesta por campo. La demo limpia espacios, sin LLM ni información inventada. Permite editar la propuesta, elegir campos explícitamente, aceptarlos al borrador o descartarla. No se guarda ni aplica al catálogo automáticamente; después hay que usar Guardar cambios. Cambiar de juego descarta la propuesta pendiente de revisión. Las propuestas no se persisten ni contienen secretos.

Quedan pendientes el intercambio OAuth de IGDB, llamadas LLM, edición manual del arte publicado y procedencia por campo de metadatos. ScreenScraper y SteamGridDB cuentan con búsquedas, descarga validada y procedencia por imagen; requieren configurar credenciales propias para validar las consultas reales. Referencias: https://api-docs.igdb.com/, https://www.screenscraper.fr/webapi2.php y https://www.steamgriddb.com/api/v2.

## Gestión para televisión

El flujo dedicado de arte está especificado en [Obtener y elegir arte con mando](artwork-flow.md): entrada desde la ficha y Gestión → Metadatos, búsqueda, selección de imágenes, comparación y guardado; lotes para completar faltantes. La búsqueda, elección y guardado individual, completar faltantes, cancelación y deshacer ya están implementados. La importación de Steam sigue siendo otra vía de descarga.

La gestión abre un menú de seis tareas: Juegos, Emuladores, Metadatos, Proveedores, Controles y Asistente. El área de trabajo ocupa el espacio entre el encabezado y la barra fija de acciones, sin desplazamiento vertical de la página. Las rutas de edición sustituyen el contenido en vez de acumular tarjetas. Se conservan el esquema de persistencia y las preferencias previas.

Los juegos se eligen en una cuadrícula de seis fichas. Cada ficha separa información, arte y mejora. Los emuladores separan rutas, pantalla y controles; el mapeador separa perfil, asignación, calibración y resumen. Los proveedores muestran una conexión por pantalla y separan preferencias de credenciales. La revisión de propuestas muestra un campo por vez y mantiene selección explícita antes de aceptar al borrador.

Los ajustes son filas grandes que abren un editor de un único valor. Las listas de opciones usan botones navegables y páginas de hasta seis opciones; no dependen de desplegables nativos para el mando. La edición de texto y secretos todavía requiere teclado: queda pendiente integrar entrada de texto nativa / teclado en pantalla con el host. El mando puede salir del campo hacia Aceptar o Cancelar. B y Escape cancelan primero la edición, después vuelven por los niveles de gestión y finalmente a Jugar. Al cancelar se recupera el foco del ajuste. Guardar y Volver permanecen visibles abajo.

Verificación: compilación y suite de pruebas. Pendiente revisión visual a 720p/1080p/4K, lectura a distancia y recorrido con dispositivo físico. El entorno de desarrollo bloquea abrir el puerto local, por lo que no se afirma validación visual del diseño. Los textos largos de una propuesta conservan desplazamiento interno como respaldo; la navegación general ya no usa una página larga.

### Feedback visual de televisión

El foco incluye borde lima, glow exterior e interior y una ampliación leve en tarjetas y filas. Se aplica también al foco programático del mando. Las selecciones confirmadas conservan un halo de menor intensidad para distinguirlas del objetivo actual. La pulsación da una respuesta breve de escala; el cambio de pantalla/editor usa fundido y el encabezado una entrada corta. Propuestas y grupos de ajustes repiten la transición al cambiar de campo o página. Los botones del dibujo mantienen su posición absoluta.

El modo de movimiento reducido elimina animaciones y cambios de escala, conservando el borde y el glow estáticos. No hay pulsación luminosa continua. Compilación verificada; ajuste visual del halo en un televisor físico pendiente.

El gestor central se encuentra en Gestión → Credenciales. Detalles de protección, uso y recuperación en [Gestor seguro de credenciales](credentials.md).
