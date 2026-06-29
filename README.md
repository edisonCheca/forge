# Forge - AI-Powered Software Engineering Assistant

**Forge** es un asistente de ingeniería de software inteligente de interfaz de línea de comandos (CLI) diseñado para agilizar tu flujo de desarrollo local. En su primera fase, Forge analiza de manera inteligente los cambios en el *staging area* de Git y genera mensajes de commit precisos, descriptivos y con formato perfecto.

---

## ✨ Características (Fase 1)

- **⚡ Conventional Commits Estrictos**: Genera mensajes que cumplen a cabalidad con el estándar Conventional Commits (`feat(auth): ...`), garantizando compatibilidad total con linters estrictos como `commitlint` y herramientas de automatización de versiones.
- **🌐 Agnóstico al Proveedor de IA**: No estás atado a un solo proveedor. Configura y utiliza cualquier API compatible con OpenAI, incluyendo **OpenRouter**, **OpenAI**, **Groq**, **Anthropic** (vía gateways) o modelos ejecutados localmente (como Ollama o LM Studio).
- **🧙‍♂️ Asistente Interactivo (Wizard)**: Olvídate de configurar variables de entorno complejas a mano. Al ejecutar Forge por primera vez, un asistente guiado te ayudará a configurar tu proveedor, modelo e idioma favorito, guardando todo de forma segura en `~/.forge.json` (con permisos protegidos `0600`).
- **🌍 Soporte Multilenguaje Nativo**: Genera propuestas de commit directamente en tu idioma local (**Español**, Inglés, etc.) con formato conciso y viñetas cortas que respetan el límite de 80 caracteres por línea.
- **🛡️ Human-in-the-Loop Estricto**: La IA propone, pero tú siempre tienes el control final. Revisa la propuesta generada y acéptala con un simple `Enter` o cáncelala instantáneamente.

---

## 🚀 Instalación

### Prerrequisitos
Asegúrate de tener [Go](https://go.dev/doc/install) instalado en tu sistema (versión 1.21 o superior).

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

## 💡 Uso

El flujo de trabajo es extremadamente rápido y natural:

1. **Prepara tus cambios en Git** sumando los archivos que deseas incluir en tu próximo commit:
   ```bash
   git add .
   ```

2. **Ejecuta el generador inteligente de commits**:
   ```bash
   forge commit
   ```

Si es tu primera vez ejecutando el comando, el **Wizard interactivo** te solicitará los datos de configuración básica. Tras la configuración (o en ejecuciones posteriores), Forge analizará el diff y te presentará una propuesta lista para confirmar:

```text
Analizando cambios en staging...

Propuesta de Commit:

feat(cli): integrar wizard interactivo de configuración y soporte en español

- añadir persistencia segura de credenciales en ~/.forge.json
- implementar selección automática de idioma por defecto
- reestructurar adaptador http para dar soporte agnóstico a openrouter

¿Aceptar este commit? [Y/n]: Y
✅ Commit creado exitosamente.
```

---

## 🏗️ Arquitectura

Forge está construido bajo los principios de la **Arquitectura Hexagonal Pragmática (Puertos y Adaptadores)**. El núcleo computacional y las reglas de negocio (`core`) se encuentran completamente desacoplados de la infraestructura externa (como el sistema de archivos, el CLI de Git y los clientes HTTP de IA).

Para conocer a fondo las decisiones arquitectónicas, principios de ingeniería y el diseño de la base del código, consulta nuestro documento de arquitectura:

👉 **[Documento de Arquitectura de Forge](docs/ARCHITECTURE.md)**

---

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Consulta el archivo [LICENSE](LICENSE) para más detalles.

Copyright (c) 2026 Edison Checa.
