# RetroArch: instalación y configuración asistidas

En Windows de 64 bits, **Gestión → Juegos → RetroArch** detecta instalaciones existentes y ofrece instalar una copia portable oficial cuando hace falta. El mismo asistente está disponible desde **Gestión → Emuladores → RetroArch** y desde la configuración inicial. Se puede repetir para cambiar carpetas, añadir consolas o actualizar el catálogo.

## Recorrido

1. **Detectar:** busca el ejecutable configurado, copias administradas por Gemu, ubicaciones convencionales, PATH, App Paths del registro y bibliotecas Steam. Se puede elegir un ejecutable con el diálogo nativo o escribir la ruta con el teclado virtual de Gemu. La detección comprueba que sea un ejecutable Windows x64; no lo ejecuta. Una carpeta de núcleos existente es opcional.
2. **Consolas:** elegir los sistemas que se van a utilizar. Se reutilizan núcleos válidos de la carpeta indicada, los administrados por Gemu o la carpeta `cores` junto al ejecutable. Solo se descargan los que falten.
3. **Componentes:** consultar la versión estable en el servidor oficial, revisar destino, lista de descargas, tamaño y espacio recomendado. **Descargar e instalar** es la acción que inicia la instalación. RetroArch procede del canal estable; los núcleos individuales proceden del canal del actualizador oficial, que publica compilaciones actuales. La información y licencia de cada núcleo se conserva en sus archivos `.info`. Sin descargas pendientes se puede continuar sin conexión.
4. **Carpetas:** elegir una carpeta por consola, con subcarpetas opcionales. **Crear y usar carpetas de Gemu** prepara directorios privados de juegos y BIOS dentro del perfil. También se puede utilizar una biblioteca externa sin mover archivos. La carpeta de BIOS puede apuntar a una ubicación existente.
5. **Experiencia:** configurar ventana o pantalla completa sin bordes, rebobinado y controlador de mando Automático/XInput/SDL2/DirectInput. El botón de prueba abre el menú real de RetroArch; permite comprobar dependencias, vídeo, sonido y controles. Es necesario cerrar esa prueba antes de iniciar otra partida desde Gemu.
6. **Revisar:** comprobar juegos, núcleos, BIOS e incidencias; guardar configuración e incorporar juegos. El guardado actualiza únicamente RetroArch para ese jugador y conserva las demás fuentes. También admite configurar carpetas aún vacías para añadir juegos después y repetir el asistente.

## Sistemas y formatos

| Sistema | Núcleo preparado | Formatos indexados |
| --- | --- | --- |
| NES | FCEUmm | NES, UNF, UNIF |
| Super Nintendo | Snes9x | SFC, SMC |
| Game Boy / Game Boy Color | Gambatte | GB / GBC |
| Game Boy Advance | mGBA | GBA |
| Mega Drive / Genesis | Genesis Plus GX | MD, GEN, BIN, SMD |
| Master System / Game Gear | Genesis Plus GX | SMS / GG |
| Nintendo 64 | Mupen64Plus-Next | N64, Z64, V64 |
| PlayStation | Beetle PSX HW | CUE, CHD, PBP, M3U |

Los sistemas de cartucho también admiten ZIP cuando contiene un único juego del formato seleccionado. Otros sistemas, núcleos alternativos, ROMs en 7z, escaneo por DAT y bibliotecas arcade quedan para otros incrementos; no se presentan como compatibles automáticamente por tener RetroArch instalado.

Las listas M3U de PlayStation agrupan los discos y evitan publicar también sus discos individuales. Los CUE deben referenciar pistas existentes. Se rechazan descriptores incompletos y referencias que escapen de la carpeta de la colección; no se siguen enlaces fuera de ella. Los BIN de PlayStation no se indexan por separado. Los archivos auxiliares como SBI siguen siendo responsabilidad del usuario y del núcleo.

El escaneo admite hasta 20 000 juegos y 100 000 entradas por colección, con un plazo de dos minutos. Las incidencias de archivos individuales permiten revisar resultados parciales; un núcleo ausente impide publicar. La revisión expira tras quince minutos y vuelve a comprobar el contenido antes de confirmar para evitar incorporar juegos que no se hayan revisado.

## BIOS y controles

Gemu no proporciona juegos ni BIOS. Comprueba los nombres y MD5 conocidos de `scph5500.bin`, `scph5501.bin` y `scph5502.bin`, e informa de su región. Un archivo presente con checksum incorrecto bloquea la publicación y el lanzamiento hasta corregirlo. Una BIOS ausente se muestra como ausente: el soporte sin BIOS depende de la versión de Beetle PSX HW y del juego y no se da por comprobado. La documentación actual del núcleo describe OpenBIOS como alternativa; la versión instalada debe probarse.

Los mandos utilizan los perfiles de autoconfiguración incluidos en RetroArch. El asistente permite elegir otro controlador si es necesario; su funcionamiento debe comprobarse con el dispositivo y transporte reales. Los mapeos del editor de Gemu siguen siendo independientes y no se convierten automáticamente en perfiles de RetroArch.

RetroArch carga una configuración base privada por jugador y un archivo complementario generado por lanzamiento. Para conservar cambios hechos en su menú, se utiliza **Guardar configuración actual**; el archivo base se reutiliza en la siguiente partida. Las opciones controladas por Gemu y las carpetas de guardado se vuelven a aplicar mediante el archivo complementario. No se sobrescribe la configuración de una instalación externa ni se guardan automáticamente cambios al salir.

**L3 + R3** o **F1** abre el menú propio de RetroArch. **F10** o **View + Menu** abre el menú externo de Gemu. Son menús distintos: las funciones de guardar estados, cambiar disco y ajustar el núcleo se encuentran en RetroArch. El rebobinado depende del soporte del núcleo y de los recursos disponibles.

