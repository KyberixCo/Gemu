# Movimiento y sonido de Gemu

Estado: contrato de diseño con implementación parcial. Ya existen bienvenida animada, nueve sonidos WAV locales, motor Web Audio compartido, preferencias persistentes, foco audible, escenarios estabilizados y entrada/salida de paneles. Siguen pendientes los eventos de sesión nativa, el coordinador completo con IDs de acción y la calibración física. Los tiempos y niveles son valores iniciales a ajustar en TV y mando.

## Dirección creativa

Gemu debe sentirse como un dispositivo de juego: preciso al navegar, tranquilo al esperar y contundente al confirmar. El movimiento explica qué cambió; el sonido confirma que la acción fue aceptada. La biblioteca mantiene protagonismo sobre ambos.

La firma sonora es electrónica, cálida y espacial: parciales cristalinos ligeramente desafinados, base grave, aire suave, reflejos estéreo y reverberación difusa. La referencia solicitada es la sensación de profundidad de una consola como PS2; el motivo y las muestras de Gemu son originales. Evitar sonidos de monedas, clics estridentes y voz. La bienvenida puede tener una cola larga; la navegación usa ataques redondeados de 35 ms, volumen discreto y una nota mi consonante con la armonía común. Mantener el mismo timbre al cambiar de juego; las escenas pueden variar visualmente sin cambiar el significado de los controles.

La revisión de producto incorpora ambiente continuo original: una composición procedural local basada en el texto del juego, con armonía mayor, voces suaves, campanas y frases de descanso, sin canción ni ritmo percusivo. La identidad mantiene selección, confirmación y retorno. Por petición del producto, este corte incorpora un motivo de bienvenida como preferencia independiente, activada por defecto a volumen bajo. Se inicia con Enter/A/Entrar; B/Escape/Omitir lo descarta. En navegador, un mando puede no habilitar audio sin un gesto adicional aceptado. La animación dura 1,85 segundos (1,3 con movimiento reducido) y es omisible en cualquier momento.

## Catálogo de efectos

Los sonidos se emiten por el resultado de la acción, no directamente por el botón físico. Las duraciones de audio incluyen su cola.

| Evento semántico | Movimiento | Sonido | Condición |
| --- | --- | --- | --- |
| `welcome.started` | Fragmentos del emblema, letras y línea de señal; salida a la biblioteca en 1,85 s. Reducido: composición estática y salida en 1,3 s. | `welcome`: motivo espacial de 3,92 s, con ecos y cola reverberante que termina sobre la biblioteca. | Inicio explícito; omitir detiene voces y descarta desbloqueos pendientes. |
| `focus.changed` | Contorno inmediato; elevación de 3 px en 100 ms. Desplazamiento de la banda en 160 ms solo si hace falta revelar el elemento. | `nav.tick`: ataque cristalino de 95 ms; duración total de 415 ms con reflejos y cola suave. | Solo si el foco realmente cambia por teclado/mando. Sin sonido al enfocar inicialmente o restaurar el foco. |
| `selection.settled` | Tras 120 ms sin otro cambio, fundido entre escenarios en 320 ms. Título entra con opacidad y 6 px en 160 ms. | Ninguno. | El título y metadatos se actualizan inmediatamente; el arte espera. Conservar el fondo anterior hasta que el nuevo esté listo. |
| `section.changed` | Indicador de la barra en 140 ms; biblioteca en fundido de 160 ms. | `section.shift`: transición suave con cola de sala; 715 ms en total. | Solo si cambia la sección; sustituye al tick de foco en esa acción. |
| `panel.opened` | Panel desde 32 px y opacidad, 220 ms. Velo de fondo en 180 ms. | `panel.open`: notas ascendentes y reflejos; 940 ms en total. | Un efecto por apertura efectiva. El foco entra de inmediato, sin esperar el movimiento. |
| `panel.closed` | Panel sale 24 px en 160 ms; velo desaparece en 160 ms. | `panel.close`: variante descendente; 790 ms en total. | Escape/B fuera de un panel no produce efecto. Restaurar foco al solicitar cierre, manteniendo la salida visual sin interceptar entrada. |
| `favorite.saved` | Corazón crece hasta 1.12 y vuelve, 180 ms. | `favorite.on`: dos notas cristalinas con cola; 940 ms en total. | Después de confirmar persistencia. |
| `favorite.removed` | Corazón cambia a contorno en 100 ms. | `favorite.off`: una nota con cola corta; 680 ms en total. | Después de confirmar persistencia. |
| `launch.accepted` | Acción pasa a «Preparando»; progreso estático por estado. Contraer acciones en 140 ms. | `launch.commit`: motivo de dos notas, 160–220 ms. | Cuando el backend acepta una solicitud validada, una vez por solicitud. No en el catálogo demo. |
| `launch.handed_off` | Atenuar la interfaz en 200 ms al ceder foco. | Ninguno. | Handoff confirmado al cliente o proceso; no equivale a confirmar que el juego está ejecutándose. |
| `session.returned` | Recuperar selección y fondo con fundido de 220 ms. | `return.home`: motivo de confirmación cálido, 120–160 ms. | Solo en retorno verificado de una sesión y con ventana visible/en foco. No al volver de un Alt+Tab cualquiera. |
| `operation.failed` | Mensaje inline o aviso en fundido de 120 ms. Conservar la selección. | `error.soft`: nota grave y reverberación discreta; 600 ms en total. | Un sonido por error visible; sin sacudidas, rojo intermitente ni alarmas. |
| `controller.connected` / `controller.disconnected` | Actualizar icono y aviso discreto en 120 ms. | `device.on` / `device.off`: variante breve de 70–100 ms. | Solo cambios reales y con Gemu activo. Ninguno en sondeo inicial. |

