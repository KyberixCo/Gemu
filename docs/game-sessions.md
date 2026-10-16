# Lanzamiento y menú de partida

La opción **Asistencia de escritorio** abre el
[asistente independiente con bandeja y herramienta flotante](desktop-assistance.md)
para usar teclado y ratón nativos ante un menú o login externo. Su ciclo de vida
es independiente del de Gemu. Mientras responde, coordina las combinaciones
globales del mando; Home solo abre el menú al soltarlo para distinguir Home + X.

El menú de partida usa la ventana externa de Gemu. Alcance y límites de plataforma y modos de pantalla en [Diseño y compatibilidad del menú de partida](game-menu.md).

RetroArch utiliza el mismo menú y ciclo de retorno. Se inicia con ejecutable, núcleo y ROM validados, mediante argumentos directos y configuración privada por jugador. El seguimiento usa el PID, fecha de creación y ruta del proceso que Gemu inició; no incorpora otros RetroArch ni auxiliares de su carpeta. Dos lecturas sin ese proceso confirman su salida. Una instancia ya abierta impide lanzar otra. Configuración y requisitos en [RetroArch](retroarch.md).

Durante el juego Gemu entra en bajo consumo: pausa animaciones CSS y los bucles de efectos visuales, detiene la lectura del mando en el navegador y suspende el AudioContext. El estado de partida llega por eventos del host, sin consultas periódicas desde la interfaz en el escritorio. Se conserva la biblioteca y selección en memoria para volver sin recargar. Abrir el menú o volver a Gemu reactiva la interfaz; la música solo se activa si Gemu tiene foco y las preferencias la permiten.

El host procesa los atajos cada 50 ms y busca procesos cada segundo durante la partida (cada 500 ms al iniciar/cerrar). En Windows, F10 se registra con RegisterHotKey y MOD_NOREPEAT durante la sesión: las pulsaciones breves quedan en la cola de mensajes y mantener la tecla no alterna repetidamente el menú. El registro y la lectura usan el mismo hilo de Windows; F10 se libera al terminar la sesión o salir de Gemu. Si otra aplicación ya ocupa F10, se registra el fallo en el log y se conserva el sondeo de teclado como alternativa. Los canales XInput desconectados se comprueban cada dos segundos. No se suspende el proceso de Gemu ni se cambia la prioridad del juego. Estas medidas reducen trabajo de CPU/GPU; no garantizan un consumo de RAM concreto y requieren medición con WebView2 y juegos reales.

En Windows, «Jugar ahora» abre la instalación de Steam mediante `steam://run/<appid>`. Solo se aceptan juegos de la biblioteca seleccionada, instalados y con una ruta absoluta. Abrir Steam no confirma que el juego esté ejecutándose: Gemu espera procesos nuevos cuyo ejecutable esté dentro de la carpeta de instalación. No se inicia otra sesión mientras haya una activa.

El menú ofrece continuar, volver al menú principal dejando el juego abierto y solicitar el cierre con confirmación. Al terminar los procesos detectados, Gemu recupera su ventana y conserva la selección. Durante la partida se suspende la navegación de Gemu; al volver voluntariamente a su biblioteca se habilita nuevamente.

Gemu inicia en pantalla completa nativa de Wails y la recupera al abrir el menú de partida o volver a la biblioteca. El ajuste de pantalla completa también usa el host en escritorio, conservando la API del navegador para las previsualizaciones. En Windows solo se restaura una ventana nativa si está minimizada, evitando cambiar el tamaño de una ventana ya visible.

La recuperación localiza la ventana principal `wailsWindow` del proceso de Gemu, incluso oculta o minimizada, y restaura el HWND antes de aplicar pantalla completa. Esto evita depender de `WindowUnminimise`, que Wails omite si conserva su indicador de fullscreen, o de `WindowShow`, que encola trabajo en el hilo de interfaz. La prioridad de la ventana se aplica al final con SetWindowPos y SWP_NOACTIVATE; retirarla no devuelve el foco a Gemu después de continuar el juego.

El menú ocupa toda la pantalla: acciones en una columna lateral, fondo y título del juego activo, estado de ejecución y ayudas de entrada al pie. Comparte las superficies planas, esquinas rectas, acento ácido y foco en un único borde del diseño de Gemu. El foco comienza en continuar (o en biblioteca mientras arranca), Tab queda dentro del menú y Escape/B permite regresar al juego. La confirmación de cierre da prioridad a cancelar; cancelar devuelve el foco a cerrar. Las peticiones pendientes bloquean acciones repetidas y los errores se muestran dentro del menú. No se presentan horas ni progreso ficticios. El aviso de partida en segundo plano aparece únicamente al volver a la biblioteca.

El menú se abre con Home / Guide / Xbox, F10 o View + Menu (Back + Start). El acceso global con Home en Windows utiliza la extensión opcional XInputGetStateEx cuando está disponible; de lo contrario, se conservan F10 y View + Menu. En el navegador se reconoce Home como botón estándar 16 cuando Gemu tiene foco. Home + X abre el teclado virtual sobre un campo de texto. Steam y Game Bar también pueden responder a Home: Gemu no modifica su configuración ni bloquea sus overlays.

Si la ventana de desarrollo queda minimizada u oculta y no puedes recuperarla, ejecuta `powershell -NoProfile -File scripts/restore-window.ps1` desde este repositorio. El script restaura la ventana existente de Gemu y solicita traerla al frente; no inicia otra instancia ni cambia su prioridad a siempre visible. Solo actúa sobre los ejecutables de este checkout en `build/bin`. `-InspectOnly` muestra el estado sin modificar la ventana. Debe ejecutarse en el escritorio del usuario, ya que un entorno aislado no puede enumerar sus ventanas.

El menú es la ventana de Gemu traída al frente y mantenida encima mientras está abierto; al continuar, volver a la biblioteca o terminar la partida se retira esa prioridad. El juego continúa ejecutándose mientras el menú está abierto. F10 queda reservado para Gemu mientras la sesión está activa; los botones del mando siguen llegando al juego. El fullscreen exclusivo, juegos elevados y políticas de foco de Windows requieren validación con hardware real. Si Windows rechaza traer Gemu al frente, el fallo queda en el log. Si rechaza enfocar el juego, la interfaz indica usar Alt+Tab.

Cerrar envía WM_CLOSE a ventanas de procesos detectados, validando PID, fecha de creación y ruta antes de actuar. No mata Steam ni fuerza la terminación. Si el juego pide guardar o confirmar, hay que continuar jugando para responder. Procesos auxiliares sin ventanas se omiten cuando existe otra ventana a la que solicitar cierre. Un juego que ignora WM_CLOSE puede permanecer abierto.

El seguimiento Steam por carpeta es una heurística: no cubre ejecutables externos, algunas cadenas de launchers ni procesos inaccesibles. Tras dos minutos sin detectar un proceso se muestra un error, sin declarar una partida real ni inventar su duración. RetroArch utiliza seguimiento directo de su proceso. No se incluye historial de tiempo de juego, Epic ni otros emuladores en esta implementación.

Prueba manual: lanzar un juego Steam; abrir y recorrer el menú con mando; volver a Gemu; reabrir el menú y continuar; solicitar cierre y aceptar cualquier diálogo del juego; verificar regreso a la selección anterior. Repetir con salida desde el propio juego, mando desconectado, Game Bar activo y pantalla completa exclusiva.