## Instalación y recuperación

El host solo construye URLs HTTPS de `buildbot.libretro.com` y de la carpeta oficial `libretro-core-info` en GitHub. No recibe URLs arbitrarias de la interfaz. Un plan revisado tiene token de un solo uso, propietario y caducidad; otra sesión de jugador no puede iniciarlo ni cancelarlo.

Los paquetes se descargan a un área temporal privada, con progreso, plazo máximo de treinta minutos y cancelación. La extracción valida nombres, tipos, tamaño, CRC y arquitectura PE; rechaza rutas que escapen del destino, enlaces, archivos especiales y duplicados. El lector 7z está incluido en Go: no se necesita instalar 7-Zip ni ejecutar un instalador externo.

La publicación utiliza archivos completos y renombrados dentro de la misma unidad. Un mutex de Windows evita instalaciones simultáneas entre procesos. Al reintentar se recupera el bloqueo y se limpian únicamente los directorios temporales propios; las versiones completas y las instalaciones externas se conservan. Una copia administrada dañada puede repararse si conserva el recibo de Gemu; la anterior se mueve a `recovery/`. Las carpetas sin ese recibo no se reemplazan. Los componentes publicados antes de un fallo posterior siguen siendo reutilizables al reintentar.

El recibo conserva URLs, versión y SHA-256 de las descargas. El hash sirve como registro de la descarga; no se presenta como una firma del editor ni se compara con un digest independiente del distribuidor. Las licencias incluidas en el paquete se conservan. No se cambia el registro, el shell ni el inicio automático del equipo.

## Persistencia y lanzamiento

La migración SQLite 3 incorpora `retroarch_setups`. La configuración y el catálogo se confirman en una sola transacción, incluidos los cambios del índice FTS5. Se comprueba que la configuración no haya cambiado desde la revisión. Una reexploración conserva metadatos, arte elegido y progreso ya guardados de los mismos IDs.

Dentro del directorio de datos de cada entorno:

- `emulators/retroarch/releases/<versión>/`: paquetes completos administrados.
- `emulators/retroarch/cores/` e `info/`: núcleos e información compartidos entre jugadores del entorno.
- `emulators/retroarch/recovery/`: copias administradas anteriores reparadas.
- `profiles/<id>/retroarch/roms/` y `system/`: carpetas opcionales creadas por el asistente.
- `profiles/<id>/retroarch/saves/<juego>/` y `states/<juego>/`: partidas y estados separados por jugador y juego.
- `profiles/<id>/retroarch/`: remaps, opciones de núcleos, capturas, logs y configuraciones de lanzamiento.

La identidad del juego depende del sistema y la ruta normalizada del archivo; dos juegos con el mismo nombre en carpetas distintas tienen partidas separadas. Mover un archivo cambia su identidad; este incremento no migra automáticamente sus partidas a la nueva ruta.

Cada lanzamiento crea un archivo de configuración complementario y utiliza argumentos directos `--config`, `--appendconfig`, `-L` y la ruta del juego, sin construir comandos de shell. No se modifica el `retroarch.cfg` de una instalación existente. La sesión se confirma por el PID, fecha de creación y ruta del proceso iniciado; no por procesos ajenos de la misma carpeta. Si RetroArch ya está abierto, se pide cerrarlo o continuar la sesión existente.

Continuar, volver a biblioteca y solicitar cierre reutilizan el [menú de partida](game-sessions.md). El cierre es una solicitud normal a su ventana, no una terminación forzada. Al terminar el proceso se restaura Gemu y su selección.

Los metadatos iniciales son locales. La búsqueda de arte con SteamGridDB se mantiene disponible desde la biblioteca; ScreenScraper no busca juegos retro como si fueran PC Windows. Su mapeo de sistemas retro, la importación de logros y las configuraciones avanzadas de cada núcleo quedan pendientes.

## Verificación

Las pruebas usan archivos de muestra y transporte HTTP simulado para instalación, cancelación, corrupción, reparación, bloqueo, rutas de archivo y extracción 7z/ZIP. Comprueban discos y playlists, pérdida de requisitos, configuración concurrente, guardado atómico, aislamiento por jugador y seguimiento exacto del proceso. Los ejecutables de prueba estructurales nunca se ejecutan como emuladores.

`GEMU_RETROARCH_LIVE_CHECK=1 go test ./internal/retroarch -run TestOfficialEndpointsLive -v` consulta versiones y cabeceras de paquetes oficiales sin instalarlos. `GEMU_RETROARCH_PACKAGE_CHECK=1 go test ./internal/retroarch -run TestOfficialPackageLive -v -timeout 12m` descarga y extrae la distribución completa y FCEUmm en un directorio temporal. La prueba real del 15 de octubre de 2026 resolvió 1.22.2, descargó y extrajo el paquete oficial, validó su ejecutable y el núcleo y terminó correctamente. No ejecuta RetroArch; la prueba ordinaria permanece offline.

Pendiente validar ejecución gráfica, ROMs reales, mandos físicos, BIOS y retorno en Windows/WebView2. Las pruebas de instalación e integración no sustituyen ese recorrido.

Fuentes oficiales consultadas: [distribución de RetroArch](https://www.retroarch.com/?page=platforms), [versiones estables](https://buildbot.libretro.com/stable/), [línea de comandos](https://docs.libretro.com/guides/cli-intro/), [directorios](https://docs.libretro.com/guides/change-directories/), [Beetle PSX HW](https://docs.libretro.com/library/beetle_psx_hw/) y [combinaciones de mando](https://github.com/libretro/RetroArch/blob/v1.22.2/input/input_defines.h).
