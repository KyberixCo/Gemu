# Búsqueda de biblioteca con FTS5

La búsqueda de juegos publicados se ejecuta en SQLite desde el campo Buscar de la biblioteca. Consulta título, descripción y género, ignora mayúsculas y acentos y ordena por relevancia, dando más peso al título y después al género. La demo conserva su filtro local de título y género.

## Uso

- `explorac` encuentra palabras que comienzan por ese prefijo, como Exploración.
- `oceano profundo` exige ambos términos y permite que aparezcan separados o en campos distintos.
- `"océano profundo"` busca esa frase completa, con sus palabras consecutivas en un mismo campo.
- Varias frases y palabras se combinan exigiendo todas las condiciones. Las comillas sin cerrar incluyen el resto como frase.
- La puntuación separa palabras. `OR`, `NOT`, nombres de columnas y otros operadores se tratan como texto; no se permite ejecutar sintaxis FTS arbitraria.

Los filtros de fuente y favoritos se aplican en el host antes de contar y paginar. La interfaz espera 180 ms de escritura estable y carga 48 resultados por página mediante Cargar más juegos. Muestra progreso y errores con reintento. Una consulta vacía vuelve al catálogo sin filtrar; una consulta de solo puntuación no produce resultados.

El cambio de consulta, fuente, favoritos, catálogo o jugador inicia una búsqueda nueva. Las respuestas tardías se descartan. Los errores al cargar otra página conservan los resultados anteriores y reintentan esa página.

## Esquema y actualización

La migración 2 comprueba soporte real de FTS5 creando una tabla virtual temporal, insertando un texto y consultándolo mediante MATCH. Después crea `library_documents` y `library_documents_fts`, con contenido externo, `unicode61 remove_diacritics 2` e índices de prefijos. Copia los juegos existentes y reconstruye el índice. Datos, triggers, índice y versión se confirman en la misma transacción. Un fallo conserva el esquema anterior para poder reintentar.

Cada documento contiene propietario, ID estable del juego, fuente, tipo `game`, título, descripción y género. La columna `notes` queda reservada y vacía: no se incorporan aún notas de fuentes o emuladores ni los borradores de metadatos que no estén publicados en el catálogo.

Los triggers de juegos mantienen los documentos en altas y modificaciones; las bajas eliminan documentos mediante clave foránea. Los triggers de documentos mantienen FTS5 en inserciones, actualizaciones y borrados. La importación, reimportación y el guardado de arte usan esas mismas transacciones. El índice es derivado del catálogo y puede reconstruirse con el comando FTS5 `rebuild`; la migración lo ejecuta y las pruebas verifican su integridad.

## Contrato

`App.SearchLibrary(profileID, request)` valida un propietario explícito y consulta únicamente su ámbito, aunque haya cambiado el jugador activo. `request` contiene `query`, `sourceId`, `favoritesOnly`, `limit` y `offset`. Devuelve `hits`, `total` y `hasMore`; cada coincidencia incluye `gameId`, `title`, `sourceId`, `kind` y `score` para conservar identidad y procedencia.

Las consultas usan parámetros SQL. Un constructor convierte palabras Unicode y frases a cláusulas literales y aplica prefijos únicamente a palabras sueltas. Límites: 2048 bytes de texto, 32 cláusulas, entre 1 y 100 resultados por página (48 por defecto) y desplazamiento entre 0 y 100000. El ranking usa `bm25` con desempate por título e ID. El cliente no interpreta ni inserta HTML procedente del índice.

## Validación

Las pruebas cubren migración desde esquema 1 con juegos existentes, fallo y reintento sin esquema parcial, acentos, frases, prefijos, relevancia, actualizaciones, borrado, reconstrucción, integridad del índice, rollback de una publicación fallida, aislamiento por jugador, filtros previos a paginación y consultas literales. El frontend comprueba debounce, respuestas tardías, cambios de jugador, paginación y recuperación de errores.

Pendiente recorrido visual en WebView2 y con mando físico. No se afirma una medición de rendimiento con bibliotecas grandes. Referencia técnica: [SQLite FTS5](https://www.sqlite.org/fts5.html).
