import re
import os
import ast

# Common Secret Patterns
SECRET_PATTERNS = {
    "Generic API Key": r'(?i)(api_key|secret|password|token|auth)\s*[:=]\s*["\'][a-zA-Z0-9_\-]{16,}["\']',
    "AWS Access Key": r'AKIA[0-9A-Z]{16}',
    "Slack Webhook": r'https://hooks.slack.com/services/T[A-Z0-9]+/[A-Z0-9]+/[A-Za-z0-9]+'
}

DANGEROUS_FUNCTIONS = ['eval', 'exec', 'os.system']

def check_secrets(content):
    findings = []
    for name, pattern in SECRET_PATTERNS.items():
        if re.search(pattern, content):
            findings.append(f"Potential {name} found in plain text.")
    return findings

def check_insecure_logic(content):
    findings = []
    try:
        tree = ast.parse(content)
        for node in ast.walk(tree):
            if isinstance(node, ast.Call):
                func = node.func
                func_name = None
                if isinstance(func, ast.Name):
                    func_name = func.id
                elif isinstance(func, ast.Attribute):
                    val = func.value
                    if isinstance(val, ast.Name):
                        func_name = f"{val.id}.{func.attr}"
                
                if func_name and func_name in DANGEROUS_FUNCTIONS:
                    findings.append(f"Dangerous function usage: `{func_name}()` detected.")
    except SyntaxError:
        pass # Handle non-python files differently or skip
    return findings

# Logic to iterate through workspace files and print warnings
# ... (standard file traversal logic)