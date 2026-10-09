# Desarrollo de Gemu

Un launcher de videojuegos para convertir el PC en una experiencia de consola. Implementación inicial de [idea.md](../idea.md).

## Ejecutar la interfaz

Requisitos: Node.js 22.12+ y npm.

```sh
cd frontend
npm install
npm run dev
```

Abre la URL que imprime Vite. `npm run build` comprueba TypeScript y genera el bundle; `npm test` prueba los filtros de biblioteca.

La interfaz funciona offline después de cargar los recursos locales. Incluye seis juegos ficticios con escenarios vectoriales propios, búsqueda, filtros, favoritos guardados en el navegador, detalles y pantalla completa. La pantalla de consola tiene un escenario a todo el fondo, una banda horizontal de juegos, barra lateral de sistema y paneles laterales de detalles y ajustes. El foco inicial cae sobre un juego; al moverte cambia el escenario y el título. Los juegos de demostración no se pueden ejecutar. Flechas/Enter/Escape y mandos compatibles con la Gamepad API permiten navegar. Un mando requiere interacción inicial con el navegador; esta API no sustituye XInput nativo.

## Bienvenida, movimiento y sonido

Al abrir, pulsa **Enter / A / Entrar** para iniciar la bienvenida animada y su motivo sonoro. **Escape / B / Omitir** entra directamente a la biblioteca y descarta el audio pendiente. Una segunda confirmación durante la animación también la omite. La nueva firma sonora dura 3,92 segundos y su reverberación termina sobre la biblioteca cuando la entrada es normal; omitir corta el sonido. En navegador, el sonido necesita una interacción que habilite Web Audio; si un mando no logra habilitarlo, usa el botón Entrar o Probar sonido con ratón/teclado.

La navegación tiene ticks suaves solo al cambiar de foco, con sonidos distintos para secciones, paneles y favoritos. El escenario espera 120 ms de selección estable antes de una transición con desplazamiento, escala y fundido; el foco y los metadatos responden inmediatamente. Los paneles tienen entrada y salida interrumpibles y restauran el foco sin otro tick.

El fondo seleccionado mantiene un desplazamiento y zoom suaves con recorridos de 36 segundos. Cada título admite una galería `artworks` en `frontend/src/catalog.ts`; la demo incluye tres variantes procedurales por juego. Las imágenes se alternan tras 18 segundos de permanencia con transiciones de 1,4 segundos. Las imágenes externas se cargan antes de la rotación y se omiten si fallan; una imagen principal fallida conserva el escenario procedural. La rotación no cambia el título ni reinicia la música. Perder foco pausa el desplazamiento y detiene la rotación; movimiento reducido o iluminación dinámica desactivada dejan el fondo estático.

Para usar imágenes locales, guarda los archivos en `frontend/public/art/` y añade, por ejemplo, `artworks: [{id:'panorama',type:'image',src:'/art/panorama.webp',position:'70% center'}, {id:'noche',type:'image',src:'/art/noche.webp'}]`. La primera imagen también se usa en la tarjeta. Los títulos sin galería mantienen su escenario procedural.

**Ajustes** incluye silencio, volumen, sonido de bienvenida independiente, movimiento Sistema/Reducido/Completo, Probar sonido y Repetir bienvenida. Las preferencias son locales. El movimiento reducido elimina desplazamientos y conserva el motivo sonoro; el audio se corta al perder foco o visibilidad de ventana. No se reproducen sonidos de lanzamiento ni de sesiones ficticias.

El **ambiente de fondo** ahora es una composición procedural derivada del título, subtítulo, género y descripción del juego. Voces suaves, campanas y pulsaciones forman frases con presentación, respuesta, desarrollo y descanso. La armonía permanece luminosa, sin subgraves; motivos, densidad y timbre varían con el texto.

Al mantener una selección 600 ms, se solicita un cambio musical que entra en la siguiente frontera de frase corta. Las notas actuales completan su salida y las nuevas entran suavemente, manteniendo el reloj y la reverberación. Navegar rápidamente conserva un único destino pendiente. La iluminación sigue las notas reales y la energía de la composición; el arte, las portadas y los controles permanecen estables.

En ajustes puedes cambiar **Ambiente de fondo**, **Volumen ambiente** e **Iluminación dinámica** por separado. El ambiente baja bajo confirmaciones y pausa al perder foco; recupera posición y armonía mediante fundido al volver. Movimiento reducido congela la luz y conserva el sonido. La síntesis es local y no descarga pistas ni envía los textos a terceros. Más detalles en [Composición procedural](procedural-music.md).

Nueve WAV estéreo originales con reverberación y eco (aprox. 1,93 MB en total) se distribuyen localmente; fuente y procedencia en [el generador](../scripts/generate-ui-audio.py) y [el manifiesto](../frontend/public/audio/ui/README.md). Regenera con `python3 scripts/generate-ui-audio.py`. Pendiente escuchar/calibrar en TV y auriculares y revisar visualmente en un navegador: este entorno no permite abrir el puerto de Vite.

## Núcleo Go

Requisitos: Go 1.23+.

```sh
go test ./...
go run ./cmd/discover -steam-root '/ruta/a/Steam'
```

El inspector lee `libraryfolders.vdf` y manifiestos ACF, consulta múltiples bibliotecas, distingue instalaciones incompletas y devuelve diagnósticos sin descartar los juegos válidos. El dominio separa juegos de instalaciones. El preparador valida AppIDs y construye URIs Steam con confianza baja: todavía no ejecuta procesos ni contabiliza sesiones.

## Estado del primer corte

Frontend y dominio Go están separados; todavía no están conectados. Pendientes para el siguiente corte vertical: host Wails con versión fijada, bindings, SQLite, descubrimiento de Steam en Windows, lanzamiento real, XInput, seguimiento del juego y recuperación del foco. No hay cambios al shell, registro ni inicio automático del equipo.

Consulta [las decisiones de diseño](design.md), [el plan de movimiento y sonido](feedback.md) y [el siguiente sprint](roadmap.md).

## Gestores y asistente · vista previa

**Gestión** abre fuentes (Steam, Epic Games, GOG y manual), perfiles de emuladores (RetroArch, Dolphin, PCSX2 y PPSSPP), edición de metadatos y configuración del asistente (MiniMax M3, OpenRouter gratuito, Ollama y APIs compatibles). Guarda borradores localmente mediante **Guardar configuración**; las claves API solo viven en memoria. Las operaciones están señaladas como demo y las fichas todavía no se aplican al catálogo. Contratos de backend y diseño de búsqueda SQLite FTS5 en [Centro de gestión](management.md).

**Gestión → Controles** añade un mapeador visual de Xbox, Nintendo Switch, PS4 y PS5 con asignaciones por emulador, ajustes de sticks y prueba simulada. Cada perfil de emulador incluye acceso directo al editor. La lectura física y la aplicación de perfiles requieren backend.

## Licencia

Gemu se distribuye bajo la GNU General Public License, versión 3 exclusivamente (`GPL-3.0-only`). Consulta el texto completo en [LICENSE](../LICENSE). Las dependencias y los recursos de terceros conservan sus respectivas licencias.
