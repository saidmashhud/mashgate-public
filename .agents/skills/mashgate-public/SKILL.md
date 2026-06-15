```markdown
# mashgate-public Development Patterns

> Auto-generated skill from repository analysis

## Overview
This skill introduces the core development patterns and conventions used in the `mashgate-public` Python repository. While the project does not use a specific framework, it follows clear conventions for file naming, imports, exports, commit messages, and testing. This guide will help you quickly understand and contribute effectively to the codebase.

## Coding Conventions

### File Naming
- Files use **camelCase** naming.
  - **Example:** `userProfile.py`, `dataLoader.py`

### Import Style
- **Relative imports** are preferred.
  - **Example:**
    ```python
    from .utils import parseData
    ```

### Export Style
- **Named exports** are used to explicitly specify what is available from a module.
  - **Example:**
    ```python
    def processData():
        pass

    __all__ = ['processData']
    ```

### Commit Messages
- Use **conventional commit** style.
- Prefixes: `fix`
- Average commit message length: 72 characters.
  - **Example:**  
    ```
    fix: resolve issue with data parsing in userProfile module
    ```

## Workflows

### Making a Fix
**Trigger:** When you need to fix a bug or issue in the codebase  
**Command:** `/fix`

1. Identify the bug or issue.
2. Create a new branch for your fix.
3. Make the necessary code changes following the coding conventions.
4. Write or update tests as needed.
5. Commit your changes using the `fix:` prefix in your commit message.
6. Push your branch and open a pull request.

    **Example commit:**
    ```
    fix: correct import path in dataLoader.py
    ```

### Adding a New Module
**Trigger:** When adding a new feature or module  
**Command:** `/add-module`

1. Create a new file using camelCase naming.
2. Implement the module, using relative imports and named exports.
3. Add or update relevant tests in a corresponding `*.test.*` file.
4. Commit your changes with an appropriate message.
5. Push and open a pull request.

    **Example:**
    ```python
    # newFeature.py
    def newFeature():
        pass

    __all__ = ['newFeature']
    ```

### Writing Tests
**Trigger:** When adding or updating tests  
**Command:** `/write-tests`

1. Create or update files matching the `*.test.*` pattern.
2. Implement test cases for your modules.
3. Run tests to ensure correctness.

    **Example:**
    ```python
    # userProfile.test.py
    from .userProfile import getUser

    def test_getUser():
        assert getUser('alice') == {'name': 'Alice'}
    ```

## Testing Patterns

- Test files follow the `*.test.*` naming pattern (e.g., `module.test.py`).
- The testing framework is not specified; follow existing patterns in the repo.
- Place tests close to the modules they cover.
- Use relative imports in test files as well.

## Commands
| Command      | Purpose                                  |
|--------------|------------------------------------------|
| /fix         | Start the workflow for fixing a bug      |
| /add-module  | Add a new module following conventions   |
| /write-tests | Add or update tests for your modules     |
```