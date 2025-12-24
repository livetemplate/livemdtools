---
title: "Markdown Data Todo"
sources:
  tasks:
    type: markdown
    anchor: "#data-section"
---

# Todo List from Markdown

This example demonstrates reading AND writing task list data from a markdown section in the same file.

## Interactive Todo Display

```lvt
<main lvt-source="tasks">
    <h3>My Tasks</h3>
    {{if .Error}}
    <p><mark>Error: {{.Error}}</mark></p>
    {{else}}
    <ul style="list-style: none; padding-left: 0;">
        {{range .Data}}
        <li style="display: flex; align-items: center; gap: 8px; padding: 4px 0;">
            <input type="checkbox" {{if .Done}}checked{{end}}
                   lvt-click="Toggle" lvt-data-id="{{.Id}}">
            <span {{if .Done}}style="text-decoration: line-through; opacity: 0.7"{{end}}>{{.Text}}</span>
            <button lvt-click="Delete" lvt-data-id="{{.Id}}"
                    style="margin-left: auto; padding: 2px 8px; color: red; border: 1px solid red; background: transparent; border-radius: 4px; cursor: pointer;">
                x
            </button>
        </li>
        {{end}}
    </ul>
    <p><small>Total: {{len .Data}} tasks</small></p>
    {{end}}

    <hr style="margin: 16px 0;">

    <form lvt-submit="Add" style="display: flex; gap: 8px;">
        <input type="text" name="text" placeholder="Add new task..." required
               style="flex: 1; padding: 8px; border: 1px solid #ccc; border-radius: 4px;">
        <button type="submit"
                style="padding: 8px 16px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">
            Add
        </button>
    </form>

    <button lvt-click="Refresh" style="margin-top: 8px;">Refresh</button>
</main>
```

---

## Data Section {#data-section}

Edit this task list directly in the markdown file:

- [ ] Buy groceries <!-- id:task1 -->
- [x] Clean the house <!-- id:task2 -->
- [ ] Walk the dog <!-- id:task3 -->
- [ ] Send emails <!-- id:task4 -->
- [x] Finish project report <!-- id:task5 -->
