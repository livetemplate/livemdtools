---
title: "My Bookmarks"
description: "Self-contained bookmark manager using markdown data source"
sources:
  bookmarks:
    type: markdown
    anchor: "#links"
    readonly: false
---

# {{.Config.Title}}

A self-contained bookmark manager that stores data in this markdown file.

## Add New Bookmark

```lvt
<form lvt-submit="Add" lvt-source="bookmarks" style="display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 16px;">
    <input name="Name" placeholder="Name" required style="padding: 8px; border: 1px solid #ccc; border-radius: 4px;">
    <input name="URL" type="url" placeholder="https://..." required style="flex: 1; min-width: 200px; padding: 8px; border: 1px solid #ccc; border-radius: 4px;">
    <input name="Tags" placeholder="Tags (comma-separated)" style="padding: 8px; border: 1px solid #ccc; border-radius: 4px;">
    <button type="submit" style="padding: 8px 16px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">
        Add Bookmark
    </button>
</form>
```

## Bookmark List

```lvt
<main lvt-source="bookmarks">
    {{if .Error}}
    <p><mark>Error: {{.Error}}</mark></p>
    {{else if eq (len .Data) 0}}
    <p><em>No bookmarks yet. Add your first bookmark above!</em></p>
    {{else}}
    <table style="width: 100%; border-collapse: collapse;">
        <thead>
            <tr style="background: #f5f5f5;">
                <th style="text-align: left; padding: 12px; border-bottom: 2px solid #ddd;">Name</th>
                <th style="text-align: left; padding: 12px; border-bottom: 2px solid #ddd;">URL</th>
                <th style="text-align: left; padding: 12px; border-bottom: 2px solid #ddd;">Tags</th>
                <th style="width: 50px; padding: 12px; border-bottom: 2px solid #ddd;"></th>
            </tr>
        </thead>
        <tbody>
        {{range .Data}}
            <tr style="border-bottom: 1px solid #eee;">
                <td style="padding: 12px;">{{.Name}}</td>
                <td style="padding: 12px;"><a href="{{.URL}}" target="_blank" rel="noopener">{{.URL}}</a></td>
                <td style="padding: 12px;">
                    {{if .Tags}}
                    <span style="display: inline-flex; gap: 4px; flex-wrap: wrap;">
                        {{range (split .Tags ", ")}}
                        <span style="background: #e3f2fd; color: #1565c0; padding: 2px 8px; border-radius: 12px; font-size: 0.85em;">{{.}}</span>
                        {{end}}
                    </span>
                    {{end}}
                </td>
                <td style="padding: 12px; text-align: center;">
                    <button lvt-click="Delete" lvt-data-id="{{.Id}}"
                            style="padding: 4px 8px; color: #dc3545; border: 1px solid #dc3545; background: transparent; border-radius: 4px; cursor: pointer;"
                            title="Delete bookmark">
                        x
                    </button>
                </td>
            </tr>
        {{end}}
        </tbody>
    </table>
    <p style="margin-top: 16px; color: #666;"><small>Total: {{len .Data}} bookmarks</small></p>
    {{end}}
</main>
```

---

## Links {#links}

| Name | URL | Tags |
|------|-----|------|
| GitHub | https://github.com | dev, code | <!-- id:bm_001 -->
| Hacker News | https://news.ycombinator.com | news, tech | <!-- id:bm_002 -->
| MDN Web Docs | https://developer.mozilla.org | docs, web | <!-- id:bm_003 -->
| Go Documentation | https://go.dev/doc | docs, golang | <!-- id:bm_004 -->
