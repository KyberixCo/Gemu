# Perfiles de jugadores

Gemu distingue el perfil local de jugador de la cuenta de una fuente como Steam. Cada perfil local tiene su biblioteca publicada, asociaciones a fuentes, logros, favoritos, ajustes de sonido y movimiento, y configuración de gestión. Las instalaciones físicas y la caché de imágenes se pueden compartir; los datos del jugador no.

## Arranque

- Sin perfiles: asistente de dos pasos, nombre y revisión. Crear el perfil lo guarda antes de entrar a la consola.
- Un perfil: selección automática y animación del logo, sin pulsar Entrar.
- Varios perfiles: selector de jugador antes de la animación.
- Gestión → Jugadores y el botón del perfil en la cabecera permiten crear otro perfil o cambiar de jugador. Al cambiar se desmonta la sesión anterior, se descartan sus asistentes y se carga el espacio del jugador elegido.

Importar Steam desde el perfil activo permite elegir una cuenta de Steam. Esa asociación y sus logros se guardan exclusivamente en la biblioteca de ese perfil. Una importación pendiente de otra sesión no puede publicarse después del cambio de jugador. Gemu no cambia la cuenta iniciada en el cliente Steam.

## Persistencia

Se mantiene el almacenamiento JSON atómico existente; esta entrega no incorpora SQLite. El directorio de datos contiene `profiles.json`, `profiles/<id>/library.json` y la caché compartida `art/`. Las preferencias del frontend usan claves `gemu.profile.<id>.*` de localStorage. Los IDs son aleatorios y el backend solo acepta IDs presentes en el registro.

La biblioteca global anterior y sus ajustes se conservan, pero no se asignan automáticamente a un jugador porque no conocemos su propietario. Es necesario importar Steam en cada perfil nuevo.

## Probar desde cero

`dev.ps1` usa `artifacts/development-data`, separado de los datos normales de la aplicación. Cierra Gemu y ejecuta:

```powershell
.\dev.ps1 -ResetData
```

El reinicio mueve los datos de desarrollo a una copia con fecha dentro de `artifacts` y crea un almacén vacío. Los perfiles nuevos tienen IDs nuevos, por lo que no reutilizan preferencias antiguas del WebView. No se borra Steam ni el almacenamiento de producción. La variable `GEMU_DATA_DIR` permite seleccionar un almacén en ejecuciones del binario; `dev.ps1` fija el almacén de desarrollo.

Comprueba: primer arranque y creación; reinicio con un solo perfil; creación de un segundo perfil y selector al reiniciar; favoritos y ajustes distintos; importaciones Steam con cuentas distintas; cambio de perfil mientras una importación está pendiente.

Pruebas automatizadas cubren persistencia, nombres inválidos y duplicados, aislamiento de bibliotecas y cuentas, rechazo de tokens de otra sesión, reglas de arranque y separación de claves. La prueba con mando físico y WebView2 queda pendiente.
