# Documento de Arquitectura Física: Fase 1 (CLI Inteligente Local)

**Proyecto:** Forge  
**Rol:** Arquitectura de Software Senior  
**Estado:** Definición de Arquitectura Física Inicial (Fase 1)  
**Clasificación:** Documento Técnico Fundamental  

---

## 1. Justificación de la Organización

El objetivo primordial al diseñar la arquitectura física de **Forge** en su Fase 1 es resolver la tensión fundamental entre **simplicidad inicial** y **capacidad de evolución a largo plazo**. 

Muchos proyectos Open Source orientados a herramientas de terminal cometen dos errores en extremos opuestos:
1. **El antipatrón del Monolito de Script:** estructurar todo el código en torno a los comandos del CLI (`commit.py`, `utils.py`), acoplando la interacción de la consola, las llamadas al sistema y la lógica de inteligencia artificial. Cuando el proyecto intenta crecer hacia integraciones externas o una biblioteca central (Core), el costo de refactorización es tan alto que suele requerir una reescritura total.
2. **El antipatrón de la Sobregeneración Prematura:** crear árboles profundos de directorios para plugins, agentes, marketplaces y redes que estarán vacíos o contendrán implementaciones ficticias durante meses o años, aumentando la carga cognitiva de los nuevos colaboradores.

La organización propuesta para Forge adopta una **Arquitectura Hexagonal Pragmática (Puertos y Adaptadores)** adaptada a una aplicación de línea de comandos. Esta estructura justifica su diseño al trazar límites físicos infranqueables entre tres responsabilidades excluyentes:
* **La Presentación (`cli`):** cómo interactúa el usuario con el sistema.
* **El Núcleo de Negocio (`core`):** qué hace el sistema y cuáles son sus reglas de validación y flujo.
* **La Infraestructura (`adapters`):** cómo se comunica el sistema con el mundo exterior (sistema de archivos, Git, APIs de Inteligencia Artificial).

Esta separación estructural garantiza que el flujo especificado para la Fase 1 (`forge commit`) se ejecute con máxima eficiencia local, mientras que las fronteras de los módulos preparan el terreno para las transiciones arquitectónicas de las Fases 2, 3 y 4 sin reorganizar ni un solo directorio existente.

---

## 2. Organización Física del Repositorio (Árbol de Fase 1)

El siguiente árbol de directorios refleja **únicamente** aquello que existe y tiene responsabilidad funcional activa en la Fase 1 del proyecto. Se excluyen deliberadamente carpetas vacías o especulativas.

```text
forge/
├── docs/                 # Documentación técnica, visión y arquitectura del proyecto
├── tests/                # Suite de pruebas automatizadas aisladas por módulo
└── src/
    └── forge/
        ├── cli/          # Capa de presentación: interfaz y comandos de terminal
        ├── core/         # Lógica de dominio, flujos de trabajo y reglas de validación
        ├── adapters/     # Integraciones y comunicación con sistemas externos
        │   ├── git/      # Adaptador para introspección del repositorio Git local
        │   └── ai/       # Adaptadores y abstracciones para proveedores de IA
        └── config/       # Carga, validación y gestión de preferencias locales
```

---

## 3. Responsabilidad de Cada Directorio

En lugar de describir archivos individuales, la arquitectura define **módulos funcionales** con responsabilidades estrictamente delimitadas:

### `docs/`
* **Responsabilidad:** Albergue de la base de conocimiento estática, decisiones arquitectónicas registradas (ADRs), visión del producto (`PROJECT_VISION.md`) y guías de contribución Open Source.
* **Límite:** No contiene código ejecutable ni scripts de automatización.

### `tests/`
* **Responsabilidad:** Aseguramiento de la calidad mediante pruebas unitarias, de integración y de contrato. Reproduce la topología de `src/forge/` para verificar cada módulo de forma aislada (por ejemplo, simulando respuestas de Git o de APIs de IA sin realizar llamadas de red reales).
* **Límite:** Código estrictamente de diagnóstico y verificación; no puede ser importado por el código de producción.

### `src/forge/`
* **Responsabilidad:** Directorio raíz del paquete ejecutable y distribuible de la aplicación. Actúa como el contenedor de espacio de nombres (namespace) de todo el software.