La búsqueda, el reloj, los cambios de hover, los filtros sin cambios efectivos y la navegación contra un borde permanecen silenciosos. Al cambiar de filtro se permite el efecto de sección si el conjunto cambia, sin repetir un tick adicional. En ratón, la confirmación explícita puede sonar; mover el puntero nunca produce una secuencia de ticks.

## Tokens de movimiento

Centralizar estos valores para evitar duraciones distintas por componente:

| Token | Valor inicial | Uso |
| --- | --- | --- |
| `motion.instant` | 0 ms | Foco, estado semántico, entrada aceptada |
| `motion.quick` | 100 ms | Respuesta del elemento |
| `motion.short` | 160 ms | Banda, título, cierre |
| `motion.panel` | 220 ms | Apertura y retorno |
| `motion.scene` | 320 ms | Fundido de escenarios |
| `motion.settleDelay` | 120 ms | Espera antes de cambiar el arte |
| `motion.easeOut` | `cubic-bezier(0.2, 0.8, 0.2, 1)` | Entradas y selección |
| `motion.easeIn` | `cubic-bezier(0.4, 0, 1, 1)` | Salidas |

Animar principalmente `transform` y `opacity`. El contorno de foco aparece inmediatamente. No animar dimensiones, blur de pantalla completa, sombras grandes, desplazamiento de cámara, parallax ni objetos geométricos flotantes. El ambiente visual solo modula capas de luz difusa, reflejos y profundidad a un ritmo lento; se suspende al perder foco y con movimiento reducido. El velo puede cambiar opacidad; el desenfoque, si se conserva, será fijo y debe medirse en WebView2.

## Repetición e interrupciones

- El motor de entrada decide cuándo repetir una acción; el motor de efectos no añade repetición propia.
- La selección semántica siempre sigue al mando. Los fondos conservan solo el último destino pendiente; nunca encolar escenas intermedias.
- Si se interrumpe un fundido, reutilizar su composición visible como punto de partida y limitar a dos capas de escenario. No remontar y reiniciar todo el fondo por cada tick.
- Abrir/cerrar un panel durante su transición invierte desde la posición actual. No bloquear A/B hasta terminar ni reproducir efectos de una acción cancelada antes de ser aceptada.
- Máximo un `nav.tick` cada 90 ms; sin sonido si no cambia el foco. Confirmaciones eliminan ticks programados en el mismo ciclo.
- Máximo dos voces principales de efectos de interfaz más un transporte musical procedural independiente limitado a doce voces. Prioridad: error visible > lanzamiento/retorno > confirmación > navegación. Los ticks pueden descartarse; nunca se ponen en cola para sonar después.
- Al perder visibilidad/foco de ventana o al ceder entrada al juego, detener audio, cancelar efectos pendientes y suspender movimiento decorativo. El retorno no reproduce eventos acumulados.

## Sonido, niveles y recursos

Efectos locales estéreo, con reverberación y ecos horneados en los archivos, sin diferencias de volumen perceptual grandes. Usar WAV PCM a 48 kHz y 16 bits. El paquete espacial ocupa 1,93 MB; fijar un presupuesto de 2 MB para estos nueve efectos; documentar autoría y licencia de cada archivo. El paquete debe ser original o tener licencia explícita compatible con distribución comercial.

