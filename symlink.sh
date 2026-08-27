cd to project root

for item in .scratch docs ROADMAP.md NOTES.md GEMINI.md DESIGN.md CONTEXT.md AGENTS.md; do
    if [ -e "$item" ] || [ -L "$item" ]; then
        mv "$item" "reqly-docs-branch/$item"
    fi
    ln -s "reqly-docs-branch/$item" "$item"
done
