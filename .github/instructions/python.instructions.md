---
description: "Python coding conventions and guidelines"
applyTo: "**/*.py"
---

# Python Coding Conventions

## Python Instructions

- Comments: use comments to explain the purpose of code blocks and any non-obvious logic.
- Typing: Use the `typing` module for type annotations (e.g., `List[str]`, `Dict[str, int]`).
- Type hints: always include type hints for function parameters and return types.
- Modularity: break down complex functions into smaller, more manageable functions.
- Extensibility: design code that can be easily extended in the future.
- Readability and clarity: write code that is easy to read and understand.
- Safety: write code that minimizes the risk of errors and bugs.
- Maintainability: Write code with good maintainability practices, including comments on why certain design decisions were made.
- Testing: include unit tests for functions and document them with docstrings explaining the test cases.
- Testing: use Pytest for writing and running tests.
- Favour explicitness over implicitness.
- For algorithm-related code, include explanations of the approach used.
- Handle edge cases and write clear exception handling.
- For libraries or external dependencies, mention their usage and purpose in comments.
- Use consistent naming conventions and follow language-specific best practices.
- Write concise, efficient, and idiomatic code that is also easily understandable.
- Configuration: use constants from a global config file, and use classes to divide types of constants.

## Code Style and Formatting

- Follow the PEP 8 style guide for Python.
- Provide docstrings following PEP 257 conventions.
- Maintain proper indentation (use 4 spaces for each level of indentation).
- Ensure lines do not exceed 79 characters.
- Place function and class docstrings immediately after the `def` or `class` keyword.
- Use blank lines to separate functions, classes, and code blocks where appropriate.
- Use google-style docstrings with imperative mood.
- Multi-line docstring summary should start at the first line.
- Use British English spelling conventions.

## Edge Cases and Testing

- Always include test cases for critical paths of the application.
- Account for common edge cases like empty inputs, invalid data types, and large datasets.
- Include comments for edge cases and the expected behaviors in those cases.
- Write unit tests for functions and document them with docstrings explaining the test cases.

## Example of Proper Documentation

```python
def calculate_area(radius: float) -> float:
    """Calculate the area of a circle given the radius.

    Args:
        radius (float): The radius of the circle.

    Returns:
        float: The area of the circle, calculated as π * radius^2.
    """
    import math
    return math.pi * radius ** 2
```