### `src/forge/cli/`
* **Responsabilidad:** Gestión exclusiva de la interfaz de línea de comandos. Interpreta los argumentos introducidos por el usuario, formatea la salida visual en la terminal, gestiona la interactividad (solicitud de confirmación `y/n` para ejecutar el commit) y traduce las excepciones técnicas en mensajes legibles para el desarrollador.
* **Límite:** No toma decisiones de negocio, no inspecciona archivos locales directamente ni se comunica con APIs externas. Solo invoca flujos expuestos por el módulo `core`.

### `src/forge/core/`
* **Responsabilidad:** El corazón del sistema. Contiene los casos de uso y la orquestación del flujo de la Fase 1 (`commit workflow`). Solicita la información del diff al adaptador Git, construye el contexto estructurado, lo envía a evaluar al adaptador de IA, ejecuta las reglas deterministas de validación sobre la propuesta recibida y devuelve el resultado final listo para confirmación.
* **Límite:** Es completamente agnóstico respecto a cómo se imprimen los datos en la pantalla, qué librería específica interactúa con Git o qué proveedor HTTP se utiliza para conectar con el modelo de IA.

### `src/forge/adapters/git/`
* **Responsabilidad:** Introspección y control del sistema de control de versiones local. Se encarga de ejecutar comandos o consultar librerías subyacentes para obtener los archivos modificados en el área de preparación (staging area), leer el historial reciente y, tras la autorización superior, ejecutar la orden formal de `git commit`.
* **Límite:** Desconoce por completo la existencia de motores de Inteligencia Artificial o de la interfaz gráfica de terminal. Traduce la salida cruda de Git a estructuras de datos puras entendibles por `core`.

### `src/forge/adapters/ai/`
* **Responsabilidad:** Comunicación con los proveedores de Inteligencia Artificial. Implementa las abstracciones requeridas para serializar el contexto del repositorio, formatear los prompts del sistema, gestionar la autenticación de red, aplicar estrategias de reintento ante fallos HTTP y normalizar las respuestas del modelo en texto estructurado.
* **Límite:** No evalúa si el commit generado cumple con las normas del proyecto de ingeniería (esa tarea pertenece a `core`). Solo garantiza que la comunicación con el proveedor sea robusta y confiable.

### `src/forge/config/`
* **Responsabilidad:** Centralización de la lectura, resolución y validación de parámetros de configuración (por ejemplo, claves de API, modelo de IA seleccionado, idioma de preferencia o convenciones de commit elegidas por el usuario).
* **Límite:** Provee datos estáticos o resueltos al resto de la aplicación, pero no ejecuta lógica operacional ni flujos de trabajo.

---

## 4. Dependencias entre Módulos

La salud arquitectónica de Forge depende del cumplimiento riguroso de una regla de dependencia unidireccional: **las capas externas dependen de las internas, nunca al revés**.

```text
       ┌─────────────┐
       │     cli     │
       └──────┬──────┘
              │ (Invoca casos de uso)
              ▼
       ┌─────────────┐      (Lee preferencias)      ┌─────────────┐
       │    core     │ ───────────────────────────► │   config    │
       └──────┬──────┘                              └─────────────┘
              │ (Consume interfaces / contratos)
              ▼
       ┌─────────────┐
       │  adapters   │
       └─────────────┘
```

### Dependencias Permitidas y Obligatorias:
1. **`cli` -> `core` & `config`:** La capa de presentación importa el orquestador del caso de uso (`core`) para iniciar el proceso de commit y consulta `config` para ajustar parámetros visuales o de interacción.
2. **`core` -> `config`:** El núcleo consulta las convenciones y reglas de negocio activas cargadas por el módulo de configuración.
3. **`core` -> `adapters` (mediante interfaces/contratos):** El núcleo invoca los servicios de introspección de Git y generación de IA. *Nota de ingeniería:* En términos de diseño de código, `core` definirá los contratos o interfaces abstractas que los módulos dentro de `adapters` deben satisfacer, garantizando un bajo acoplamiento real.

