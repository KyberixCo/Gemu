# ScreenScraper y SteamGridDB

Las integraciones amplían los metadatos y el arte durante el asistente de importación. No son fuentes de instalaciones: Steam sigue proporcionando los juegos, el perfil y los logros. Las consultas pasan por Go; el frontend no envía credenciales directamente a las APIs.

## Uso

La portada seleccionada también se usa en la rotación del escenario si su tamaño real es al menos 1280 × 720 y es horizontal. Se combina con los fondos elegidos sin duplicar URLs; los logotipos, iconos y portadas pequeñas quedan fuera. Las portadas locales antiguas se miden al cargar la biblioteca. Las transiciones conservan las preferencias de movimiento y movimiento reducido.

**Para elegir arte:** Gestión → Metadatos → juego → Arte → Buscar imágenes. Elige Portada/Fondo/Logotipo, abre una imagen y confirma Usar imagen; después Revisar cambios → Guardar arte. Para varios juegos, usa Completar arte faltante desde Metadatos.

**Alternativa durante la importación:** con los proveedores activados, ve a Gestión → Juegos → Steam → Actualizar Steam, elige la cuenta, pulsa Obtener metadatos y continuar y confirma Finalizar e incorporar juegos. El arte queda visible en Gestión → Metadatos → juego → Arte → Galería. Configurar o comprobar una conexión por sí solo no actualiza juegos existentes.

El recorrido dedicado de búsqueda, elección y publicación con mando está diseñado en [Obtener y elegir arte con mando](artwork-flow.md). El flujo individual y completar faltantes ya están implementados.

La configuración ahora se realiza mediante [asistentes reutilizables](setup-wizards.md), disponibles tanto en Gestión como en la primera configuración. ScreenScraper solicita únicamente la cuenta personal. El acceso de aplicación se configura durante el [build de Gemu](screenscraper-build.md); si falta, la conexión permanece pendiente.

1. En Gestión → Información y arte → Arte, elige SteamGridDB o ScreenScraper para abrir su asistente.
2. Lee la introducción y avanza por los pasos de acceso y credenciales.
3. SteamGridDB requiere una clave API. En ScreenScraper, usuario y contraseña de la cuenta personal son opcionales, pero deben introducirse juntos. Las credenciales de aplicación las proporciona el build. Puedes guardar las credenciales por etapas mediante el gestor seguro.
4. En Comprobación, pulsa Probar acceso. Una prueba correcta confirma el acceso al catálogo, no que exista arte para todos los juegos. Sin comprobación, el proveedor queda pendiente y desactivado.
5. Elige activación/prioridad y finaliza la revisión. Las preferencias se guardan al terminar. Vuelve a Juegos → Steam → Actualizar Steam y completa el asistente, o continúa con la biblioteca desde el maestro de primera configuración.
6. Revisa las incidencias y finaliza para publicar. El arte guardado se puede inspeccionar en Metadatos → juego → Arte → Galería, con resolución y procedencia.

Las credenciales nuevas permanecen en memoria hasta guardarlas explícitamente. El [gestor seguro](credentials.md) las cifra con DPAPI en un almacén por perfil; no se guardan en localStorage, biblioteca, borradores ni registros. Las consultas reutilizan lo guardado desde Go y los campos de la interfaz muestran solo su estado. Las preferencias actuales del borrador se aplican a la importación iniciada desde esa misma sesión.

ScreenScraper comparte las credenciales entre los menús de Metadatos y Arte dentro de la misma sesión de Gestión. El usuario y contraseña de la cuenta ScreenScraper no sustituyen el ID y contraseña de desarrollador de la aplicación: son dos pares distintos exigidos por el contrato de la API.

Si aparece `CheckContentProvider is not a function`, la ventana usa un host anterior a esta integración. La recarga de Vite solo actualiza la interfaz. Cierra Gemu y reinicia `dev.ps1` o abre el ejecutable de producción actualizado; los valores nuevos que no se hayan guardado deben introducirse otra vez. El frontend comprueba esta capacidad tanto al probar acceso como al importar con proveedores habilitados, para evitar consultas fallidas o importaciones que ignoren silenciosamente la configuración nueva.

## Selección y calidad

