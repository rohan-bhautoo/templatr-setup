# `.templatr.toml` Manifest Specification

This document is the complete reference for the `.templatr.toml` manifest file format. Every Templatr template includes this file at its root to describe what the template needs.

## Overview

The manifest tells the [Templatr Setup Tool](https://templatr.co/tools/setup) what runtimes, packages, environment variables, and configuration files your template requires. The tool reads this file, scans the user's system, installs what's missing, and helps configure everything.

**Format**: [TOML v1.0](https://toml.io/) - chosen for comment support, human-readability, unambiguous types, and industry adoption in modern tooling (Cargo.toml, mise.toml, pyproject.toml).

**Location**: Root of every template directory as `.templatr.toml` (the leading dot makes it a hidden file on Unix systems).

## Quick Example

```toml
[template]
name = "SaaS Landing Page Template"
version = "1.0.0"
tier = "business"
category = "website"
slug = "saas-landing-template"

[runtimes]
node = ">=20.0.0"

[packages]
manager = "npm"
install_command = "npm install"

[[env]]
key = "NEXT_PUBLIC_SITE_URL"
label = "Site URL"
description = "Your production website URL"
default = "http://localhost:3000"
required = true
type = "url"

[post_setup]
commands = ["npm run build"]
message = "Your template is ready! Run 'npm run dev' to start."

[meta]
min_tool_version = "1.0.0"
```

## Sections

### `[template]` - Template Identity (required)

Describes the template itself. Used for display purposes in the TUI and web dashboard.

| Field      | Type   | Required | Description                                                                   |
| ---------- | ------ | -------- | ----------------------------------------------------------------------------- |
| `name`     | string | Yes      | Human-readable template name                                                  |
| `version`  | string | Yes      | Template version (semver recommended)                                         |
| `tier`     | string | No       | Pricing tier (e.g., `starter`, `business`, `agency`, `enterprise`)            |
| `category` | string | No       | Product category (e.g., `website`, `mobileApp`, `crm`, `cms`, `landingPages`) |
| `slug`     | string | No       | Product slug on Templatr (used for README/docs links)                         |

**Validation**: `name` and `version` must be non-empty.

```toml
[template]
name = "SaaS Landing Page Template"
version = "1.0.0"
tier = "business"
category = "website"
slug = "saas-landing-template"
```

### `[runtimes]` - Runtime Requirements (optional)

Specifies which language runtimes need to be installed on the user's system. Each key is a runtime name, and the value is a [semver constraint](#version-ranges).

The tool checks if these are already installed and only installs what's missing or outdated.

| Runtime Key | Installs                  | Official Source                                                                |
| ----------- | ------------------------- | ------------------------------------------------------------------------------ |
| `node`      | Node.js                   | [nodejs.org](https://nodejs.org/dist/)                                         |
| `python`    | Python (standalone build) | [python-build-standalone](https://github.com/indygreg/python-build-standalone) |
| `flutter`   | Flutter SDK               | [flutter.dev](https://flutter.dev)                                             |
| `java`      | Java (Adoptium Temurin)   | [adoptium.net](https://adoptium.net)                                           |
| `go`        | Go                        | [go.dev](https://go.dev/dl/)                                                   |
| `rust`      | Rust (via rustup)         | [rustup.rs](https://rustup.rs)                                                 |
| `ruby`      | Ruby                      | Manual install guidance                                                        |
| `php`       | PHP                       | Manual install guidance                                                        |
| `dotnet`    | .NET                      | Manual install guidance                                                        |

**Validation**: Each key must be one of the valid runtime names listed above.

```toml
[runtimes]
node = ">=20.0.0"          # Node.js 20 or higher
python = ">=3.12.0"        # Python 3.12 or higher
flutter = ">=3.22.0"       # Flutter 3.22 or higher
java = ">=21"              # Java 21 or higher (major version shorthand)
go = ">=1.22.0"            # Go 1.22 or higher
rust = "latest"            # Latest stable Rust
```

### Version Ranges

Runtime versions use [semver](https://semver.org/) constraints powered by [Masterminds/semver](https://github.com/Masterminds/semver):

| Constraint   | Meaning                                      | Example Match   |
| ------------ | -------------------------------------------- | --------------- |
| `">=20.0.0"` | Version 20.0.0 or higher                     | 20.0.0, 22.14.0 |
| `"^20.0.0"`  | Compatible with 20.x.x (same major)          | 20.1.0, 20.17.0 |
| `"~20.0.0"`  | Patch-level changes only (20.0.x)            | 20.0.0, 20.0.5  |
| `">=21"`     | Major version 21+ (useful for Java)          | 21.0.0, 21.0.2  |
| `"latest"`   | Always satisfied; installs latest if missing | Any version     |
| `"20.0.0"`   | Exact version match                          | 20.0.0 only     |

If the user already has a version installed that satisfies the constraint, the tool skips installation.

### `[packages]` - Package Manager (optional)

Configures how project dependencies are installed after runtimes are set up.

| Field             | Type     | Required | Description                                            |
| ----------------- | -------- | -------- | ------------------------------------------------------ |
| `manager`         | string   | No       | Package manager identifier                             |
| `install_command` | string   | No       | Command to run for installing project dependencies     |
| `global`          | string[] | No       | Global packages to install before project dependencies |

**Valid managers**: `npm`, `pnpm`, `yarn`, `bun`, `pip`, `pub`, `composer`, `cargo`, `go`

The `install_command` is executed as-is (split by spaces). For global packages, the tool prepends the appropriate global install prefix based on the manager:

| Manager | Global install prefix |
| ------- | --------------------- |
| `npm`   | `npm install -g`      |
| `pnpm`  | `pnpm add -g`         |
| `yarn`  | `yarn global add`     |
| `bun`   | `bun add -g`          |
| `pip`   | `pip install`         |

```toml
[packages]
manager = "npm"
install_command = "npm install"
global = ["typescript", "tsx"]
```

```toml
[packages]
manager = "pip"
install_command = "pip install -r requirements.txt"
```

```toml
[packages]
manager = "pub"
install_command = "flutter pub get"
```

### `[[env]]` - Environment Variables (optional, array)

Each `[[env]]` entry defines an environment variable that the tool will prompt the user for and write to a `.env` file. This is a TOML [array of tables](https://toml.io/en/v1.0.0#array-of-tables).

| Field         | Type   | Required | Description                                      |
| ------------- | ------ | -------- | ------------------------------------------------ |
| `key`         | string | Yes      | Environment variable name (e.g., `DATABASE_URL`) |
| `label`       | string | No       | Human-readable label shown in the form           |
| `description` | string | No       | Help text explaining what this variable is for   |
| `default`     | string | No       | Default value pre-filled in the form             |
| `required`    | bool   | No       | Whether this variable must have a value          |
| `type`        | string | No       | Field type for input rendering and validation    |
| `docs_url`    | string | No       | Link to documentation for getting this value     |
| `file`        | string | No       | Target env file path (default: `.env`)           |

**Validation**: `key` must be non-empty. `type`, if provided, must be one of the [field types](#field-types).

```toml
[[env]]
key = "NEXT_PUBLIC_SITE_URL"
label = "Site URL"
description = "Your production website URL (e.g., https://yourdomain.com)"
default = "http://localhost:3000"
required = true
type = "url"

[[env]]
key = "RESEND_API_KEY"
label = "Resend API Key"
description = "Get your API key at https://resend.com/api-keys"
default = ""
required = false
type = "secret"
docs_url = "https://resend.com/docs"

[[env]]
key = "DEBUG"
label = "Debug Mode"
description = "Enable debug mode (disable in production)"
default = "true"
required = false
type = "boolean"
```

#### Multiple Env Files

Use the `file` field to target different env files. Variables without a `file` field default to `.env`. The tool groups variables by file and writes each file separately, with a section header in the TUI and web UI for each file.

```toml
# Written to .env (default)
[[env]]
key = "NEXT_PUBLIC_SITE_URL"
label = "Site URL"
default = "http://localhost:3000"
type = "url"

# Written to .env.production
[[env]]
key = "NEXT_PUBLIC_SITE_URL"
file = ".env.production"
label = "Production Site URL"
default = "https://your-domain.com"
type = "url"

# Written to .env.production
[[env]]
key = "DATABASE_URL"
file = ".env.production"
label = "Production Database URL"
type = "secret"
```

#### Generated `.env` Output

The tool writes the `.env` file with comments from the manifest:

```bash
# Generated by templatr-setup
# See .templatr.toml for field descriptions

# Your production website URL (e.g., https://yourdomain.com)
NEXT_PUBLIC_SITE_URL=http://localhost:3000

# Get your API key at https://resend.com/api-keys
# Docs: https://resend.com/docs
RESEND_API_KEY="re_abc123..."

# Enable debug mode (disable in production)
DEBUG=true
```

Values containing spaces, tabs, quotes, backslashes, `#`, or `$` are automatically quoted.

### `[[config]]` - Configuration Files (optional, array)

Each `[[config]]` entry defines a file with editable fields. The tool reads the file, presents a form for each field, and writes the values back.

| Field         | Type   | Required | Description                                         |
| ------------- | ------ | -------- | --------------------------------------------------- |
| `file`        | string | Yes      | Path to the config file (relative to template root) |
| `label`       | string | No       | Human-readable label for this config group          |
| `description` | string | No       | Help text for this config group                     |
| `fields`      | array  | No       | List of editable fields within this file            |

Each field in `fields`:

| Field         | Type   | Required | Description                                              |
| ------------- | ------ | -------- | -------------------------------------------------------- |
| `path`        | string | Yes      | Dot-notation path to the value (e.g., `siteConfig.name`) |
| `label`       | string | No       | Human-readable label                                     |
| `description` | string | No       | Help text                                                |
| `type`        | string | No       | Field type (see [field types](#field-types))             |
| `default`     | string | No       | Default value                                            |

**Validation**: `file` must be non-empty. Each field's `path` must be non-empty.

**How config editing works**: The tool uses regex-based pattern matching to find `key: "value"` or `key: 'value'` patterns in the file. It replaces only the value while preserving the original quote style, surrounding code, comments, and formatting. The last segment of the dot-notation path is used as the key (e.g., `siteConfig.contact.email` matches the key `email`).

```toml
[[config]]
file = "src/config/site.ts"
label = "Site Configuration"
description = "Core identity and branding for your website"

  [[config.fields]]
  path = "siteConfig.name"
  label = "Site Name"
  description = "Your brand/company name"
  type = "text"
  default = "SaaSify"

  [[config.fields]]
  path = "siteConfig.description"
  label = "Site Description"
  description = "A short description for SEO and social sharing"
  type = "text"
  default = "Ship your SaaS faster."

  [[config.fields]]
  path = "siteConfig.url"
  label = "Production URL"
  type = "url"
  default = "https://your-domain.com"

  [[config.fields]]
  path = "siteConfig.contact.email"
  label = "Contact Email"
  type = "email"
  default = "hello@your-domain.com"
```

### Field Types

Used by `[[env]]` and `[[config.fields]]` to determine how the field is rendered and validated:

| Type      | TUI Rendering         | Web UI Rendering        | Behavior                       |
| --------- | --------------------- | ----------------------- | ------------------------------ |
| `text`    | Text input            | Text input              | Required check only            |
| `url`     | Text input            | URL input               | Valid URL format               |
| `email`   | Text input            | Email input             | Valid email format             |
| `secret`  | Masked input (`****`) | Password input (masked) | Value is never written to logs |
| `number`  | Text input            | Number input            | Numeric validation             |
| `boolean` | Text input            | Toggle/checkbox         | true/false                     |

If `type` is omitted, defaults to `text`.

### `[post_setup]` - Post-Setup Commands (optional)

Commands to run after runtimes are installed and packages are set up.

| Field      | Type     | Required | Description                                           |
| ---------- | -------- | -------- | ----------------------------------------------------- |
| `commands` | string[] | No       | Commands to run sequentially (stops on first failure) |
| `message`  | string   | No       | Success message shown after all commands complete     |

Commands are executed in the template directory with the user's shell. Each command is split by spaces and run via `exec.Command`.

```toml
[post_setup]
commands = [
  "npm run build",
  "npm run seed",
]
message = """
Your template is ready! Run 'npm run dev' to start the development server.
Visit http://localhost:3000 to see your site.
"""
```

Multi-line strings use TOML's `"""..."""` syntax.

### `[meta]` - Tool Metadata (optional)

Configuration for the setup tool itself.

| Field              | Type   | Required | Description                                                      |
| ------------------ | ------ | -------- | ---------------------------------------------------------------- |
| `min_tool_version` | string | No       | Minimum `templatr-setup` version required to parse this manifest |
| `docs`             | string | No       | URL to the template's documentation on Templatr                  |

```toml
[meta]
min_tool_version = "1.0.0"
docs = "https://templatr.co/saas-landing-template"
```

## Complete Examples

### Next.js Template

A typical Next.js SaaS template with npm, environment variables, and site configuration:

```toml
# .templatr.toml - Next.js SaaS Landing Page Template

[template]
name = "SaaS Landing Page Template"
version = "1.0.0"
tier = "business"
category = "website"
slug = "saas-landing-template"

[runtimes]
node = ">=20.0.0"

[packages]
manager = "npm"
install_command = "npm install"

[[env]]
key = "NEXT_PUBLIC_SITE_URL"
label = "Site URL"
description = "Your production website URL (e.g., https://yourdomain.com)"
default = "http://localhost:3000"
required = true
type = "url"

[[env]]
key = "RESEND_API_KEY"
label = "Resend API Key"
description = "Get your API key at https://resend.com/api-keys"
default = ""
required = false
type = "secret"
docs_url = "https://resend.com/docs"

[[env]]
key = "CONTACT_EMAIL"
label = "Contact Form Recipient"
description = "Email address that receives contact form submissions"
default = ""
required = false
type = "email"

[[config]]
file = "src/config/site.ts"
label = "Site Configuration"
description = "Core identity and branding for your website"

  [[config.fields]]
  path = "siteConfig.name"
  label = "Site Name"
  description = "Your brand/company name"
  type = "text"
  default = "SaaSify"

  [[config.fields]]
  path = "siteConfig.description"
  label = "Site Description"
  description = "A short description for SEO and social sharing"
  type = "text"
  default = "Ship your SaaS faster with a beautiful, modern landing page."

  [[config.fields]]
  path = "siteConfig.url"
  label = "Production URL"
  type = "url"
  default = "https://your-domain.com"

  [[config.fields]]
  path = "siteConfig.contact.email"
  label = "Contact Email"
  type = "email"
  default = "hello@your-domain.com"

[post_setup]
commands = ["npm run build"]
message = """
Your template is ready! Run 'npm run dev' to start the development server.
Visit http://localhost:3000 to see your site.
"""

[meta]
min_tool_version = "1.0.0"
docs = "https://templatr.co/saas-landing-template"
```

### Django Template

A Python Django template with multiple runtimes, pip, and database migrations:

```toml
# .templatr.toml - Django CRM Template

[template]
name = "CRM Dashboard Template"
version = "1.0.0"
tier = "business"
category = "crm"
slug = "crm-dashboard"

[runtimes]
python = ">=3.12.0"
node = ">=20.0.0"

[packages]
manager = "pip"
install_command = "pip install -r requirements.txt"

[[env]]
key = "SECRET_KEY"
label = "Django Secret Key"
description = "A random string used for cryptographic signing"
default = ""
required = true
type = "secret"

[[env]]
key = "DATABASE_URL"
label = "Database URL"
description = "PostgreSQL connection string"
default = "postgresql://localhost:5432/crm"
required = true
type = "url"

[[env]]
key = "DEBUG"
label = "Debug Mode"
description = "Enable Django debug mode (disable in production)"
default = "true"
required = false
type = "boolean"

[post_setup]
commands = [
  "python manage.py migrate",
  "python manage.py collectstatic --noinput",
]
message = """
Your CRM template is ready! Run 'python manage.py runserver' to start.
Visit http://localhost:8000 to see your dashboard.
"""

[meta]
min_tool_version = "1.0.0"
docs = "https://templatr.co/crm-dashboard"
```

### Flutter Template

A Flutter mobile app template with Java dependency (for Android builds):

```toml
# .templatr.toml - Flutter Mobile App Template

[template]
name = "E-Commerce Mobile App"
version = "1.0.0"
tier = "starter"
category = "mobileApp"
slug = "ecommerce-mobile-app"

[runtimes]
flutter = ">=3.22.0"
java = ">=17"

[packages]
manager = "pub"
install_command = "flutter pub get"

[[env]]
key = "API_BASE_URL"
label = "API Base URL"
description = "Backend API URL for the mobile app"
default = "http://localhost:8080"
required = true
type = "url"

[[env]]
key = "FIREBASE_PROJECT_ID"
label = "Firebase Project ID"
description = "Your Firebase project ID for push notifications"
default = ""
required = false
type = "text"
docs_url = "https://firebase.google.com/docs"

[post_setup]
commands = ["flutter build apk --debug"]
message = """
Your app is ready! Run 'flutter run' to launch on a connected device or emulator.
"""

[meta]
min_tool_version = "1.0.0"
docs = "https://templatr.co/ecommerce-mobile-app"
```

### Java Spring Boot Template

A Java template with Maven wrapper and multiple secret environment variables:

```toml
# .templatr.toml - Java Spring Boot CMS Template

[template]
name = "CMS Admin Panel"
version = "1.0.0"
tier = "enterprise"
category = "cms"
slug = "cms-admin-panel"

[runtimes]
java = ">=21"

[packages]
manager = "go"
install_command = "./mvnw install -DskipTests"

[[env]]
key = "SPRING_DATASOURCE_URL"
label = "Database URL"
description = "JDBC connection string for your database"
default = "jdbc:postgresql://localhost:5432/cms"
required = true
type = "url"

[[env]]
key = "SPRING_DATASOURCE_USERNAME"
label = "Database Username"
default = "postgres"
required = true
type = "text"

[[env]]
key = "SPRING_DATASOURCE_PASSWORD"
label = "Database Password"
default = ""
required = true
type = "secret"

[[env]]
key = "JWT_SECRET"
label = "JWT Secret"
description = "Secret key for JWT token signing"
default = ""
required = true
type = "secret"

[post_setup]
commands = ["./mvnw spring-boot:run"]
message = """
Your CMS template is ready! The application is starting on http://localhost:8080.
Default admin credentials: admin@example.com / admin123 (change immediately).
"""

[meta]
min_tool_version = "1.0.0"
docs = "https://templatr.co/cms-admin-panel"
```

## Minimal Manifest

The absolute minimum valid manifest only requires `template.name` and `template.version`:

```toml
[template]
name = "My Template"
version = "1.0.0"
```

This is useful when the template has no special runtime requirements and only needs the configure step or just wants to be recognized by the setup tool.

## Validation Rules

The tool validates the manifest before proceeding. All validation errors are collected and reported together (it does not stop at the first error).

| Rule                                            | Error If Violated                      |
| ----------------------------------------------- | -------------------------------------- |
| `template.name` must be non-empty               | `template.name is required`            |
| `template.version` must be non-empty            | `template.version is required`         |
| Runtime keys must be valid                      | `unknown runtime: "{key}"`             |
| `packages.manager` must be valid (if set)       | `unknown package manager: "{manager}"` |
| `env[].key` must be non-empty                   | `env entry missing key`                |
| `env[].type` must be valid (if set)             | `unknown env type: "{type}"`           |
| `config[].file` must be non-empty               | `config entry missing file`            |
| `config[].fields[].path` must be non-empty      | `config field missing path`            |
| `config[].fields[].type` must be valid (if set) | `unknown config field type: "{type}"`  |

## Tips for Template Authors

1. **Always specify version ranges, not exact versions** - `">=20.0.0"` is better than `"20.0.0"` because it allows newer compatible versions.

2. **Use `latest` sparingly** - It always satisfies the check but installs the newest version when missing, which may not be what you tested against.

3. **Mark API keys as `type = "secret"`** - This ensures they are masked in the TUI, never written to log files, and displayed as password fields in the web dashboard.

4. **Include `docs_url` for third-party services** - Helps users know where to get API keys (e.g., Resend, Firebase, Stripe).

5. **Keep `post_setup.commands` minimal** - Only include commands that verify the template works. Long build commands slow down the setup experience.

6. **Use `post_setup.message` to guide next steps** - Tell users what command to run, what URL to visit, and any first-time setup tasks.

7. **Test your manifest** - Run `templatr-setup setup --dry-run -f .templatr.toml` to verify the plan looks correct without installing anything.

8. **Include comments** - TOML supports `#` comments. Use them to explain non-obvious configuration choices.
