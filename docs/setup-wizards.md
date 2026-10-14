# Configuración guiada

La configuración de proveedores usa asistentes reutilizables. El mismo `ProviderSetupWizard` aparece en Gestión → Información y arte, Gestión → Credenciales y el asistente de primera configuración. `WizardFrame` mantiene cabecera, progreso, contenido desplazable y acciones de atrás/continuar/finalizar. Los pasos y requisitos se definen en `setup-providers.ts` sin depender del lugar desde donde se abre el asistente.

## Primera configuración

Antes de este recorrido, el asistente de perfiles permite crear el perfil inicial desde una cuenta Steam local. El nombre se sugiere de forma editable y se usa el avatar en caché si está disponible; la creación manual sigue disponible cuando Steam no está instalado o no se detectan cuentas.

Después de crear el perfil, `FirstRunSetup` guía por cuatro etapas:

1. **Tu consola:** explica qué se puede configurar y permite posponerlo.
2. **Arte y metadatos:** abre los asistentes compartidos de SteamGridDB y ScreenScraper.
3. **Tus juegos:** reutiliza `SourceWizard`, con detección, cuenta, enriquecimiento y publicación de la biblioteca.
4. **Revisión:** muestra biblioteca/proveedores y guarda las preferencias antes de entrar en Gemu.

El flujo admite teclado y mando. B/Escape vuelve dentro del asistente activo antes de abandonar el maestro; durante la publicación de la biblioteca no se permite salir. Los perfiles que ya tengan una biblioteca publicada no reciben el recorrido automáticamente. Gestión → Configuración guiada permite volver a ejecutarlo. La marca `gemu.setup.v1` es propia de cada perfil; no guarda secretos. Si no se puede persistir la configuración, el usuario puede entrar sin guardarla.

## Proveedores

Los proveedores recorren introducción, credenciales, comprobación cuando existe conexión real, preferencias cuando corresponde y revisión. Las credenciales se administran con el gestor seguro y pueden guardarse por etapas. Los valores nuevos se guardan explícitamente al finalizar; cancelar descarta el recorrido, conservando únicamente los campos guardados previamente por el usuario.

En Información y arte, finalizar guarda las preferencias del proveedor sin publicar otros borradores de Gestión. La activación de ScreenScraper y SteamGridDB exige una comprobación correcta en ese recorrido. Guardar credenciales incompletas o continuar sin comprobar conserva el proveedor desactivado. IGDB y las conexiones aún no implementadas permanecen pendientes; no se presenta una comprobación ficticia.

## ScreenScraper

Su recorrido distingue cinco etapas:

1. **Introducción:** explica el propósito y los requisitos.
2. **Tu cuenta:** usuario y contraseña personales, compartidos por metadatos y arte.
3. **Comprobación:** valida la cuenta o acceso anónimo y también la consulta del catálogo/sistema PC Windows.
4. **Preferencias:** activación y prioridad.
5. **Revisión:** activa únicamente si la comprobación fue correcta; de lo contrario guarda para retomar después.

La [documentación oficial de ScreenScraper](https://www.screenscraper.fr/webapi2.php) indica que el desarrollador debe presentar su aplicación en el foro para obtener credenciales de API. Una cuenta personal no sustituye ese registro. El [proceso de build](screenscraper-build.md) permite incorporar las credenciales asignadas a Gemu desde variables de entorno o secretos de CI. El usuario final solo configura su cuenta personal; si el build no incluye acceso, la comprobación deja la conexión pendiente. No se reutilizan identidades de otras aplicaciones.

El backend distingue HTTP 403 (credenciales de desarrollador), 401 (acceso restringido, cuenta o saturación), 423 (servicio cerrado), 426 (aplicación bloqueada), 429 (límite o saturación) y 430/431 (cuota diaria). Muestra instrucciones según el caso sin exponer el cuerpo de respuesta ni secretos. Una prueba correcta confirma acceso al catálogo; no garantiza que exista una coincidencia o un fondo de alta resolución para todos los títulos.

## Verificación y límites

Se prueban los pasos, requisitos de credenciales guardadas/temporales, cuentas personales incompletas, activación solo tras comprobar, preservación de la comprobación ante rerenders, política de primera configuración y errores HTTP redactados. Las consultas autenticadas con credenciales reales y el recorrido visual en WebView2/mando siguen pendientes. La creación del perfil y la importación de fuentes conservan sus asistentes existentes; emuladores y mapeo de controles mantienen sus pantallas de tarea actuales.
