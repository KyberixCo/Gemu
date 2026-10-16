# Gemu · by Kyberix

**Tu PC, tu biblioteca, tu experiencia de consola.**

Español · [English](README.en.md)

Gemu tiene como objetivo transformar un PC con Windows en una consola de videojuegos para el salón: encender el equipo, elegir un juego con el mando y comenzar a jugar desde una única interfaz diseñada para televisores.

Este documento describe la visión del producto y sus objetivos.

## Jugar desde el sofá

La experiencia debe ser cómoda a distancia: textos legibles, portadas grandes, foco evidente y navegación predecible con cruceta o stick. Cada pantalla debe concentrarse en una tarea, con acciones accesibles y el mínimo desplazamiento posible.

Las transiciones, el glow de selección y el sonido deben ayudar a orientarse y confirmar acciones. El usuario debe poder ajustar o desactivar el movimiento y el audio sin perder claridad al navegar.

## Una biblioteca unificada

Gemu busca reunir juegos de Steam, Epic Games, GOG, ejecutables independientes y bibliotecas de emulación en un mismo catálogo. Los conectores deben ser intercambiables y permitir elegir qué fuentes utilizar.

Un juego debe tener una identidad propia, aunque exista en varias tiendas o instalaciones. Favoritos, búsqueda y colecciones deben ayudar a encontrar qué jugar, sin exigir recorrer cada cliente por separado. La biblioteca y los datos guardados localmente deben seguir disponibles sin conexión.

El lanzamiento debe respetar los clientes oficiales y sus mecanismos de autenticación y DRM. Los emuladores, las ROMs y las BIOS se configuran como recursos externos aportados por el usuario.

## Entrar al juego y volver a Gemu

El ciclo completo de una partida es una prioridad: preparar el lanzamiento, ceder el foco al juego y recuperar la interfaz al terminar, conservando la selección anterior.

Gemu debe dejar de interpretar la navegación del mando durante la partida. Los errores de lanzamiento o la desconexión de un dispositivo deben ofrecer una salida clara. El historial y la duración de las sesiones deben basarse en señales verificables de ejecución.

## Gestión pensada para televisión

La configuración de bibliotecas, emuladores y controles debe seguir la misma lógica de consola: grupos pequeños de ajustes, edición de un valor a la vez y acciones de volver y guardar siempre accesibles.

El objetivo incluye perfiles de mando por destino, asignación de botones y calibración. La entrada de texto debe contar con alternativas adecuadas para televisión y permitir usar teclado cuando resulte más cómodo.

## Metadatos, arte y asistencia

Las fichas deben poder enriquecerse mediante proveedores de metadatos y arte configurables, con prioridades, procedencia y gestión de credenciales cuando una API lo requiera.

El asistente debe ayudar a organizar la biblioteca, preparar configuraciones y proponer mejoras. Cada propuesta debe mostrar qué cambia y permitir aceptar campos concretos o descartarla antes de guardar. El usuario conserva el control de sus datos y de los cambios.

## Base técnica y fiabilidad

La arquitectura objetivo combina una aplicación de escritorio con Wails, servicios en Go, una interfaz en React y TypeScript y almacenamiento local con SQLite. El dominio, los conectores y la integración con Windows deben mantenerse separados para facilitar pruebas y ampliar proveedores.

Las prioridades son una navegación rápida, consumo moderado de recursos, recuperación del foco y conservación de los datos. El inicio automático y un posible modo shell deben ser opcionales, reversibles y contar con una vía de recuperación.

## Documentación

- [Visión y arquitectura](idea.md)
- [Diseño de la experiencia](docs/design.md)
- [Movimiento y sonido](docs/feedback.md)
- [Gestión y contratos de integración](docs/management.md)
- [Diseño del flujo de arte con mando](docs/artwork-flow.md)
- [Guía de desarrollo](docs/development.md)
- [Persistencia y migración a SQLite](docs/persistence.md)
- [Búsqueda de biblioteca](docs/search.md)
- [Instalar y configurar RetroArch](docs/retroarch.md)

## Licencia

Gemu se distribuye bajo la **GNU General Public License, versión 3 exclusivamente** (`GPL-3.0-only`). Consulta [LICENSE](LICENSE). Las dependencias y los recursos de terceros conservan sus respectivas licencias.

## Acceso de ScreenScraper en el build

Las credenciales de aplicación se cargan automáticamente desde `.env` o variables de entorno. Copia `.env.example` a `.env` y completa los valores antes de usar `./build.ps1` o `./dev.ps1`; no se necesita un flag. Parámetros y manejo de secretos en [Acceso de aplicación a ScreenScraper](docs/screenscraper-build.md).
