# Persistencia con SQLite

La aplicación de escritorio guarda perfiles, bibliotecas, instalaciones, favoritos y preferencias en `gemu.sqlite3`, dentro del directorio de datos del entorno. En Windows, producción usa `%APPDATA%/Gemu`; `dev.ps1` usa `artifacts/development-data`. `GEMU_DATA_DIR` selecciona otro directorio. Los entornos tienen bases independientes.

## Datos y aislamiento

`internal/storage` mantiene el esquema y las migraciones numeradas mediante `PRAGMA user_version`. El driver `modernc.org/sqlite` no requiere CGO. Cada operación abre una conexión con claves foráneas, espera de bloqueo de cinco segundos, WAL y sincronización FULL. Las conexiones se cierran al terminar la operación.

Las tablas `profiles`, `sources`, `games`, `installations`, `libraries` y `preferences` conservan los datos. Metadatos, logros, configuración de fuentes y deshacer de arte se serializan en columnas JSON para conservar los contratos actuales. Juegos e instalaciones tienen filas independientes. Las referencias incluyen el propietario para impedir asociaciones entre jugadores.

La migración 3 añade `retroarch_setups`: configuración por jugador que se confirma en la misma transacción que sus juegos e instalaciones. La publicación comprueba revisiones del catálogo y que la configuración no haya cambiado desde su revisión. Los ejecutables y núcleos administrados se comparten dentro del entorno; partidas y estados de RetroArch permanecen separados por jugador y juego. Detalles en [RetroArch](retroarch.md).

Cada jugador tiene un ámbito identificado por su ID de perfil. Biblioteca, cuenta de Steam, logros, favoritos, preferencias de audio y movimiento, configuración de gestión, atajos de teclado y estado del asistente pertenecen a ese ámbito. El ámbito global contiene únicamente idioma e identidades de mandos, además del registro de perfiles. El backend comprueba que el jugador exista y recibe su ID explícito al guardar preferencias: una escritura pendiente conserva su propietario aunque cambie el perfil activo.

La caché compartida `art/` continúa en archivos. Las credenciales permanecen en el almacén cifrado por perfil y entorno; no se trasladan a SQLite. Los ajustes rechazan campos de credenciales.

## Migración de datos existentes

1. Al consultar perfiles, se importa `profiles.json` conservando IDs, orden, nombres y avatares.
2. Al seleccionar un jugador, se importa su `profiles/<id>/library.json`, incluidas fuentes, instalaciones, metadatos, logros, elecciones de arte y deshacer. Las bibliotecas se migran cuando se utilizan; no se necesita visitar todos los perfiles en el primer arranque.
3. Antes de montar la interfaz se importan idioma e identidades de mandos desde el WebView. Antes de entrar en cada perfil se importan sus claves `gemu.profile.<id>.*`: favoritos, feedback, gestión, configuración inicial y atajos.

Cada importación guarda los datos y su marca en `legacy_imports` dentro de una transacción. Se ejecuta una sola vez por entorno y ámbito. Las preferencias que ya existen en SQLite tienen prioridad. Una consulta sin datos del WebView no marca la migración como completada; un mapa vacío explícito representa una instalación nueva. Después de la migración, el escritorio lee SQLite sin depender de localStorage. La previsualización sin host conserva localStorage.

Los JSON y las claves del WebView anteriores se conservan como respaldo y dejan de actualizarse. No son una copia actualizada de los datos posteriores. La biblioteca global anterior no se asigna automáticamente a un jugador: su propietario es desconocido. La herramienta `cmd/library` utiliza un ámbito independiente `standalone` dentro de la base del directorio indicado; no publica en bibliotecas de jugadores.

Un archivo inválido, una versión incompatible o un error de escritura aborta la importación sin marcarla como completada. La interfaz muestra el error y permite reintentar después de corregir el origen. No se sustituye una base dañada por una vacía ni se borran automáticamente datos anteriores. El host debe reiniciarse para exponer los nuevos métodos `LoadPreferences` y `SavePreference`.

## Interrupciones y concurrencia

Las migraciones de esquema, importaciones y publicación de bibliotecas usan transacciones. El catálogo, instalaciones y deshacer de arte se confirman juntos. Una revisión de biblioteca impide sobrescribir cambios confirmados por otra instancia entre lectura y escritura; el conflicto requiere volver a cargar y reintentar.

El frontend hidrata una caché de lectura antes de entrar y serializa los guardados. Actualiza esa caché después de la confirmación del host y muestra los errores de guardado. Los cambios que aún no hayan recibido confirmación pueden perderse si se cierra el proceso.

WAL permite recuperar transacciones confirmadas y descartar las incompletas al abrir de nuevo. Para respaldar manualmente los datos, cierra Gemu antes de copiar el directorio completo. No elimines los archivos `gemu.sqlite3-wal` o `gemu.sqlite3-shm` de una sesión activa. La recuperación transaccional no sustituye un respaldo ante daños físicos del almacenamiento.

## Verificación y alcance

Las pruebas cubren migración y reintento de JSON inválido, importación única sin modificar respaldos, aislamiento de jugadores y entornos, conflictos de revisión, rollback de publicaciones incompletas y deshacer de arte. Una prueba abre una transacción en un subproceso y lo termina sin ejecutar limpieza: al reabrir se conserva el valor confirmado y se descartan el valor, la marca y el cambio de esquema incompletos. Esto comprueba interrupción de proceso, no un corte eléctrico real.

Las pruebas del frontend comprueban hidratación, importación por jugador, independencia del WebView después de migrar, orden de escrituras, errores del disco y propietario de operaciones pendientes. Queda pendiente recorrer la migración completa con una ventana nativa WebView2 y mando físico.

La migración 2 incorpora FTS5 y la búsqueda de biblioteca, con comprobación de soporte, indexación de juegos existentes y actualización transaccional del índice. Contratos y uso en [Búsqueda de biblioteca](search.md).

Referencias: [driver SQLite para Go](https://pkg.go.dev/modernc.org/sqlite), [pragmas de SQLite](https://www.sqlite.org/pragma.html).
