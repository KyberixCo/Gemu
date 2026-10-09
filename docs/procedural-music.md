# Composición procedural por juego

Estado: primera implementación funcional en el frontend. Sustituye el acompañamiento WAV de 24 segundos. La procedencia de los nueve efectos de interfaz sigue siendo independiente.

## Identidad musical

Carácter luminoso y tranquilo: voces sostenidas, campanas suaves y pulsaciones ligeras. Sin subgraves ni percusión dominante. Los cambios de juego alteran la composición sin convertir cada selección en un nuevo arranque musical.

El texto incluye título, subtítulo, género y descripción. Se normaliza y se analiza localmente con un vocabulario bilingüe; una semilla del texto define el motivo y las variaciones. No se envía texto a servicios externos. Esta interpretación es una heurística explícita, no un modelo semántico ni una composición producida por un LLM.

| Carácter del texto | Respuesta musical |
| --- | --- |
| Exploración, universo, planetas | Frases espaciadas, mayor apertura estéreo, campanas |
| Valle, senderos, hogar | Pulsaciones suaves, armonía cálida, mayor cercanía |
| Océano, islas, marea | Líneas fluidas, ecos y espacio entre notas |
| Carreras, curvas, arcade | Más movimiento melódico y pulsaciones, sin acelerar el pulso global |
| Recuerdos, historias, decisiones | Menor densidad, frases reflexivas y descanso |

Los títulos con el mismo carácter conservan motivos distintos gracias a la semilla textual. El título o descripción editados pueden cambiar ese motivo.

## Forma y variación

Transporte compartido de 56 BPM. Frases de 16 pulsos forman un arco de cuatro frases: presentación, respuesta, desarrollo y descanso. Los acordes cambian cada ocho pulsos siguiendo ocho pasos armónicos que cubren el arco completo y regresan al acorde inicial. Los siguientes arcos introducen sustituciones compatibles en los pasos interiores.

La melodía usa cuatro patrones rítmicos con entradas en pulsos pares e impares. Cada presentación garantiza entre cuatro y seis notas según el perfil; la respuesta invierte el motivo, baja su registro objetivo y desplaza algunas entradas medio pulso. El desarrollo eleva el registro objetivo e intercala respuestas del segundo instrumento. El descanso conserva solo dos notas más suaves y omite una entrada armónica. Los patrones cambian de forma determinista por arco, conservando una relación reconocible entre sus frases.

Paletas mayor, menor y dórica, con tres variantes de progresión determinadas por el título. Las melodías permanecen entre Do4 y Do6; el bajo suave aporta raíces desde Do3. La conducción de voces busca desplazamientos pequeños entre acordes. La melodía deriva de un motivo de ocho grados del texto y se adapta a la escala diatónica: favorece movimientos pequeños y resuelve en notas de acorde en posiciones fuertes y al terminar cada frase.

La secuencia no reproduce una grabación corta en bucle: el índice de frase produce nuevas decisiones. Los sonidos sostenidos usan tablas periódicas internas como instrumento, no como pista de acompañamiento. La variedad y la comodidad requieren escuchar durante varios minutos; pruebas automáticas no establecen calidad artística.

## Cambio de juego sin cortes

1. El foco, la portada y el texto responden inmediatamente.
2. La selección debe mantenerse 600 ms antes de solicitar un perfil musical.
3. El perfil entra en la siguiente frontera de cuatro pulsos: espera máxima aproximada de 4,3 segundos, más estabilización.
4. Las voces ya emitidas completan su envolvente; nunca se detienen por una selección nueva.
5. Las voces nuevas entran con ataques suaves y las anteriores se desvanecen. El filtro cambia gradualmente durante 2,4 segundos.
6. El reloj, la sala y el eco permanecen compartidos. Solo existe un destino pendiente; recorrer muchos juegos no encola muchas músicas.

Se conserva el pulso y la compatibilidad armónica; no se pretende cambiar tonalidad ni BPM entre juegos en este corte. Los cruces son por capas musicales, no por reproducción simultánea de dos pistas completas.

## Síntesis y recursos

`src/features/feedback/music/` contiene análisis textual, planificación musical, síntesis y transporte. Instrumentos originales generados una sola vez: voz suave sostenida, campana, pulsación, viento, cristal, bajo, flauta, mallet y chip retro. Sala estéreo con impulso sintetizado de 2,6 s, más eco de 370 ms con feedback limitado. Muestras y grafo se reutilizan entre selecciones.

El viento usa ruido estéreo filtrado, sin retumbo grave, con entrada y salida lentas: una ráfaga variable por frase activa, ligeramente más presente en perfiles espaciales y acuáticos. Durante el descanso no entra viento nuevo. Los cristales son resonancias de vidrio con fundamental afinada al acorde; dos destellos separados por medio pulso responden una vez por frase, más suaves durante el descanso. Ambas texturas comparten volumen, silencio, atenuación, sala y límite de voces con la música. El viento no dispara reflejos visuales; los cristales producen reflejos discretos.

