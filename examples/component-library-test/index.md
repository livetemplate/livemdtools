---
title: "Auto-Rendering Components"
sources:
  users:
    type: json
    file: users.json
  countries:
    type: json
    file: countries.json
  empty_source:
    type: json
    file: empty.json
  tags:
    type: json
    file: tags.json
  tasks:
    type: json
    file: tasks.json
---

# Auto-Rendering Components

This page demonstrates auto-rendering for tables, selects, and lists.

## Simple Tables

### Test 1: Table with Explicit Columns

```lvt
<table lvt-source="users" lvt-columns="name:Name,email:Email">
</table>
```

### Test 2: Table with Actions

```lvt
<table lvt-source="users" lvt-columns="name:Name,role:Role" lvt-actions="delete:Delete,edit:Edit">
</table>
```

### Test 3: Table with Empty State

```lvt
<table lvt-source="empty_source" lvt-columns="name:Name,email:Email" lvt-empty="No users found">
</table>
```

### Test 4: Table with Auto-Discovery (No Columns)

```lvt
<table lvt-source="users">
</table>
```

---

## Auto Select Dropdown

```lvt
<div class="select-container">
  <label>Select a country:</label>
  <select lvt-source="countries" lvt-value="code" lvt-label="name" class="test-select">
  </select>
</div>
```

---

## Auto Lists

### Test 5: Object Array List

```lvt
<ul lvt-source="tags" lvt-field="name" class="test-tags">
</ul>
```

### Test 6: Object Array with Field

```lvt
<ul lvt-source="tasks" lvt-field="title" class="test-tasks">
</ul>
```

### Test 7: List with Actions

```lvt
<ul lvt-source="tasks" lvt-field="title" lvt-actions="delete:×,edit:Edit" class="test-actions-list">
</ul>
```

### Test 8: Ordered List with Empty State

```lvt
<ol lvt-source="empty_source" lvt-field="name" lvt-empty="No items available" class="test-empty-list">
</ol>
```
