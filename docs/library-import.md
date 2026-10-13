# Importación de bibliotecas

El proceso común a las fuentes vive en `internal/library`. Steam es el primer adaptador; el asistente `SourceWizard` y el catálogo principal consumen el contrato común, sin consultar APIs de Steam directamente.

ScreenScraper y SteamGridDB enriquecen este proceso según la activación y prioridad configuradas. `ContentProvider` define consulta, comprobación de acceso y hosts de arte; `Request.contentProviders` contiene opciones y credenciales temporales. `Metadata.artworks` conserva la galería local y procedencia por imagen. El escenario usa únicamente sus fondos y mantiene compatibilidad con las bibliotecas anteriores. Uso, calidad y límites en [Integraciones de contenido](content-providers.md).

## Flujo

1. **Juegos:** detección automática o carpeta manual. El adaptador devuelve instalaciones, diagnósticos y perfiles locales disponibles.
2. **Perfil:** elegir el jugador que se asociará a los títulos. La clave API es opcional y solo se usa durante la consulta.
3. **Metadatos, arte y logros:** preparar una importación con descripción, géneros, portada y fondo. Cuando hay clave y perfil, consultar primero la API; el perfil local y los datos anteriormente guardados sirven de respaldo. Un logro desconocido no se marca como bloqueado.
4. **Biblioteca:** revisar y finalizar. El backend publica la importación y la pantalla principal recibe el nuevo catálogo. Cancelar descarta el borrador; finalizar es la única operación que modifica la biblioteca publicada.

La pantalla principal ofrece un cambio explícito entre **Mis juegos** y **Demo**. Las fichas reales muestran instalación, perfil y estado de sincronización de logros. La sección Metadatos usa el catálogo seleccionado. El lanzamiento todavía está pendiente.

## Contratos y persistencia

`Source` define `Name`, `Discover`, `Metadata`, `Player` y `ArtworkHosts`. Para añadir una fuente se implementa este contrato y se registra en el mapa del servicio en `main.go`; la lista de adaptadores se expone mediante `GetSources`. Los adaptadores normalizan identidades estables y mantienen las instalaciones separadas del catálogo.

Wails expone `PreviewSource`, `PrepareLibraryImport`, `CommitLibraryImport`, `DiscardLibraryImport` y `GetLibrary`. Los borradores tienen tokens de un solo uso, expiran tras 15 minutos y no contienen credenciales. Reimportar reemplaza los títulos de esa fuente sin borrar otras fuentes. Los logros pertenecen al perfil seleccionado; la caché de otro perfil no se reutiliza. El [gestor seguro](credentials.md) permite conservar claves cifradas por perfil; Go las resuelve al iniciar la consulta, dando prioridad a los valores temporales no vacíos.

El catálogo se guarda bajo `%APPDATA%/Gemu/library.json` en Windows, mediante archivo temporal, sincronización y reemplazo. Este primer almacenamiento usa JSON versionado; SQLite sigue pendiente. Las imágenes se guardan en la carpeta `art` y se sirven mediante `/library-assets/`, también durante desarrollo. Se validan HTTPS, hosts declarados por el adaptador, redirecciones, tamaño y formato de imagen. Los fallos de red conservan los metadatos y logros anteriormente publicados cuando corresponde.

## Adaptador Steam

Lee manifiestos y `config/loginusers.vdf`. Solo extrae SteamID y nombre visible: no usa contraseñas, cookies ni tokens del cliente. La tienda aporta metadatos y arte; su endpoint `appdetails` es una dependencia sin contrato público documentado y puede cambiar. Los fallos quedan como diagnósticos y permiten una importación parcial.

Con clave, consulta `GetPlayerSummaries`, `GetPlayerAchievements` y `GetSchemaForGame`. Referencias oficiales: [perfil](https://partner.steamgames.com/doc/webapi/ISteamUser) y [logros](https://partner.steamgames.com/doc/webapi/ISteamUserStats). Si falla la API, se conserva el perfil local o el último progreso sincronizado de ese mismo perfil. Los archivos locales leídos no ofrecen un registro fiable de logros: sin datos previos quedan pendientes. La clave no se guarda en JSON, localStorage ni registros.

## Verificación

En esta máquina se prepararon e importaron a una biblioteca de prueba BeamNG.drive, City Car Driving y Crysis Remastered, con metadatos, seis imágenes y perfil local, sin diagnósticos. La sincronización autenticada de logros se probó con respuestas HTTP simuladas; queda pendiente probarla con una clave real y revisar visualmente el flujo en WebView2.

```powershell
# Preparar sin publicar, usando una caché de prueba
go run ./cmd/library -data-dir artifacts/library
# Publicar solo en esa biblioteca de prueba
go run ./cmd/library -data-dir artifacts/library -commit
```

La CLI usa `GEMU_STEAM_API_KEY` si está definida y permite indicar `-profile`, `-root` y `-source`. Su salida resume estados sin imprimir la clave ni el identificador del perfil. `artifacts/library` se excluye de Git.
