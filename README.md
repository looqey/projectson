# projectson - project JSON collection tool

**Projectson** is a desktop application and command-line tool designed to flexibly collect project files based on a YAML configuration and consolidate them into a compact JSON output. This JSON is ideal for various purposes, including feeding project context to Large Language Models (LLMs) for analysis, code generation, or documentation tasks.

The application provides a user-friendly GUI built with [Fyne](https://fyne.io/) for managing configurations, previewing file selections, and running the collection process, alongside a powerful CLI for automation and scripting.

**UPD: Added CLI!** Now you can automate tasks directly from your terminal. See the [CLI Usage](#cli-usage) section below.

![ProjectSon Screenshot](https://raw.githubusercontent.com/looqey/projectson/master/Icon.png) <!-- Assuming this icon is still relevant or you might want a more generic one -->

---

## Table of Contents

-   [Core Idea](#core-idea)
-   [Key Features](#key-features)
-   [Pros and Cons](#pros-and-cons)
-   [Basic GUI Usage](#basic-gui-usage)
-   [CLI Usage](#cli-usage)
    -   [General Flags](#general-flags)
    -   [Commands](#commands)
        -   [`init`](#init)
        -   [`validate`](#validate)
        -   [`preview`](#preview)
        -   [`run`](#run)
-   [How It Works](#how-it-works)
-   [Installation](#installation)
    -   [From Releases (Recommended)](#from-releases-recommended)
    -   [From Source](#from-source)
-   [Configuration](#configuration)
    -   [`root`](#root-1)
    -   [`include`](#include-1)
    -   [`formats`](#formats-1)
    -   [`output`](#output-1)
    -   [`exclude_patterns`](#exclude_patterns-1)
    -   [`content_exclusions`](#content_exclusions-1)
-   [Applying AI-Generated Changes (GUI)](#applying-ai-generated-changes-gui)
-   [Contributing](#contributing)
-   [License](#license)

---

## Core Idea

The primary goal of ProjectSon is to provide a highly configurable way to extract relevant information (file paths and/or content) from a software project. Instead of manually copying and pasting code or struggling with complex shell scripts, ProjectSon allows you to define:

1.  **What to include**: Specific directories, files, and file types.
2.  **How to include it**: Whether to grab just the file path, its content, or both.
3.  **What to exclude**: Files, directories, or patterns (like `node_modules` or test files).
4.  **How to clean content**: Rules to strip out comments, boilerplate, or specific blocks of text/code *before* it's added to the output.

The result is a single, structured JSON file. This JSON can then be easily parsed or used as input for other tools, most notably LLMs, providing them with a curated and concise view of your project's structure and content.

---

## Key Features

*   **multi-interface**:
    *   **GUI Interface**: Easy-to-use graphical interface for managing all settings.
    *   **CLI Tool**: Powerful command-line interface for automation, scripting, and headless environments.
*   **YAML Configuration**: Human-readable and version-controllable configuration files.
*   **Flexible Inclusion Rules**: Specify precisely which files and directories to include, with different modes (path, content, both).
*   **Powerful Exclusion Capabilities**: Exclude unwanted files and directories using glob patterns or regular expressions.
*   **Content Stripping**: Define rules (delimiters or regex) to remove irrelevant sections from file content (e.g., comments, specific code blocks).
*   **File Preview**: See which files will be included (both GUI and CLI `preview` command). In GUI, inspect original and modified content.
*   **Run Statistics (GUI)**: View stats about the last collection run.
*   **AI Change Application (GUI)**: A dedicated tab to apply file modifications (create, update, delete) based on a JSON response from an AI.
*   **Cross-Platform**: Builds for Windows, macOS, and Linux (both GUI and CLI).
*   **Persistent Settings (GUI)**: Remembers the last used configuration file.

---

## Pros and Cons

**Pros:**

*   **Highly Configurable**: Tailor the data collection precisely to your needs.
*   **Reduces Manual Labor**: Automates the process of gathering project context.
*   **Compact Output**: The JSON output is designed to be concise, especially after content exclusions and space normalization.
*   **GUI for Ease of Use**: No need to be a command-line wizard for visual configuration and review.
*   **CLI for Automation**: Integrate projectson into your development workflows and CI/CD pipelines.
*   **Version Controllable Configs**: Store your collection profiles (`.yaml` files) in your project repository.
*   **Improved LLM Context**: Provides cleaner, more relevant input to LLMs, potentially leading to better results and reduced token usage.

**Cons:**

*   **Space-dependent languages (e.g., Python):** Currently, the content processing (space normalization and some content exclusions) might break the indentation of languages like Python. *A future update will introduce a setting to disable space normalization or handle it more carefully for such languages.*

---

## Basic GUI Usage

1.  **Launch ProjectSon GUI.**
2.  **Configure:**
    *   Go to the **Config** tab.
    *   Set the **Project Root Path** to your project's main directory.
    *   Specify **File Formats** (e.g., `go`, `js`, `py`, `md`).
    *   Define the **Output JSON Path** (where the `output.json` will be saved).
    *   Add **Include Paths & Modes** if you only want specific sub-folders or files.
    *   Add **Exclude Patterns** for things like `node_modules`, `.git`, build artifacts, etc.
3.  **(Optional) Content Exclusions:**
    *   Go to the **Exclusions** tab.
    *   Add rules to remove comments, log statements, or other boilerplate from file content.
4.  **Preview:**
    *   Go to the **Preview** tab.
    *   Click "Refresh File List".
    *   Inspect the list of files. Select a file to see its original and modified content (after exclusions).
5.  **Run:**
    *   Go to the **Run** tab.
    *   Click "Run Collection Process".
6.  **Output:**
    *   An `output.json` file (or the name you specified) will be created with the collected project data.
7.  **(Optional) Save Configuration:**
    *   Click the "Save" icon in the toolbar to save your current settings to a `.yaml` file for future use. You can load it using the "Open" icon.

---

## CLI Usage

The projectson CLI (`projectson-cli`) allows you to perform operations from the command line.

### General Flags

These flags can be used with most commands to override values from the configuration file or provide them if no config file is used.

*   `-c, --config <path>`: Path to the YAML configuration file (default: `projectson_config.yaml` in the current directory).
*   `-r, --root <path>`: Project root directory.
*   `-o, --output <path>`: Output JSON file path.
*   `-f, --formats <ext1,ext2,...>`: Comma-separated list of file formats/extensions (e.g., `go,vue,ts`).
*   `-e, --exclude <pattern1,pattern2,...>`: Comma-separated list of exclude patterns (e.g., `node_modules,*.log`).
*   `--include <path1,path2,...>`: (Primarily for `init`) Comma-separated list of paths to include relative to root.

### Commands

#### `init`
Creates a default configuration file (`projectson_config.yaml`).

**Usage:**
```bash
projectson-cli init [flags]
```

**Flags for `init`:**
*   `--force`: Overwrite `projectson_config.yaml` if it already exists.
*   Can also use general flags like `--root`, `--output`, `--formats`, `--include` to pre-fill the new config file.

**Example:**
```bash
# Create a default projectson_config.yaml
projectson-cli init

# Create a config and pre-fill some values
projectson-cli init --root "/path/to/my/project" --formats "js,ts,html" --output "data/context.json"
```

#### `validate`
Validates the configuration file.

**Usage:**
```bash
projectson-cli validate [flags]
```

**Example:**
```bash
# Validate the default config file
projectson-cli validate

# Validate a specific config file
projectson-cli validate --config "my_custom_config.yaml"
```

#### `preview`
Shows a list of files that would be collected based on the current configuration, without actually processing their content or writing an output file.

**Usage:**
```bash
projectson-cli preview [flags]
```

**Example:**
```bash
# Preview using projectson_config.yaml
projectson-cli preview

# Preview using a specific config and overriding the formats
projectson-cli preview --config "prod.yaml" --formats "go,mod,sum"
```

#### `run`
Runs the full file collection process and generates the output JSON file.

**Usage:**
```bash
projectson-cli run [flags]
```

**Example:**
```bash
# Run collection using projectson_config.yaml
projectson-cli run

# Run collection with a specific config and output file
projectson-cli run --config "docs_config.yaml" --output "documentation_context.json"

# Run collection by specifying all parameters via flags (no config file needed)
projectson-cli run --root "./my_awesome_project" \
                   --formats "py,md" \
                   --include "src,README.md" \
                   --exclude "*.tmp,build/*" \
                   --output "awesome_project_dump.json"
```

---

## How It Works

1.  **Configuration Loading**: ProjectSon loads settings from a `yaml` file (or uses defaults/CLI flags).
2.  **File Discovery (Preview/Run)**:
    *   It starts with the `root` directory.
    *   If `include` paths are defined, it processes only those. Otherwise, it scans the `root`.
    *   It filters files based on the specified `formats`.
    *   It applies `exclude_patterns` (glob and regex) to further filter out unwanted files and directories.
3.  **Content Processing (if mode is `content` or `both`)**:
    *   For each selected file whose content is to be included:
        *   The file content is read.
        *   `content_exclusions` rules are applied sequentially to strip out defined sections.
        *   Multiple spaces/newlines are compressed into single spaces to reduce output size (this might be made optional in the future for languages like Python).
4.  **JSON Generation**:
    *   The collected data (paths and/or processed content) is structured into a JSON format:
        ```json
        {
          "project_files": [
            { "path": "project_root_name/src/main.go", "content": "package main..." },
            { "path": "project_root_name/README.md" }
            // ... more files
          ]
        }
        ```
    *   The `path` field always includes the basename of your project root to help LLMs understand the context, e.g., `myproject/src/file.js`.
5.  **Output Saving**: The generated JSON is saved to the specified `output` file.

---

## Installation

You can install ProjectSon in one of two ways:

### From Releases (Recommended)

This is the easiest way for most users.

1.  Go to the [**Releases** page](https://github.com/looqey/projectson/releases) on GitHub.
2.  Download the appropriate pre-compiled binary for your operating system and desired interface:
    *   **GUI Application:**
        *   Windows: `projectson-gui-windows-amd64.exe`
        *   macOS: `projectson-gui-macos-amd64.app.zip` (unzip and run the `.app` file)
        *   Linux: `projectson-gui-linux-amd64` (make it executable: `chmod +x projectson-gui-linux-amd64`)
    *   **CLI Tool:**
        *   Windows: `projectson-cli-windows-amd64.exe`
        *   macOS: `projectson-cli-macos-amd64`
        *   Linux: `projectson-cli-linux-amd64` (make it executable: `chmod +x projectson-cli-linux-amd64`)
3.  Place the CLI tool in a directory included in your system's PATH for easy access (e.g., `/usr/local/bin` or `C:\Windows\System32`). Run the GUI application directly.

### From Source

If you are a developer or want the latest (potentially unstable) version, you can build from source.

**Prerequisites:**

*   Go (version 1.24.1 or newer, check `go.mod` for the exact version)
*   Git
*   For GUI: Fyne dependencies for your OS (see [Fyne documentation](https://developer.fyne.io/started/)).
    *   **Linux (Ubuntu/Debian):** `sudo apt-get install -y libgl1-mesa-dev xorg-dev gcc pkg-config libgtk-3-dev`
    *   **Linux (Fedora):** `sudo dnf install -y libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel gcc pkgconfig gtk3-devel`
    *   **macOS:** Xcode Command Line Tools (`xcode-select --install`)
    *   **Windows:** A C compiler like TDM-GCC (usually handled by Fyne tooling if MSYS2/Mingw-w64 is set up).

**Steps:**

1.  Clone the repository:
    ```bash
    git clone https://github.com/looqey/projectson.git
    cd projectson
    go mod tidy
    ```
2.  Build and run:
    *   **CLI:**
        ```bash
        go build -o projectson-cli ./cmd/cli
        ./projectson-cli --help
        ```
    *   **GUI:**
        ```bash
        # For development run:
        go run ./cmd/gui/main.go 
        # Or build:
        # fyne package -src ./cmd/gui/ -name projectson-gui # (will create OS specific package)
        ```

---

## Configuration

The behavior of ProjectSon is controlled by a YAML configuration file. You can load and save these configurations via the GUI or create/edit them manually. Here's a detailed breakdown of the fields:

### `root`
-   **Type**: `String`
-   **Required**: Yes
-   **Description**: The absolute path to the root directory of the project you want to analyze.
-   **Example**:
    ```yaml
    root: "/path/to/your/project"
    # On Windows:
    # root: "C:\\Users\\YourName\\Projects\\MyProject"
    ```

### `include`
-   **Type**: `List of Strings`
-   **Required**: No (Defaults to scanning the entire `root` directory, respecting `formats` and `exclude_patterns`)
-   **Description**: A list of specific files or directories to include for processing. Paths are relative to the `root` directory. Each entry can define a path and an optional collection mode.
-   **Syntax per entry**:
    -   `"path/to/item"`: Collects both path and content (default mode). If `item` is a directory, it's scanned recursively.
    -   `"path/to/item:path"`: Collects only the path string for the item(s).
    -   `"path/to/item:content"`: Collects only the content for the item(s).
    -   `"path/to/item:both"`: Explicitly collects both path and content.
    -   `"path/to/directory/*"`: Collects files *directly within* `path/to/directory` (non-recursively). Default mode is `both`.
    -   `"path/to/directory/*:mode"`: Collects files *directly within* `path/to/directory` with the specified `mode`.
-   **Examples**:
    ```yaml
    include:
      - "src"  # Include everything in src/, recursively, path & content
      - "README.md:content" # Include only content of README.md
      - "assets/*:path"     # Include only paths of files directly in assets/
      - "docs/api.md"       # Include path & content of docs/api.md
    ```

### `formats`
-   **Type**: `List of Strings`
-   **Required**: Yes
-   **Description**: A list of file extensions (without the leading dot) to be included in the collection. Only files matching these extensions will be considered.
-   **Example**:
    ```yaml
    formats:
      - "go"
      - "html"
      - "css"
      - "js"
      - "vue"
      - "ts"
      - "py"
    ```

### `output`
-   **Type**: `String`
-   **Required**: Yes
-   **Description**: The full path (or path relative to where ProjectSon is run) for the output JSON file where the collected data will be saved. If the directory does not exist, ProjectSon will attempt to create it.
-   **Example**:
    ```yaml
    output: "project_data_output.json"
    # output: "/tmp/my_project_collection.json"
    ```

### `exclude_patterns`
-   **Type**: `List of Strings`
-   **Required**: No
-   **Description**: A list of patterns to exclude files or directories. These patterns are applied *after* `include` rules and `formats` have identified potential candidates.
    -   **Glob Patterns**: Standard file globbing (e.g., `node_modules`, `*.log`, `dist/*`). These match against:
        1.  The base name of the file or directory (e.g., `*.log` matches `debug.log`).
        2.  The path relative to `root` (e.g., `src/tests/*` matches `src/tests/test_helper.go` if the pattern contains a path separator).
    -   **Regular Expressions**: Go-compatible regular expressions. Must be enclosed in forward slashes (e.g., `/\.git/`, `/private_.*\.key$/`). These also match against basenames or relative paths.
-   **Note**: If a directory is excluded, its contents will not be scanned.
-   **Examples**:
    ```yaml
    exclude_patterns:
      - "node_modules"  # Exclude the entire node_modules directory
      - ".git"          # Exclude .git directory
      - "*.min.js"      # Exclude minified JavaScript files
      - "dist"          # Exclude build output directory
      - "/test_data/"   # Exclude directories named test_data (regex matching the name)
      - "/^\\.(svn|hg|DS_Store)/" # Exclude common VCS and OS files (regex)
      - "target/*"      # Exclude contents of a target directory (glob)
    ```

### `content_exclusions`
-   **Type**: `List of Objects`
-   **Required**: No
-   **Description**: A list of rules to remove specific sections of content from files *before* they are added to the output. This is useful for removing comments, boilerplate, sensitive data, or irrelevant code blocks. Each rule object has the following fields:
    -   `type` (String, Required): Specifies the type of exclusion.
        -   `"delimiters"`: Uses start and end string tags to identify content to remove.
        -   `"regexp"`: Uses a regular expression to identify content to remove.
    -   `file_pattern` (String, Required): A glob pattern that specifies which files this rule applies to, based on their extension.
        -   It is tested against the file's extension string (e.g., `file_pattern: "vue"` would match files with a `.vue` extension).
        -   It is also tested against the file's extension string prefixed with a dot (e.g., `file_pattern: "*.vue"` would match files with a `.vue` extension).
        -   Use `"*"` to apply the rule to all files that match the global `formats` list.
    -   `start` (String, Optional): Used when `type` is `"delimiters"`. The starting string tag of the content to exclude.
    -   `end` (String, Optional): Used when `type` is `"delimiters"`. The ending string tag of the content to exclude.
    -   `pattern` (String, Optional): Used when `type` is `"regexp"`. The Go-compatible regular expression. The regex should be crafted to match the content you want to remove. It's often useful to use the `(?s)` flag (dot matches newline) for multi-line patterns.
-   **Examples**:
    ```yaml
    content_exclusions:
      - type: "delimiters"
        file_pattern: "*.vue" # Or "vue"
        start: "<style>"
        end: "</style>"
      - type: "delimiters"
        file_pattern: "html"
        start: "<!-- BEGIN_EXCLUDE -->"
        end: "<!-- END_EXCLUDE -->"
      - type: "regexp"
        file_pattern: "*.js" # Or "js"
        pattern: "(//.*)|(/\\*[\\s\\S]*?\\*/)" # Remove JS line and block comments
      - type: "regexp"
        file_pattern: "*" # Apply to all matched file types
        pattern: "SECRET_API_KEY = '.*?'" # Remove a line containing a secret key
    ```
You can find more detailed documentation on these fields within the GUI application itself, under the **"Config Docs"** tab.
---

## Contributing

Contributions are welcome! Whether it's bug reports, feature requests, documentation improvements, or code contributions, please feel free to:

1.  **Open an Issue**: For bugs, suggestions, or discussions.
2.  **Fork the Repository**: Create your own copy of the project.
3.  **Create a Pull Request**: For submitting your changes. Please try to follow existing code style and include tests if applicable.