SteamGridDB usa el AppID de Steam para obtener la identidad del juego. Para instalaciones sin identidad Steam, el adaptador admite únicamente una coincidencia exacta y única de título. Obtiene imágenes estáticas PNG/JPEG: heroes para fondos, grids horizontales para portadas, logos e iconos. Los candidatos se ordenan por puntuación y resolución.

ScreenScraper consulta la lista de sistemas para identificar PC Windows por sus nombres regionales en `noms` (por ejemplo, `noms.nom_eu`) y busca por título dentro del ID devuelto por el catálogo. Conserva compatibilidad con el campo simple `nom`; no fija el ID ni confunde DOS o las versiones antiguas de Windows con PC Windows. Solo acepta una coincidencia exacta y única, considerando los nombres regionales. Prefiere español, inglés y francés para descripción y géneros. Usa fanart como fondo, box-2D y steamgrid como portada, wheel-hd y wheel como logotipo. Las capturas de baja resolución y los vídeos no se convierten en fondos. El soporte de sistemas retro y búsqueda por hashes de ROM queda pendiente del adaptador de emulación; este incremento cubre la biblioteca PC disponible.

Cada proveedor aporta hasta tres fondos, una portada, un logo y un icono por juego, con un máximo de diez intentos por tipo. Se respeta la prioridad para escoger portada y fondo principales. Solo los fondos forman parte de la galería del escenario. Los logotipos e iconos se conservan y se muestran en la revisión de arte; su uso en la composición principal queda pendiente.

La descarga aplica HTTPS, hosts permitidos, redirecciones restringidas y un máximo de 8 MiB. Los fondos deben tener al menos 1280 × 720 píxeles y anchura mayor que altura. El backend verifica las dimensiones del archivo PNG/JPEG descargado, además de las declaradas por la API; el frontend conserva su comprobación al mostrar y rotar el fondo. No se aceptan fondos panorámicos de menos de 720 píxeles de altura aunque tengan una anchura elevada. WebP y arte animado de estos proveedores quedan fuera de este incremento.

## Fallos y persistencia

Un proveedor sin acceso, sin coincidencias o sin imágenes válidas no impide importar Steam. Los errores aparecen como incidencias. Autenticación, bloqueo y cuota detienen ese proveedor durante el resto de la importación. No se realizan reintentos automáticos que consuman más cuota. Las peticiones son secuenciales y tienen timeout; el asistente conserva su descarte de respuestas tardías y borradores cancelados.

La biblioteca guarda URLs locales de caché, ID, tipo, proveedor, dimensiones, autor cuando existe y página pública de procedencia. Las URLs de descarga de ScreenScraper con autenticación no se publican. Si no se obtiene nuevo arte externo de un tipo, se conservan las imágenes locales anteriores de ese tipo; una actualización de iconos no descarta los fondos. La fuente Steam continúa como respaldo; sin fondo válido se muestra el escenario procedural.

Los borradores siguen publicándose únicamente al finalizar. La galería es de inspección: la elección manual de candidatos individuales y la aplicación de rutas propias al catálogo son trabajo posterior.

## Referencias y validación

Contratos consultados: [SteamGridDB API v2](https://www.steamgriddb.com/api/v2), [cliente oficial de SteamGridDB](https://github.com/SteamGridDB/node-steamgriddb), [ScreenScraper API v2](https://www.screenscraper.fr/webapi2.php).

Las pruebas cubren autenticación, identidad Steam, separación de tipos de arte, coincidencias ambiguas, sistema PC, localización, cuotas, errores sin credenciales, prioridades, comprobación de dimensiones reales, publicación y recuperación de caché. No se han realizado consultas autenticadas a cuentas reales: se necesitan las credenciales del usuario y la validación en WebView2. No se afirma disponibilidad de arte para los títulos instalados hasta realizar esa consulta.

## Presentación de SteamGridDB

Los grids son las portadas de las tarjetas, preservando su proporción sin recorte; la consulta admite distintos tamaños y orientaciones. Los heroes proporcionan fondos de escenario. El logotipo elegido funciona como título visual con texto de respaldo si falla la imagen. El icono elegido aparece junto al nombre en el catálogo, el escenario y Gestión. Los cuatro tipos son seleccionables en Metadatos → juego → Arte → Buscar imágenes y se guardan con procedencia y dimensiones. Los iconos de galerías antiguas se recuperan al cargar la biblioteca.
