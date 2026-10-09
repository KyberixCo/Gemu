# Sprint siguiente · Console Core

Primer incremento completado: identidad visual, prototipo navegable, catálogo demo, favoritos, búsqueda, núcleo de lectura Steam y preparación de URI con validación.

1. Fijar y validar una versión de Wails v3 en Windows antes de adoptarla para shell. Agregar host y bindings que conecten la biblioteca real; mantener demo como modo explícito.
2. SQLite con migraciones: juegos, instalaciones, favoritos y ajustes. Probar escritura atómica y recuperación tras cierre.
3. Detectar Steam por rutas y registro de Windows; importar ejecutables con validación backend. Agregar fixtures anonimizados de instalaciones reales.
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