Pico máximo del archivo: −6 dBFS. Ganancia inicial del bus: 20 %, pendiente de calibración auditiva; no normalizar cada tick al máximo. Usar envolventes suaves de entrada y salida para evitar clics eléctricos. Navegación debe sentirse más baja que confirmación y error. Estos valores no garantizan un nivel acústico fijo: validar con altavoces de TV y auriculares.

Precargar/decodificar una sola vez al habilitar audio. Reutilizar un contexto de audio y buffers; no crear un contexto ni leer un archivo por pulsación. La falta de un recurso deja la UI operativa y silenciosa.

En la vista de navegador, activar o reanudar audio solo tras una interacción permitida. Si el contexto sigue suspendido —incluida navegación con mando sin gesto aceptado por el navegador— permanecer en silencio y ofrecer activación explícita en ajustes. Nunca reproducir después los eventos anteriores al desbloqueo. Validar este comportamiento en el host Wails; no asumir que permite autoplay.

## Preferencias y accesibilidad

Ajustes de producto a persistir localmente:

- **Sonidos de interfaz:** activados a volumen bajo; interruptor independiente. Antes de un gesto permitido, silencio efectivo.
- **Volumen de interfaz:** 0–100 %, valor inicial 20 %. Silenciar corta voces activas y pendientes. 0 equivale a silencio.
- **Probar sonido:** acción explícita que reproduce una confirmación al volumen elegido. Disponible sin salir de ajustes.
- **Movimiento:** Sistema (predeterminado), Reducido o Completo. Sistema sigue `prefers-reduced-motion`; Reducido elimina elevación, desplazamientos y zoom, permitiendo solo fundidos de hasta 80 ms. Si la preferencia del sistema es reducida, Completo requiere una elección explícita del usuario.

La reducción de movimiento no silencia audio y el silencio no elimina confirmaciones visuales. Ambiente, volumen ambiente e iluminación dinámica se controlan de forma independiente; apagar sonidos de interfaz no apaga el ambiente. El silencio de bienvenida desactiva ambos buses. Ningún estado se comunica solo mediante sonido, color o movimiento. Mantener contornos y textos estáticos en modo reducido. La configuración queda disponible con teclado/mando y conserva valores tras reinicio.

## Integración técnica prevista

`Input → acción semántica → cambio aceptado de estado → evento de feedback → movimiento/audio`.

Rutas propuestas en frontend:

- `src/features/feedback/events.ts`: unión tipada de eventos, origen (`keyboard`, `gamepad`, `pointer`, `system`), ID de acción/solicitud y motivo (`user`, `initial`, `restore`).
- `src/features/feedback/motion.ts`: tokens y resolución de la preferencia de movimiento.
- `src/features/feedback/audio.ts`: contexto compartido, buffers, bus de volumen, límite de voces, prioridades y desbloqueo.
- `src/features/feedback/coordinator.ts`: deduplicación, supresión por visibilidad/sesión y emisión de efectos.
- `src/features/settings/feedback.ts`: preferencias persistentes; almacenamiento del navegador en demo y servicio de ajustes Go en escritorio.
- `public/audio/ui/`: paquete de sonidos y manifiesto de procedencia.

Un evento aceptado recibe un ID estable para que clic, tecla, bindings y render no produzcan sonidos duplicados. Nunca emitir audio desde el render de React ni desde un efecto que vuelva a ejecutarse por StrictMode. La restauración de foco y el montaje inicial están marcados como silenciosos. El resultado de `moveFocus` deberá indicar si cambió el destino; no emitir por cada intento de desplazamiento.

El frontend resuelve los efectos locales. Go conserva autoridad sobre lanzamiento, persistencia y sesión. Los eventos nativos incluyen ID de solicitud y evidencia/estado: no convertir un URI entregado a Steam en `running` o `session.returned`. Hasta conectar persistencia confirmada, no presentar el sonido de favorito como garantía de guardado ante un fallo de almacenamiento.

## Orden de implementación

1. Tokens, preferencias y coordinador tipado; sustituir duraciones CSS sueltas y respetar movimiento reducido. Mantener audio silencioso mientras no exista el paquete.
2. Foco con resultado, selección estabilizada, fundido de escenarios con dos capas y salida de paneles interrumpible. Conservar el foco instantáneo.
3. Producir y revisar paquete original. Conectar audio precargado, silencio, volumen, prueba y desbloqueo. Afinar primero navegación/confirmación; después favoritos, paneles y avisos.
4. Conectar estados nativos: lanzamiento, retorno, error y cambios de mando. Silenciar y suspender al entrar al juego.
5. Calibrar en Windows/WebView2, TV 1080p y 4K, mando y auriculares. Medir el coste antes de ampliar los efectos.

## Criterios de aceptación

