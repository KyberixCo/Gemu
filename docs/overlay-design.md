# Menú durante la partida: compatibilidad y alcance

La ruta normal conserva el menú de Gemu como una ventana externa sin bordes. Para alternar sin cambios de modo de pantalla, el juego debe ejecutarse también sin bordes a la resolución del escritorio. Gemu no cambia la resolución global ni reescribe los ajustes de cada juego.

Un juego en pantalla completa exclusiva puede perder ese modo al mostrar otra ventana. Igualar el tamaño de las ventanas no garantiza evitarlo. Esta limitación se explica en la [documentación de DXGI](https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/d3d10-graphics-programming-guide-dxgi). El menú externo es una solución sencilla, pero no equivale a dibujar dentro del juego.

## Estado real

| Entorno | Menú externo | Overlay dentro del renderizado |
| --- | --- | --- |
| Windows, DX9/DX9Ex, x86 o x64 | Ruta normal de Gemu; exclusiva puede cambiar de modo | Sin implementar; necesita un adaptador propio |
| Windows, DX10/DX10.1 | Ruta normal de Gemu; misma limitación | Sin implementar; no usa el renderer DX11 |
| Windows, DX11 x64 | Ruta normal de Gemu | Prototipo escrito, aún sin validación gráfica; desactivado por defecto |
| Windows, DX11 x86, DX12, OpenGL o Vulkan | Ruta normal de Gemu | Sin implementar |
| Linux, incluido Wine/Proton | Host de lanzamiento/atajos/foco todavía sin implementar | Sin implementar; la DLL Windows no se aplica directamente |
| macOS | Host de lanzamiento/atajos/foco todavía sin implementar | Sin implementar; requiere integración propia |
| Emuladores retro externos | Lanzamiento de emuladores todavía sin implementar | Preferir interfaces del emulador cuando estén disponibles |

La compatibilidad retro depende del emulador, su backend gráfico, la arquitectura del proceso y su modo de pantalla; no de la consola emulada. Un emulador con un menú o canal de comandos oficial puede ofrecer acciones sin interceptar su renderizado. Por ejemplo, [RetroArch expone una interfaz de comandos](https://docs.libretro.com/development/retroarch/network-control-interface/); abrir su menú nativo no debe presentarse como si fuera el menú de Gemu.

## Decisión de diseño

El gestor de sesión coordina continuar, volver a la biblioteca y cerrar. Las capacidades de plataforma, renderizado, entrada e integración con emuladores deben ser independientes. Un adaptador de renderizado se selecciona solo para una combinación comprobada de sistema, arquitectura y API; una biblioteca gráfica cargada en el proceso es un indicio, no una confirmación de compatibilidad.

Cada adaptador debe confirmar que dibuja frames antes de anunciar disponibilidad. Los estados necesarios son no disponible, candidato, inicializando, listo y fallido, con una razón de diagnóstico. Una pérdida del renderer debe cerrar el menú integrado y ofrecer la ruta externa, informando de su limitación en exclusiva. El prototipo actual solo ofrece un indicador de disponibilidad; ese contrato detallado sigue pendiente.

DX9 requiere pruebas específicas de pérdida del dispositivo y Reset, además de Present; véase [Lost Devices](https://learn.microsoft.com/en-us/windows/win32/direct3d9/lost-devices). DX10 necesita sus propios recursos y restauración de estado. Vulkan tiene su [mecanismo de capas](https://docs.vulkan.org/guide/latest/layers.html); no permite reutilizar automáticamente un hook DXGI. Linux también distingue X11, Wayland y Wine/Proton. macOS requiere un backend y una estrategia de distribución propios. Compartir comandos y diseño visual no elimina esas diferencias.

Antes de activar un adaptador por defecto se deben medir resolución, frecuencia, modo exclusivo, foco, estado gráfico, entrada retenida, cambios de tamaño, recuperación del dispositivo, cierre de Gemu y coste cuando el menú está oculto. Las pruebas del coordinador Go no sustituyen estas pruebas gráficas.

## Prototipo DX11 aislado

La compilación normal no descarga Zig/ImGui/MinHook ni compila o empaqueta la DLL. El experimento se construye explícitamente con `scripts/build-overlay.ps1 -DownloadDependencies` o `build.ps1 -ExperimentalOverlay`. La conexión a juegos requiere además `GEMU_EXPERIMENTAL_OVERLAY=dx11` en el entorno del proceso de Gemu. La mera presencia de la DLL no activa la conexión.

`native/overlay` y `cmd/overlay-check` conservan el trabajo experimental y una escena de prueba. La DLL y la prueba gráfica todavía no se consideran compiladas ni verificadas. DX9, DX10 y el soporte multiplataforma no se dan por implementados.
