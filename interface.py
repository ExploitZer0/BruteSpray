# hacker_ui.py
import sys
import time
import random
from rich.console import Console
from rich.text import Text

console = Console()

def hacker_print(line, success=False, fail=False):
    text = Text(line)
    if success:
        text.stylize("bold green")
    elif fail:
        text.stylize("bold red")
    else:
        text.stylize("bold cyan")
    console.print(text)
    time.sleep(random.uniform(0.05, 0.15))

if __name__ == "__main__":
    console.print("[bold magenta]=== Hacker BruteSpray UI ===[/bold magenta]\n")
    for line in sys.stdin:
        line = line.strip()
        if "SUCCESSO" in line or "[+]" in line:
            hacker_print(line, success=True)
        elif "FALLITO" in line or "
