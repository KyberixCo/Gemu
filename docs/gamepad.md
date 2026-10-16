# Soporte de gamepad

Los indicadores siguen el último método usado, no la conexión de un mando. Una tecla activa teclado; clic, rueda o movimiento significativo activa ratón; una pulsación o navegación del mando activa su familia detectada. Un mando conectado en reposo no sustituye las ayudas de teclado/ratón. Al desconectar el mando activo se vuelve a teclado.

Las ayudas de confirmación y regreso usan Enter/Esc (o los atajos guardados), clic al usar ratón y las posiciones físicas del mando: A/B en Xbox, ×/○ en PlayStation y B/A en Nintendo. Los mandos sin familia conocida muestran símbolos genéricos.

En Gestión → Controles, «Método de entrada» permite elegir Automático, Teclado y ratón o Mando. Automático sigue el método activo; fija Mando para editarlo con el ratón. La vista de teclado permite reasignar la navegación y búsqueda de Gemu por perfil. Una tecla ocupada intercambia las dos asignaciones. Los cambios se guardan inmediatamente; Tab y F10 permanecen como funciones fijas. Los campos de texto conservan su entrada normal. Los clics y la rueda mantienen el comportamiento habitual del ratón. Estas asignaciones de teclado no se exportan a juegos ni emuladores.

