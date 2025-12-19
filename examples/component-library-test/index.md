---
title: "Component Library Test"
sources:
  users:
    type: json
    file: users.json
  countries:
    type: json
    file: countries.json
---

# Component Library Test

This page demonstrates smart table and select auto-generation using the components library.

## Test 1: Datatable with Explicit Columns

```lvt
<table lvt-source="users" lvt-columns="name:Name,email:Email">
</table>
```

## Test 2: Datatable with Different Columns

```lvt
<table lvt-source="users" lvt-columns="name:Name,role:Role">
</table>
```

## Test 3: Auto Select Dropdown

```lvt
<div class="select-container">
  <label>Select a country:</label>
  <select lvt-source="countries" lvt-value="code" lvt-label="name" class="test-select">
  </select>
</div>
```

## Test 4: Auto Table with Auto-Discovery (No Columns)

```lvt
<table lvt-source="users">
</table>
```
