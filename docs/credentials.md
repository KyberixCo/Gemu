# Gestor seguro de credenciales

Gestión → Credenciales permite guardar, reemplazar y eliminar credenciales de ScreenScraper, SteamGridDB, Steam, IGDB, MiniMax, OpenRouter y API compatible. Los mismos controles aparecen en la configuración de cada proveedor, la importación Steam y la conexión del asistente.

## Uso

- Introducir un valor crea un borrador en memoria. Se puede probar o usar en la consulta actual sin guardarlo.
- Guardar credenciales persiste únicamente los campos nuevos no vacíos. Un campo vacío conserva lo guardado; no muestra ni recupera la contraseña anterior en la interfaz.
- Eliminar guardadas elimina las credenciales de ese proveedor en el perfil activo y descarta sus valores nuevos en ese editor.
- Descartar valores nuevos limpia el borrador; conserva lo guardado.
- Probar acceso y la importación resuelven las credenciales guardadas en Go. Un valor nuevo no vacío tiene prioridad durante esa consulta. No se persiste automáticamente.

Los campos muestran «Guardada de forma segura», «Valor nuevo en sesión» o «Sin credencial». El editor informa el nombre y disponibilidad del almacén seleccionado. Guardar no activa un proveedor ni verifica sus credenciales. IGDB y las APIs del asistente pueden almacenar sus claves, pero sus consultas siguen pendientes de implementación.

## Contrato independiente del sistema operativo

La aplicación depende de `credentials.Vault`, no del almacén de Windows:

```go
type Vault interface {
    Get(provider string) (map[string]string, error)
    Status(provider string) (Status, error)
    Save(Request) error
    Delete(provider string) error
    Capabilities() Capabilities
}
```

`Get` entrega secretos únicamente a los consumidores de Go. `Status` expone los nombres de campos guardados y `Capabilities` el nombre/disponibilidad del backend; estas son las únicas consultas de estado que llegan a la UI. Se mantienen el reemplazo parcial, separación por perfil, escrituras serializadas y ausencia de almacenamiento en texto plano como requisitos del contrato.

`credentials.New(profileDir)` selecciona el adaptador mediante build tags. `backend_windows.go` devuelve el almacén DPAPI privado; `backend_other.go` devuelve un adaptador explícitamente no disponible. Este último entrega cero credenciales guardadas para permitir valores temporales y devuelve `ErrUnavailable` al guardar o eliminar. La UI deshabilita la persistencia cuando `available` es falso.

Para añadir un almacén nativo de macOS o Linux se implementa `Vault`, se añade su fábrica con el build tag correspondiente y se excluye ese sistema de `backend_other.go`. No hace falta modificar App, las consultas de proveedores o los métodos Wails. Las implementaciones de esos almacenes nativos quedan pendientes; se ha comprobado la compilación del paquete genérico para `linux` y `darwin`, sin afirmar soporte completo de Gemu en esas plataformas.

## Protección

El adaptador actual de Windows usa [DPAPI](https://learn.microsoft.com/en-us/windows/win32/api/dpapi/nf-dpapi-cryptprotectdata) en el contexto del usuario, sin `CRYPTPROTECT_LOCAL_MACHINE`, y sin prompts. Windows administra las claves de cifrado y la protección de integridad. Gemu no incorpora una clave de cifrado fija ni usa localStorage para persistir secretos.

Cada perfil tiene un archivo binario versionado `credentials.bin` dentro de su directorio de datos. El contexto adicional de DPAPI deriva del directorio absoluto del perfil; desarrollo y producción quedan separados. El archivo se escribe mediante temporal, sincronización y reemplazo, siempre después de cifrar. No hay alternativa en texto plano: si Windows no puede proteger o descifrar, la operación falla y conserva el archivo existente. Fuera de Windows este backend no ofrece persistencia de secretos.

El frontend recibe solo el proveedor y los nombres de campos guardados. No existe un método Wails para leer los valores almacenados. Las consultas capturan biblioteca y almacén bajo el mismo bloqueo para evitar mezclar perfiles si el jugador cambia durante una operación. Los borradores de biblioteca, el catálogo, los diagnósticos y las preferencias no contienen los secretos guardados.

Los buffers de serialización y los devueltos por DPAPI se limpian después de usarlos. Los valores descifrados necesitan existir brevemente en memoria de Go mientras se prepara una consulta; los valores nuevos existen en memoria de React hasta guardarlos, descartarlos o salir de su pantalla. No se garantiza eliminación física de todas las copias administradas por los runtimes.

DPAPI protege datos en reposo según la cuenta de Windows. Los perfiles de Gemu son separación lógica de bibliotecas, no cuentas de seguridad del sistema operativo. Un proceso con acceso a esa misma cuenta de Windows puede tener capacidad de descifrar; no se presenta este almacén como aislamiento frente a procesos comprometidos dentro de la cuenta.

## Copias y recuperación

La copia de respaldo debe conservar el archivo cifrado y restaurarse en el directorio original del perfil, dentro del contexto de Windows que lo protegió. Mover el directorio, cambiar de cuenta Windows o perder sus claves DPAPI puede impedir descifrar; hay que volver a configurar las credenciales. Gemu no exporta secretos en texto plano ni reemplaza automáticamente un almacén corrupto.

## Validación

Pruebas con credenciales ficticias y DPAPI real: cifrado sin texto plano en disco, recuperación tras crear una nueva instancia del almacén, reemplazo parcial, eliminación, actualizaciones concurrentes, separación de perfiles, corrupción y preservación del archivo ante un fallo. Pruebas de App verifican resolución automática de Steam y proveedores, prioridad de valores de sesión, ausencia de secretos en el estado devuelto a la UI y que probar acceso no guarda los valores temporales.

Las pruebas DPAPI requieren el contexto normal del usuario Windows; el token del entorno aislado de Codex no puede utilizar esa protección. El flujo visual con credenciales reales sigue pendiente de revisión en WebView2. Después de actualizar los métodos nativos hay que reiniciar `dev.ps1` o abrir el ejecutable actualizado; una recarga Vite no cambia el host.
