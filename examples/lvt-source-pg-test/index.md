---
title: "Users from PostgreSQL"
sources:
  users:
    type: pg
    query: "SELECT id, name, email FROM users ORDER BY id"
---

# Users List

Test for `lvt-source` attribute that fetches data from PostgreSQL.

```lvt
<main lvt-source="users">
    <h2>Users from Database</h2>

    {{if .Error}}
    <p><mark>Error: {{.Error}}</mark></p>
    {{else}}
    <table>
        <thead>
            <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Email</th>
            </tr>
        </thead>
        <tbody>
            {{range .Data}}
            <tr data-user-id="{{.id}}">
                <td>{{.id}}</td>
                <td>{{.name}}</td>
                <td>{{.email}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    {{end}}

    <button lvt-click="Refresh">Refresh Data</button>
</main>
```