- Mantener dirección pulsada durante 10 segundos: selección inmediata, sin cola de fondos, sin sonidos tardíos y sin crecimiento continuo de voces/capas.
- Pulsar A y B alternadamente durante una transición: el panel refleja el último estado aceptado, no se bloquea ni pierde el foco.
- Entrar, restaurar foco, escribir búsqueda y llegar a un borde: ningún tick accidental. Una acción por ratón/Enter/A produce una sola confirmación.
- Desactivar sonido, poner volumen a cero o perder foco durante un efecto: salida cortada suavemente en un máximo de 30 ms, sin retorno de sonidos pendientes.
- Habilitar movimiento reducido: sin traslaciones ni zoom; selección legible y todos los flujos disponibles.
- Recurso de audio ausente o bloqueo de autoplay: biblioteca funcional, sin excepciones no gestionadas ni reintentos por pulsación.
- Juego activo: cero audio de interfaz y cero animación decorativa. Recuperación verificada: selección conservada y una sola señal de retorno.
- Objetivo de respuesta de foco: menos de 100 ms; objetivo de fluidez: 60 FPS a 1080p en hardware de referencia. Son metas a medir, no prestaciones demostradas.

Pruebas automatizadas: reloj y audio simulados para deduplicación, prioridades, límites, interrupciones, silencio y descarte al suspender; integración para foco y paneles. Pruebas físicas: sonoridad, fatiga al navegar, recuperación tras juego y coste de composición en TV/WebView2.

## Implementación de este corte

- `features/feedback/audio.ts`, `policy.ts`: un contexto/buffers compartidos, precarga, volumen, prioridades, ticks limitados y cancelación de previsualizaciones pendientes. Máximo dos voces principales; una voz desalojada deja una cola de apagado de hasta 25 ms.
- `preferences.ts`, `useFeedback.ts`: validación y persistencia de ajustes, preferencia del sistema y suspensión por foco/visibilidad.
- `Welcome.tsx`: entrada explícita, omisión y repetición desde ajustes. `input.ts` conserva el estado de botón mantenido entre cambios de pantalla.
- `WorldBackdrop.tsx`: selección estabilizada y máximo dos escenarios. Los fundidos nuevos parten de la última escena elegida; la continuidad exacta de la composición visible al interrumpir requiere pulido adicional.
- Paneles con transición CSS reversible, cierre sin bloqueo y restauración silenciosa de foco.
- `scripts/generate-ui-audio.py`: síntesis reproducible de nueve WAV propios, con picos verificados por debajo de −6 dBFS.

Las pruebas automáticas cubren preferencias, admisión de ticks, voces prioritarias, silencio, archivos ausentes, desbloqueo fallido, descarte de bienvenida y repetición del mando. No sustituyen escucha, inspección visual ni pruebas físicas.

## Revisión de la paleta espacial

La síntesis se amplió con parciales desafinados, ruido filtrado tenue, cuatro reflejos alternados para la bienvenida y una sala de combs amortiguados con difusión allpass. Procesamiento offline: cero coste de reverberación adicional durante navegación. Los efectos habituales tienen menor intensidad y parciales agudos atenuados; navegación y paneles conservan el volumen del fondo, mientras confirmación, bienvenida y error lo reducen suavemente un 18%; se mantienen dos voces principales y prioridad de confirmación sobre ticks.

La finalización visual natural invalida previsualizaciones pendientes sin detener la voz ya iniciada. Omitir, volumen cero, silencio o pérdida de foco detienen tanto ataque como reverberación. Movimiento reducido mantiene el mismo audio y no alarga la espera visual. La prueba de ajustes usa la nueva confirmación espacial.

## Composición procedural e iluminación

La pista repetitiva fue sustituida por [Composición procedural por juego](procedural-music.md). El análisis local de título, subtítulo, género y descripción define densidad, brillo, apertura y motivo. Una estructura de presentación, respuesta, desarrollo y descanso evita repetir un ciclo corto sin cambios.

El cambio de selección espera 600 ms, luego entra en una frontera común de cuatro pulsos, conservando reloj, armonía compatible, voces y reverberación. Nuevas notas con entrada suave y salida natural de las anteriores; la luminosidad cambia con notas efectivamente emitidas y energía del bus musical. Un solo destino pendiente evita encadenar selecciones rápidas.

Volumen, silencio y movimiento reducido siguen siendo independientes. Las pruebas cubren planificación musical, continuidad, límites, pausa/reanudación y forma de onda de instrumentos; falta evaluar musicalidad, comodidad y rendimiento mediante escucha prolongada y pruebas reales.