La interfaz utiliza la Gamepad API de navegador/WebView2 y acepta dispositivos con `mapping: standard`. El mapeo sigue las posiciones físicas del [estándar Gamepad](https://www.w3.org/TR/gamepad/): botón inferior para confirmar, derecho para volver, cruceta y stick izquierdo para navegar. La marca del mando no garantiza que el host lo exponga con este mapeo; los dispositivos sin mapeo estándar no producen acciones.

Una lectura compartida sirve a perfiles, configuración inicial y biblioteca. Cualquiera de los mandos compatibles conectados puede tomar el control al producir una acción. Confirmar y volver requieren soltar antes de repetirse; las direcciones se repiten cada 240 ms. Al perder foco o esconder la ventana se pausa la entrada y, al volver, se espera a que los controles de navegación estén en reposo.

En Gestión → Controles → Asignación → Probar visualmente se muestran el identificador del dispositivo y sus botones físicos pulsados, incluidos gatillos y botones de stick. Durante la prueba se captura la navegación para evitar activar opciones accidentalmente. B / Círculo sale de la prueba; también se puede usar teclado o ratón.

Los perfiles de destino siguen siendo configuraciones guardadas: no se inyectan en juegos ni se exportan a emuladores. La zona muerta, sensibilidad e inversión de esos perfiles no modifican la navegación de Gemu, que usa un umbral de 0,5 para el stick. La navegación continúa usando Gamepad API; XInput nativo se utiliza para batería y número de canal. No se implementan vibración ni giroscopio.

## Identificación y arte

El mapeador detecta automáticamente Xbox, DualShock 4, DualSense (incluido Edge) y Switch Pro / Joy-Con combinados. Primero compara los identificadores de fabricante y producto expuestos por Chromium o Firefox; después reconoce nombres explícitos. Los identificadores se contrastaron con la [lista de mandos de Chromium](https://chromium.googlesource.com/chromium/src/+/HEAD/device/gamepad/gamepad_id_list.h). Un nombre ambiguo como «Wireless Controller» sin identificadores no se clasifica como PlayStation.

La familia detectada selecciona su perfil y una ilustración SVG original: Xbox con sticks asimétricos, DualShock 4 con touchpad y sticks simétricos, DualSense con carcasa blanca y Switch Pro con etiquetas Nintendo. Los puntos de asignación se colocan según cada dibujo. El arte es local y no requiere descargar imágenes ni recursos de Steam.

La opción Dispositivo permite elegir entre los mandos conectados; Automático sigue al mando activo. Cambiar la familia manualmente desactiva la detección automática hasta reactivarla. Un dispositivo desconocido conserva la plantilla elegida y se identifica como desconocido. Detectar su familia no convierte un dispositivo sin mapeo estándar en uno compatible. Joy-Con individuales no se presentan como un Switch Pro.

Si Steam Input, DS4Windows u otro controlador virtual expone un mando PlayStation como Xbox/XInput, la interfaz identifica el dispositivo expuesto como Xbox. Se puede elegir la plantilla PlayStation manualmente.

## Identidad, jugadores y mapeos propios

Gestión → Controles → Identidad permite registrar un mando con un UUID local de Gemu, darle un nombre (hasta 64 caracteres) y asociarlo a un perfil de jugador. El registro se comparte entre los jugadores del mismo almacenamiento local (`gemu.controllers.v1`). Los mapeos se guardan en la configuración del jugador bajo `destino:familia:uuid`; los perfiles existentes por familia sirven de plantilla cuando el mando no tiene un mapeo propio. Registrar, renombrar y asociar un jugador se guarda inmediatamente; las asignaciones se guardan con la barra de gestión.

El UUID local no es un número de serie del hardware. Los índices de navegador y slots XInput son temporales; tampoco se debe identificar un dispositivo por su nombre de modelo, porque dos mandos idénticos pueden reportar el mismo nombre. La asociación de conexión se descarta al observar una desconexión. Al reconectar se elige el registro desde «Mando guardado»; un registro no puede asociarse a dos dispositivos conectados a la vez. Un proveedor futuro puede aportar `stableId` para recuperar automáticamente una identidad única. Los IDs no se usan como credenciales ni provocan un cambio de jugador automático: al volver al selector de perfiles se muestra el jugador sugerido si la asociación de esa conexión sigue disponible.

## Batería y LEDs

El proveedor nativo Windows enumera los cuatro canales XInput y consulta [XInputGetBatteryInformation](https://learn.microsoft.com/en-us/windows/win32/api/xinput/nf-xinput-xinputgetbatteryinformation). La UI distingue cable, batería desconocida y niveles vacía/baja/media/completa. XInput no ofrece porcentaje, por lo que no se convierte el nivel a un porcentaje. PS4, PS5 y Switch conectados directamente por Gamepad API no tienen telemetría de batería en este proveedor.

El canal XInput no se empareja automáticamente con el índice del navegador. «Canal de batería» exige elegirlo explícitamente, solo para la conexión actual, y se limpia al cambiar de dispositivo. Una desconexión observada genera un nuevo identificador de sesión XInput. Una desconexión y reconexión que ocurra íntegramente entre dos consultas puede no observarse; si existe duda, debe seleccionarse de nuevo el canal. El número mostrado es el slot XInput (1–4), no un número físico garantizado del mando.

La gestión de LEDs queda condicionada a la capacidad `player-led`: el editor se habilita únicamente cuando un proveedor la anuncia. Los proveedores actuales no la anuncian: Gamepad API y la API pública XInput no permiten cambiar estos LEDs. El método `SetControllerFeature` ya enruta operaciones soportadas a proveedores específicos y rechaza dispositivos desconectados o capacidades no disponibles. Añadir protocolos HID para LEDs de Sony/Nintendo requiere un proveedor específico y validación con ese hardware; no se envían informes HID supuestos.

## Contratos para nuevos proveedores

- Go: `internal/controllers.Provider` implementa `ID`, `Devices` y `SetFeature`. `Device` incluye identidad de conexión, identidad persistente opcional, familia, batería y un diccionario de capacidades. `Service` combina proveedores y comprueba capacidad/conexión antes de mutaciones. Registrar nuevas implementaciones en `controllers.go`.
- TypeScript: `ControllerProvider` y `ControllerProviders` ofrecen enumeración y envío genérico de features. El adaptador nativo usa los métodos Wails `GetControllers` / `SetControllerFeature`; la ausencia del host produce una lista vacía en la previsualización web. Los errores de un proveedor se separan de sus resultados.
- La identificación del modelo y su arte permanecen en los adaptadores de presentación existentes. Una nueva familia con botones o arte distinto debe extender ese catálogo; nuevas capacidades como color, giroscopio o gatillos pueden añadirse al diccionario y al proveedor sin cambiar el registro de UUID.

Para usar los nuevos métodos Wails hay que reiniciar la aplicación de escritorio. La lectura de capacidades nativas se consulta cada dos segundos durante la gestión y se pausa cuando la ventana pierde foco o está oculta. La compatibilidad con batería y LEDs no puede verificarse únicamente con pruebas unitarias: debe probarse con el modelo, firmware y transporte reales.

## Validación física

### Entrada virtual en Gemu

Para controlar ventanas externas de juegos y launchers existe además el
[asistente de escritorio independiente](desktop-assistance.md), con herramienta
flotante, entrada nativa y bandeja del sistema. Los controles descritos debajo
pertenecen a la interfaz HTML de Gemu.

No hace falta conectar teclado ni ratón para introducir texto en la aplicación. Enfoca un campo y pulsa Home + X en Xbox (Home + □ en PlayStation; Home + Y en Nintendo) para abrir o cerrar el teclado virtual. Son los botones estándar 16 + 2. Hay que soltar ambos botones antes de repetir la combinación. Steam y Game Bar también pueden responder a Home. Llegar al campo con la cruceta o hacer clic sobre él no abre el teclado ni captura la navegación. El panel compacto se ancla al borde inferior, muestra el texto que estás editando y oculta las contraseñas. Funciona desde la creación de perfiles, el asistente inicial, las búsquedas y la gestión de rutas o credenciales. Los cambios se aplican al campo; «Listo» cierra el teclado y devuelve el foco, sin enviar formularios ni guardar ajustes por sí solo.

- Cruceta o stick izquierdo: elegir una tecla; botón inferior: escribir; botón derecho: cerrar.
- Botón izquierdo: borrar; botón superior: alternar mayúsculas.
- LB: letras/símbolos (incluye acentos y caracteres de rutas); RB: espacio.
- R3 (pulsar el stick derecho): activar/desactivar el ratón virtual; si el teclado estaba abierto, lo cierra. También funciona con un campo de texto enfocado. B / Esc sale del ratón virtual.
- En modo ratón, stick izquierdo: mover el cursor con aceleración y zona muerta; stick derecho vertical: desplazar el panel bajo el cursor; botón inferior: clic; botón izquierdo: menú contextual; botón derecho: salir. Tras enfocar un selector o deslizador, la cruceta ajusta su valor.

La entrada virtual se pausa al perder foco y requiere soltar los controles al volver. Respeta la captura del menú de partida y de la prueba de mandos. Utiliza el mapeo estándar del navegador y actúa dentro de Gemu; no emula dispositivos del sistema ni envía entrada a juegos, emuladores o diálogos nativos de Windows. El navegador no permite detectar con fiabilidad si hay teclado o ratón físicos conectados, por lo que la función se ofrece al utilizar el mando sin depender de esa detección.

Para validar con hardware, crear un perfil, buscar un juego y editar una ruta usando solo el mando; comprobar contraseñas, mayúsculas, selección de texto y borrado. Pasar por búsqueda con la cruceta debe permitir seguir navegando sin abrir el teclado. Abrirlo con Home + X, cerrarlo con B y seguir navegando; repetir con Home + X para cerrarlo y soltar Home antes que X para comprobar que no cambia favoritos. Comprobar Home para recuperar el menú de Gemu durante una partida, además de View + Menu y F10. Activar R3, hacer clic y desplazar un panel largo; repetir con un campo enfocado y con el teclado abierto. Mantener un botón al cambiar de ventana o abrir/cerrar el teclado no debe producir una segunda activación.

1. En Windows/WebView2, conectar un mando y pulsar un botón para que el host lo exponga.
2. Navegar por perfiles, asistente, biblioteca y ajustes con cruceta y stick; mantener confirmar al abrir una pantalla no debe activar otra opción.
3. Conectar un segundo mando y usarlo sin desconectar el primero; desconectar el activo y continuar con el otro.
4. Cambiar de ventana manteniendo un botón: volver a Gemu no debe activarlo hasta soltar y pulsar otra vez.
5. Abrir la prueba visual, comprobar botones y salir con B / Círculo.

Las pruebas automatizadas cubren lectura, umbral analógico, prioridad de cruceta, varios mandos, desconexión y recuperación del foco. La compatibilidad física debe verificarse con hardware real.

## Acciones de la biblioteca

Sobre un juego seleccionado, el botón inferior (A en Xbox) lo lanza directamente, el izquierdo (X) añade o quita favoritos y el superior (Y) abre el submenú opcional. En PlayStation corresponden a ×, □ y △; en Nintendo, B, Y y X. Con teclado se usan Enter, F e I; con ratón, clic para jugar y clic derecho o «Ver juego» para abrir el submenú. Los títulos sin instalación muestran un aviso al intentar jugar.

El submenú enfoca primero «Jugar ahora» cuando está disponible; de lo contrario, favoritos. La cruceta recorre las acciones en orden y B / Esc cierra el panel y devuelve el foco a la biblioteca. Mantener X o Y pulsado no repite la acción.