Programación con reloj de Web Audio, anticipación de 180 ms y revisión cada 50 ms. Si el hilo principal se retrasa, se saltan pulsos vencidos en vez de emitirlos juntos. Límite de veinticuatro voces musicales; al alcanzarlo se omiten nuevas voces y se conservan las colas existentes. Los dos efectos de interfaz conservan su límite y prioridades independientes.

El ambiente mantiene volumen, silencio y atenuación bajo confirmaciones. Al perder foco se cancela el programador y se corta la salida suavemente; al volver se recupera posición musical y armonía mediante fundido. Una plataforma sin síntesis compatible deja disponibles los efectos de interfaz.

No se descarga una pista ambiental ni se requiere API musical. El generador Python de la versión anterior queda como referencia histórica; su WAV ya no forma parte de los recursos distribuidos.

## Iluminación

La respiración sigue el transporte musical. Los reflejos responden a las notas de campana y pulsación efectivamente programadas y a sus ecos; la energía se mide exclusivamente en el bus de música. Las notas omitidas no producen reflejos. Sin audio, permanece una respiración visual suave sin campanas inventadas. Movimiento reducido detiene la animación y conserva el sonido.

Las portadas de fondo también reaccionan al audio: brillo y saturación siguen la energía y los pulsos, las voces sostenidas desplazan el matiz hasta veintiocho grados en cada dirección, los cristales iluminan una zona según su posición estéreo y el viento introduce luz difusa durante su envolvente. Se usan únicamente voces emitidas, sin anticipar entradas futuras ni reflejar notas descartadas por el límite del conjunto. Las señales suaves usan una curva perceptual para mantener visibles los efectos tras la sombra del menú. Los reflejos combinan un halo amplio y una franja de luz. Los parámetros visuales se suavizan a 20 actualizaciones por segundo; se aplican al arte bajo la capa de sombra para mantener legible el menú. Al detener la música vuelven a valores neutros. Desactivar iluminación dinámica, perder foco o activar movimiento reducido cancela este programador y restaura el arte original.

## Validación

Pruebas de perfiles textuales, motivos diferenciados, variación de frases, descanso, registros, conducción de voces, sustitución de destinos pendientes, continuidad del transporte, notas futuras, límite de voces, suspensión, controles independientes y muestras finitas con finales suaves.

Pendiente escuchar transiciones entre los seis juegos, navegar rápidamente durante varios minutos y medir CPU en el navegador/host. Este entorno bloquea el puerto de Vite; no se afirma calibración auditiva ni rendimiento nativo verificado.

## Metadatos y diversidad de composición

El catálogo de demostración añade `releaseYear`, `platform`, `emulated` y `tags` opcionales. Son datos ficticios de estos títulos de ejemplo. Los importadores futuros pueden aportar estos campos sin modificar el motor. Su ausencia mantiene un perfil determinista. Las etiquetas se normalizan y ordenan para que su orden no cambie la música; la tienda y el arte no influyen en la identidad sonora.

El título tiene un flujo aleatorio independiente que escoge variante armónica y familia rítmica. El texto completo, las etiquetas, el año, la plataforma y la emulación forman la semilla del motivo y las variaciones. El género guía el carácter; exploración y océano usan modo dórico, el perfil reflexivo menor y los demás mayor. Emulación o un año anterior a 1996 eligen chip retro, un filtro más abierto y acompañamiento a doble subdivisión. Plataforma y año exactos también modifican las decisiones por semilla; no se simula el hardware de cada consola.

Las frases activas añaden un bajo de raíz y una voz interior arpegiada cada dos o cuatro pulsos, con mayor presencia durante el desarrollo y silencio durante el descanso. Aventura orgánica utiliza flauta y mallet; carreras, pulsación; el ejemplo emulado, chip. Los nuevos sonidos son síntesis local con envolventes suaves. El límite de 24 voces deja espacio para las nuevas capas y conserva las colas durante transiciones. El transporte sigue a 56 BPM.

Validación de esta revisión: pruebas automáticas de diversidad estructural por título y metadatos, compatibilidad sin metadatos, orden de etiquetas, registros y transporte, además de compilación de producción. Sigue pendiente la escucha para ajustar balance y calidad artística; las pruebas no equivalen a una evaluación auditiva.

## Voces incidentales

Además del viento sostenido, entran ráfagas de aire con filtro móvil, oleaje suave en perfiles acuáticos y susurros de hojas en perfiles orgánicos. Cada frase activa puede emitir como máximo una de estas capas, con probabilidad, entrada y posición estéreo derivadas del juego y la frase. No entran durante el descanso. Son ruido filtrado sintetizado localmente, con modulación y envolventes suaves, sin archivos externos. Comparten volumen, silencio, límites de voces e iluminación difusa del viento; no generan pulsos melódicos visuales.
