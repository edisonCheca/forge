# Forge - AI-Powered Software Engineering Assistant

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Status](https://img.shields.io/badge/Status-Phase_1_Active-success)

**Forge** es un asistente de ingeniería de software inteligente de interfaz de línea de comandos (CLI) diseñado para agilizar tu flujo de desarrollo local. Forge analiza de manera inteligente los cambios en tu entorno de trabajo de Git para generar mensajes de commit precisos bajo estándares de la industria y automatizar la creación de Pull Requests ejecutivos listos para producción.

---

## Características Principales

- **Creación Inteligente de Ramas (`forge start`)**: Genera ramas de trabajo estandarizadas vinculadas a Historias de Usuario e Issues de GitHub (`feat/5-hu-01-control-de-acceso`), resolviendo metadatos y activando la rama automáticamente.
- **Auditoría Híbrida de Seguridad y Calidad (`forge review`)**: Inspecciona cambios en staging o ramas antes de comitear. Detecta archivos sensibles (`.env`), secretos filtrados (`sk-...`), llamadas de depuración olvidadas (`console.log`, `fmt.Println`) y ofrece recomendaciones semánticas con IA.
- **Conventional Commits Estrictos (`forge commit`)**: Genera mensajes que cumplen a cabalidad con el estándar Conventional Commits (`feat(auth): ...`), garantizando compatibilidad total con linters estrictos como `commitlint` y herramientas de automatización de versiones.
- **Creación Automatizada de Pull Requests (`forge pr`)**: Sintetiza el historial de commits de tu rama contra la rama base, estructurando historias de usuario, subtareas resueltas y notas de diseño, publicando directamente en GitHub mediante la integración con GitHub CLI (`gh`).
- **Trazabilidad de Tareas e Issues (`--issue`)**: Vincula identificadores de tareas o subtareas (ej. `#10`, `PROJ-123`) a tus commits y PRs para mantener trazabilidad con gestores como Jira o GitHub Issues.
- **Agnóstico al Proveedor de IA**: Compatible con cualquier API basada en OpenAI, incluyendo **OpenRouter**, **OpenAI**, **Groq**, **Anthropic** (vía gateways) o modelos ejecutados localmente (**Ollama**, **LM Studio**).
- **Asistente Interactivo (Wizard)**: Configuración guiada en la primera ejecución que almacena tus preferencias de forma segura en `~/.forge.json` (con permisos protegidos `0600`).
- **Terminal UI Moderna y Elegante**: Mensajes estilizados con colores TrueColor ANSI, íconos descriptivos y banners informativos.
- **Soporte Multilenguaje Nativo**: Genera propuestas de commit y PRs en tu idioma preferido (**Español**, Inglés, etc.).
- **Human-in-the-Loop Estricto**: Tú siempre tienes el control final. Revisa la propuesta generada y confírmala interactivamente con un simple `Enter` o cancélala en cualquier momento.

---

## Instalación

### Prerrequisitos
- [Go](https://go.dev/doc/install) (versión 1.21 o superior).
- [Git](https://git-scm.com/) configurado en el sistema.
- [GitHub CLI (`gh`)](https://cli.github.com/) *(opcional, necesario para `forge pr`)* autenticado con `gh auth login`.

### Instalación desde el código fuente

Clona el repositorio e instala el binario en tu sistema:

```bash
git clone git@github.com:edisonCheca/forge.git
cd forge
go install .
```

O instálalo directamente vía `go install`:

```bash
go install github.com/edisonCheca/forge@latest
```

> **Nota sobre el PATH**: Asegúrate de que el directorio de binarios de Go esté incluido en la variable de entorno `PATH` de tu sistema:
> - **Linux/macOS**: `export PATH=$PATH:$(go env GOPATH)/bin`
> - **Windows (PowerShell)**: `$env:Path += ";$env:USERPROFILE\go\bin"`

---

## Uso

### 1. Creación Estandarizada de Ramas (`forge start`)

Inicia el desarrollo de una nueva Historia de Usuario o incidencia creando una rama con la convención óptima. Forge se conecta con GitHub para resolver el título y las etiquetas del issue automáticamente:

```bash
# Modo Ultra Rápido (Consulta automática del Issue en GitHub):
forge start 5

# Modo con argumentos explícitos:
forge start 5 "Control de Acceso RBAC" --hu HU-01 --type feat

# Modo interactivo guiado:
forge start
```

*(Opciones disponibles)*:
- `--type` / `-t`: Tipo de cambio (`feat`, `fix`, `docs`, `refactor`, `test`, `chore`). Por defecto: `feat` o el inferido del issue.
- `--hu` / `-u`: Código opcional de la Historia de Usuario (ej. `HU-01` o `1`).
- `--base` / `-b`: Rama base origen (por defecto: rama actual).

#### Ejemplo de Salida (Modo Ultra Rápido con Issue):
```text
→ Consultando detalles del Issue #5 en GitHub...
✓ Issue #5 encontrado: "[HU-01] Control de Acceso Basado en Roles (RBAC)"

─────────────────────────────────────────────────────────────────
◆ Propuesta de Nueva Rama:
─────────────────────────────────────────────────────────────────
Rama: feat/5-hu-01-control-de-acceso-basado-en-roles-rbac
Base: main
─────────────────────────────────────────────────────────────────
» ¿Crear y cambiar a esta rama? [Y/n]: Y

✓ Rama 'feat/5-hu-01-control-de-acceso-basado-en-roles-rbac' creada y activada exitosamente.
ℹ Flujo sugerido: realiza tus cambios, ejecuta 'forge commit -i 5' y al finalizar 'forge pr'.
```

---

### 2. Auditoría Previa de Código (`forge review`)

Audita los cambios preparados en staging antes de comitear o revisa los commits acumulados en tu rama frente a la rama base:

```bash
# Auditar cambios en staging con reglas estáticas e IA:
forge review

# Modo rápido sin costo (solo reglas estáticas deterministas, 0 tokens):
forge review --static-only

# Auditar la rama completa frente a develop/main:
forge review --branch --base main
```

*(Opciones disponibles)*:
- `--static-only` / `-s`: Ejecuta únicamente auditoría estática determinista sin invocar la API de IA (0 costo).
- `--branch` / `-b`: Audita los cambios de la rama frente a la rama base (útil antes de abrir un PR).
- `--base`: Rama base de comparación en modo rama (por defecto: `develop`).
- `--extra` / `-e`: Pasa contexto sobre la intención del código para que la IA refine su análisis.

#### Ejemplo de Salida:
```text
→ Iniciando auditoría de código sobre Área de Staging (cambios preparados)...

─────────────────────────────────────────────────────────────────
◆ Resultados de la Auditoría (forge review)
─────────────────────────────────────────────────────────────────
Alcance:               Área de Staging (cambios preparados)
Archivos auditados:    3
─────────────────────────────────────────────────────────────────

🟡 ADVERTENCIAS DE CALIDAD Y LIMPIEZA:
  1. ⚠ [Calidad] Archivo: src/service.go (Línea ~34)
     Mensaje: Llamada 'fmt.Println' en paquete de lógica/dominio. Utilice un logger estructurado.

◆ Análisis Semántico y Sugerencias de Arquitectura (IA)
- Asegurar validación de punteros nulos en `ProcessTransaction`.
- Manejar adecuadamente el timeout del contexto en llamadas a repositorios externos.

─────────────────────────────────────────────────────────────────
✓ AUDITORÍA EXITOSA: Código limpio, seguro y listo para producción.
```

---

### 3. Generador Inteligente de Commits (`forge commit`)

1. **Prepara tus cambios en el staging area de Git**:
   ```bash
   git add .
   ```

2. **Ejecuta Forge Commit**:
   ```bash
   forge commit
   ```

   *(Opciones disponibles)*:
   - `--issue` / `-i`: Asocia un ID de issue o subtarea (ej. `10` o `#10`).
   - `--extra` / `-e`: Pasa notas técnicas o contexto de intención directamente al prompt de la IA para enriquecer el análisis del diff al generar el commit.
   - `--concise`: Genera una propuesta de commit ultra concisa (máximo 1-2 viñetas breves).
   - `--detailed`: Genera una propuesta de commit exhaustiva y detallada (4-7 viñetas estructuradas).

   ```bash
   # Flujo estándar automático (proporcional al diff):
   forge commit

   # Con contexto y subtarea:
   forge commit -i 10 -e "Refactorización del cliente HTTP para soporte de reintentos"

   # Forzando modo conciso o detallado:
   forge commit --concise
   forge commit --detailed
   ```

#### Ejemplo de Salida:
```text
→ Analizando cambios en staging...

============================================================
  Propuesta de Commit (openrouter/free)
============================================================

feat(cli): integrar wizard interactivo de configuración y soporte en español

- añadir persistencia segura de credenciales en ~/.forge.json
- implementar selección automática de idioma por defecto
- reestructurar adaptador http para dar soporte agnóstico a openrouter

============================================================
? ¿Aceptar este commit? [Y/n]: Y

✔ Commit creado exitosamente
```

---

### 4. Creación Automatizada de Pull Requests (`forge pr`)

Forge analiza automáticamente los commits realizados en tu rama actual frente a la rama base, sintetiza un resumen ejecutivo con IA y crea el Pull Request en GitHub:

1. **Ejecuta el comando desde tu rama de funcionalidad**:
   ```bash
   forge pr
   ```

2. **Opciones disponibles**:
   - `--base` / `-b`: Especifica la rama base destino *(por defecto: `develop`)*.
   - `--extra` / `-e`: Proporciona notas técnicas o decisiones de arquitectura para que la IA las integre orgánicamente en el resumen y cuerpo del PR.

   ```bash
   forge pr --base main --extra "Se optimizó el adaptador HTTP con reintentos exponenciales."
   ```

#### Ejemplo de Salida:
```text
→ Analizando commits en la rama 'feature/10-auth-flow'...
→ Sintetizando resumen ejecutivo del PR con Inteligencia Artificial...

============================================================
  Propuesta de Pull Request:
============================================================
Título: feat(auth): implementar flujo de autenticación JWT (#10)
Base:   main <- feature/10-auth-flow

Cuerpo:
## Historia de Usuario
Closes #10

## Subtareas Completadas
- Resolves #11
- Resolves #12

## Notas de Diseño / Decisiones
Se optimizó el adaptador HTTP con reintentos exponenciales.
============================================================

? ¿Crear este Pull Request en GitHub? [Y/n]: Y

→ Sincronizando rama 'feature/10-auth-flow' con origin...
→ Invocando GitHub CLI (gh pr create)...

✔ Pull Request creado exitosamente: https://github.com/edisonCheca/forge/pull/15
```

---

## Configuración

Forge guarda su configuración en el archivo `~/.forge.json`. Puedes editarlo directamente o dejar que el wizard interactivo lo genere al ejecutar `forge commit` por primera vez:

```json
{
  "base_url": "https://openrouter.ai/api/v1/chat/completions",
  "model": "openrouter/free",
  "api_key": "sk-or-v1-...",
  "language": "es"
}
```

### Variables de Entorno (Opcional)
También puedes sobreescribir la configuración usando variables de entorno:
- `FORGE_AI_BASE_URL`: URL del endpoint compatible con OpenAI.
- `FORGE_MODEL`: Modelo a utilizar (ej. `openrouter/free`, `openai/gpt-4o-mini`).
- `FORGE_LANGUAGE`: Idioma de generación (`es`, `en`, etc.).
- `FORGE_CONVENTIONAL`: `true` o `false` para exigir Conventional Commits.

---

## Arquitectura

Forge está construido bajo los principios de la **Arquitectura Hexagonal Pragmática (Puertos y Adaptadores)**. El núcleo computacional y las reglas de negocio (`core`) se encuentran completamente desacoplados de la infraestructura externa (como el sistema de archivos, el CLI de Git y los clientes HTTP de IA).

Para conocer a fondo las decisiones arquitectónicas, principios de ingeniería y el diseño de la base del código, consulta nuestro documento de arquitectura:

[Documento de Arquitectura de Forge](docs/ARCHITECTURE.md)

---

## Licencia

Este proyecto está bajo la Licencia MIT. Consulta el archivo [LICENSE](LICENSE) para más detalles.

Copyright (c) 2026 Edison Checa.
