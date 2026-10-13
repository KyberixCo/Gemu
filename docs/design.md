# Identidad de Gemu

## Referencias estudiadas

`../home/src/app/[lang]/globals.css`, `components/header.tsx` y `tailwind.config.ts`: negro #050505, papel #f2efe6, ácido #d9ff3f, microtipografía monoespaciada y Brolink como firma de marca.

## Pantalla de consola

La primera composición imitaba una web: cabecera, hero promocional, cuadrícula de tarjetas y pie corporativo. La revisión sustituye esa estructura por una única pantalla de uso: barra lateral de sistema, escenario de fondo, información del juego seleccionado, banda horizontal de biblioteca y ayudas de control abajo. No hay bloques publicitarios ni eslóganes.

El escenario es el fondo de toda la aplicación y responde a la selección con una transición breve. Cada juego tiene un entorno vectorial original: paisaje planetario, carretera nocturna, valle, ruinas, órbita o costa. La iluminación y el título diferencian los juegos. El ácido identifica selección, foco y acción principal. La firma de Kyberix queda como marca discreta en la barra de sistema.

La pantalla adapta las proporciones al viewport. En escritorio la navegación sucede dentro de la banda horizontal, sin scroll vertical del documento. Biblioteca y favoritos mantienen el contexto del juego y añaden filtros. Los detalles y ajustes aparecen en un panel lateral; el foco vuelve al elemento anterior al cerrarlo. En tamaños pequeños la banda conserva el desplazamiento horizontal.

## Interacción

Al entrar, el primer juego recibe el foco. Flechas, cruceta y stick usan navegación espacial; Enter o A activa, Escape o B cierra el panel. La selección cambia el escenario mediante foco, evitando saltos por mover el ratón sobre portadas. Búsqueda con `/`, favoritos persistentes, aislamiento de foco en paneles y movimiento reducido.

Los títulos y escenas son ficticios. No se muestran horas jugadas ni instalaciones inventadas; la acción es «Ver juego», con el estado «Sin instalación · Demo». Las portadas reales requerirán procedencia y caché.

La Gamepad API es provisional. El producto nativo deberá transformar XInput en acciones semánticas y suspenderlas durante una sesión real. Pendiente probar mandos físicos, escalado en TV y revisión visual en navegador: el entorno de desarrollo bloquea puertos y no dispone de navegador conectado.

## Movimiento y sonido

El contrato de eventos, tiempos, firma sonora, interrupciones, preferencias y criterios de aceptación se define en [Movimiento y sonido de Gemu](feedback.md). Los efectos acompañan cambios aceptados de estado y mantienen el foco inmediato; la especificación distingue lo ya construido de lo pendiente.

## Ambiente e iluminación dinámica

La referencia de consola se expresa mediante profundidad y espacio sonoro. No se reproduce una órbita con objetos flotantes: el movimiento pertenece a la iluminación del escenario. Tres capas de luz difusa varían levemente en opacidad y profundidad, y un reflejo tenue acompaña campanas y ecos. Ninguna portada, control o texto oscila con el sonido.

Audio y luz comparten el transporte de la composición procedural a 56 BPM. La respiración acompaña la armonía y los reflejos siguen las notas realmente emitidas, en lugar de campanas fijas cada seis segundos. El reloj de Web Audio marca los eventos y un analizador exclusivo del bus ambiente aporta intensidad. Las actualizaciones visuales se limitan a 20 Hz sobre transform y opacidad, sin rerenders de React por frame; se detienen con movimiento reducido o ventana inactiva. No se considera rendimiento verificado hasta medirlo en el host y TV.

Ambiente, volumen ambiental e iluminación tienen preferencias independientes. Silenciar ambiente no obliga a detener la atmósfera visual; sin audio usa un reloj visual suave. El volumen de interfaz no cambia el volumen ambiental. El ambiente baja temporalmente bajo confirmaciones.

El análisis textual y las reglas de transición están en [Composición procedural por juego](procedural-music.md). Se conserva la atmósfera angelical con variación y espacios de descanso.

## Superficies e interacción · revisión del 12 de octubre de 2026

La interfaz adopta superficies planas y esquinas rectas. Perfiles, gestión y ajustes se organizan mediante espacio, jerarquía tipográfica y separadores tenues, sin marcos alrededor de cada tarjeta o formulario. Los campos editables usan un fondo ligeramente distinto; la acción principal concentra el relleno ácido. Las formas circulares se reservan para avatares y representaciones de botones físicos del mando.

Hover, foco y selección tienen funciones distintas:

- Hover: cambio tenue de luminosidad o fondo, sin elevar ni ampliar el control.
- Foco de navegación: una marca ácida en un único borde acompañada de glow suave, visible con teclado y mando, sin contorno completo. Sobre la acción principal la marca es oscura para conservar contraste.
- Selección guardada: texto ácido, glow de menor intensidad y, en opciones y pestañas, un subrayado. No compite en intensidad con el foco actual.

Las portadas mantienen su arte sin marcos; el foco se dibuja sobre su borde inferior con luz difusa. Se oculta el índice decorativo del escenario, conservando el indicador de posición de la biblioteca. Los perfiles se presentan como una lista abierta y comparten los colores de la consola. El glow se conserva como parte de la identidad gaming, sin escalado ni pulsaciones continuas.

Las portadas importadas se reservan para las miniaturas y no participan en la rotación del escenario. Los fondos rasterizados se validan por sus dimensiones reales al cargar: al menos 1280 × 720 píxeles y anchura mayor que altura. Las imágenes pequeñas, verticales o que no se pueden cargar se descartan como fondo, incluso en bibliotecas ya guardadas. Si no existe un fondo válido se conserva el escenario procedural de respaldo. Este umbral es un mínimo de calidad; la calibración en pantallas 4K sigue pendiente.

Validación: compilación de producción y 88 pruebas del frontend correctas. La inspección visual sigue pendiente: el servidor local funciona, pero esta sesión no dispone de navegador conectado. Revisar en WebView2 la legibilidad a distancia y el recorrido con mando antes de considerar cerrado el pulido visual.