### Dependencias Prohibidas (Anti-patrones Arquitectónicos):
* **`core` -> `cli` [PROHIBIDO]:** El núcleo jamás debe importar formateadores de terminal, librerías de colores de consola o funciones de captura de teclado. Hacerlo imposibilitaría ejecutar Forge en un proceso automatizado o servidor sin pantalla (Fase 5).
* **`adapters/git` <-> `adapters/ai` [PROHIBIDO]:** Los adaptadores son horizontales y paralelos. El módulo de Git no debe comunicarse bajo ninguna circunstancia con el módulo de IA. La transferencia de información entre el código fuente local y el modelo inteligente es responsabilidad exclusiva de coordinación del `core`.
* **`adapters/*` -> `cli` [PROHIBIDO]:** Los adaptadores devuelven estructuras de datos, excepciones técnicas o estados; jamás imprimen directamente en la consola estándar del usuario.

---

## 5. Evolución Futura sin Reorganización

La validación suprema de esta estructura inicial es su capacidad para absorber la evolución planificada en la visión del producto manteniendo su esqueleto intacto:

### Hacia la Fase 2 (Asistente de Desarrollo Local)
En la Fase 2 se incorporan capacidades de revisión diferencial (`forge review`) y explicación de código (`forge explain`).
* **Cómo se integra:** Se añaden nuevos sub-módulos de lógica dentro de `src/forge/core/` (por ejemplo, flujos de análisis estático o generadores de revisión). El directorio `src/forge/cli/` simplemente registra los nuevos comandos para invocar estas nuevas rutinas del `core`. Si se requiere inspeccionar el árbol de sintaxis abstracta (AST) de los archivos, se añade un nuevo módulo en `src/forge/adapters/ast/`, sin tocar las carpetas existentes.

### Hacia la Fase 3 (Separación entre CLI y Núcleo Reutilizable)
En la Fase 3 se busca que Forge pueda ser consumido como biblioteca por terceras herramientas.
* **Cómo se integra:** Gracias a la estricta separación física lograda desde la Fase 1, el directorio `src/forge/core/` y `src/forge/adapters/` pueden empaquetarse y publicarse de manera autónoma como un paquete independiente (ej. `forge-core`). El directorio `src/forge/cli/` se transforma en un paquete consumidor (ej. `forge-cli`) que importa el núcleo. No hay que desenredar código porque la frontera física preexistía desde el día uno.

### Hacia la Fase 4 (Integraciones con Plataformas Externas)
En la Fase 4 el sistema interactúa con repositorios remotos y gestores de incidencias (GitHub, GitLab, Jira).
* **Cómo se integra:** Se crean nuevas carpetas hermanas bajo `src/forge/adapters/` (por ejemplo, `src/forge/adapters/github/`, `src/forge/adapters/jira/`). El núcleo (`core`) define qué datos necesita de una incidencia o pull request, y los nuevos adaptadores implementan esa obtención sin alterar los flujos de Git local o las interacciones de terminal existentes.

---

## 6. Riesgos Arquitectónicos Identificados y Mitigados

Durante el diseño de la Fase 1, se han evaluado y neutralizado proactivamente los siguientes riesgos de ingeniería:

1. **Riesgo de Fuga de Lógica de Negocio al CLI:**  
   * *Problema:* Que el código que gestiona los argumentos de la terminal empiece a filtrar el diff de Git o decida si un mensaje de commit es válido o no.  
   * *Mitigación:* Asignación estricta del CLI como capa tonta ("Humble Object"). Su única labor es capturar entradas, invocar una sola función del orquestador en `core` y renderizar el resultado que este devuelva.

2. **Riesgo de Acoplamiento Fuerte a un Proveedor de IA Específico:**  
   * *Problema:* Diseñar estructuras de datos o flujos pensando exclusivamente en el formato de API de un único modelo comercial.  
   * *Mitigación:* Aislar toda la comunicación dentro de `adapters/ai/` y obligar al `core` a trabajar con una abstracción de solicitud/respuesta agnóstica (ej. `PromptRequest` -> `CompletionResponse`). Cambiar de proveedor será un detalle de implementación invisible para el núcleo.

