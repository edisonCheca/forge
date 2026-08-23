# Forge - AI-Powered Software Engineering Assistant

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Status](https://img.shields.io/badge/Status-Phase_1_Active-success)

**Forge** es un asistente de ingeniería de software inteligente de interfaz de línea de comandos (CLI) diseñado para agilizar tu flujo de desarrollo local. Forge analiza de manera inteligente los cambios en tu entorno de trabajo de Git para generar mensajes de commit precisos bajo estándares de la industria y automatizar la creación de Pull Requests ejecutivos listos para producción.

---

## Características Principales

- **Conventional Commits Estrictos**: Genera mensajes que cumplen a cabalidad con el estándar Conventional Commits (`feat(auth): ...`), garantizando compatibilidad total con linters estrictos como `commitlint` y herramientas de automatización de versiones.
- **Creación Inteligente de Pull Requests (`forge pr`)**: Sintetiza el historial de commits de tu rama contra la rama base, estructurando historias de usuario, subtareas resueltas y notas de diseño, publicando directamente en GitHub mediante la integración con GitHub CLI (`gh`).
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

### 1. Generador Inteligente de Commits (`forge commit`)

1. **Prepara tus cambios en el staging area de Git**:
   ```bash
   git add .
   ```

2. **Ejecuta Forge Commit**:
   ```bash
   forge commit
   ```

   *(Opcional) Puedes asociar directamente un ID de issue o subtarea*:
   ```bash
   forge commit --issue 10
   # o usando el alias corto:
   forge commit -i 10
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

### 2. Creación Automatizada de Pull Requests (`forge pr`)

Forge analiza automáticamente los commits realizados en tu rama actual frente a la rama base, sintetiza un resumen ejecutivo con IA y crea el Pull Request en GitHub:

1. **Ejecuta el comando desde tu rama de funcionalidad**:
   ```bash
   forge pr
   ```

2. **Opciones disponibles**:
   - `--base` / `-b`: Especifica la rama base destino *(por defecto: `develop`)*.
   - `--extra` / `-e`: Agrega contexto adicional, notas técnicas o decisiones de arquitectura al PR.

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
