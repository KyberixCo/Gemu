# Menú durante la partida: compatibilidad y alcance

El menú de partida usa la ventana externa de Gemu sin bordes. El gestor de sesión coordina continuar, volver a la biblioteca y solicitar el cierre del juego.

Para alternar sin cambios de modo de pantalla, el juego debe ejecutarse también sin bordes a la resolución del escritorio. Gemu no cambia la resolución global ni reescribe los ajustes de cada juego. Un juego en pantalla completa exclusiva puede perder ese modo al mostrar otra ventana; igualar el tamaño de las ventanas no garantiza evitarlo.

## Estado actual

| Entorno | Menú de partida |
| --- | --- |
| Windows, juegos Steam | Ventana externa de Gemu; pendiente completar validación física de foco, atajos y modos de pantalla |
| Linux, incluido Wine/Proton | Host de lanzamiento, atajos y foco todavía sin implementar |
| macOS | Host de lanzamiento, atajos y foco todavía sin implementar |
| RetroArch en Windows x64 | Lanzamiento directo y menú externo implementados; pendiente validación con emulación y mando reales |
| Otros emuladores externos | Lanzamiento todavía sin implementar |

La integración futura con emuladores debe utilizar sus interfaces oficiales cuando estén disponibles. Abrir el menú propio de un emulador debe identificarse como tal en la interfaz.

## Validación pendiente

Probar con juegos reales el lanzamiento, la apertura del menú, continuar, volver a la biblioteca y solicitar el cierre. Comprobar recuperación del foco y de la selección, resolución y frecuencia de pantalla, modo exclusivo, juegos elevados, desconexión del mando y cierre de Gemu.

Medir el consumo de recursos durante la partida y con el menú abierto. Las pruebas automatizadas del gestor de sesiones no sustituyen la validación con Windows, WebView2 y hardware real.

El funcionamiento y el recorrido de prueba se detallan en [Lanzamiento y menú de partida](game-sessions.md).