3. **Riesgo de Contaminación de Contexto Git con Lógica IA:**  
   * *Problema:* Que el módulo encargado de leer comandos Git intente truncar o preparar el texto específicamente para un modelo de lenguaje.  
   * *Mitigación:* El módulo `adapters/git/` entrega datos estructurales puros (archivos modificados, líneas añadidas/eliminadas, metadatos puros). Es el `core` quien toma esa información cruda y la procesa para construir el contexto semántico adecuado para la inteligencia artificial.

4. **Riesgo de Complejidad Prematura (Sobrediseño):**  
   * *Problema:* Crear capas abstractas de buses de eventos intermedios, inyectores de dependencias dinámicos o sistemas de plugins complejos en la Fase 1 para justificar la "extensibilidad".  
   * *Mitigación:* Aplicación rigurosa del principio YAGNI ("You Aren't Gonna Need It"). Se utilizan llamadas a funciones limpias y directas dentro de los límites modulares descritos. La arquitectura modular proporciona extensibilidad estructural sin sobrecarga computacional o cognitiva.

---

## 7. Decisiones Tomadas y Justificación Técnica

| Decisión Arquitectónica | Justificación Técnica | Beneficio Inmediato (Fase 1) | Beneficio Evolutivo (Fases 2-6) |
| :--- | :--- | :--- | :--- |
| **Uso del patrón de layout `src/` (`src/forge/`)** | Evita que el entorno de ejecución importe accidentalmente el código raíz no instalado y separa limpiamente los metadatos del repositorio (`docs`, `tests`) del código empaquetable. | Previene errores sutiles de rutas de importación durante el desarrollo local en la estación de trabajo. | Facilita la empaquetación limpia y distribución en repositorios de paquetes formales cuando el Core y el CLI se separen. |
| **Separación estricta entre `cli/` y `core/` desde el primer comando** | Rompe la tendencia natural a escribir scripts de terminal monolíticos separando la presentación del dominio. | Permite someter todo el flujo de generación de commits a pruebas unitarias rigurosas sin tener que simular interacciones de consola TTY. | Habilita la Fase 3 (separación Core/CLI) de forma casi instantánea y permite ejecutar Forge en pipelines sin interfaz gráfica (Fase 5). |
| **Agrupación de integraciones bajo el namespace `adapters/`** | Aplica el principio de inversión de dependencias aislando el dominio de bibliotecas o APIs externas inestables o cambiantes. | Permite simular (mockear) Git y los proveedores de IA con una sola línea en las pruebas automatizadas, haciendo los tests ultrarrapidos. | Permite agregar soporte para GitHub, GitLab o Jira en la Fase 4 agregando subcarpetas en `adapters/` sin tocar la lógica fundacional. |
| **Exclusión de directorios vacíos para funcionalidades futuras** | Respeta la carga cognitiva de colaboradores Open Source mostrando únicamente componentes con responsabilidades reales e implementadas. | El repositorio es limpio, fácil de entender y auditable por cualquier programador que se sume al proyecto en su etapa inicial. | Evita el mantenimiento de abstracciones especulativas incorrectas. Cuando la funcionalidad nace, el directorio se crea con el diseño exacto requerido. |
| **Centralización del módulo `config/` como dependencia hoja** | Evita la dispersión de lectura de variables de entorno o archivos de usuario en múltiples puntos del programa. | Unifica la validación de credenciales de IA y preferencias de formato de commit antes de ejecutar cualquier llamada costosa de red. | Permite que en el futuro la configuración provenga de variables institucionales CI/CD o políticas corporativas sin alterar los analizadores. |

---

## 8. Conclusión de Arquitectura

La estructura física inicial diseñada para **Forge Fase 1** representa un equilibrio exacto entre rigor arquitectónico y pragmatismo operativo. Al confinar la presentación en `cli`, proteger las reglas de ingeniería y orquestación dentro de `core` y encapsular la volátil interacción con Git y los modelos de Inteligencia Artificial en `adapters`, el proyecto adquiere una base estructural inmune a la degradación por crecimiento.

Esta organización cumple la promesa fundamental del documento de visión del producto: nacer como una herramienta de línea de comandos ágil y comprensible, dotada desde su primer día de un esqueleto arquitectónico preparado para escalar orgánicamente hasta convertirse en un ecosistema universal de asistencia para el desarrollo de software.
